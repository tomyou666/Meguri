# .scrb v1（Scraper Bot プロジェクトファイル）

1 ワークスペース分のグラフ・設定を ZIP で運ぶ。SQLite DB ファイルや crawl 結果は含めない。

## レイアウト

```
manifest.json
workspace.json
nodes.json
edges.json
ui_state.json
```

## manifest.json

```json
{
  "formatVersion": 1,
  "exportedAt": "2026-01-01T00:00:00Z",
  "app": "scraper-bot",
  "workspaceName": "My Workspace"
}
```

- `formatVersion` が **1 以外**の場合、インポートは拒否する。

## 各 JSON

| ファイル | 内容 |
|----------|------|
| `workspace.json` | `workspaces` 行（`baseline_run_id` はエクスポート時 null） |
| `nodes.json` | `graph_nodes[]` |
| `edges.json` | `graph_edges[]` |
| `ui_state.json` | `graph_ui_state` 行 |

## インポート

- 新規ワークスペースとして取り込む（**ID 再採番**）。
- `crawl_runs` / `node_results` は含まれないため、ノード status は DB 上の値のまま（通常 idle）。
- 旧形式の `domain_settings.json` が含まれていても **無視**する。

## 実装

Go: [`front/internal/infrastructure/scrb/scrb.go`](../../front/internal/infrastructure/scrb/scrb.go)  
Wails: `ProjectService.OpenScrb` / `ProjectService.SaveScrb`
