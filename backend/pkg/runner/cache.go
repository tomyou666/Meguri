package runner

import (
	"context"
	"fmt"
	"net/url"
	"sync"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

const defaultCacheMaxEntries = 8

// cachedRunner は再利用する Kernel と Pipeline の組。
type cachedRunner struct {
	hash     string
	kernel   *core.Kernel
	pipeline *core.Pipeline
}

// RunnerCache は cfg hash 単位で Kernel を再利用する LRU キャッシュ。
type RunnerCache struct {
	mu         sync.Mutex
	maxEntries int
	order      []string
	entries    map[string]*cachedRunner
}

// NewRunnerCache は RunnerCache を構築する。
func NewRunnerCache() *RunnerCache {
	return &RunnerCache{
		maxEntries: defaultCacheMaxEntries,
		entries:    make(map[string]*cachedRunner),
	}
}

// ScrapeWithConfig はキャッシュ済み Kernel で 1 URL を実行する。
func (c *RunnerCache) ScrapeWithConfig(
	ctx context.Context,
	rawURL string,
	cfg *model.Config,
	progress ProgressSink,
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
func (c *RunnerCache) CloseAll() {
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

func (c *RunnerCache) getOrCreate(ctx context.Context, hash string, cfg *model.Config) (*cachedRunner, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e, ok := c.entries[hash]; ok {
		c.touchLocked(hash)
		return e, nil
	}

	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
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

func (c *RunnerCache) touchLocked(hash string) {
	for i, h := range c.order {
		if h == hash {
			c.order = append(c.order[:i], c.order[i+1:]...)
			c.order = append(c.order, hash)
			return
		}
	}
	c.order = append(c.order, hash)
}

func (c *RunnerCache) evictIfNeededLocked() {
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
