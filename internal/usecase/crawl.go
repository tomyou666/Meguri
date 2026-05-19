package usecase

import (
	"context"
	"fmt"
	"net/url"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

// Crawl はクローリングシナリオを束ねるユースケース。
type Crawl struct {
	Kernel  *core.Kernel
	Fetcher core.Fetcher
	Robots  core.RobotsChecker
	Sink    core.ResultSink
}

func NewCrawl(k *core.Kernel, f core.Fetcher, robots core.RobotsChecker, sink core.ResultSink) *Crawl {
	return &Crawl{Kernel: k, Fetcher: f, Robots: robots, Sink: sink}
}

func (c *Crawl) Run(ctx context.Context, targets []string) (*core.CrawlStats, []*model.Result, error) {
	if len(targets) == 0 {
		return nil, nil, fmt.Errorf("no target URLs")
	}
	seeds := make([]*url.URL, 0, len(targets))
	for _, t := range targets {
		u, err := url.Parse(t)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid seed url %q: %w", t, err)
		}
		seeds = append(seeds, u)
	}

	var collected []*model.Result
	sink := c.Sink
	if sink == nil {
		sink = func(r *model.Result) { collected = append(collected, r) }
	} else {
		original := sink
		sink = func(r *model.Result) {
			collected = append(collected, r)
			original(r)
		}
	}

	crawler := core.NewCrawler(c.Kernel, c.Fetcher, c.Robots, sink)
	stats, err := crawler.Run(ctx, seeds)
	return stats, collected, err
}
