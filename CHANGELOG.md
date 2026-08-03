# 変更履歴

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/).

## [Unreleased]

## [0.11.0] - 2026-08-04

### 追加

- `.crawlproj` 保存時に最新成功のクロール結果を含められるようにした（開くと結果も復元）
- 左サイドバーでワークスペース名を変更できる編集ダイアログを追加
- エクスポート順ツリーにラベル / URL 検索を追加
- `content.include_tags` 用 Filter `include_tags` を追加し、指定時にパイプラインへ自動組み込みするようにした
- `content.exclude_tags` 用 Filter `exclude_tags` を追加し、指定時にパイプラインへ自動組み込みするようにした
- デフォルト `exclude_tags` に `dialog` を追加した
- robots タブの各 host 行に host 名コピーボタンを追加
- ワークスペースコピーで「設定のみ / 設定とノード構成」を選べるようにした
- 左サイドバーにノードツリー（検索・status フィルタ・複数選択・URL コピー・コンテキストメニュー）と robots タブを追加
- `content.exclude_selectors`（CSS セレクタ除外）と Filter `exclude_selectors`、フロント設定・CLI `--exclude-selector` を追加
- グラフツール切替ショートカット（手のひら `H` / 矩形選択 `V`）を追加
- ノード右クリックメニューで「クロールしない／クロールする」をトグルできるようにした
- グラフのノード単一クリックで、ツリー側の該当行を展開してスクロール表示するようにした
- ノード個別設定の作成（未作成時は作成ボタンのみ）と削除を追加
- `crawl.include_hosts` / `crawl.exclude_hosts`（ホスト完全一致フィルタ）と CLI `--include-host` / `--exclude-host`、設定フォームを追加
- 設定の複数値入力（TagList）バッジに全文ツールチップを表示するようにした

### 修正

- 結果なしの skipped / クロールしないノードが「変更を確認済みにする」後も fetch 差分として残っていたのを修正した
- 「変更を確認済みにする」実行中のローディング表示と二重実行防止を追加した
- mode4 連打時に `TrimCrawlRuns` が古い `crawl_runs` を消すと CASCADE で無関係ノードの `node_results` まで消えていたのを修正
  - `node_results.run_id` → `crawl_runs` の `ON DELETE CASCADE` を外すマイグレーションを追加した
- グラフノード選択時の再レンダーを抑え、ノード数が多いときのクリック FPS を改善した
  - ノード上の Radix Tooltip をネイティブ `title` に置換した
  - `data.selected` 二重持ちをやめ、`memo` で React Flow の `width`/`height` 計測ノイズを無視するようにした
- ノードツリー選択時の再レンダーを抑え、クリック FPS を改善した
  - `TreeRow` を `memo` 化し、行内 Tooltip をネイティブ `title` に置換した
  - グラフクリックの選択・ツリー追従・RF sync 抑制を同一 `set` にまとめた
- 再取得 OFF 時の既存 success 展開を高速化した
  - skip scrape ジョブでは `request_delay` を待たないようにした
  - `SkipScrapeURLs` の enqueue では robots 判定を省略するようにした
  - 既存向け `already_success` / `duplicate_existing` の linkSkipped UI 連打を抑え、集計のみ残すようにした
  - リンク発見ログを Debug に下げた
- chromium 取得で Navigate・待機・HTML 取得を同一 `chromedp.Run` にまとめ、`waitCtx` cancel 後の別 Run によるタブデッドロックを解消した
  - HTML 取得を `OuterHTML` から `Evaluate(document.documentElement.outerHTML)` に変更
  - リクエスト期限切れが `context canceled` に化ける問題を `preferRequestContextError` で修正
- Export ツリーの初期チェックで `crawlExclude` と非 success ノードを OFF にするようにした
- ノード選択時に `lastResult` があれば再取得せず、DB 側は nodeIDs 絞り込みで `GetNodeResult(s)` を軽くした
  - クロール開始時に対象ノードの `lastResult`（選択中なら `loadedNodeResult` も）を無効化し、再クロール後の古い表示を防ぐ
  - repo メソッド名を `GetNodeResultsByNodeIDs` にリネームした
- `content.only_main_content` を Filter `maincontent` の起動と連動させ、`selector` 指定時は maincontent を入れないように修正
- Filter チェーンを固定順（selector → maincontent → exclude_selectors → include_tags → exclude_tags）で組み立てるように修正
- `maincontent` の script/style/noscript ハードコード除去をやめ、`exclude_tags` に一本化した
- Biome 2.5 向けに `biome.json` を移行し、`useSortedClasses` の class 並びを修正
- `content.selector` 指定時に Filter `selector` を自動でパイプラインへ組み込むように修正
- `content.exclude_selectors` 指定時に Filter `exclude_selectors` を自動でパイプラインへ組み込むように修正
- 既存ノード再取得 OFF でも保存リンクから BFS を続け、増やした `max_pages` 分の未訪問をクロールするように修正
- ノード設定の保存を replace にし、編集可能でないキーがマージされて意図せぬ動作になる問題を修正
- クロール時のノード設定マージで `content` 以外を無視するように修正
- ノード設定のリセット先を空オブジェクトにし、デフォルト設定の丸コピーをやめた

## [0.10.0] - 2026-07-28

### 追加

- ノード status 一括照会 API `GetGraphNodeStatuses` を追加
- ノード結果取得中のスケルトン表示を追加

### 修正

- クロール完了後にノードが running のまま残る問題を修正（成功 Event の軽量化・完了時 reconcile・runId 確定前キュー）]
- 脆弱性修正

## [0.9.0] - 2026-07-13

### 修正

- クロール enqueue で安い判定と `max_pages` を robots より先に行い、枠外 URL の robots 取得を避けるように変更
- robots.txt キャッシュのロック範囲を縮小し、同一ホストの同時ミスを singleflight で 1 回にまとめるように変更
- robots.txt キャッシュキーを host 単位にし、http/https で共有するように変更

## [0.8.0] - 2026-07-11

### 追加

- chromium フェッチャーの `wait_until=load` 成功後に SPA 等のレンダリング完了を待つ `wait_after_load` 設定を追加（wait_timeout とは独立）
- ドメインステータスの robots.txt 取得結果をキャッシュ（24時間TTL）。再起動後はキャッシュがあれば再取得しない
- 設定エディタにリセットボタンを追加（app は既定値、workspace/node はデフォルト設定へドラフトを戻す）

### 修正

- コントロールバーの再生モード・既存ノード再取得の設定をキャッシュに保存し、再起動後も保持されるように変更

### 削除

- 設定画面の `crawl.enabled` チェックボックス（実行時は常に上書きされ効果がなかったため）

## [0.7.0] - 2026-07-10

### 追加

- ステルス設定の `lang` / `accept_language` を主な国のプリセットからセレクト選択可能に（カスタム自由入力あり）
- `plugins.stealth` ステルス設定（`http` / `chromium`）。取得方法タブにステルス対策グループを追加
- chromium フェッチャーに `wait_until`（`none` / `load` / `network_idle` / `selector`）によるページ読み込み待機を追加。`wait_timeout` を待機フェーズに配線
- `network_idle_request_max_age`（通信の打ち切り時間）設定を追加。`wait_until=network_idle` 時に終わらない通信を諦める上限を指定可能
- UI 設定の `fetcher_config`（待機設定含む）を backend に正しく反映
- ルート・backend・front の Makefile に `make generate` を追加（codegen 一括実行）
- クロール時の URL 正規化を info ログ出力（raw / normalized）

### 変更

- UA / headless を `fetcher_config` / `request.headers` から `plugins.stealth` へ移動（**互換破壊**）

### 削除

- CLI `--fetcher-user-agent` / `--fetcher-headless`
- `plugins.stealth.chromium.disable_infobars`（Chromium から `--disable-infobars` が削除済みのため。情報バー非表示は `hide_automation` で対応）

### 修正

- chromium 共有ブラウザが最初のリクエスト context キャンセルで終了しないよう修正（後続取得の `context canceled` を解消）
- `network_idle` 待機をメインフレームの通信のみ監視するよう変更。iframe 配下・長寿命接続を除外し、終わらない通信は `network_idle_request_max_age` で諦める
- SQLite 接続に WAL・synchronous(NORMAL)・busy_timeout(5000) を適用し、crawl 中の UI 読み取りと書き込み競合を緩和
- chromium `hide_automation` が `--enable-automation` を外すよう修正（`excludeSwitches` は CLI 非対応のため）
- Windows 等で不要な `--no-sandbox` 付与をやめ、サポート外フラグの infobar 表示を抑制

## [0.6.0] - 2026-07-05

### 追加

- ワークスペース新規作成で、アプリ設定をコピーするように変更
- robots.txtの取得失敗時は手動で取得できるように変更
- ノード結果パネルの URL 横にコピーアイコンを追加
- アプリ終了時に active crawl を停止し chromium 共有プールを強制クローズする ServiceShutdown を追加

### 修正

- ノード結果パネルのエラー表示が枠をはみ出す問題を修正
- chromium クロール中の robots.txt 取得で User-Agent ヘッダが付与されず `inconsistent chromium user-agent` になる問題を修正
- chromium PDF 取得で HTTP 403 等の非 PDF 応答をパースしようとする問題を修正
  - 取得段階で HTTP ステータスと content-type を含むエラーを返すように変更
- front の golangci-lint が node_modules 内の Go コードを走査して失敗する問題を修正

## [0.5.0] - 2026-07-02

### 修正

- 設定系の入力UXを改善
- ノードを右クリックした際のメニューを最適化
- ミニマップの状態を保持するように変更
- アプリを閉じた際にノードの状態を保存するように変更
- chrominiumを利用した際にプロセスが残る不具合を修正
- デフォルト設定で出力先を削除

## [0.4.0] - 2026-06-30

### 修正

- テキストをctrl + c でコピーできない問題を修正
- PDFのFetch方法について、Chromiumを選択した場合にnet/httpを利用してしまう問題を修正
  - CDPを使用するように修正

## [0.3.0] - 2026-06-29

### 修正

- PDF URL 取得: PDF取得がうまくできていなかった問題を修正

## [0.2.0] - 2026-06-29

### 追加
- 自動更新ダイアログを追加
- CHANGELOG を追加

### 修正
- 自動更新機能が正常に動作していない問題を修正

### その他
- 古いドキュメントを削除

## [0.1.0] - 2026-06-27

### 追加
- 初回リリース

### 修正
-