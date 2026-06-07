# Scraper UI API

Wails メソッド名と将来の HTTP REST を併記。TypeScript 契約は `ScraperPort`（[`front/frontend/src/types/adapter.ts`](../front/frontend/src/types/adapter.ts)）。

**命名**: Wails 公開メソッドは **PascalCase**。TS `ScraperPort` は **camelCase**（Wails bindings が変換）。

## フェーズとサービス境界

| フェーズ | Go Wails サービス | クロール実行 | 進捗配信 |
|----------|-------------------|--------------|----------|
| **Phase 3（現行）** | **`StoreService`** + **`ProjectService`** + **`ScraperService`** | Go `backend` in-process（`scraperbot/pkg/runner`） | Wails Event → TS adapter コールバック |
| Phase 2（完了） | `StoreService` + `ProjectService` | TS `crawlStub`（削除済み） | コールバック |

## 通信モデル

| 種別 | デスクトップ（Wails） | HTTP REST（将来） |
|------|----------------------|-------------------|
| 設定・WS・結果の CRUD | Wails メソッド（同期 RPC） | 対応する REST |
| クロール開始 | `ScraperService.StartCrawl`（非同期） | `POST .../crawl` |
| クロール進捗 | Wails Event（`scraper:crawl:*`） | 別設計 |
| pause / resume / stop | `ScraperService.PauseCrawl` / `ResumeCrawl` / `StopCrawl` | 対応 REST |

## ScraperService（Phase 3）

| Wails | 説明 |
|-------|------|
| `StartCrawl` | ワークスペーススナップショット + マージ済み設定でクロール開始（goroutine） |
| `PauseCrawl` | 実行中ジョブを一時停止（v1: オーケストレーション層） |
| `ResumeCrawl` | 一時停止を解除 |
| `StopCrawl` | context cancel で停止 |

### Wails Event トピック

| topic | 説明 |
|-------|------|
| `scraper:crawl:nodeStarted` | ノード処理開始 |
| `scraper:crawl:nodeSucceeded` | 成功 + result |
| `scraper:crawl:nodeFailed` | 失敗 + error |
| `scraper:crawl:nodeSkipped` | スキップ + reason |
| `scraper:crawl:edgeDiscovered` | リンク発見（sourceId / targetId / targetUrl） |
| `scraper:crawl:completed` | ジョブ完了 + summary |
| `scraper:crawl:error` | ジョブ全体エラー |

payload: `CrawlEventPayload`（`workspaceId`, `runId` + イベント固有フィールド）。TS adapter が `runId` でフィルタし `StoreService` 永続化 + `StartCrawlParams` コールバックへ橋渡し。

## App config

| Wails (`StoreService`) | ScraperPort | HTTP | 説明 |
|------------------------|-------------|------|------|
| `GetAppDefaults` | `getAppDefaults` | `GET /api/v1/app/defaults` | アプリ既定設定 |
| `SetAppDefaults` | `setAppDefaults` | `PUT /api/v1/app/defaults` | 既定設定更新 |
| `SaveAppDefaults` | `saveAppDefaults` | `PUT /api/v1/app/defaults` | 既定設定の保存 |

## Workspace

| Wails | ScraperPort | HTTP | 説明 |
|-------|-------------|------|------|
| `ListWorkspaces` | `listWorkspaces` | `GET /api/v1/workspaces` | WS 一覧（id / name / updatedAt） |
| `LoadWorkspace` | `loadWorkspace` | `GET /api/v1/workspaces/{id}` | WS + グラフ + 設定 |
| `SaveWorkspace` | `saveWorkspace` | `PUT /api/v1/workspaces/{id}` | 永続化 |
| `DuplicateWorkspace` | `duplicateWorkspace` | `POST /api/v1/workspaces/{id}/duplicate` | 複製 |

## Settings save

| Wails | ScraperPort | 説明 |
|-------|-------------|------|
| `SaveWorkspaceSettings` | `saveWorkspaceSettings` | WS 設定 |
| `SaveDomainSettings` | `saveDomainSettings` | ドメイン設定 |
| `SaveNodeSettings` | `saveNodeSettings` | ノード設定 |

## Results

| Wails | ScraperPort | 説明 |
|-------|-------------|------|
| `GetNodeResult` | `getNodeResult` | 最新成功結果 |
| `GetNodeResults` | `getNodeResults` | 複数取得 |
| `MergeResults` | `mergeResults` | マージ表示 |
| `SaveResults` | `saveResults` | baseline 用 |
| `DeleteResults` | `deleteResults` | 最新 1 件削除 |
| `SaveResultsSnapshot` | `saveResultsSnapshot` | baseline snapshot |

## Crawl 永続化（StoreService）

`compositeScraperAdapter.startCrawl` が Event 受信時に呼び出す:

| Wails | タイミング |
|-------|-----------|
| `BeginCrawlRun` | crawl 開始（adapter が `runId` 生成） |
| `AppendNodeResult` | ノード成功/失敗 |
| `PatchGraphNodeStatus` | status 更新 |
| `FinishCrawlRun` | 完了/エラー |

## Project (.scrb)

| Wails (`ProjectService`) | 説明 |
|--------------------------|------|
| `OpenScrb` | ネイティブダイアログで開く → 新規 WS import |
| `SaveScrb` | アクティブ WS をエクスポート |

形式: [`docs/formats/scrb-v1.md`](../formats/scrb-v1.md)

## Diff (Phase 4)

| Wails | ScraperPort |
|-------|-------------|
| `GetWorkspaceDiff` | `getWorkspaceDiff` |

## DTO

- 設定 JSON: `AppConfig` / `PartialConfig`（`config.example.yaml` 準拠）
- 永続層: [`front/storage/schema.sql`](../front/storage/schema.sql)
- Wails DTO: `front/internal/model/api.go` → bindings `frontend/bindings/.../models.ts`
- グラフノード `origin`: `crawl` | `manual`（マイグレーション `000002_origin`）
