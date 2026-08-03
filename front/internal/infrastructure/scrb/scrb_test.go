package scrb

import (
	"testing"

	"meguri-app/internal/model"
)

// TestExportImport は .scrb / .crawlproj 形式のエクスポート・インポート往復を検証する。
func TestExportImport(t *testing.T) {
	t.Run("正常系: エクスポートしたバンドルをインポートして内容を復元できる", func(t *testing.T) {
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID: model.StrPtr("old-id"), Name: "Demo", SeedURL: "https://example.com",
				SettingsJSON: `{}`, ExcludeUrlsJSON: `[]`, GraphLayoutDirection: model.StrPtr("LR"),
				CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
			},
			Nodes: []model.GraphNode{{
				WorkspaceID: "old-id", ID: "n1", URLNormalized: "https://example.com",
				Label: "ex", PositionX: 0, PositionY: 0, NodeSettingsJSON: `{}`, Status: model.StrPtr("idle"),
			}},
		}
		data, err := Export(bundle)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Import(data)
		if err != nil {
			t.Fatal(err)
		}
		if got.Workspace.Name != "Demo" {
			t.Fatalf("name=%q", got.Workspace.Name)
		}
		if len(got.Results) != 0 {
			t.Fatalf("results=%d want 0", len(got.Results))
		}
	})

	t.Run("正常系: results.json がある場合は最新成功結果を復元する", func(t *testing.T) {
		md := "# hello"
		html := "<p>hello</p>"
		raw := "<html><body><p>hello</p></body></html>"
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID: model.StrPtr("old-id"), Name: "WithResults", SeedURL: "https://example.com",
				SettingsJSON: `{}`, ExcludeUrlsJSON: `[]`, GraphLayoutDirection: model.StrPtr("LR"),
				CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
			},
			Nodes: []model.GraphNode{{
				WorkspaceID: "old-id", ID: "n1", URLNormalized: "https://example.com",
				Label: "ex", PositionX: 0, PositionY: 0, NodeSettingsJSON: `{}`, Status: model.StrPtr("success"),
			}},
			Results: []model.NodeResult{{
				ID: model.StrPtr("r1"), RunID: "run-1", WorkspaceID: "old-id", NodeID: "n1",
				URL: "https://example.com", Markdown: &md, HTML: &html, RawHTML: &raw,
				FetchedAt: "2026-01-02T00:00:00Z", ManuallyEdited: 0,
			}},
		}
		data, err := Export(bundle)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Import(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Results) != 1 {
			t.Fatalf("results=%d want 1", len(got.Results))
		}
		if got.Results[0].NodeID != "n1" {
			t.Fatalf("node_id=%q", got.Results[0].NodeID)
		}
		if got.Results[0].Markdown == nil || *got.Results[0].Markdown != md {
			t.Fatalf("markdown=%v", got.Results[0].Markdown)
		}
		if got.Results[0].HTML == nil || *got.Results[0].HTML != html {
			t.Fatalf("html=%v", got.Results[0].HTML)
		}
		if got.Results[0].RawHTML == nil || *got.Results[0].RawHTML != raw {
			t.Fatalf("raw_html=%v", got.Results[0].RawHTML)
		}
	})

	t.Run("正常系: results.json が無い旧 ZIP は結果なしで読める", func(t *testing.T) {
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID: model.StrPtr("old-id"), Name: "Legacy", SeedURL: "https://example.com",
				SettingsJSON: `{}`, ExcludeUrlsJSON: `[]`,
				CreatedAt: "2026-01-01T00:00:00Z", UpdatedAt: "2026-01-01T00:00:00Z",
			},
			Nodes: []model.GraphNode{{
				WorkspaceID: "old-id", ID: "n1", URLNormalized: "https://example.com",
				Label: "ex", PositionX: 0, PositionY: 0, NodeSettingsJSON: `{}`, Status: model.StrPtr("idle"),
			}},
		}
		data, err := Export(bundle)
		if err != nil {
			t.Fatal(err)
		}
		got, err := Import(data)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Results) != 0 {
			t.Fatalf("results=%v want empty", got.Results)
		}
	})
}
