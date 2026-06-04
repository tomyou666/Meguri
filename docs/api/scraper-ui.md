# Scraper UI API

Wails メソッド名と将来の HTTP REST を併記。TypeScript 契約は `ScraperPort`（`front/frontend/src/types/adapter.ts`）。

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

## Crawl

| Wails | HTTP | 説明 |
|-------|------|------|
| `StartCrawl` | `POST /api/v1/workspaces/{wsId}/crawl` | body: `{ mode, startNodeId?, nodeIds? }` |
| `PauseCrawl` | `POST /api/v1/workspaces/{wsId}/crawl/pause` | |
| `StopCrawl` | `POST /api/v1/workspaces/{wsId}/crawl/stop` | |

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
