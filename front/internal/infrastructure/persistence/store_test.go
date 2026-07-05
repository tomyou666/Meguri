package persistence

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"meguri-app/internal/model"
	"meguri-app/internal/sqlitedsn"

	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
)

func applyTestSchema(db *gorm.DB) error {
	for _, name := range []string{
		"000001_init.up.sql",
		"000002_origin.up.sql",
		"000005_node_result_manual_edit.up.sql",
	} {
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

		db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
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

	t.Run("正常系: PatchGraphNodePositions で座標のみ更新できる", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
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

		wsID := "ws-pos"
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID:                   model.StrPtr(wsID),
				Name:                 "Pos",
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
					Label: "a", PositionX: 100, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
			},
		}
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("save ws: %v", err)
		}

		err = store.PatchGraphNodePositions(ctx, wsID, []model.NodePositionPatchDTO{
			{
				NodeID:         "n1",
				Position:       model.PositionDTO{X: 42, Y: 84},
				UserPositioned: true,
			},
			{
				NodeID:         "n2",
				Position:       model.PositionDTO{X: 200, Y: 50},
				UserPositioned: true,
			},
		})
		if err != nil {
			t.Fatalf("patch positions: %v", err)
		}

		loaded, err := store.LoadWorkspaceBundle(ctx, wsID)
		if err != nil || loaded == nil {
			t.Fatalf("load: %v", err)
		}
		byID := map[string]model.GraphNode{}
		for _, n := range loaded.Nodes {
			byID[n.ID] = n
		}
		if byID["n1"].PositionX != 42 || byID["n1"].PositionY != 84 || byID["n1"].UserPositioned != 1 {
			t.Fatalf("n1 position: %+v", byID["n1"])
		}
		if byID["n2"].PositionX != 200 || byID["n2"].PositionY != 50 || byID["n2"].UserPositioned != 1 {
			t.Fatalf("n2 position: %+v", byID["n2"])
		}
		if byID["n2"].Status == nil || *byID["n2"].Status != "success" {
			t.Fatalf("n2 status should be unchanged: %+v", byID["n2"].Status)
		}
	})

	t.Run("正常系: SaveWorkspaceBundle 再保存後も node_results が残る", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
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

		wsID := "ws-results"
		runID := "run-1"
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID:                   model.StrPtr(wsID),
				Name:                 "Results",
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
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
			},
		}
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("save ws: %v", err)
		}
		if err := store.BeginCrawlRun(ctx, model.CrawlRun{
			ID:          model.StrPtr(runID),
			WorkspaceID: wsID,
			Mode:        1,
			Status:      model.StrPtr("running"),
			StartedAt:   "2026-01-01T00:00:00Z",
		}); err != nil {
			t.Fatalf("begin crawl run: %v", err)
		}
		markdown := "# hello"
		if err := store.AppendNodeResult(ctx, model.NodeResult{
			ID:          model.StrPtr("nr-1"),
			RunID:       runID,
			WorkspaceID: wsID,
			NodeID:      "n1",
			URL:         "https://example.com",
			Markdown:    &markdown,
			FetchedAt:   "2026-01-01T00:00:01Z",
		}); err != nil {
			t.Fatalf("append node result: %v", err)
		}

		bundle.Nodes[0].PositionX = 42
		bundle.Nodes[0].PositionY = 84
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("resave ws: %v", err)
		}

		results, err := store.GetNodeResults(ctx, wsID)
		if err != nil {
			t.Fatalf("get node results: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 node result, got %d", len(results))
		}
		if results[0].NodeID != "n1" || model.StrVal(results[0].Markdown) != markdown {
			t.Fatalf("unexpected node result: %+v", results[0])
		}
	})

	t.Run("正常系: SaveWorkspaceBundle でノード除外時のみ node_results が削除される", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
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

		wsID := "ws-prune"
		runID := "run-1"
		bundle := model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID:                   model.StrPtr(wsID),
				Name:                 "Prune",
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
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
				{
					WorkspaceID: wsID, ID: "n2", URLNormalized: "https://example.com/a",
					Label: "a", PositionX: 100, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
			},
		}
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("save ws: %v", err)
		}
		if err := store.BeginCrawlRun(ctx, model.CrawlRun{
			ID:          model.StrPtr(runID),
			WorkspaceID: wsID,
			Mode:        1,
			Status:      model.StrPtr("running"),
			StartedAt:   "2026-01-01T00:00:00Z",
		}); err != nil {
			t.Fatalf("begin crawl run: %v", err)
		}
		m1, m2 := "# one", "# two"
		for _, tc := range []struct {
			id, nodeID, md string
		}{
			{"nr-1", "n1", m1},
			{"nr-2", "n2", m2},
		} {
			md := tc.md
			if err := store.AppendNodeResult(ctx, model.NodeResult{
				ID:          model.StrPtr(tc.id),
				RunID:       runID,
				WorkspaceID: wsID,
				NodeID:      tc.nodeID,
				URL:         "https://example.com",
				Markdown:    &md,
				FetchedAt:   "2026-01-01T00:00:01Z",
			}); err != nil {
				t.Fatalf("append node result %s: %v", tc.id, err)
			}
		}

		bundle.Nodes = bundle.Nodes[:1]
		if err := store.SaveWorkspaceBundle(ctx, bundle); err != nil {
			t.Fatalf("resave ws without n2: %v", err)
		}

		results, err := store.GetNodeResults(ctx, wsID)
		if err != nil {
			t.Fatalf("get node results: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 node result, got %d", len(results))
		}
		if results[0].NodeID != "n1" {
			t.Fatalf("expected n1 result only, got %+v", results[0])
		}
	})
}
