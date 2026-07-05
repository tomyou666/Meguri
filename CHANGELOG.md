# 変更履歴

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/ja/1.1.0/).

## [Unreleased]

### 修正

- SQLite 接続に WAL・synchronous(NORMAL)・busy_timeout(5000) を適用し、crawl 中の UI 読み取りと書き込み競合を緩和

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