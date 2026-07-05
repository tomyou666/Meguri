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

	hash, err := cfgHash(cfg)
	if err != nil {
		return nil, err
	}

	runner, err := c.getOrCreate(ctx, hash, cfg)
	if err != nil {
		return nil, err
	}

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

func (c *ScrapeCache) getOrCreate(ctx context.Context, hash string, cfg *model.Config) (*cachedRunner, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[hash]; ok {
		c.touchLocked(hash)
		return e, nil
	}

	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
	if c.fetchLimiter != nil {
		k.SetFetchLimiter(c.fetchLimiter)
	}
	if err := k.Init(ctx); err != nil {
		return nil, fmt.Errorf("kernel init: %w", err)
	}
	e := &cachedRunner{
		hash:     hash,
		kernel:   k,
		pipeline: core.NewPipeline(k),
	}
	c.evictIfNeededLocked()
	c.entries[hash] = e
	c.order = append(c.order, hash)
	return e, nil
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

func (c *ScrapeCache) evictIfNeededLocked() {
	for len(c.entries) >= c.maxEntries && len(c.order) > 0 {
		oldest := c.order[0]
		c.order = c.order[1:]
		if e, ok := c.entries[oldest]; ok {
			if e.kernel != nil {
				_ = e.kernel.Close(context.Background())
			}
			delete(c.entries, oldest)
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
