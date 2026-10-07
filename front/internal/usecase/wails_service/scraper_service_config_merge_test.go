package wails_service

import (
	"encoding/json"
	"testing"
	"time"

	"meguri-app/internal/model"
	"meguri/pkg/runner"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMergedConfigLayers は実行時マージがモードに依らず app → workspace → node content であることを検証する。
func TestMergedConfigLayers(t *testing.T) {
	st := newCrawlState(model.StartCrawlRequest{
		AppDefaults: json.RawMessage(`{"content":{"selector":"#app"},"request":{"timeout":"30s"},"crawl":{"max_depth":7,"fetch_limits":{"http_max_inflight":3}}}`),
		Workspace: model.WorkspaceDTO{
			Settings: json.RawMessage(`{"content":{"selector":"#ws"},"crawl":{"max_depth":4,"fetch_limits":{"http_max_inflight":9}}}`),
			Nodes: []model.GraphNodeDTO{
				{
					ID:            "n1",
					URLNormalized: "https://example.com/a",
					NodeSettings:  json.RawMessage(`{"content":{"selector":"#node","formats":["html"]},"request":{"timeout":"1s"},"crawl":{"max_depth":1}}`),
				},
				{
					ID:            "n2",
					URLNormalized: "https://example.com/b",
				},
			},
			ExcludeURLs: []string{"https://example.com/skip"},
		},
	})

	t.Run("正常系: ノード content が app と workspace を上書きし request と crawl は捨てる", func(t *testing.T) {
		cfg, err := st.mergedConfig(st.nodeByID["n1"])
		require.NoError(t, err)
		assert.Equal(t, "#node", cfg.Content.Selector)
		assert.Equal(t, 4, cfg.Crawl.MaxDepth)
		assert.Equal(t, 9, cfg.Crawl.FetchLimits.HTTPMaxInflight)
		assert.Equal(t, 30*time.Second, cfg.Request.Timeout)
		assert.NotContains(t, cfg.Content.Formats, "html")
		assert.ElementsMatch(t, []string{"https://example.com/skip"}, cfg.Crawl.ExcludeURLs)
	})

	t.Run("正常系: ラン設定は workspace の crawl と content でノードを含まない", func(t *testing.T) {
		cfg, err := st.runConfig()
		require.NoError(t, err)
		assert.Equal(t, "#ws", cfg.Content.Selector)
		assert.Equal(t, 4, cfg.Crawl.MaxDepth)
		assert.Equal(t, 9, cfg.Crawl.FetchLimits.HTTPMaxInflight)
	})

	t.Run("正常系: limiter も app → workspace", func(t *testing.T) {
		cfg, err := st.limiterConfig()
		require.NoError(t, err)
		assert.Equal(t, 9, cfg.Crawl.FetchLimits.HTTPMaxInflight)
		assert.Equal(t, 4, cfg.Crawl.MaxDepth)
	})

	t.Run("正常系: content のある既存ノードだけページ設定を返す", func(t *testing.T) {
		withNode, err := st.pageConfig("https://example.com/a")
		require.NoError(t, err)
		require.NotNil(t, withNode)
		assert.Equal(t, "#node", withNode.Content.Selector)

		empty, err := st.pageConfig("https://example.com/b")
		require.NoError(t, err)
		assert.Nil(t, empty)

		unknown, err := st.pageConfig("https://example.com/new")
		require.NoError(t, err)
		assert.Nil(t, unknown)
	})
}

// TestRunCrawlLimiterConfigError は limiter 用マージの失敗をクロール開始前に返すことを検証する。
func TestRunCrawlLimiterConfigError(t *testing.T) {
	s := &scraperService{
		job: &activeCrawlJob{opts: &runner.RunOptions{}},
	}
	err := s.runCrawl(t.Context(), model.StartCrawlRequest{
		AppDefaults: json.RawMessage(`{`),
		Workspace:   model.WorkspaceDTO{},
	})
	require.Error(t, err)
}
