package runner

import (
	"context"

	"meguri/internal/usecase"
)

//go:generate go tool gowrap gen -p meguri/pkg/runner -i Scraper -t templates/slog_debug.gotmpl -o scraper_with_debug_log.go

// Scraper は単一 URL スクレイプを抽象化する。
type Scraper interface {
	// ScrapeWithConfig は 1 URL を任意 Config で実行する。
	ScrapeWithConfig(ctx context.Context, rawURL string, cfg *Config, progress ProgressSink, opts *RunOptions) (*Result, error)
}

type scraperImpl struct{}

func (scraperImpl) ScrapeWithConfig(ctx context.Context, rawURL string, cfg *Config, progress ProgressSink, opts *RunOptions) (*Result, error) {
	return usecase.NewScrape(nil).RunWithConfig(ctx, rawURL, cfg, progress, opts)
}

// ScrapeWithConfig は 1 URL を任意 Config で実行する。
func ScrapeWithConfig(ctx context.Context, rawURL string, cfg *Config, progress ProgressSink, opts *RunOptions) (*Result, error) {
	return defaultScraper.ScrapeWithConfig(ctx, rawURL, cfg, progress, opts)
}
