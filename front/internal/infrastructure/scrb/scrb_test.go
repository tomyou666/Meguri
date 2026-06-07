package scrb

import (
	"testing"

	"scraperbot-front/internal/model"
)

func TestExportImportRoundTrip(t *testing.T) {
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
}
