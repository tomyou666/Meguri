package core

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"scraperbot/internal/domain/model"
)

// RobotsChecker は robots.txt ベースの許可判定を抽象化する。
// テストでフェイク差し込みできるよう interface としている。
type RobotsChecker interface {
	Allowed(ctx context.Context, u *url.URL, userAgent string) bool
}

// ResultSink はクロール中に得られた Result を受け取るシンク。
type ResultSink func(res *model.Result)

// Crawler は BFS で URL を巡回し、各 URL でパイプラインを実行する。
type Crawler struct {
	// cfg はクロール設定を含む実行設定。
	cfg *model.Config
	// kernel はプラグインカーネル。
	kernel *Kernel
	// pipeline は 1 URL あたりの処理パイプライン。
	pipeline *Pipeline
	// robots は robots.txt 判定（nil 可）。
	robots RobotsChecker

	// includeRe は許可パス正規表現（コンパイル済み）。
	includeRe []*regexp.Regexp
	// excludeRe は除外パス正規表現（コンパイル済み）。
	excludeRe []*regexp.Regexp

	// sink は各ページの Result 受け取り先。
	sink ResultSink
}

// CrawlStats はクロールの最終サマリ。
type CrawlStats struct {
	// Enqueued はキューに投入した URL 数。
	Enqueued int
	// Succeeded はパイプライン成功した URL 数。
	Succeeded int
	// Failed はパイプライン失敗した URL 数。
	Failed int
	// Skipped は重複・フィルタ・上限でスキップした URL 数。
	Skipped int
}

// NewCrawler はクローラを構築する。robots は nil 可（その場合は判定をスキップ）。
func NewCrawler(k *Kernel, robots RobotsChecker, sink ResultSink) *Crawler {
	cfg := k.Config()
	c := &Crawler{
		cfg:      cfg,
		kernel:   k,
		pipeline: NewPipeline(k),
		robots:   robots,
		sink:     sink,
	}
	for _, p := range cfg.Crawl.IncludePaths {
		if re, err := regexp.Compile(p); err == nil {
			c.includeRe = append(c.includeRe, re)
		}
	}
	for _, p := range cfg.Crawl.ExcludePaths {
		if re, err := regexp.Compile(p); err == nil {
			c.excludeRe = append(c.excludeRe, re)
		}
	}
	return c
}

// job はクロールキュー内の 1 件分の作業単位。
type job struct {
	// url は処理対象 URL。
	url *url.URL
	// depth はシードからの深度。
	depth int
}

// Run は与えられたシード URL から BFS でクロールを実行する。
// crawl.enabled=false の場合は単一 URL モードとして seed[0] のみを処理する。
func (c *Crawler) Run(ctx context.Context, seeds []*url.URL) (*CrawlStats, error) {
	stats := &CrawlStats{}

	if !c.cfg.Crawl.Enabled {
		if len(seeds) == 0 {
			return stats, nil
		}
		stats.Enqueued = 1
		if c.runOne(ctx, job{url: seeds[0], depth: 0}, nil) {
			stats.Succeeded++
		} else {
			stats.Failed++
		}
		return stats, nil
	}

	workerN := c.cfg.Crawl.MaxConcurrency
	if c.cfg.Crawl.RequestDelay > 0 {
		workerN = 1
	}

	jobs := make(chan job, workerN*2)
	pushQ := make(chan job, 256)

	var (
		stateMu sync.Mutex
		seen    = map[string]struct{}{}
		visited int
		pending int
		closed  bool
	)

	// dispatcher: pushQ → jobs を中継しつつ無制限キューを内部に持つ
	queueDone := make(chan struct{})
	go func() {
		defer close(queueDone)
		q := make([]job, 0, 64)
		for {
			var (
				out  chan job
				head job
			)
			if len(q) > 0 {
				out = jobs
				head = q[0]
			}
			select {
			case <-ctx.Done():
				close(jobs)
				return
			case j, ok := <-pushQ:
				if !ok {
					for len(q) > 0 {
						select {
						case <-ctx.Done():
							close(jobs)
							return
						case jobs <- q[0]:
							q = q[1:]
						}
					}
					close(jobs)
					return
				}
				q = append(q, j)
			case out <- head:
				q = q[1:]
			}
		}
	}()

	finishOne := func(ok bool) {
		stateMu.Lock()
		defer stateMu.Unlock()
		pending--
		if ok {
			stats.Succeeded++
		} else {
			stats.Failed++
		}
		if pending == 0 && !closed {
			closed = true
			close(pushQ)
		}
	}

	enqueue := func(u *url.URL, depth int) bool {
		normalized := normalizeURL(u)
		key := normalized.String()

		stateMu.Lock()
		if !c.shouldVisit(ctx, normalized, depth, seeds[0]) {
			stats.Skipped++
			stateMu.Unlock()
			return false
		}
		if _, dup := seen[key]; dup {
			stats.Skipped++
			stateMu.Unlock()
			return false
		}
		if visited >= c.cfg.Crawl.MaxPages {
			stats.Skipped++
			stateMu.Unlock()
			return false
		}
		seen[key] = struct{}{}
		visited++
		if closed {
			stateMu.Unlock()
			return false
		}
		pending++
		stats.Enqueued++
		stateMu.Unlock()

		pushQ <- job{url: normalized, depth: depth}
		return true
	}

	var wg sync.WaitGroup
	for i := 0; i < workerN; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				ok := c.runOne(ctx, j, enqueue)
				if c.cfg.Crawl.RequestDelay > 0 {
					select {
					case <-ctx.Done():
					case <-time.After(c.cfg.Crawl.RequestDelay):
					}
				}
				finishOne(ok)
			}
		}()
	}

	for _, s := range seeds {
		enqueue(s, 0)
	}
	wg.Wait()
	<-queueDone
	return stats, ctx.Err()
}

// runOne は 1 ジョブ分のパイプラインを実行し、結果を sink に渡し、抽出リンクを enqueue する。
// enqueue が nil の場合（単一URLモード）は次URLを追加しない。
func (c *Crawler) runOne(ctx context.Context, j job, enqueue func(*url.URL, int) bool) bool {
	req := model.NewRequest(j.url, j.depth)
	out, err := c.pipeline.Run(ctx, req)
	if err != nil {
		slog.Warn("pipeline failed", "url", j.url.String(), "err", err.Error())
		return false
	}
	if c.sink != nil && out.Result != nil {
		c.sink(out.Result)
	}
	if enqueue != nil {
		for _, link := range out.Links {
			enqueue(link, j.depth+1)
		}
	}
	return true
}

// shouldVisit はクロール対象 URL の事前フィルタ。
// 既訪問判定は別 (seen map) で行うので、ここでは登録前のチェックのみ。
func (c *Crawler) shouldVisit(ctx context.Context, u *url.URL, depth int, base *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if depth > c.cfg.Crawl.MaxDepth {
		return false
	}

	// PDF 無効化フィルタ
	if !c.cfg.PDF.Enabled && strings.HasSuffix(strings.ToLower(u.Path), ".pdf") {
		return false
	}

	// ドメイン制限
	if base != nil {
		if !c.cfg.Crawl.AllowExternal {
			if !sameRegisteredDomain(u.Host, base.Host, c.cfg.Crawl.AllowSubdomains) {
				return false
			}
		}
	}

	// path パターン
	if len(c.includeRe) > 0 {
		ok := false
		for _, re := range c.includeRe {
			if re.MatchString(u.Path) {
				ok = true
				break
			}
		}
		if !ok {
			return false
		}
	}
	for _, re := range c.excludeRe {
		if re.MatchString(u.Path) {
			return false
		}
	}

	// robots.txt
	if c.cfg.Crawl.RespectRobotsTxt && c.robots != nil {
		ua := c.cfg.Request.Headers["User-Agent"]
		if !c.robots.Allowed(ctx, u, ua) {
			return false
		}
	}
	return true
}

// sameRegisteredDomain はホストが同一登録ドメインかを判定する。
// allowSubdomains=false の場合は完全一致を要求し、true の場合は末尾一致で許可する。
// 厳密な PSL 検査ではなく、テスト・開発で十分な簡易判定。
func sameRegisteredDomain(a, b string, allowSubdomains bool) bool {
	a = strings.ToLower(a)
	b = strings.ToLower(b)
	if a == b {
		return true
	}
	if !allowSubdomains {
		return false
	}
	// 末尾を ".base" で許容する
	return strings.HasSuffix(a, "."+b) || strings.HasSuffix(b, "."+a)
}

// normalizeURL はクロールフロンティアに入れる前の URL 正規化。
func normalizeURL(u *url.URL) *url.URL {
	cp := *u
	cp.Scheme = strings.ToLower(cp.Scheme)
	cp.Host = strings.ToLower(cp.Host)
	cp.Fragment = ""
	// デフォルトポート除去
	host := cp.Host
	switch {
	case cp.Scheme == "http" && strings.HasSuffix(host, ":80"):
		cp.Host = strings.TrimSuffix(host, ":80")
	case cp.Scheme == "https" && strings.HasSuffix(host, ":443"):
		cp.Host = strings.TrimSuffix(host, ":443")
	}
	return &cp
}
