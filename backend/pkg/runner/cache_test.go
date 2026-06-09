package runner_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/domain/model"
	"scraperbot/pkg/runner"
)

func testScrapeCfg() *model.Config {
	return &model.Config{
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
}

// TestRunnerCache は RunnerCache のキー生成・再利用・LRU 退避を検証する。
func TestRunnerCache(t *testing.T) {
	t.Run("正常系: targets と exclude_urls はキャッシュキーに含めない", func(t *testing.T) {
		a := testScrapeCfg()
		b := testScrapeCfg()
		a.Targets = []string{"https://a.example"}
		b.Targets = []string{"https://b.example"}
		a.Crawl.ExcludeURLs = []string{"https://skip.example"}
		b.Crawl.ExcludeURLs = nil

		ha, err := runner.CfgHashForTest(a)
		require.NoError(t, err)
		hb, err := runner.CfgHashForTest(b)
		require.NoError(t, err)
		assert.Equal(t, ha, hb)
	})

	t.Run("正常系: 同一設定ではカーネルを再利用する", func(t *testing.T) {
		cfg := testScrapeCfg()
		cache := runner.NewRunnerCache()
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
		cache := runner.NewRunnerCacheWithMaxForTest(2)
		defer cache.CloseAll()

		ctx := context.Background()
		makeCfg := func(ua string) *model.Config {
			c := testScrapeCfg()
			c.Request.Headers["User-Agent"] = ua
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
		cache := runner.NewRunnerCache()

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
