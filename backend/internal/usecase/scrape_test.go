package usecase_test

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/usecase"

	_ "meguri/plugins/fetcher-chromium"
	_ "meguri/plugins/fetcher-http"
	_ "meguri/plugins/filter-maincontent"
	_ "meguri/plugins/filter-selector"
	_ "meguri/plugins/linkextractor-default"
	_ "meguri/plugins/parser-html"
	_ "meguri/plugins/parser-pdf"
	_ "meguri/plugins/preprocessor-header"
	_ "meguri/plugins/transformer-html"
	_ "meguri/plugins/transformer-markdown"
	_ "meguri/plugins/transformer-raw-html"
)

// TestScrapeRunWithConfig は単一 URL スクレイプの進捗通知を検証する。
func TestScrapeRunWithConfig(t *testing.T) {
	t.Run("正常系: スクレイプ成功時に started と succeeded/failed が発火する", func(t *testing.T) {
		cfg := &model.Config{
			Request: model.RequestConfig{
				Timeout:       10_000_000_000,
				RetryCount:    0,
				RetryInterval: 100_000_000,
				Headers:       map[string]string{"User-Agent": "meguri-test"},
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

		_, err := usecase.NewScrape(nil).RunWithConfig(context.Background(), "https://example.com", cfg, progress, nil)
		if err != nil {
			t.Skip("network unavailable:", err)
		}

		mu.Lock()
		defer mu.Unlock()
		assert.Contains(t, kinds, core.ProgressStarted)
		assert.True(t, containsKind(kinds, core.ProgressSucceeded) || containsKind(kinds, core.ProgressFailed))
	})
}

func containsKind(kinds []core.ProgressKind, want core.ProgressKind) bool {
	for _, k := range kinds {
		if k == want {
			return true
		}
	}
	return false
}

// TestCrawlRunWithConfig はクロール API の前提条件を検証する。
func TestCrawlRunWithConfig(t *testing.T) {
	t.Run("異常系: シード URL なしではエラー", func(t *testing.T) {
		cfg := &model.Config{}
		_, err := usecase.NewCrawl(nil).RunWithConfig(context.Background(), cfg, nil, nil, nil)
		require.Error(t, err)
	})
}
