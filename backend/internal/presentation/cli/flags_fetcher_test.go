package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
)

// TestFetcherFlags は CLI の fetcher 関連フラグのパースと YAML へのマージを検証する。
func TestFetcherFlags(t *testing.T) {
	t.Run("正常系: fetcher 関連フラグをパースできる", func(t *testing.T) {
		t.Parallel()
		f, err := ParseArgs([]string{
			"--url", "https://example.com/",
			"--fetcher", "chromium",
			"--fetcher-browser-path", "/usr/bin/chromium",
		})
		require.NoError(t, err)
		assert.Equal(t, "chromium", f.Fetcher)
		assert.Equal(t, "/usr/bin/chromium", f.FetcherBrowserPath)
	})

	t.Run("正常系: CLI フラグが YAML の fetcher 設定を上書きする", func(t *testing.T) {
		t.Parallel()
		cfg := model.Default()
		cfg.Plugins.Fetcher = model.FetcherHTTP
		cfg.Plugins.FetcherConfig.BrowserPath = "/from/yaml"

		Merge(&cfg, &Flags{
			Fetcher:            "chromium",
			FetcherBrowserPath: "/from/cli",
		})

		assert.Equal(t, model.FetcherChromium, cfg.Plugins.Fetcher)
		assert.Equal(t, "/from/cli", cfg.Plugins.FetcherConfig.BrowserPath)
	})
}
