package usecase

import (
	"context"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/infrastructure/chromium"
)

// PrepareFetchLimiter は設定から FetchLimiter を構築し、必要ならキャリブレーションと動的監視を開始する。
//
// opts.FetchLimiter が非 nil の場合はそれをそのまま返す。
func PrepareFetchLimiter(ctx context.Context, cfg *model.Config, opts *RunOptions) *fetchlimit.FetchLimiter {
	if opts != nil && opts.FetchLimiter != nil {
		return opts.FetchLimiter
	}
	if cfg == nil {
		return nil
	}
	lim := fetchlimit.NewFromConfig(cfg.Crawl.FetchLimits)
	fl := cfg.Crawl.FetchLimits
	if fl.AutoCalibrate {
		browserPath, err := chromium.ResolveBrowserPath(cfg.Plugins.FetcherConfig.BrowserPath)
		if err != nil {
			browserPath = ""
		}
		_ = fetchlimit.CalibrateChromium(ctx, lim, browserPath)
	}
	if fl.DynamicChromium {
		fetchlimit.StartDynamicChromium(lim, fl)
	}
	if opts != nil {
		opts.FetchLimiter = lim
	}
	return lim
}
