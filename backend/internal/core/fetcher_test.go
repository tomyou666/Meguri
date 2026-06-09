package core_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/logger"

	_ "scraperbot/plugins/fetcher-chromium"
	_ "scraperbot/plugins/fetcher-http"
)

// TestKernel_InitFetcher は設定に応じた Fetcher 初期化の成否を検証する。
func TestKernel_InitFetcher(t *testing.T) {
	t.Parallel()

	t.Run("正常系: デフォルトで http Fetcher が初期化される", func(t *testing.T) {
		t.Parallel()
		logger.Init(io.Discard, slog.LevelInfo)
		cfg := model.Default()
		cfg.Targets = []string{"https://example.com/"}

		host := core.NewHost(&cfg)
		k := core.NewKernel(&cfg, host, core.Default())
		err := k.Init(context.Background())
		require.NoError(t, err)
		defer k.Close(context.Background())

		assert.NotNil(t, k.Fetcher())
		assert.Equal(t, string(model.FetcherHTTP), k.Fetcher().Metadata().Name)
	})

	t.Run("異常系: chromium でブラウザパスが無効だと Init が失敗する", func(t *testing.T) {
		t.Parallel()
		logger.Init(io.Discard, slog.LevelInfo)
		cfg := model.Default()
		cfg.Targets = []string{"https://example.com/"}
		cfg.Plugins.Fetcher = model.FetcherChromium
		cfg.Plugins.FetcherConfig.BrowserPath = "/nonexistent/chromium-binary"

		host := core.NewHost(&cfg)
		k := core.NewKernel(&cfg, host, core.Default())
		err := k.Init(context.Background())
		require.Error(t, err)
	})

	t.Run("異常系: 未知の fetcher 名だとエラー", func(t *testing.T) {
		t.Parallel()
		logger.Init(io.Discard, slog.LevelInfo)
		cfg := model.Default()
		cfg.Targets = []string{"https://example.com/"}
		cfg.Plugins.Fetcher = "selenium"

		host := core.NewHost(&cfg)
		k := core.NewKernel(&cfg, host, core.Default())
		err := k.Init(context.Background())
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fetcher not found")
	})
}
