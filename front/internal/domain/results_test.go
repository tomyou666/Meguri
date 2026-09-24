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
		"000006_node_results_drop_run_cascade.up.sql",
		"000007_node_results_reorder_small_cols.up.sql",
		"000008_node_result_bodies.up.sql",
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
			URL: "https://example.com", FetchedAt: "2026-01-01T00:00:01Z",
		}, &model.NodeResultBody{ID: model.StrPtr("nr-1"), Markdown: &m1}))
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-2"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n2",
			URL: "https://example.com/a", FetchedAt: "2026-01-01T00:00:02Z",
		}, &model.NodeResultBody{ID: model.StrPtr("nr-2"), Markdown: &m2}))
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-3"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n3",
			URL: "https://example.com/b", Error: &errMsg, FetchedAt: "2026-01-01T00:00:03Z",
		}, nil))

		out, err := svc.GetNodeResults(ctx, wsID, []string{"n2", "n3", "n1", "missing"})
		require.NoError(t, err)
		require.Len(t, out, 2)
		require.Equal(t, m2, out[0].Markdown)
		require.Equal(t, m1, out[1].Markdown)
	})
}

// TestSaveResultsSnapshot は baseline 確定を検証する。
func TestSaveResultsSnapshot(t *testing.T) {
	t.Run("正常系: クロール runId 指定では UNIQUE せずその run を baseline にする", func(t *testing.T) {
		ctx, store, svc := setupResultsTest(t)
		wsID := "ws-snap"
		require.NoError(t, store.SaveWorkspaceBundle(ctx, model.WorkspaceBundle{
			Workspace: model.Workspace{
				ID:                   model.StrPtr(wsID),
				Name:                 "Snap",
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
			},
		}))
		require.NoError(t, store.BeginCrawlRun(ctx, model.CrawlRun{
			ID:          model.StrPtr("run-1"),
			WorkspaceID: wsID,
			Mode:        1,
			Status:      model.StrPtr("completed"),
			StartedAt:   "2026-01-01T00:00:00Z",
		}))
		md := "# body"
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("nr-1"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n1",
			URL: "https://example.com", FetchedAt: "2026-01-01T00:00:01Z",
		}, &model.NodeResultBody{ID: model.StrPtr("nr-1"), Markdown: &md}))

		got, err := svc.SaveResultsSnapshot(ctx, wsID, "run-1")
		require.NoError(t, err)
		require.Equal(t, "run-1", got)

		bundle, err := store.LoadWorkspaceBundle(ctx, wsID)
		require.NoError(t, err)
		require.Equal(t, "run-1", model.StrVal(bundle.Workspace.BaselineRunID))

		metas, err := store.GetNodeResultMetas(ctx, wsID)
		require.NoError(t, err)
		require.Len(t, metas, 1)
		require.Equal(t, "run-1", metas[0].RunID)

		preview, err := svc.GetNodeResult(ctx, wsID, "n1")
		require.NoError(t, err)
		require.NotNil(t, preview)
		require.Equal(t, md, preview.Markdown)
	})
}
