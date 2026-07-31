package domain_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/libtnb/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"meguri-app/internal/domain"
	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
	"meguri-app/internal/sqlitedsn"
)

func applyResultsTestSchema(db *gorm.DB) error {
	for _, name := range []string{
		"000001_init.up.sql",
		"000002_origin.up.sql",
		"000005_node_result_manual_edit.up.sql",
	} {
		path := filepath.Join("..", "..", "internal", "app", "migrations", name)
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

func setupResultsTest(t *testing.T) (context.Context, persistence.Repository, *domain.ResultsService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, applyResultsTestSchema(db))
	sqlDB, _ := db.DB()
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = os.Remove(dbPath)
	})
	ctx := context.Background()
	store := persistence.NewStore(db)
	wsSvc := domain.NewWorkspaceService(store)
	return ctx, store, domain.NewResultsService(store, wsSvc)
}

// TestGetNodeResults は複数 ID の最新成功結果取得を検証する。
func TestGetNodeResults(t *testing.T) {
	t.Run("正常系: リクエスト順で成功結果のみ返す", func(t *testing.T) {
		ctx, store, svc := setupResultsTest(t)
		wsID := "ws-results"
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
					Label: "n1", PositionX: 0, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
				{
					WorkspaceID: wsID, ID: "n2", URLNormalized: "https://example.com/a",
					Label: "n2", PositionX: 100, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("success"),
				},
				{
					WorkspaceID: wsID, ID: "n3", URLNormalized: "https://example.com/b",
					Label: "n3", PositionX: 200, PositionY: 0,
					NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("error"),
				},
			},
		}
		require.NoError(t, store.SaveWorkspaceBundle(ctx, bundle))
		require.NoError(t, store.BeginCrawlRun(ctx, model.CrawlRun{
			ID:          model.StrPtr("run-1"),
			WorkspaceID: wsID,
			Mode:        1,
			Status:      model.StrPtr("running"),
			StartedAt:   "2026-01-01T00:00:00Z",
		}))

		m1, m2 := "# one", "# two"
		errMsg := "boom"
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-1"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n1",
			URL: "https://example.com", Markdown: &m1, FetchedAt: "2026-01-01T00:00:01Z",
		}))
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-2"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n2",
			URL: "https://example.com/a", Markdown: &m2, FetchedAt: "2026-01-01T00:00:02Z",
		}))
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-3"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n3",
			URL: "https://example.com/b", Error: &errMsg, FetchedAt: "2026-01-01T00:00:03Z",
		}))

		out, err := svc.GetNodeResults(ctx, wsID, []string{"n2", "n3", "n1", "missing"})
		require.NoError(t, err)
		require.Len(t, out, 2)
		require.Equal(t, m2, out[0].Markdown)
		require.Equal(t, m1, out[1].Markdown)
	})
}
