package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"meguri/internal/core"
	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
)

const defaultCacheMaxEntries = 8

// scrapeConfigFingerprint は Scrape 実行に影響する設定の正規化サブセット。
type scrapeConfigFingerprint struct {
	Request model.RequestConfig   `json:"request"`
	Content model.ContentConfig   `json:"content"`
	PDF     model.PDFConfig       `json:"pdf"`
	Plugins model.PluginSelection `json:"plugins"`
}

// cachedRunner は再利用する Kernel と Pipeline の組。
type cachedRunner struct {
	hash     string
	kernel   *core.Kernel
	pipeline *core.Pipeline
	// refs は実行中の利用者数。0 のエントリだけ LRU で Close する。
	refs int
}

// ScrapeCache は cfg hash 単位で Kernel を再利用する LRU キャッシュ。
type ScrapeCache struct {
	mu           sync.Mutex
	maxEntries   int
	order        []string
	entries      map[string]*cachedRunner
	fetchLimiter *fetchlimit.FetchLimiter
}

// SetFetchLimiter はジョブ共有の取得並列上限を設定する（Init 前に各 Kernel へ伝播）。
func (c *ScrapeCache) SetFetchLimiter(l *fetchlimit.FetchLimiter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.fetchLimiter = l
}

// NewScrapeCache は ScrapeCache を構築する。
func NewScrapeCache() *ScrapeCache {
	return &ScrapeCache{
		maxEntries: defaultCacheMaxEntries,
		entries:    make(map[string]*cachedRunner),
	}
}

// PipelineFor は cfg に対応するパイプラインを返す。同一 hash は再利用する。
//
// release は使い終わったら 1 回だけ呼ぶ。使用中の Kernel は LRU で Close しない。
func (c *ScrapeCache) PipelineFor(ctx context.Context, cfg *model.Config) (*core.Pipeline, func(), error) {
	runner, err := c.acquire(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}
	return runner.pipeline, func() { c.release(runner.hash) }, nil
}

// ScrapeWithConfig はキャッシュ済み Kernel で 1 URL を実行する。
func (c *ScrapeCache) ScrapeWithConfig(
	ctx context.Context,
	rawURL string,
	cfg *model.Config,
	progress core.ProgressSink,
	pause *PauseController,
) (*model.Result, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid url %q: %w", rawURL, err)
	}

	runner, err := c.acquire(ctx, cfg)
	if err != nil {
		return nil, err
	}
	defer c.release(runner.hash)

	urlStr := u.String()
	if pause != nil {
		if err := pause.WaitIfPaused(ctx); err != nil {
			return nil, err
		}
	}

	core.EmitProgress(progress, core.ProgressEvent{
		Kind: core.ProgressStarted,
		URL:  urlStr,
	})

	req := model.NewRequest(u, 0)
	out, err := runner.pipeline.Run(ctx, req)
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

// CloseAll は全キャッシュエントリの Kernel を Close する。
func (c *ScrapeCache) CloseAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, e := range c.entries {
		if e.kernel != nil {
			_ = e.kernel.Close(context.Background())
		}
	}
	c.entries = make(map[string]*cachedRunner)
	c.order = nil
}

func (c *ScrapeCache) acquire(ctx context.Context, cfg *model.Config) (*cachedRunner, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is nil")
	}
	hash, err := cfgHash(cfg)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	if e, ok := c.entries[hash]; ok {
		e.refs++
		c.touchLocked(hash)
		c.mu.Unlock()
		return e, nil
	}
	lim := c.fetchLimiter
	c.mu.Unlock()

	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
	if lim != nil {
		k.SetFetchLimiter(lim)
	}
	if err := k.Init(ctx); err != nil {
		return nil, fmt.Errorf("kernel init: %w", err)
	}

	c.mu.Lock()
	if e, ok := c.entries[hash]; ok {
		e.refs++
		c.touchLocked(hash)
		c.mu.Unlock()
		_ = k.Close(context.Background())
		return e, nil
	}
	e := &cachedRunner{
		hash:     hash,
		kernel:   k,
		pipeline: core.NewPipeline(k),
		refs:     1,
	}
	doomed := c.evictIdleLocked(c.maxEntries - 1)
	c.entries[hash] = e
	c.order = append(c.order, hash)
	c.mu.Unlock()
	closeKernels(doomed)
	return e, nil
}

func (c *ScrapeCache) release(hash string) {
	c.mu.Lock()
	e, ok := c.entries[hash]
	if !ok {
		c.mu.Unlock()
		return
	}
	if e.refs > 0 {
		e.refs--
	}
	var doomed []*core.Kernel
	if e.refs == 0 {
		doomed = c.evictIdleLocked(c.maxEntries)
	}
	c.mu.Unlock()
	closeKernels(doomed)
}

func (c *ScrapeCache) touchLocked(hash string) {
	for i, h := range c.order {
		if h == hash {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, hash)
			return
		}
	}
	c.order = append(c.order, hash)
}

// evictIdleLocked はエントリ数が keep を超える未使用分を外し、Close すべき Kernel を返す。
// 使用中（refs > 0）は残す。その場合キャッシュは一時的に keep を超える。
func (c *ScrapeCache) evictIdleLocked(keep int) []*core.Kernel {
	if c.maxEntries < 1 {
		return nil
	}
	if keep < 0 {
		keep = 0
	}
	var doomed []*core.Kernel
	for len(c.entries) > keep {
		idx := -1
		for i, h := range c.order {
			e := c.entries[h]
			if e != nil && e.refs == 0 {
				idx = i
				break
			}
		}
		if idx < 0 {
			return doomed
		}
		hash := c.order[idx]
		c.order = append(c.order[:idx], c.order[idx+1:]...)
		e := c.entries[hash]
		delete(c.entries, hash)
		if e.kernel != nil {
			doomed = append(doomed, e.kernel)
		}
	}
	return doomed
}

func closeKernels(kernels []*core.Kernel) {
	for _, k := range kernels {
		if k != nil {
			_ = k.Close(context.Background())
		}
	}
}

// cfgHash は scrape 用設定の SHA-256 十六進ハッシュを返す。
//
// Targets・ExcludeURLs・Crawl 巡回パラメータは hash 対象外。
func cfgHash(cfg *model.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}
	fp := scrapeConfigFingerprint{
		Request: cfg.Request,
		Content: cfg.Content,
		PDF:     cfg.PDF,
		Plugins: cfg.Plugins,
	}
	data, err := json.Marshal(fp)
	if err != nil {
		return "", fmt.Errorf("marshal scrape config: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
