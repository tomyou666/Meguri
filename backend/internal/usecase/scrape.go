// Package usecase はプレゼンテーション層からカーネルへのシナリオを束ねる。
package usecase

import (
	"context"
	"fmt"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// Scrape は単一URLの取得→出力までを実行するユースケース。
type Scrape struct {
	// Pipeline は 1 URL あたりの処理パイプライン（Wire 注入時に使用）。
	Pipeline *core.Pipeline
}

// NewScrape は単一 URL スクレイプ用ユースケースを構築する。
func NewScrape(pipeline *core.Pipeline) *Scrape {
	return &Scrape{Pipeline: pipeline}
}

// Run は与えられた target URL に対してパイプラインを1回走らせる。
func (s *Scrape) Run(ctx context.Context, target string) (*model.Result, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target url: %w", err)
	}
	req := model.NewRequest(u, 0)

	out, err := s.Pipeline.Run(ctx, req)
	if err != nil {
		return nil, err
	}
	return out.Result, nil
}

// RunWithConfig は 1 URL を任意 Config で実行する。
//
// opts.Cache が非 nil の場合はキャッシュ経由で実行する。
// それ以外は呼び出しごとに Kernel を Init し、完了後に Close する。
//
// opts が nil の場合は pause なしで実行する。
func (s *Scrape) RunWithConfig(
	ctx context.Context,
	rawURL string,
	cfg *model.Config,
	progress core.ProgressSink,
	opts *RunOptions,
) (*model.Result, error) {
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
