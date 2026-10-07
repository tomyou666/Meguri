# クロールモードの設定マージ

モード 1〜4 の実行時に、設定をどの順で重ねるかを調べ、実行経路を一つのルールに揃えた記録。

## 結論（採用）

実行（`StartCrawl`）は全モードで次の順にマージする。

1. Go の組み込み既定（`model.Default()`）
2. app（`app_config.defaults_json`）
3. workspace（`workspaces.settings_json`）
4. その URL のノード（`graph_nodes.node_settings_json` の **content のみ**。`formats` は除く）

未登録 URL は 4 を付けない。ノード層は content だけなので、次はラン全体で 1 セット（app → workspace）のまま。

- 探索の上限（`max_depth` / `max_pages` など）
- fetch limiter（`crawl.fetch_limits`）
- request / pdf / plugins

`exclude_urls` は JSON マージのあと、ワークスペースの除外 URL 集合で置き換える。ネストしたオブジェクト（`crawl.fetch_limits` など）はフィールド単位で重ね、配列は置き換える。ページごとの Kernel は使用中は LRU で閉じない。結果プレビューと robots 取得はこの変更の外。

モードの差は起点と探索範囲だけ。

| モード | 起点 | 探索 |
| --- | --- | --- |
| 1 起点 URL から開始 | ワークスペースの起点 URL | リンクを辿る |
| 2 選択ノードから開始 | 選択ノード | リンクを辿る |
| 3 選択ノードから既存ノードのみ | 選択ノード | 既存ノードだけ |
| 4 選択ノードのみ取得 | 選択ノード | なし |

## 変更前

実行時の正は `crawlState.mergedConfig`（`front/internal/usecase/wails_service/scraper_service.go`）。フロントの `mergeConfig` はプレビュー用で、実行には使っていなかった。

| 経路 | 変更前 |
| --- | --- |
| モード 2 | app のみ。workspace と node をスキップ。UI は「デフォルト設定」 |
| モード 1 / 3 / 4 | app → workspace → node content |
| モード 1 / 2 の BFS | 起点ノードで 1 回マージした設定を、探索した全ページに使う |
| fetch limiter | app のみ。workspace の `crawl.fetch_limits` は無視 |
| モード 3 / 4 と manual 再取得 | ノードごとに `mergedConfig` |

ノード JSON の request / crawl などは `FilterNodeUIConfigLayer` が実行前に捨てる。UI の保存も content（formats 除く）だけ。

## 変更後の適用箇所

- `mergedConfig` はモード分岐なし。常に app → workspace → node content
- モード 1 / 2 の BFS は `runConfig`（app → workspace）で上限を決め、ページごとに `pageConfig`
  - 既存ノードに content があればその設定でパイプラインを組む
  - 無い、または未登録 URL はラン共通パイプライン
- `runCrawl` の fetch limiter は `limiterConfig`（app → workspace）を 1 回だけ使う
- モード 2 の表示は「選択ノードから開始」

プレビュー（`mergedPreviewSettings`）は従来どおり app → workspace → ノード全フィールド。robots は app → workspace のまま。
