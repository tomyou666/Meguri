package runner_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/pkg/runner"
)

func TestScrapeWithConfigEmitsProgress(t *testing.T) {
	_ = runner.ProgressStarted // ensure package links

	cfg := &model.Config{
		Request: model.RequestConfig{
			Timeout:       10_000_000_000,
			RetryCount:    0,
			RetryInterval: 100_000_000,
			Headers:       map[string]string{"User-Agent": "scraperbot-test"},
		},
		Content: model.ContentConfig{
			Formats:         []model.OutputFormat{model.FormatMarkdown},
			ExtractLinks:    true,
			ExtractMetadata: true,
		},
		Plugins: model.PluginSelection{
			Fetcher:       model.FetcherHTTP,
			PreProcessors: []string{"header"},
			Parsers:       []string{"html"},
			Filters:       []string{"maincontent"},
			LinkExtractor: "default",
			Transformer:   "markdown",
		},
	}

	var mu sync.Mutex
	var kinds []core.ProgressKind
	progress := func(ev core.ProgressEvent) {
		mu.Lock()
		defer mu.Unlock()
		kinds = append(kinds, ev.Kind)
	}

	// plugins registered via plugins.go blank imports when testing runner package
	_, err := runner.ScrapeWithConfig(context.Background(), "https://example.com", cfg, progress, nil)
	// network may fail in CI; we only assert progress shape when scrape runs
	if err != nil {
		t.Skip("network unavailable:", err)
	}

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, kinds, core.ProgressStarted)
	assert.True(t, containsKind(kinds, core.ProgressSucceeded) || containsKind(kinds, core.ProgressFailed))
}

func containsKind(kinds []core.ProgressKind, want core.ProgressKind) bool {
	for _, k := range kinds {
		if k == want {
			return true
		}
	}
	return false
}

func TestCrawlWithProgressRequiresSeeds(t *testing.T) {
	cfg := &model.Config{}
	_, err := runner.CrawlWithProgress(context.Background(), cfg, nil, nil, nil)
	require.Error(t, err)
}
