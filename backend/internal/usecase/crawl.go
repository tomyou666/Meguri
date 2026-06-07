package usecase

import (
	"context"
	"fmt"
	"net/url"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

// CrawlerFactory は実行時の ResultSink と ProgressSink を受け取り Crawler を生成する。
type CrawlerFactory func(sink core.ResultSink, progress core.ProgressSink) *core.Crawler

// Crawl はクローリングシナリオを束ねるユースケース。
type Crawl struct {
	// Factory は実行時 sink 付きの Crawler を生成する。
	Factory CrawlerFactory
	// Sink は各ページの Result 受け取り先（nil 可）。
	Sink core.ResultSink
	// Progress は URL 単位の進捗通知先（nil 可）。
	Progress core.ProgressSink
}

// NewCrawl はクロール用ユースケースを構築する。
//
// sink は nil の場合、収集した Result は戻り値のスライスにのみ格納される。
func NewCrawl(factory CrawlerFactory, sink core.ResultSink) *Crawl {
	return &Crawl{Factory: factory, Sink: sink}
}

// Run はシード URL 一覧からクロールを実行し、統計と収集結果を返す。
func (c *Crawl) Run(ctx context.Context, targets []string) (*core.CrawlStats, []*model.Result, error) {
	return c.RunWithProgress(ctx, targets, c.Progress)
}

// RunWithProgress は ProgressSink を指定してクロールを実行する。
func (c *Crawl) RunWithProgress(ctx context.Context, targets []string, progress core.ProgressSink) (*core.CrawlStats, []*model.Result, error) {
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

	crawler := c.Factory(sink, progress)
	stats, err := crawler.Run(ctx, seeds)
	return stats, collected, err
}
