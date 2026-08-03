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

// TestWorkspaceServiceExportImportResults は結果付き .crawlproj 相当の Export/Import を検証する。
func TestWorkspaceServiceExportImportResults(t *testing.T) {
	t.Run("正常系: includeResults で最新成功のみ載せ Import 後 Load で lastResult が復元される", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		wsID := "ws-src"
		seedDuplicateSource(t, ctx, store, wsID)

		require.NoError(t, store.BeginCrawlRun(ctx, model.CrawlRun{
			ID: model.StrPtr("run-1"), WorkspaceID: wsID, Mode: 1,
			Status: model.StrPtr("completed"), StartedAt: "2026-01-01T01:00:00Z",
			FinishedAt: model.StrPtr("2026-01-01T01:01:00Z"),
		}))
		md := "# ok"
		html := "<p>ok</p>"
		raw := "<html>ok</html>"
		errMsg := "fail"
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("r-ok"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n1",
			URL: "https://example.com/", Markdown: &md, HTML: &html, RawHTML: &raw,
			FetchedAt: "2026-01-01T01:00:30Z",
		}))
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("r-fail"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n2",
			URL: "https://example.com/a", Error: &errMsg,
			FetchedAt: "2026-01-01T01:00:40Z",
		}))

		exported, err := svc.ExportBundle(ctx, wsID, true)
		require.NoError(t, err)
		require.Len(t, exported.Results, 1)
		assert.Equal(t, "n1", exported.Results[0].NodeID)

		newID, err := svc.ImportBundle(ctx, *exported)
		require.NoError(t, err)

		loaded, err := svc.Load(ctx, newID)
		require.NoError(t, err)
		require.NotNil(t, loaded)
		require.Len(t, loaded.Nodes, 2)

		var withResult, withoutResult int
		for _, n := range loaded.Nodes {
			if n.LastResult != nil {
				withResult++
				assert.Equal(t, md, n.LastResult.Markdown)
				assert.Equal(t, html, n.LastResult.HTML)
				assert.Equal(t, raw, n.LastResult.RawHTML)
			} else {
				withoutResult++
			}
		}
		assert.Equal(t, 1, withResult)
		assert.Equal(t, 1, withoutResult)
	})

	t.Run("正常系: includeResults=false では Results が空", func(t *testing.T) {
		ctx, store, svc := setupWorkspaceTestService(t)
		wsID := "ws-src"
		seedDuplicateSource(t, ctx, store, wsID)
		require.NoError(t, store.BeginCrawlRun(ctx, model.CrawlRun{
			ID: model.StrPtr("run-1"), WorkspaceID: wsID, Mode: 1,
			Status: model.StrPtr("completed"), StartedAt: "2026-01-01T01:00:00Z",
		}))
		md := "# ok"
		require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
			ID: model.StrPtr("r-ok"), RunID: "run-1", WorkspaceID: wsID, NodeID: "n1",
			URL: "https://example.com/", Markdown: &md, FetchedAt: "2026-01-01T01:00:30Z",
		}))

		exported, err := svc.ExportBundle(ctx, wsID, false)
		require.NoError(t, err)
		assert.Empty(t, exported.Results)
	})

	t.Run("異常系: 結果挿入失敗時は補償削除で WS を残さない", func(t *testing.T) {
		// 同一 node_id の成功結果を2件載せ、合成 run 上で UNIQUE(run_id, node_id) 衝突させる。
		ctx, store, svc := setupWorkspaceTestService(t)
		wsID := "ws-src"
		seedDuplicateSource(t, ctx, store, wsID)

		md := "# ok"
		bundle, err := svc.ExportBundle(ctx, wsID, false)
		require.NoError(t, err)
		bundle.Results = []model.NodeResult{
			{
				ID: model.StrPtr("r1"), RunID: "old-run", WorkspaceID: wsID, NodeID: "n1",
				URL: "https://example.com/", Markdown: &md, FetchedAt: "2026-01-01T01:00:00Z",
			},
			{
				ID: model.StrPtr("r2"), RunID: "old-run", WorkspaceID: wsID, NodeID: "n1",
				URL: "https://example.com/", Markdown: &md, FetchedAt: "2026-01-01T02:00:00Z",
			},
		}

		before, err := store.ListWorkspaces(ctx)
		require.NoError(t, err)

		newID, err := svc.ImportBundle(ctx, *bundle)
		require.Error(t, err)
		assert.Empty(t, newID)

		after, err := store.ListWorkspaces(ctx)
		require.NoError(t, err)
		assert.Len(t, after, len(before))
	})
}
