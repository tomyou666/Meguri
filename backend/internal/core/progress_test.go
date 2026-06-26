package core_test

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// TestCrawlerProgress はクロール中の進捗イベントとスキップ理由の通知を検証する。
func TestCrawlerProgress(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

	t.Run("正常系: started/succeeded/linkDiscovered/completed が発火する", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 1
		cfg.Crawl.MaxPages = 10
		cfg.Crawl.MaxConcurrency = 1

		var mu sync.Mutex
		var events []core.ProgressEvent
		progress := func(ev core.ProgressEvent) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, ev)
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, nil, progress)

		seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
		require.NoError(t, err)

		stats, err := c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)
		require.NotNil(t, stats)

		mu.Lock()
		defer mu.Unlock()

		kinds := make([]core.ProgressKind, 0, len(events))
		for _, ev := range events {
			kinds = append(kinds, ev.Kind)
		}
		assert.Contains(t, kinds, core.ProgressStarted)
		assert.Contains(t, kinds, core.ProgressSucceeded)
		assert.Contains(t, kinds, core.ProgressLinkDiscovered)
		assert.Contains(t, kinds, core.ProgressCompleted)

		var startedWithParent bool
		for _, ev := range events {
			if ev.Kind == core.ProgressStarted && ev.URL != seed.String() && ev.ParentURL != "" {
				startedWithParent = true
			}
		}
		assert.True(t, startedWithParent, "child started events should include parentURL")
	})

	t.Run("正常系: exclude_urls に一致する URL は exclude_urls でスキップされる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.ExcludeURLs = []string{srv.URL + "/docs/page-a.html"}

		var mu sync.Mutex
		var skipped []core.ProgressEvent
		progress := func(ev core.ProgressEvent) {
			if ev.Kind != core.ProgressSkipped {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			skipped = append(skipped, ev)
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, nil, progress)

		seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
		require.NoError(t, err)

		_, err = c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()

		found := false
		for _, ev := range skipped {
			if ev.SkipReason == "exclude_urls" {
				found = true
				break
			}
		}
		assert.True(t, found, "page-a should be skipped with exclude_urls reason")
	})

	t.Run("正常系: skip_scrape_urls は取得せず already_success でスキップされる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.SkipScrapeURLs = []string{srv.URL + "/docs/page-a.html"}

		var mu sync.Mutex
		var skipped []core.ProgressEvent
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}
		progress := func(ev core.ProgressEvent) {
			if ev.Kind != core.ProgressSkipped {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			skipped = append(skipped, ev)
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, progress)

		seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
		require.NoError(t, err)

		_, err = c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()

		found := false
		for _, ev := range skipped {
			if ev.SkipReason == "already_success" {
				found = true
				break
			}
		}
		assert.True(t, found, "page-a should be skipped with already_success reason")
		for _, u := range collected {
			assert.NotContains(t, u, "/docs/page-a.html", "skip scrape URLs must not be fetched")
		}
	})

	t.Run("正常系: max_depth でスキップされた URL には linkDiscovered を出さない", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 1
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.MaxConcurrency = 1

		var mu sync.Mutex
		var events []core.ProgressEvent
		progress := func(ev core.ProgressEvent) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, ev)
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, nil, progress)

		seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
		require.NoError(t, err)

		_, err = c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()

		var skippedDepth3 []core.ProgressEvent
		var linkDiscoveredDepth3 []core.ProgressEvent
		for _, ev := range events {
			if ev.Kind == core.ProgressSkipped && ev.SkipReason == "max_depth" {
				skippedDepth3 = append(skippedDepth3, ev)
			}
			if ev.Kind == core.ProgressLinkDiscovered && ev.Depth > cfg.Crawl.MaxDepth {
				linkDiscoveredDepth3 = append(linkDiscoveredDepth3, ev)
			}
		}
		assert.NotEmpty(t, skippedDepth3, "max_depth でスキップされた URL があること")
		assert.Empty(t, linkDiscoveredDepth3, "enqueue されない URL には linkDiscovered を出さない")
	})

	t.Run("正常系: 訪問済み URL の深い経路からの再 enqueue は duplicate になる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.MaxConcurrency = 4

		targetURL := srv.URL + "/docs/page-a.html"

		var mu sync.Mutex
		var events []core.ProgressEvent
		progress := func(ev core.ProgressEvent) {
			mu.Lock()
			defer mu.Unlock()
			events = append(events, ev)
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, nil, progress)

		seed, err := url.Parse(srv.URL + "/parallel_dup_seed.html")
		require.NoError(t, err)

		_, err = c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()

		succeeded := map[string]bool{}
		for _, ev := range events {
			if ev.Kind == core.ProgressSucceeded {
				succeeded[ev.URL] = true
			}
		}
		require.True(t, succeeded[targetURL], "page-a should be fetched via shallow path")

		dupSkipCount := 0
		for _, ev := range events {
			if ev.Kind != core.ProgressSkipped || ev.URL != targetURL {
				continue
			}
			assert.Equal(t, "duplicate", ev.SkipReason,
				"visited URL should skip as duplicate, not %s", ev.SkipReason)
			dupSkipCount++
		}
		assert.Greater(t, dupSkipCount, 0, "deep path should produce duplicate skip for page-a")
	})

	t.Run("正常系: クロール無効時も sink に結果が渡される", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = false

		var collected []*model.Result
		sink := func(r *model.Result) { collected = append(collected, r) }

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, err := url.Parse(srv.URL + "/links_with_pdf.html")
		require.NoError(t, err)

		_, err = c.Run(context.Background(), []*url.URL{seed})
		require.NoError(t, err)
		assert.Len(t, collected, 1)
	})
}
