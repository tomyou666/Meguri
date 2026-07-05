package runner

import (
	"context"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// ScrapeWithConfig は 1 URL を任意 Config で実行する。
func ScrapeWithConfig(ctx context.Context, rawURL string, cfg *model.Config, progress ProgressSink, opts *RunOptions) (*Result, error) {
	return usecase.NewScrape(nil).RunWithConfig(ctx, rawURL, cfg, progress, opts)
}
