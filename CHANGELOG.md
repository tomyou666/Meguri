# 変更履歴

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/).

## [Unreleased]

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