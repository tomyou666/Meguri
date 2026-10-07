# node_results 本文分離

`GetNodeResults` などの SELECT ALL が、巨大 TEXT を毎回メモリと Wails IPC に載せて遅くなる問題の調査と対策。実装前の設計メモ。採用案は **1. 表分割**。以降は棄却した代替を、次点から順に残す。

## 結論（採用）

同一 SQLite で表を分ける。


| 表                    | 役割                                       |
| -------------------- | ---------------------------------------- |
| `node_results`       | メタのみ。今後 `Find()` / `SELECT *` しても本文に届かない |
| `node_result_bodies` | 本文。明示 API だけが読む                          |


- Load では本文を載せない。`lastResult` は slim の `CrawlResultMetaDTO`
- 本文行はノードあたり **latest 成功 1 件 + baseline 1 件**（同一行なら 1 件）
- メタ行は現状どおりノードあたり最大 20 件（`MaxNodeResultsPerNode`）
- `content_hash` は canonical Markdown のみ。HTML 変化は差分に出さない
- links 判定用にメタへ `links_hash` を追加。`links_json` 本体は本文表

列の `Omit` や列順入替だけだと、次の `Find()` で再発する。

## 現状



### SQLite packed record

SQLite は行を定義順の 1 レコードで持つ。大きな TEXT は overflow page に続く。**後段の列だけ欲しくても、手前の巨大 TEXT の overflow を辿る。** 値のバイト数ではなく、列の位置がコストになる。

実測（`GetNodeResults` 相当）: `content_hash` を SELECT から外すと 200ms 以下、入れると 400〜500ms。`content_hash` 自体は SHA-256 の 64 文字だが、当時は `markdown` / `html` / `raw_html` の後ろにあった。

`000007_node_results_reorder_small_cols` で `content_hash` と `manually_edited` を巨大 TEXT より前へ移済み。hash 単体の SELECT は軽くなったが、次は残る。

- GORM `Find()` は全列を読む
- `error` / `fetched_at` はまだ巨大 TEXT の後ろ
- `WorkspaceService.Load` は全履歴本文を `lastResult` に載せて IPC する



### 重い列


| 列                                | 大きさ     | 備考                                 |
| -------------------------------- | ------- | ---------------------------------- |
| `raw_html` / `html` / `markdown` | ページ本文   | 主因                                 |
| `json_body`                      | 大きいことあり | クロール persist は現状ほとんど書かない。手動編集で入り得る |
| `links_json` / `metadata_json`   | 中程度     | HTML よりは小さい                        |
| `content_hash`                   | 64 文字   | 値は軽い。**同じ行の overflow を辿ると重い**      |




### 増幅している呼び出し

`Store.GetNodeResults` は WS の全 `node_results` を `SELECT *` する（ノードあたり最大 20 行 × 本文）。呼び出し側は「最新成功 1 件」しか使わないことが多い。


| 呼び出し                                          | 実際に要るデータ                        | 今やっていること                              |
| --------------------------------------------- | ------------------------------- | ------------------------------------- |
| `WorkspaceService.Load`                       | ノードごとの最新成功メタ（と、現状は本文）           | 全履歴の全列 → 各ノード `lastResult` に本文付き DTO  |
| `GetNodeResult` / Wails `GetNodeResults(ids)` | 指定ノードの最新成功本文                    | `GetNodeResultsByNodeIDs` が履歴 20 件の全列 |
| `GetWorkspaceDiff`                            | baseline と最新成功の hash、fetch 状態   | 全列。links は `links_json` をパース          |
| `GetNodeDiffDetail`                           | 上記 + markdown / links の old/new | 同じく WS 全件の全列                          |
| `SaveResults` / `SaveResultsSnapshot`         | 最新成功のコピー                        | 全列読み → 挿入                             |
| `ExportBundle(includeResults)`                | 最新成功本文                          | 全列から最新成功だけ残す                          |
| `BuildSkipScrapeLinkMap`                      | 対象 URL の `links_json`           | 全列                                    |


フロントは既に `lastResult` が無いと `fetchSelectedNodeResult` する。Load で全ノード本文を先読みする必要はない。

### 差分の仕様（変更しない）

差分 kind は 3 つのまま。


| kind    | 判定                                       | 詳細画面               |
| ------- | ---------------------------------------- | ------------------ |
| content | `content_hash`                           | Markdown の old/new |
| links   | 今は `links_json` の正規化比較。採用後は `links_hash` | リンク配列の old/new     |
| fetch   | 最新行の成功 / 失敗 / なし                         | 状態ラベル              |


`content_hash` は `CanonicalizeMarkdown`（CRLF→LF、前後 trim）の SHA-256。HTML / RawHTML / JSON は見ない。パネルで HTML だけ編集しても content 差分には出ない。これは仕様として維持する。

HTML を hash に混ぜると、広告やタイムスタンプで誤検知が増え、content / links タブも潰れる。

「latest」は **いちばん新しい行ではなく、いちばん新しい成功行**（`latestSuccessByNode`）。失敗しても右サイドバーの本文は前の成功結果のまま。失敗の事実はメタの `error` に残り、fetch 差分に使う。

## 採用設計: 表分割



### スキーマ

`node_results`（メタ。列はすべて小さいものだけ）

- `id`, `run_id`, `workspace_id`, `node_id`, `url`
- `content_hash`（Markdown）
- `links_hash`（後述）
- `manually_edited`
- `error`
- `fetched_at`
- 既存制約: `PRIMARY KEY (id)`, `UNIQUE (run_id, node_id)`, `graph_nodes` への FK CASCADE
- 既存インデックス: `run_id`, `(workspace_id, node_id, fetched_at DESC)`

`node_result_bodies`（本文。PK = `node_results.id`）

列順は overflow 対策として **短い / 先に読むものから**:

1. `id`（PK, FK → `node_results.id` ON DELETE CASCADE）
2. `links_json`
3. `metadata_json`
4. `markdown`
5. `html`
6. `raw_html`
7. `json_body`

本文が無いメタ行（古い履歴・失敗行）は正常。JOIN や GORM `Preload` をリスト取得のデフォルトにしない。

### 保持件数


|      | メタ                   | 本文                                        |
| ---- | -------------------- | ----------------------------------------- |
| 上限   | ノードあたり 20 行（現行 trim） | latest 成功 + そのノードの baseline 行。最大 2。同一なら 1 |
| 失敗行  | 残す（fetch 差分）         | 作らない                                      |
| 古い成功 | メタだけ残す               | 消す。ただしその行が baseline なら残す                  |


クロール成功で新しい本文を書いたら、latest でも baseline でもない本文行を消す。baseline 更新（`SaveResultsSnapshot`）後も同じ。

### ハッシュ

- `content_hash`: 現行どおり canonical Markdown の SHA-256
- `links_hash`: `canonicalLinks` と同じ入力（URL をソートした JSON 配列）の SHA-256 十六進。空は `"[]"` 相当。Go と TS で同じ算法（`content_hash` と同じ配置）

差分一覧はメタの 2 hash + `error` / 行の有無だけ。本文表を開かない。

差分詳細は対象 2 行の本文だけ読む。`SELECT markdown, links_json` で足りる（html / raw は詳細に出さない）。

### API / 型

GORM モデルを分ける。`NodeResult` に `Markdown` 等を持たせない。これが「次の人が Select All する」ことへの根治。


| 層                  | メタ                                                                          | 本文                                                                |
| ------------------ | --------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| Store              | `GetNodeResultMetas`（現行 `GetNodeResults` 相当、全列が小さい）                         | `GetNodeResultBodies(ids)` など result id 指定                        |
| Wails Load / 差分一覧  | `CrawlResultMetaDTO`（url, contentHash, linksHash, manuallyEdited。本文フィールドなし） | 呼ばない                                                              |
| Wails パネル          | —                                                                           | `GetNodeResult(ws, nodeID)` → 最新成功の `CrawlResultDTO`              |
| Wails エクスポート / マージ | —                                                                           | `GetNodeResults(ws, nodeIDs)` → **指定 ID の最新成功本文のみ**（履歴 20 件は返さない） |
| Wails 差分詳細         | メタで kind 判定                                                                 | 内部で 2 行の markdown / links_json                                    |


`LoadWorkspace` の `GraphNodeDTO.lastResult` は `*CrawlResultMetaDTO`。フロントの `fetchSelectedNodeResult` が本文を取る。

Store の全件取得メソッド名は `GetNodeResults` から変える。Wails の `GetNodeResults(ids)` は本文付きのままなので、同名のままメタ全件を返すと再発する。

### マイグレーション方針

1. `node_result_bodies` を作る
2. `node_results.links_hash` を追加し、既存 `links_json` から埋める
3. ノードごとに latest 成功と baseline 行の本文だけ `node_result_bodies` へコピー
4. `node_results` をメタ列だけに rebuild（`000007` と同様の table swap）
5. `storage/schema.sql` と GORM Gen（`tools/gen/main.go` に `node_result_bodies`）を追随

既存 DB は in-place。latest / baseline 以外の本文は捨てる（メタ 20 件は残る）。

### 実装時に直す呼び出し

- `WorkspaceService.Load`: メタのみ → slim `lastResult`
- `DiffService`: メタで一覧。詳細だけ本文
- `ResultsService.GetNodeResult(s)`: 最新成功メタ 1 件 + その body。履歴 20 件の本文を読まない
- `SaveResults` / snapshot / `ExportBundle`: 対象行のメタ + body をコピー
- `BuildSkipScrapeLinkMap`: 対象 URL の latest 成功 body から `links_json` だけ
- `UpdateLatestNodeResult`: 対象 id をメタから特定して body を部分更新。Markdown 変更時は `content_hash`、links 変更時は `links_hash`
- フロント: `CrawlResultMeta` と本文付き preview を分ける。Load 後の選択は既存の遅延取得を本線にする



## 代替案（お勧め順）

採用が 1。下がるほど「今回はやらない」。

### 2. 列 Omit + 最新 1 行だけ読む（スキーマ据え置き）

`Select` で巨大列を外し、履歴 20 件ではなく最新成功だけ取る。`Load` から本文を外す。

- 良い点: マイグレーション不要。`000007` の列順と併用しやすい
- 悪い点: モデルに本文フィールドが残る。次の `Find()` で再発する。根治にならない

部分的な先送り（クエリ修正の一部）としては、表分割 PR の途中状態として出てよい。

### 3. 列順入替だけ（`000007` 相当）

小さい列を巨大 TEXT より前へ。**既に適用済み。**

- 良い点: hash だけの SELECT は速くなる
- 悪い点: `SELECT *` は今まで通り全部読む。`Load` の IPC も変わらない。`error` / `fetched_at` / `links_json` はまだ後段

これ以上の列順いじりは採用案に吸収する（メタ表には巨大列が無い）。

### 4. 本文をファイルへ出す

SQLite にはメタ + hash。markdown/html は結果ファイル。

- 良い点: DB ファイルが膨らまない
- 悪い点: トランザクション、`ON DELETE CASCADE`、`.scrb` 書き出し、バックアップが二重管理。表分割でも足りないときまで待つかつ



### 5. 1 本の結合ハッシュ / HTML も content とみなす

md + html + links を 1 hash にする、または `html_hash` を足して kind を増やす。

- 結合: 一覧が「何か変わった」までしか分からない。content / links タブが壊れる。HTML ノイズで誤検知
- `html_hash`: 手動 HTML 編集は追えるが、クロールのたびに content 相当が立つ。今回は採用しない



### 6. Load では markdown だけ先読み

html / raw だけ遅延。

- 悪い点: ノード数 × markdown が IPC に残る。巨大サイトでは足りない。型も `CrawlResultDTO` のままなので再発しやすい



## node 以外の SELECT ALL

今回の実装対象外。本文 overflow ほどではない。


| 箇所                                           | 何が起きるか                                                            | 扱い                          |
| -------------------------------------------- | ----------------------------------------------------------------- | --------------------------- |
| `SaveWorkspaceSettings` / `SaveNodeSettings` | グラフ全体を `LoadWorkspaceBundle` → `SaveWorkspaceBundle` で全ノード upsert | 部分更新 API は別タスク              |
| `GetCrawlRuns`                               | run 存在確認なのに `summary_json` ごと全件                                   | 列投影で足りる。run は WS あたり最大 20   |
| `app_config.defaults_json`                   | singleton 1 行                                                     | 放置                          |
| `graph_nodes.node_settings_json`             | WS 読み込み時に全ノード分                                                    | PartialConfig 程度。別表化の効果は小さい |
| `workspaces.settings_json`                   | WS 1 行                                                            | 放置                          |


`graph_nodes` の `SELECT *` はノード結果の本文問題とは別。設定 JSON を `node_results` と同じ理由で恐れる必要はない。

## 確定判断

grilling で固定した分岐。

1. 本文は遅延読み込み。構造として Select All が本文に届かないこと
2. 既存 SQLite はマイグレーションで移す
3. 本文保持は latest 成功 + baseline。メタは 20 件
4. 同一 SQLite の表分割。ファイル分離はしない
5. `links_json` / `metadata_json` は本文表。メタには `links_hash`
6. Load の `lastResult` は slim。markdown 先読みはしない
7. 公開 API は用途別（パネル / 指定 ID 本文 / 差分詳細）。Load と差分一覧はメタのみ
8. `content_hash` = Markdown のみ。HTML は差分に出さない
9. DTO を `CrawlResultMetaDTO` と `CrawlResultDTO` に分ける
10. `GetCrawlRuns` や設定保存の全件 upsert は同じ対策に含めない

