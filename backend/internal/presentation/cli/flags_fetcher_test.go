package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/domain/model"
)

// TestFetcherFlags は CLI の fetcher 関連フラグのパースと YAML へのマージを検証する。
func TestFetcherFlags(t *testing.T) {
	t.Run("正常系: fetcher 関連フラグをパースできる", func(t *testing.T) {
		t.Parallel()
		f, err := ParseArgs([]string{
			"--url", "https://example.com/",
			"--fetcher", "chromium",
			"--fetcher-browser-path", "/usr/bin/chromium",
			"--fetcher-user-agent", "TestAgent/1.0",
			"--fetcher-headless=false",
		})
		require.NoError(t, err)
		assert.Equal(t, "chromium", f.Fetcher)
		assert.Equal(t, "/usr/bin/chromium", f.FetcherBrowserPath)
		assert.Equal(t, "TestAgent/1.0", f.FetcherUserAgent)
		assert.True(t, f.FetcherHeadless.set)
		assert.False(t, f.FetcherHeadless.v)
	})

	t.Run("正常系: CLI フラグが YAML の fetcher 設定を上書きする", func(t *testing.T) {
		t.Parallel()
		cfg := model.Default()
		cfg.Plugins.Fetcher = model.FetcherHTTP
		cfg.Plugins.FetcherConfig.BrowserPath = "/from/yaml"
		cfg.Plugins.FetcherConfig.UserAgent = "YAML/1.0"

		Merge(&cfg, &Flags{
			Fetcher:            "chromium",
			FetcherBrowserPath: "/from/cli",
			FetcherUserAgent:   "CLI/1.0",
		})

		assert.Equal(t, model.FetcherChromium, cfg.Plugins.Fetcher)
		assert.Equal(t, "/from/cli", cfg.Plugins.FetcherConfig.BrowserPath)
		assert.Equal(t, "CLI/1.0", cfg.Plugins.FetcherConfig.UserAgent)
	})
}
