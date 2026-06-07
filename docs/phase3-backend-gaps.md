# Phase 3 — backend 不足機能一覧

Phase 3 実装状況の記録。設計の前提は [scraper-ui.md](./api/scraper-ui.md) および grill 合意（2026-06）。

## 実装済み（Phase 3 v1）

- [x] **Progress イベント** — `core.ProgressEvent` / `ProgressSink`、`Crawler` / `usecase.Crawl.RunWithProgress`
- [x] **`exclude_urls`** — `CrawlConfig.ExcludeURLs`、Validate、config.example.yaml
- [x] **`ScrapeWithConfig`** — `pkg/runner`（案 A: 呼び出しごと Kernel Init）
- [x] **`CrawlWithProgress`** — `pkg/runner`
- [x] **公開 API** — `backend/pkg/runner`（front は `internal/` 直 import なし）
- [x] **front `ScraperService`** — オーケストレーション、Wails Event、mode 1/2/3 + manual 後段
- [x] **`GraphNode.origin`** — DDL マイグレーション `000002_origin`
- [x] **TS adapter** — Event 購読 + StoreService 永続化、`crawlStub` 削除

## v2 完了（Phase 3 v2）

- [x] **`PauseController`** — backend worker レベルの pause（`runOne` 前 `WaitIfPaused`）
- [x] **`RunnerCache` cfg hash キャッシュ** — Mode 3 / manual 後段の Kernel 再利用（LRU 8）
- [x] **CLI** — `--exclude-url` / `--progress-json`（crawl は `pkg/runner` 経由）
- [x] **Go 側永続化** — `ScraperService` が `CrawlPersistService` 経由で SQLite 更新。`StartCrawl` が `runId` を返す

## アーキテクチャ前提（変更なし）

- GraphNode / GraphEdge DTO、RunMode 訪問順 — **front `ScraperService`**
- 4 層設定マージ — front が JSON を渡し、Go で `runner.MergeUIConfigLayers`
- Wails Event 発火、url ↔ nodeId — **front `ScraperService`**
- SQLite / crawl run 永続化 — **ScraperService → CrawlPersistService**（v2）。TS adapter は UI コールバックのみ

## 実行経路（実装どおり）

| 経路 | モード | backend API |
|------|--------|-------------|
| 本流 BFS | 1, 2 | `runner.CrawlWithProgress` |
| 既存ノードのみ | 3 | `runner.ScrapeWithConfig` × N |
| manual 後段 | 1, 2, 3 | `runner.ScrapeWithConfig`（`origin=manual`、本流未到達のみ） |

## Grill 合意ログ

| 論点 | 決定 |
|------|------|
| グラフオーケストレーション | front 担当。backend は URL エンジン |
| 手動ノード | `origin: manual \| crawl`、本流後に `ScrapeWithConfig` |
| `exclude_urls` | backend に追加（完了） |
| 進捗 | backend が URL 単位で emit（完了） |
| pause | v2 `PauseController`（ScraperService がジョブごとに注入） |
| モード 3 | `Crawler.Run` 不使用、`ScrapeWithConfig` × N |
| 永続化 | TS adapter + StoreService |
| module 境界 | `pkg/runner` ファサード |
