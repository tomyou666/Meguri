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

func TestKernel_Init_httpFetcher(t *testing.T) {
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
}

func TestKernel_Init_chromiumMissingBrowser(t *testing.T) {
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
}

func TestKernel_Init_unknownFetcher(t *testing.T) {
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
}
