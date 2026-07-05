package runner

import (
	"context"

	"meguri/internal/usecase"
)

//go:generate go tool gowrap gen -p meguri/pkg/runner -i Crawler -t templates/slog_debug.gotmpl -o crawler_with_debug_log.go

// Crawler は BFS クロール実行を抽象化する。
type Crawler interface {
	// CrawlWithProgress はマージ済み Config とシード URL から BFS クロールを実行する。
	CrawlWithProgress(ctx context.Context, cfg *Config, seeds []string, progress ProgressSink, opts *RunOptions) (*CrawlStats, error)
}

type crawlerImpl struct{}

func (crawlerImpl) CrawlWithProgress(ctx context.Context, cfg *Config, seeds []string, progress ProgressSink, opts *RunOptions) (*CrawlStats, error) {
	return usecase.NewCrawl(nil).RunWithConfig(ctx, cfg, seeds, progress, opts)
}

// CrawlWithProgress はマージ済み Config とシード URL から BFS クロールを実行する。
func CrawlWithProgress(ctx context.Context, cfg *Config, seeds []string, progress ProgressSink, opts *RunOptions) (*CrawlStats, error) {
	return defaultCrawler.CrawlWithProgress(ctx, cfg, seeds, progress, opts)
}
