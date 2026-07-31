package core

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"meguri/internal/domain/model"
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
	// includeHosts は許可ホスト集合（空なら制限なし）。キーは小文字の url.Host。
	includeHosts map[string]struct{}
	// excludeHosts は除外ホスト集合。キーは小文字の url.Host。
	excludeHosts map[string]struct{}
	// excludeURLs は完全一致でスキップする正規化 URL 集合。
	excludeURLs map[string]struct{}
	// skipScrapeURLs は fetch のみスキップする正規化 URL 集合（already_success 理由で通知）。
	skipScrapeURLs map[string]struct{}
	// skipScrapeLinkMap は skipScrapeURLs の保存 outbound リンク（正規化 URL → 生リンク文字列）。
	skipScrapeLinkMap map[string][]string

	// sink は各ページの Result 受け取り先（レガシー互換）。
	sink ResultSink
	// progress は URL 単位の進捗通知先。
	progress ProgressSink
	// pause は一時停止制御（nil 可）。
	pause *PauseController
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

// NewCrawler はクローラを構築する。
//
// robots は nil 可（その場合は判定をスキップ）。
// progress は nil 可（進捗通知なし）。
func NewCrawler(k *Kernel, pipeline *Pipeline, robots RobotsChecker, sink ResultSink, progress ProgressSink) *Crawler {
	cfg := k.Config()
	c := &Crawler{
		cfg:      cfg,
		kernel:   k,
		pipeline: pipeline,
		robots:   robots,
		sink:     sink,
		progress: progress,
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
	if len(cfg.Crawl.IncludeHosts) > 0 {
		c.includeHosts = make(map[string]struct{}, len(cfg.Crawl.IncludeHosts))
		for _, h := range cfg.Crawl.IncludeHosts {
			c.includeHosts[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
		}
	}
	if len(cfg.Crawl.ExcludeHosts) > 0 {
		c.excludeHosts = make(map[string]struct{}, len(cfg.Crawl.ExcludeHosts))
		for _, h := range cfg.Crawl.ExcludeHosts {
			c.excludeHosts[strings.ToLower(strings.TrimSpace(h))] = struct{}{}
		}
	}
	if len(cfg.Crawl.ExcludeURLs) > 0 {
		c.excludeURLs = make(map[string]struct{}, len(cfg.Crawl.ExcludeURLs))
		for _, raw := range cfg.Crawl.ExcludeURLs {
			u, err := url.Parse(raw)
			if err != nil {
				continue
			}
			c.excludeURLs[normalizeURL(u).String()] = struct{}{}
		}
	}
	if len(cfg.Crawl.SkipScrapeURLs) > 0 {
		c.skipScrapeURLs = make(map[string]struct{}, len(cfg.Crawl.SkipScrapeURLs))
		for _, raw := range cfg.Crawl.SkipScrapeURLs {
			u, err := url.Parse(raw)
			if err != nil {
				continue
			}
			c.skipScrapeURLs[normalizeURL(u).String()] = struct{}{}
		}
	}
	if len(cfg.Crawl.SkipScrapeLinkMap) > 0 {
		c.skipScrapeLinkMap = make(map[string][]string, len(cfg.Crawl.SkipScrapeLinkMap))
		for raw, links := range cfg.Crawl.SkipScrapeLinkMap {
			u, err := url.Parse(raw)
			if err != nil {
				continue
			}
			key := normalizeURL(u).String()
			c.skipScrapeLinkMap[key] = append([]string(nil), links...)
		}
	}
	return c
}

// SetPauseController は一時停止制御を設定する。
func (c *Crawler) SetPauseController(p *PauseController) {
	c.pause = p
}

// job はクロールキュー内の 1 件分の作業単位。
type job struct {
	// url は処理対象 URL。
	url *url.URL
	// depth はシードからの深度。
	depth int
	// parentURL はリンク発見元 URL（シードは空）。
	parentURL string
}

// Run は与えられたシード URL から BFS でクロールを実行する。
// crawl.enabled=false の場合は単一 URL モードとして seed[0] のみを処理する。
func (c *Crawler) Run(ctx context.Context, seeds []*url.URL) (*CrawlStats, error) {
	stats := &CrawlStats{}

	defer func() {
		emitProgress(c.progress, ProgressEvent{
			Kind:  ProgressCompleted,
			Stats: stats,
		})
	}()

	if !c.cfg.Crawl.Enabled {
		if len(seeds) == 0 {
			return stats, nil
		}
		stats.Enqueued = 1
		if err := c.waitIfPaused(ctx); err != nil {
			return stats, err
		}
		seed := normalizeURL(seeds[0])
		slog.Info("crawl url normalized",
			"raw", seeds[0].String(),
			"normalized", seed.String(),
			"depth", 0,
			"parent", "",
		)
		ok, skipped := c.runOne(ctx, job{url: seed, depth: 0}, nil)
		if skipped {
			stats.Skipped++
		} else if ok {
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
		// queueMu は pushQ への send と close を直列化し、閉じた channel への send panic を防ぐ。
		queueMu sync.Mutex
		seen    = map[string]struct{}{}
		visited int
		pending int
		closed  bool
		// seeding は初期シード enqueue 中。この間は pending==0 でも pushQ を閉じない
		//（先頭シード robots 拒否で後続シードが入れなくなるのを防ぐ）。
		seeding = true
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

	// maybeClosePushQ は pending が 0 なら pushQ を 1 回だけ閉じる。
	// stateMu 保持中に呼び、close 自体は queueMu 下で行う。
	// seeding 中は閉じない（シード投入完了後に再度呼ぶ）。
	maybeClosePushQ := func() {
		if pending != 0 || closed || seeding {
			return
		}
		closed = true
		queueMu.Lock()
		close(pushQ)
		queueMu.Unlock()
	}

	finishOne := func(ok, skipped bool) {
		stateMu.Lock()
		defer stateMu.Unlock()
		pending--
		if skipped {
			stats.Skipped++
		} else if ok {
			stats.Succeeded++
		} else {
			stats.Failed++
		}
		maybeClosePushQ()
	}

	emitSkip := func(u *url.URL, depth int, parentURL, reason string) {
		stats.Skipped++
		emitProgress(c.progress, ProgressEvent{
			Kind:       ProgressSkipped,
			URL:        u.String(),
			ParentURL:  parentURL,
			Depth:      depth,
			SkipReason: reason,
		})
	}

	// releaseReserve は provisional reserve を取り消す（seen は残す）。
	// robots 拒否時のロールバック用。pending が 0 なら maybeClosePushQ する。
	releaseReserve := func() {
		visited--
		pending--
		maybeClosePushQ()
	}

	enqueue := func(u *url.URL, depth int, parentURL string) bool {
		normalized := normalizeURL(u)
		slog.Info("crawl url normalized",
			"raw", u.String(),
			"normalized", normalized.String(),
			"depth", depth,
			"parent", parentURL,
		)
		key := normalized.String()

		stateMu.Lock()
		if _, dup := seen[key]; dup {
			stateMu.Unlock()
			emitSkip(normalized, depth, parentURL, "duplicate")
			return false
		}
		if reason := c.skipReason(normalized, depth, seeds[0]); reason != "" {
			stateMu.Unlock()
			emitSkip(normalized, depth, parentURL, reason)
			return false
		}
		if visited >= c.cfg.Crawl.MaxPages {
			stateMu.Unlock()
			emitSkip(normalized, depth, parentURL, "max_pages")
			return false
		}
		if closed {
			stateMu.Unlock()
			return false
		}
		// provisional reserve: robots I/O 中の二重 enqueue を防ぐ。
		seen[key] = struct{}{}
		visited++
		pending++
		stateMu.Unlock()

		if c.cfg.Crawl.RespectRobotsTxt && c.robots != nil {
			ua := c.cfg.Plugins.Stealth.HTTP.EffectiveUserAgent()
			if !c.robots.Allowed(ctx, normalized, ua) {
				stateMu.Lock()
				releaseReserve()
				stateMu.Unlock()
				emitSkip(normalized, depth, parentURL, "robots")
				return false
			}
		}

		// stateMu → queueMu の順で取得し、閉じた channel への send を防ぐ。
		// pending 予約中は maybeClosePushQ しないため、send 完了まで close は待つ。
		stateMu.Lock()
		stats.Enqueued++
		queueMu.Lock()
		stateMu.Unlock()
		pushQ <- job{url: normalized, depth: depth, parentURL: parentURL}
		queueMu.Unlock()
		return true
	}

	var wg sync.WaitGroup
	for i := 0; i < workerN; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if err := c.waitIfPaused(ctx); err != nil {
					return
				}
				ok, skipped := c.runOne(ctx, j, enqueue)
				if c.cfg.Crawl.RequestDelay > 0 {
					select {
					case <-ctx.Done():
					case <-time.After(c.cfg.Crawl.RequestDelay):
					}
				}
				finishOne(ok, skipped)
			}
		}()
	}

	for _, s := range seeds {
		enqueue(s, 0, "")
	}
	stateMu.Lock()
	seeding = false
	maybeClosePushQ()
	stateMu.Unlock()
	wg.Wait()
	<-queueDone
	return stats, ctx.Err()
}

// waitIfPaused は一時停止中なら解除または ctx キャンセルまで待つ。
func (c *Crawler) waitIfPaused(ctx context.Context) error {
	if c.pause == nil {
		return nil
	}
	return c.pause.WaitIfPaused(ctx)
}

// runOne は 1 ジョブ分のパイプラインを実行し、結果を通知し、抽出リンクを enqueue する。
// enqueue が nil の場合（単一URLモード）は次URLを追加しない。
//
// 戻り値 ok はパイプライン成功、skipped は fetch スキップ（already_success）を表す。
// skipped=true のとき ok は無視する。
func (c *Crawler) runOne(ctx context.Context, j job, enqueue func(*url.URL, int, string) bool) (ok bool, skipped bool) {
	urlStr := j.url.String()
	if _, skipFetch := c.skipScrapeURLs[urlStr]; skipFetch {
		emitProgress(c.progress, ProgressEvent{
			Kind:       ProgressSkipped,
			URL:        urlStr,
			ParentURL:  j.parentURL,
			Depth:      j.depth,
			SkipReason: "already_success",
		})
		c.expandCachedLinks(j, enqueue)
		return false, true
	}

	emitProgress(c.progress, ProgressEvent{
		Kind:      ProgressStarted,
		URL:       urlStr,
		ParentURL: j.parentURL,
		Depth:     j.depth,
	})

	req := model.NewRequest(j.url, j.depth)
	out, err := c.pipeline.Run(ctx, req)
	if err != nil {
		slog.Warn("pipeline failed", "url", urlStr, "err", err.Error())
		emitProgress(c.progress, ProgressEvent{
			Kind:      ProgressFailed,
			URL:       urlStr,
			ParentURL: j.parentURL,
			Depth:     j.depth,
			Error:     err.Error(),
		})
		return false, false
	}
	if out.Result != nil {
		if c.sink != nil {
			c.sink(out.Result)
		}
		emitProgress(c.progress, ProgressEvent{
			Kind:      ProgressSucceeded,
			URL:       urlStr,
			ParentURL: j.parentURL,
			Depth:     j.depth,
			Result:    out.Result,
		})
	}
	if enqueue != nil {
		parent := urlStr
		for _, link := range out.Links {
			c.tryEnqueueDiscovered(link, j.depth+1, parent, enqueue)
		}
	}
	return true, false
}

// expandCachedLinks は skip scrape 対象 URL の保存リンクを enqueue する。
func (c *Crawler) expandCachedLinks(j job, enqueue func(*url.URL, int, string) bool) {
	if enqueue == nil || len(c.skipScrapeLinkMap) == 0 {
		return
	}
	links, ok := c.skipScrapeLinkMap[j.url.String()]
	if !ok || len(links) == 0 {
		return
	}
	parent := j.url.String()
	for _, raw := range links {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		c.tryEnqueueDiscovered(u, j.depth+1, parent, enqueue)
	}
}

// tryEnqueueDiscovered は発見リンクを enqueue し、成功時に ProgressLinkDiscovered を出す。
func (c *Crawler) tryEnqueueDiscovered(
	link *url.URL,
	depth int,
	parent string,
	enqueue func(*url.URL, int, string) bool,
) {
	normalizedLink := normalizeURL(link)
	slog.Info("crawl link discovered",
		"raw", link.String(),
		"normalized", normalizedLink.String(),
		"depth", depth,
		"parent", parent,
	)
	if enqueue(link, depth, parent) {
		emitProgress(c.progress, ProgressEvent{
			Kind:      ProgressLinkDiscovered,
			URL:       normalizedLink.String(),
			ParentURL: parent,
			Depth:     depth,
		})
	}
}

// skipReason はネットワーク不要の訪問不可理由を返す。訪問可能なら空文字。
// robots.txt 判定は含まない（enqueue 側で max_pages 予約後に lock 外で行う）。
// SkipScrapeURLs は enqueue 対象（枠消費）とし、fetch スキップは runOne 側で行う。
func (c *Crawler) skipReason(u *url.URL, depth int, base *url.URL) string {
	if u.Scheme != "http" && u.Scheme != "https" {
		return "invalid_scheme"
	}
	if depth > c.cfg.Crawl.MaxDepth {
		return "max_depth"
	}
	host := strings.ToLower(u.Host)
	if c.excludeHosts != nil {
		if _, ok := c.excludeHosts[host]; ok {
			return "exclude_hosts"
		}
	}
	if c.excludeURLs != nil {
		if _, ok := c.excludeURLs[u.String()]; ok {
			return "exclude_urls"
		}
	}
	if !c.cfg.PDF.Enabled && strings.HasSuffix(strings.ToLower(u.Path), ".pdf") {
		return "pdf_disabled"
	}
	if base != nil {
		if !c.cfg.Crawl.AllowExternal {
			if !sameRegisteredDomain(u.Host, base.Host, c.cfg.Crawl.AllowSubdomains) {
				return "external_domain"
			}
		}
	}
	if len(c.includeHosts) > 0 {
		if _, ok := c.includeHosts[host]; !ok {
			return "include_hosts"
		}
	}
	if len(c.includeRe) > 0 {
		ok := false
		for _, re := range c.includeRe {
			if re.MatchString(u.Path) {
				ok = true
				break
			}
		}
		if !ok {
			return "include_paths"
		}
	}
	for _, re := range c.excludeRe {
		if re.MatchString(u.Path) {
			return "exclude_paths"
		}
	}
	return ""
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
