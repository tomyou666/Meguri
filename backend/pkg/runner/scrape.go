package runner

import (
	"context"
	"fmt"
	"net/url"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

// ScrapeWithConfig は 1 URL を任意 Config で実行する。
//
// 呼び出しごとに Kernel を Init し、完了後に Close する（案 A）。
// opts が nil の場合は pause なしで実行する。
func ScrapeWithConfig(ctx context.Context, rawURL string, cfg *model.Config, progress ProgressSink, opts *RunOptions) (*model.Result, error) {
	if opts != nil && opts.Cache != nil {
		lim := PrepareFetchLimiter(ctx, cfg, opts)
		opts.Cache.SetFetchLimiter(lim)
		var pause *PauseController
		if opts.Pause != nil {
			pause = opts.Pause
		}
		return opts.Cache.ScrapeWithConfig(ctx, rawURL, cfg, progress, pause)
	}
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	lim := PrepareFetchLimiter(ctx, cfg, opts)

	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
	if lim != nil {
		k.SetFetchLimiter(lim)
	}
	if err := k.Init(ctx); err != nil {
		return nil, fmt.Errorf("kernel init: %w", err)
	}
	defer func() { _ = k.Close(ctx) }()

	urlStr := u.String()
	if opts != nil && opts.Pause != nil {
		if err := opts.Pause.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
	}

	core.EmitProgress(progress, core.ProgressEvent{
		Kind: core.ProgressStarted,
		URL:  urlStr,
	})

	pipeline := core.NewPipeline(k)
	req := model.NewRequest(u, 0)
	out, err := pipeline.Run(ctx, req)
	if err != nil {
		core.EmitProgress(progress, core.ProgressEvent{
			Kind:  core.ProgressFailed,
			URL:   urlStr,
			Error: err.Error(),
		})
		return nil, err
	}
	if out.Result == nil {
		return nil, fmt.Errorf("pipeline returned nil result for %s", urlStr)
	}
	core.EmitProgress(progress, core.ProgressEvent{
		Kind:   core.ProgressSucceeded,
		URL:    urlStr,
		Result: out.Result,
	})
	return out.Result, nil
}
