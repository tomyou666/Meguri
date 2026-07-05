package runner

import (
	"context"

	"meguri/internal/usecase"
)

// CrawlWithProgress はマージ済み Config とシード URL から BFS クロールを実行する。
func CrawlWithProgress(ctx context.Context, cfg *Config, seeds []string, progress ProgressSink, opts *RunOptions) (*CrawlStats, error) {
	return usecase.NewCrawl(nil).RunWithConfig(ctx, cfg, seeds, progress, opts)
}
