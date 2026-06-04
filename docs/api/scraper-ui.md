# Scraper UI API

Wails メソッド名と将来の HTTP REST を併記。TypeScript 契約は `ScraperPort`（`front/frontend/src/types/adapter.ts`）。

## 通信モデル

| 種別 | デスクトップ（Wails） | HTTP REST（将来） |
|------|----------------------|-------------------|
| 設定・WS・結果の CRUD | Wails メソッド（同期 RPC） | 対応する REST |
| クロール開始・一時停止・停止 | Wails メソッド（同期 RPC） | `POST .../crawl` 等 |
| **クロール進捗（ノード単位のリアルタイム更新）** | **Wails 組み込み Event（pub/sub）** | 本ドキュメントの正規経路外（必要時に別設計） |

デスクトップでは、再生中のグラフ・右 SB 更新は **WebSocket / SSE ではなく Wails Event** で行う。Go 側が `application.Events.Emit`、フロントの `WailsScraperAdapter` が `Events.On` で購読し、既存 `StartCrawlParams` のコールバック（`onNodeStarted` 等）へ渡す。`appStore.startCrawl` の UI 更新ロジックはそのまま維持する。

Phase 1（現状）の `MockScraperAdapter` は同一プロセス内で `crawlStub` がコールバックを直接呼び出すのみ（Event 未使用）。

## App config

| Wails | HTTP | 説明 |
|-------|------|------|
| `GetAppDefaults` | `GET /api/v1/app/defaults` | アプリ既定設定 |
| `SetAppDefaults` | `PUT /api/v1/app/defaults` | 既定設定更新（非推奨: `SaveAppDefaults` を使用） |
| `SaveAppDefaults` | `PUT /api/v1/app/defaults` | 既定設定の保存。body: `PartialConfig`。成功時 `{ ok: true, scope: "app" }` |

## Settings save（UI 保存ボタン）

いずれも body は `PartialConfig`（`config.example.yaml` 準拠）。クライアントは保存前に zod で検証する。

| Wails | HTTP | 説明 |
|-------|------|------|
| `SaveAppDefaults` | `PUT /api/v1/app/defaults` | アプリ全体の既定 |
| `SaveWorkspaceSettings` | `PUT /api/v1/workspaces/{id}/settings` | WS 既定（**request / content / pdf / crawl のみ**。plugins・output は送らない） |
| `SaveDomainSettings` | `PUT /api/v1/workspaces/{id}/domains/{domain}/settings` | ドメイン単位の上書き |
| `SaveNodeSettings` | `PUT /api/v1/workspaces/{id}/nodes/{nodeId}/settings` | ノード単位の上書き（plugins/output は送らない） |

レスポンス例:

```json
{ "ok": true, "scope": "workspace" }
```

エラー時は HTTP 4xx + `{ "ok": false, "errors": ["crawl.max_depth: ..."] }`（将来）。

## Workspace

| Wails | HTTP | 説明 |
|-------|------|------|
| `LoadWorkspace` | `GET /api/v1/workspaces/{id}` | WS + グラフ + ドメイン設定 |
| `SaveWorkspace` | `PUT /api/v1/workspaces/{id}` | 永続化 |
| `DuplicateWorkspace` | `POST /api/v1/workspaces/{id}/duplicate` | 複製 |

## Results

| Wails | HTTP | 説明 |
|-------|------|------|
| `GetNodeResult` | `GET /api/v1/workspaces/{wsId}/nodes/{nodeId}/result` | 最新結果 |
| `GetNodeResults` | `POST /api/v1/workspaces/{wsId}/results/batch` | 複数取得 |
| `MergeResults` | `POST /api/v1/workspaces/{wsId}/results/merge` | マージ表示用テキスト |
| `SaveResults` | `POST /api/v1/workspaces/{wsId}/results/save` | baseline 用保存 |
| `DeleteResults` | `DELETE /api/v1/workspaces/{wsId}/results` | 指定ノードの **最新 1 件のみ** 削除（ノードごとの履歴は最大 20 件を保持） |

## Crawl（コマンド: RPC / REST）

| Wails | HTTP | 説明 |
|-------|------|------|
| `StartCrawl` | `POST /api/v1/workspaces/{wsId}/crawl` | body: `{ mode, startNodeId?, nodeIds? }`。**即時 ACK のみ**（進捗は Event で配信） |
| `PauseCrawl` | `POST /api/v1/workspaces/{wsId}/crawl/pause` | |
| `StopCrawl` | `POST /api/v1/workspaces/{wsId}/crawl/stop` | |

Wails では上記は **メソッド呼び出し（RPC）**。長時間のクロール本体は戻り値に載せず、進捗は下記 Event で逐次通知する。

## Crawl progress（Wails Events — pub/sub）

Wails 組み込み Event を **クロール進捗の唯一のリアルタイム経路** とする（デスクトップ）。トピック名は文字列定数で Go / TS 共有する。

| トピック | 発火タイミング | payload（概念） |
|----------|----------------|-----------------|
| `scraper:crawl:nodeStarted` | ノード処理開始 | `{ workspaceId, nodeId, url }` |
| `scraper:crawl:nodeSucceeded` | 取得成功 | `{ workspaceId, nodeId, result: CrawlResultPreview }` |
| `scraper:crawl:nodeFailed` | 取得失敗 | `{ workspaceId, nodeId, url, error }` |
| `scraper:crawl:nodeSkipped` | スキップ | `{ workspaceId, nodeId, url, reason }` |
| `scraper:crawl:edgeDiscovered` | 新規リンクでグラフ拡張 | `{ workspaceId, sourceId, targetId, targetUrl }` |
| `scraper:crawl:completed` | 正常終了 | `{ workspaceId, summary: CrawlRunSummary（id/startedAt 除く） }` |
| `scraper:crawl:error` | 実行全体の失敗 | `{ workspaceId, message }` |

- **Go**: `ScraperService` が crawler の `ResultSink` / 将来の `ProgressEvent` を受け、`Events.Emit(ctx, topic, payload)`。
- **TS**: `WailsScraperAdapter.startCrawl` 内で `Events.On` を登録し、解除は `startCrawl` 完了または `StopCrawl` / abort 時。ハンドラから `StartCrawlParams` の同名コールバックを呼ぶ。
- **購読ライフサイクル**: `StartCrawl` 呼び出し前に `On` を張り、`completed` / `error` または stop 後に `Off`（重複購読防止）。
- **HTTP**: 上記 Event に相当する SSE / WebSocket は **未定義**。リモート API が必要になった場合のみ別途追加する。

`StartCrawlParams`（[`adapter.ts`](../front/frontend/src/types/adapter.ts)）のコールバック名・引数型と payload を一致させる。

## Diff (Phase4)

| Wails | HTTP | 説明 |
|-------|------|------|
| `GetWorkspaceDiff` | `GET /api/v1/workspaces/{wsId}/diff` | baseline vs current |
| `SaveResultsSnapshot` | `POST /api/v1/workspaces/{wsId}/snapshot` | `workspaces.baseline_run_id` を更新し、当該 `crawl_runs.id` に baseline 用 `node_results` を格納 |

### 差分の定義（3 種）

| 種別 | 比較元 |
|------|--------|
| `content` | baseline run と現在の最新成功行の `content_hash`（canonical markdown の SHA-256） |
| `links` | 同上の **`links_json` のみ**（グラフ出辺 `graph_edges` は使わない） |
| `fetch` | baseline 行の成否（`error` の有無）と現在の最新成功行 / `graph_nodes.status` |

## DTO

- 設定 JSON: `AppConfig` / `PartialConfig`（`config.example.yaml` 準拠）
- 永続層: [`front/storage/schema.sql`](../front/storage/schema.sql)
- `content_hash`: [`front/frontend/src/lib/contentHash.ts`](../front/frontend/src/lib/contentHash.ts) の算法と一致させる
