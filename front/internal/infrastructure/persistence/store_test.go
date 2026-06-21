package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"scraperbot-front/internal/model"

	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
)

func applyTestSchema(db *gorm.DB) error {
	for _, name := range []string{"000001_init.up.sql", "000002_origin.up.sql"} {
		path := filepath.Join("..", "..", "app", "migrations", name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			return err
		}
	}
	return nil
}

// TestStore は DB 初期化とワークスペースの保存・読み込みを検証する。
func TestStore(t *testing.T) {
	t.Run("正常系: Bootstrap 後にワークスペースを保存・一覧・読み込みできる", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=foreign_keys(1)"), &gorm.Config{})
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		if err := applyTestSchema(db); err != nil {
			t.Fatalf("schema: %v", err)
		}
		sqlDB, _ := db.DB()
		t.Cleanup(func() {
			_ = sqlDB.Close()
			_ = os.Remove(dbPath)
		})

		ctx := context.Background()
		store := NewStore(db)

		if err := store.BootstrapAppConfig(ctx); err != nil {
			t.Fatalf("bootstrap: %v", err)
		}
		cfg, err := store.GetAppConfig(ctx)
		if err != nil || cfg == nil {
			t.Fatalf("get app config: %v", err)
		}

		wsID := "ws-1"
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID:                   model.StrPtr(wsID),
				Name:                 "Test",
				SeedURL:              "https://example.com",
				SettingsJSON:         `{}`,
				ExcludeUrlsJSON:      `[]`,
				GraphLayoutDirection: model.StrPtr("LR"),
				CreatedAt:            "2026-01-01T00:00:00Z",
				UpdatedAt:            "2026-01-01T00:00:00Z",
			},
			Nodes: []model.GraphNode{
				{
					WorkspaceID: wsID, ID: "n1", URLNormalized: "https://example.com",
					Label: "example", PositionX: 0, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("idle"),
				},
				{
					WorkspaceID: wsID, ID: "n2", URLNormalized: "https://example.com/a",
					Label: "a", PositionX: 100, PositionY: 0, UserPositioned: 1,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
				{
					WorkspaceID: wsID, ID: "n3", URLNormalized: "https://example.com/b",
					Label: "b", PositionX: 200, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("skipped"),
				},
			},
		}
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("save ws: %v", err)
		}
		list, err := store.ListWorkspaces(ctx)
		if err != nil || len(list) != 1 {
			t.Fatalf("list: %v len=%d", err, len(list))
		}
		loaded, err := store.LoadWorkspaceBundle(ctx, wsID)
		if err != nil || loaded == nil || loaded.Workspace.Name != "Test" {
			t.Fatalf("load: %v", err)
		}
	})
}
