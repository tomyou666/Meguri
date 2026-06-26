# Scraper UI API

Wails メソッド名と将来の HTTP REST を併記。TypeScript 契約は `ScraperPort`（[`front/frontend/src/types/adapter.ts`](../front/frontend/src/types/adapter.ts)）。

**命名**: Wails 公開メソッドは **PascalCase**。TS `ScraperPort` は **camelCase**（Wails bindings が変換）。

## フェーズとサービス境界

| フェーズ | Go Wails サービス | クロール実行 | 進捗配信 |
|----------|-------------------|--------------|----------|
| **Phase 3 v2（現行）** | **`StoreService`** + **`ProjectService`** + **`ScraperService`** | Go `backend` in-process（`scraperbot/pkg/runner` + PauseController + RunnerCache） | Wails Event → TS adapter UI コールバック（永続化は Go） |
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
| `StartCrawl` | ワークスペーススナップショット + マージ済み設定でクロール開始（goroutine）。**戻り値 `runId`**。`BeginCrawlRun` は Go 内で実行 |
| `PauseCrawl` | 実行中ジョブを一時停止（backend `PauseController`） |
| `ResumeCrawl` | 一時停止を解除 |
| `StopCrawl` | context cancel で停止 |

### Wails Event トピック

| topic | 説明 |
|-------|------|
| `scraper:crawl:runStarted` | crawl 開始（workspaceId / runId） |
| `scraper:crawl:nodeStarted` | ノード処理開始 |
| `scraper:crawl:nodeSucceeded` | 成功 + result |
| `scraper:crawl:nodeFailed` | 失敗 + error |
| `scraper:crawl:nodeSkipped` | スキップ + reason |
| `scraper:crawl:linkSkipped` | 重複リンクスキップ（url=親 / targetUrl=子 / reason） |
| `scraper:crawl:edgeDiscovered` | リンク発見（sourceId / targetId / targetUrl） |
| `scraper:crawl:completed` | ジョブ完了 + summary |
| `scraper:crawl:error` | ジョブ全体エラー |

payload: `CrawlEventPayload`（`workspaceId`, `runId` + イベント固有フィールド）。TS adapter が `runId` でフィルタし `StartCrawlParams` コールバック（UI 更新）へ橋渡し。ノード永続化は `ScraperService` が Go 内で完結。

### RunMode（`StartCrawlRequest.mode`）

| mode | 意味 | 必須入力 | リンク探索 | manual 後段 |
|------|------|----------|-----------|-------------|
| 1 | 起点 URL から BFS | — | あり | あり |
| 2 | 選択ノードから BFS（app 設定のみ） | `startNodeId` | あり | あり |
| 3 | 選択ノードから有向に**既存ノードを辿る** | `startNodeId` | なし | あり |
| 4 | **選択ノードのみ**取得（辿らない） | `nodeIds`（1 件以上） | なし | なし |

`StartCrawlRequest` フィールド:

| フィールド | 説明 |
|-----------|------|
| `startNodeId` | mode 2 / 3 の起点ノード ID |
| `nodeIds` | mode 4 の訪問対象（入力順で scrape。未知 ID はスキップ） |
| `rescrapeExisting` | `false` 時、status=success のノードは fetch せず `linkSkipped` |

UI トリガー:

- ControlBar 再生ボタン: ドロップダウンで選択した mode を実行
- ControlBar「選択をスクレイプ」/ グラフ右クリック「スクレイプ」: **mode 4**（`runMode` は変更しない）

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
| `DuplicateWorkspace` | `duplicateWorkspace` | `POST /api/v1/workspaces/{id}/duplicate` | 複製。**引数 `name`**（空ならコピー元名） |
| `DeleteWorkspace` | `deleteWorkspace` | `DELETE /api/v1/workspaces/{id}` | WS 削除（CASCADE） |

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

## Crawl 永続化（ScraperService / v2）

`ScraperService` が `CrawlPersistService` 経由で SQLite を更新（TS adapter は RPC しない）:

| 操作 | タイミング |
|------|-----------|
| `BeginCrawlRun` | `StartCrawl` 同期部（`runId` 生成直後） |
| `AppendNodeResult` / `PatchGraphNodeStatus` | 各 node Event 発火前 |
| `UpsertDiscoveredGraph` | `edgeDiscovered` 発火前 |
| `FinishCrawlRun` | completed / stopped / error |

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
| `GetNodeDiffDetail` | `getNodeDiffDetail` |
| `ShowNodeDiffWindow` | `showNodeDiffWindow` |
| `GetNodeDiffViewerSession` | `getNodeDiffViewerSession` |

## DTO

- 設定 JSON: `AppConfig` / `PartialConfig`（`config.example.yaml` 準拠）
- 永続層: [`front/storage/schema.sql`](../front/storage/schema.sql)
- Wails DTO: `front/internal/model/api.go` → bindings `frontend/bindings/.../models.ts`
- グラフノード `origin`: `crawl` | `manual`（マイグレーション `000002_origin`）
