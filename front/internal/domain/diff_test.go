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
)

func applyDiffTestSchema(db *gorm.DB) error {
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

func setupDiffTestStore(t *testing.T) (context.Context, persistence.Repository, *domain.WorkspaceService, *domain.DiffService) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=foreign_keys(1)"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, applyDiffTestSchema(db))
	sqlDB, _ := db.DB()
	t.Cleanup(func() {
		_ = sqlDB.Close()
		_ = os.Remove(dbPath)
	})
	ctx := context.Background()
	store := persistence.NewStore(db)
	wsSvc := domain.NewWorkspaceService(store)
	diffSvc := domain.NewDiffService(store, wsSvc)
	return ctx, store, wsSvc, diffSvc
}

func seedDiffWorkspace(
	t *testing.T,
	ctx context.Context,
	store persistence.Repository,
	wsID, baselineRun, currentRun string,
) {
	t.Helper()
	bundle := model.WorkspaceBundle{
		Workspace: model.Workspace{
			ID:                   model.StrPtr(wsID),
			Name:                 "Diff",
			SeedURL:              "https://example.com",
			SettingsJSON:         `{}`,
			ExcludeUrlsJSON:      `[]`,
			GraphLayoutDirection: model.StrPtr("LR"),
			BaselineRunID:        model.StrPtr(baselineRun),
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
	require.NoError(t, store.SaveWorkspaceBundle(ctx, bundle))
	for _, runID := range []string{baselineRun, currentRun} {
		require.NoError(t, store.BeginCrawlRun(ctx, model.CrawlRun{
			ID:          model.StrPtr(runID),
			WorkspaceID: wsID,
			Mode:        1,
			Status:      model.StrPtr("completed"),
			StartedAt:   "2026-01-01T00:00:00Z",
		}))
	}
}

func appendResult(
	t *testing.T,
	ctx context.Context,
	store persistence.Repository,
	runID, wsID, nodeID string,
	markdown, linksJSON, contentHash string,
	errMsg *string,
	fetchedAt string,
) {
	t.Helper()
	md := markdown
	links := linksJSON
	hash := contentHash
	require.NoError(t, store.AppendNodeResult(ctx, model.NodeResult{
		ID:          model.StrPtr(runID + "-" + nodeID),
		RunID:       runID,
		WorkspaceID: wsID,
		NodeID:      nodeID,
		URL:         "https://example.com",
		Markdown:    &md,
		LinksJSON:   &links,
		ContentHash: &hash,
		Error:       errMsg,
		FetchedAt:   fetchedAt,
	}))
}

// TestDiffService は content / links / fetch 分類と GetNodeDiffDetail の old/new を検証する。
func TestDiffService(t *testing.T) {
	ctx, store, _, diffSvc := setupDiffTestStore(t)
	wsID := "ws-diff"
	baselineRun := "run-baseline"
	currentRun := "run-current"

	t.Run("content: hash 不一致で content 差分", func(t *testing.T) {
		seedDiffWorkspace(t, ctx, store, wsID, baselineRun, currentRun)
		appendResult(t, ctx, store, baselineRun, wsID, "n1", "old body", `["https://a"]`, "hash-a", nil, "2026-01-01T00:00:00Z")
		appendResult(t, ctx, store, currentRun, wsID, "n1", "new body", `["https://a"]`, "hash-b", nil, "2026-01-02T00:00:00Z")

		diff, err := diffSvc.GetWorkspaceDiff(ctx, wsID)
		require.NoError(t, err)
		assert.True(t, diff.HasDiff)
		assert.Equal(t, 1, diff.Summary.Content)
		assert.Equal(t, 0, diff.Summary.Links)
		assert.Equal(t, 0, diff.Summary.Fetch)

		detail, err := diffSvc.GetNodeDiffDetail(ctx, wsID, "n1")
		require.NoError(t, err)
		assert.Equal(t, []string{"content"}, detail.Kinds)
		require.NotNil(t, detail.Content)
		assert.Equal(t, "old body", detail.Content.Old)
		assert.Equal(t, "new body", detail.Content.New)
	})

	t.Run("links: links_json 不一致で links 差分", func(t *testing.T) {
		ctx, store, _, diffSvc := setupDiffTestStore(t)
		wsID := "ws-links"
		seedDiffWorkspace(t, ctx, store, wsID, baselineRun, currentRun)
		appendResult(t, ctx, store, baselineRun, wsID, "n1", "same", `["https://a"]`, "hash-same", nil, "2026-01-01T00:00:00Z")
		appendResult(t, ctx, store, currentRun, wsID, "n1", "same", `["https://b"]`, "hash-same", nil, "2026-01-02T00:00:00Z")

		diff, err := diffSvc.GetWorkspaceDiff(ctx, wsID)
		require.NoError(t, err)
		assert.Equal(t, 1, diff.Summary.Links)

		detail, err := diffSvc.GetNodeDiffDetail(ctx, wsID, "n1")
		require.NoError(t, err)
		assert.Contains(t, detail.Kinds, "links")
		require.NotNil(t, detail.Links)
		assert.Contains(t, detail.Links.Old, "https://a")
		assert.Contains(t, detail.Links.New, "https://b")
	})

	t.Run("fetch: baseline error → current success で fetch 差分", func(t *testing.T) {
		ctx, store, _, diffSvc := setupDiffTestStore(t)
		wsID := "ws-fetch"
		seedDiffWorkspace(t, ctx, store, wsID, baselineRun, currentRun)
		errMsg := "timeout"
		appendResult(t, ctx, store, baselineRun, wsID, "n1", "body", `[]`, "hash", &errMsg, "2026-01-01T00:00:00Z")
		appendResult(t, ctx, store, currentRun, wsID, "n1", "body", `[]`, "hash", nil, "2026-01-02T00:00:00Z")

		diff, err := diffSvc.GetWorkspaceDiff(ctx, wsID)
		require.NoError(t, err)
		assert.Equal(t, 1, diff.Summary.Fetch)

		detail, err := diffSvc.GetNodeDiffDetail(ctx, wsID, "n1")
		require.NoError(t, err)
		assert.Equal(t, []string{"fetch"}, detail.Kinds)
		require.NotNil(t, detail.Fetch)
		assert.Equal(t, "error", detail.Fetch.Old)
		assert.Equal(t, "success", detail.Fetch.New)
	})

	t.Run("fetch: baseline success → current error のみ（diffsite fetch-b 相当）", func(t *testing.T) {
		ctx, store, _, diffSvc := setupDiffTestStore(t)
		wsID := "ws-fetch-b"
		seedDiffWorkspace(t, ctx, store, wsID, baselineRun, currentRun)
		appendResult(t, ctx, store, baselineRun, wsID, "n1", "body", `[]`, "hash", nil, "2026-01-01T00:00:00Z")
		errMsg := "HTTP 500"
		appendResult(t, ctx, store, currentRun, wsID, "n1", "body", `[]`, "hash", &errMsg, "2026-01-02T00:00:00Z")

		diff, err := diffSvc.GetWorkspaceDiff(ctx, wsID)
		require.NoError(t, err)
		assert.Equal(t, 1, diff.Summary.Fetch)

		detail, err := diffSvc.GetNodeDiffDetail(ctx, wsID, "n1")
		require.NoError(t, err)
		assert.Equal(t, []string{"fetch"}, detail.Kinds)
		require.NotNil(t, detail.Fetch)
		assert.Equal(t, "success", detail.Fetch.Old)
		assert.Equal(t, "error", detail.Fetch.New)
	})
}
