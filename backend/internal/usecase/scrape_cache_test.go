package usecase_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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

func testScrapeCfg() *model.Config {
	return &model.Config{
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
}

// TestScrapeCache は ScrapeCache のキー生成・再利用・LRU 退避を検証する。
func TestScrapeCache(t *testing.T) {
	t.Run("正常系: targets と exclude_urls はキャッシュキーに含めない", func(t *testing.T) {
		a := testScrapeCfg()
		b := testScrapeCfg()
		a.Targets = []string{"https://a.example"}
		b.Targets = []string{"https://b.example"}
		a.Crawl.ExcludeURLs = []string{"https://skip.example"}
		b.Crawl.ExcludeURLs = nil

		ha, err := usecase.CfgHashForTest(a)
		require.NoError(t, err)
		hb, err := usecase.CfgHashForTest(b)
		require.NoError(t, err)
		assert.Equal(t, ha, hb)
	})

	t.Run("正常系: 同一設定ではカーネルを再利用する", func(t *testing.T) {
		cfg := testScrapeCfg()
		cache := usecase.NewScrapeCache()
		defer cache.CloseAll()

		ctx := context.Background()
		url := "https://example.com"

		_, err := cache.ScrapeWithConfig(ctx, url, cfg, nil, nil)
		if err != nil {
			t.Skip("network unavailable:", err)
		}
		assert.Equal(t, 1, cache.EntryCountForTest())

		_, err = cache.ScrapeWithConfig(ctx, url, cfg, nil, nil)
		if err != nil {
			t.Skip("network unavailable:", err)
		}
		assert.Equal(t, 1, cache.EntryCountForTest())
	})

	t.Run("正常系: maxEntries 超過で LRU 退避する", func(t *testing.T) {
		cache := usecase.NewScrapeCacheWithMaxForTest(2)
		defer cache.CloseAll()

		ctx := context.Background()
		makeCfg := func(ua string) *model.Config {
			c := testScrapeCfg()
			c.Plugins.Stealth.HTTP.UserAgent = ua
			return c
		}

		for _, ua := range []string{"ua-1", "ua-2", "ua-3"} {
			_, err := cache.ScrapeWithConfig(ctx, "https://example.com", makeCfg(ua), nil, nil)
			if err != nil {
				t.Skip("network unavailable:", err)
			}
		}

		assert.Equal(t, 2, cache.EntryCountForTest())
	})

	t.Run("正常系: CloseAll で全エントリを破棄する", func(t *testing.T) {
		cfg := testScrapeCfg()
		cache := usecase.NewScrapeCache()

		ctx := context.Background()
		_, err := cache.ScrapeWithConfig(ctx, "https://example.com", cfg, nil, nil)
		if err != nil {
			t.Skip("network unavailable:", err)
		}
		require.Equal(t, 1, cache.EntryCountForTest())

		cache.CloseAll()
		assert.Equal(t, 0, cache.EntryCountForTest())
	})
}
