package usecase

import (
	"context"
	"fmt"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/infrastructure/robots"
)

// Crawl はクローリングシナリオを束ねるユースケース。
type Crawl struct {
	// Sink は各ページの Result 受け取り先（nil 可）。
	Sink core.ResultSink
}

// NewCrawl はクロール用ユースケースを構築する。
//
// sink は nil の場合、収集した Result は戻り値のスライスにのみ格納される。
func NewCrawl(sink core.ResultSink) *Crawl {
	return &Crawl{Sink: sink}
}

// RunWithConfig はマージ済み Config とシード URL から BFS クロールを実行する。
//
// opts が nil の場合は pause なしで実行する。
func (c *Crawl) RunWithConfig(
	ctx context.Context,
	cfg *model.Config,
	seeds []string,
	progress core.ProgressSink,
	opts *RunOptions,
) (*core.CrawlStats, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if len(seeds) == 0 {
		return nil, fmt.Errorf("no seed URLs")
	}

	parsed := make([]*url.URL, 0, len(seeds))
	for _, s := range seeds {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid seed url %q: %w", s, err)
		}
		parsed = append(parsed, u)
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

	pipeline := core.NewPipeline(k)
	var robotsChecker core.RobotsChecker
	if cfg.Crawl.RespectRobotsTxt {
		robotsChecker = robots.NewCache(k.Fetcher())
	}

	crawler := core.NewCrawler(k, pipeline, robotsChecker, c.Sink, progress)
	if opts != nil && opts.Pause != nil {
		crawler.SetPauseController(opts.Pause)
	}
	stats, err := crawler.Run(ctx, parsed)
	if err != nil {
		core.EmitProgress(progress, core.ProgressEvent{
			Kind:  core.ProgressError,
			Error: err.Error(),
		})
		return stats, err
	}
	return stats, nil
}
