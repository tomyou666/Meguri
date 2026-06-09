package core_test

import (
	"context"
	"net/url"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

func TestCrawlerProgressEvents(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

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
}

func TestCrawlerExcludeURLs(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

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
}

func TestCrawlerSkipScrapeURLs(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

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
}

func TestCrawlerSkippedURLsNoLinkDiscovered(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

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
}

func TestCrawlerProgressSinkCollectsResults(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

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
}
