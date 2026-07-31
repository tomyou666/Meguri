package domain_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/libtnb/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"meguri-app/internal/domain"
	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
	"meguri-app/internal/sqlitedsn"
)

func applyWorkspaceTestSchema(db *gorm.DB) error {
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

func setupWorkspaceTestService(t *testing.T) (context.Context, persistence.Repository, *domain.WorkspaceService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(sqlitedsn.DSN(dbPath)), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, applyWorkspaceTestSchema(db))
	sqlDB, _ := db.DB()
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = os.Remove(dbPath)
	})
	ctx := context.Background()
	store := persistence.NewStore(db)
	return ctx, store, domain.NewWorkspaceService(store)
}

func seedDuplicateSource(
	t *testing.T,
	ctx context.Context,
	store persistence.Repository,
	wsID string,
) {
	t.Helper()
	bundle := model.WorkspaceBundle{
		Workspace: model.Workspace{
			ID:                   model.StrPtr(wsID),
			Name:                 "Source",
			SeedURL:              "https://example.com/",
			SettingsJSON:         `{"crawl":{"max_depth":2}}`,
			ExcludeUrlsJSON:      `["https://exclude.example/"]`,
			GraphLayoutDirection: model.StrPtr("TB"),
			BaselineRunID:        model.StrPtr("run-baseline"),
			CreatedAt:            "2026-01-01T00:00:00Z",
			UpdatedAt:            "2026-01-01T00:00:00Z",
		},
		Nodes: []model.GraphNode{
			{
				WorkspaceID: wsID, ID: "n1", URLNormalized: "https://example.com/",
				Label: "root", PositionX: 0, PositionY: 0,
				NodeSettingsJSON: `{"crawl":{"max_pages":1}}`, Origin: "crawl", Status: model.StrPtr("success"),
			},
			{
				WorkspaceID: wsID, ID: "n2", URLNormalized: "https://example.com/a",
				Label: "a", PositionX: 100, PositionY: 0,
				NodeSettingsJSON: `{}`, Origin: "crawl", Status: model.StrPtr("idle"),
			},
		},
		Edges: []model.GraphEdge{
			{WorkspaceID: wsID, ID: "e-n1-n2", SourceNodeID: "n1", TargetNodeID: "n2"},
		},
		UIState: &model.GraphUIState{
			WorkspaceID:          model.StrPtr(wsID),
			CollapsedNodeIdsJSON: `{"collapsed":["n1"],"expandedDetail":[]}`,
		},
	}
	require.NoError(t, store.SaveWorkspaceBundle(ctx, bundle))
}

// TestWorkspaceServiceDuplicate は full / settings モードの複製結果を検証する。
func TestWorkspaceServiceDuplicate(t *testing.T) {
	t.Run("正常系: full は設定・ノード・エッジをコピーし baseline をクリアする", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		seedDuplicateSource(t, ctx, store, "ws-src")

		copy, err := svc.Duplicate(ctx, model.DuplicateWorkspaceRequest{
			ID:   "ws-src",
			Name: "Copy Full",
			Mode: "full",
		})
		require.NoError(t, err)
		require.NotNil(t, copy)
		assert.Equal(t, "Copy Full", copy.Name)
		assert.NotEqual(t, "ws-src", copy.ID)
		assert.Equal(t, "https://example.com/", copy.SeedURL)
		assert.JSONEq(t, `{"crawl":{"max_depth":2}}`, string(copy.Settings))
		assert.Equal(t, []string{"https://exclude.example/"}, copy.ExcludeURLs)
		assert.Equal(t, "TB", copy.GraphLayoutDirection)
		assert.Empty(t, copy.BaselineRunID)
		require.Len(t, copy.Nodes, 2)
		require.Len(t, copy.Edges, 1)
		assert.Equal(t, []string{"n1"}, copy.CollapsedNodeIDs)
		for _, n := range copy.Nodes {
			assert.Equal(t, "idle", n.Status)
			assert.NotEqual(t, "n1", n.ID)
			assert.NotEqual(t, "n2", n.ID)
		}
	})

	t.Run("正常系: settings は設定のみコピーしシードノード1つを作る", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		seedDuplicateSource(t, ctx, store, "ws-src")

		copy, err := svc.Duplicate(ctx, model.DuplicateWorkspaceRequest{
			ID:      "ws-src",
			Name:    "Copy Settings",
			Mode:    "settings",
			SeedURL: "https://new.example.com/path",
		})
		require.NoError(t, err)
		require.NotNil(t, copy)
		assert.Equal(t, "Copy Settings", copy.Name)
		assert.JSONEq(t, `{"crawl":{"max_depth":2}}`, string(copy.Settings))
		assert.Equal(t, []string{"https://exclude.example/"}, copy.ExcludeURLs)
		assert.Equal(t, "TB", copy.GraphLayoutDirection)
		assert.Empty(t, copy.BaselineRunID)
		assert.Empty(t, copy.Edges)
		assert.Empty(t, copy.CollapsedNodeIDs)
		require.Len(t, copy.Nodes, 1)
		assert.Equal(t, copy.SeedURL, copy.Nodes[0].URLNormalized)
		assert.Equal(t, copy.SeedURL, copy.Nodes[0].Label)
		assert.Equal(t, "idle", copy.Nodes[0].Status)
		assert.JSONEq(t, `{}`, string(copy.Nodes[0].NodeSettings))
		assert.InDelta(t, 250, copy.Nodes[0].Position.X, 0.01)
		assert.InDelta(t, 200, copy.Nodes[0].Position.Y, 0.01)
	})

	t.Run("異常系: 不正な mode はエラー", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		seedDuplicateSource(t, ctx, store, "ws-src")

		_, err := svc.Duplicate(ctx, model.DuplicateWorkspaceRequest{
			ID:   "ws-src",
			Name: "Bad",
			Mode: "other",
		})
		require.Error(t, err)
	})

	t.Run("異常系: settings で不正な seedUrl はエラー", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		seedDuplicateSource(t, ctx, store, "ws-src")

		_, err := svc.Duplicate(ctx, model.DuplicateWorkspaceRequest{
			ID:      "ws-src",
			Name:    "Bad URL",
			Mode:    "settings",
			SeedURL: "not-a-url",
		})
		require.Error(t, err)
	})
}
