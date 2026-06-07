# Scraper UI API

Wails メソッド名と将来の HTTP REST を併記。TypeScript 契約は `ScraperPort`（[`front/frontend/src/types/adapter.ts`](../front/frontend/src/types/adapter.ts)）。

**命名**: Wails 公開メソッドは **PascalCase**。TS `ScraperPort` は **camelCase**（Wails bindings が変換）。

## フェーズとサービス境界

| フェーズ | Go Wails サービス | クロール実行 | 進捗配信 |
|----------|-------------------|--------------|----------|
| **Phase 2（現行）** | **`StoreService`**（SQLite CRUD）+ **`ProjectService`**（.scrb） | TS **`compositeScraperAdapter`** + `crawlStub` | コールバック + **StoreService で run/result を永続化** |
| Phase 3 | **`ScraperService`**（`StartCrawl` / `PauseCrawl` / `StopCrawl`） | Go crawler | Wails Event |

Phase 2 では `StartCrawl` 等の Go RPC は**未実装**。pause/stop は TS `appStore` が `AbortSignal` / フラグで処理する。

## 通信モデル

| 種別 | デスクトップ（Wails） | HTTP REST（将来） |
|------|----------------------|-------------------|
| 設定・WS・結果の CRUD | Wails メソッド（同期 RPC） | 対応する REST |
| クロール開始 | TS `crawlStub`（Phase 2） | `POST .../crawl`（Phase 3） |
| クロール進捗 | コールバック（Phase 2） / Wails Event（Phase 3 予定） | 別設計 |

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

## Crawl 永続化（StoreService、ScraperPort 外）

`compositeScraperAdapter.startCrawl` 内で `crawlStub` と併用:

| Wails | タイミング |
|-------|-----------|
| `BeginCrawlRun` | crawl 開始 |
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
