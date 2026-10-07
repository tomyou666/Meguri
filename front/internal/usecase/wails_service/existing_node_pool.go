package wails_service

import (
	"context"
	"sync"
	"time"

	"meguri-app/internal/model"
	"meguri/pkg/runner"
)

// existingNodeWork はワーカープールが処理する 1 ノード。
type existingNodeWork struct {
	nodeID string
	node   model.GraphNodeDTO
}

// existingNodeWorkerCount は run 設定からワーカー数を返す（request_delay>0 なら 1）。
func existingNodeWorkerCount(runCfg *runner.Config) int {
	if runCfg == nil {
		return 4
	}
	n := runCfg.Crawl.MaxConcurrency
	if n < 1 {
		n = 4
	}
	if n > 64 {
		n = 64
	}
	if runCfg.Crawl.RequestDelay > 0 {
		n = 1
	}
	return n
}

// existingNodeScrapePoolHooks はプールの外部依存（テスト差し替え用）。
type existingNodeScrapePoolHooks struct {
	waitIfPaused  func(ctx context.Context) error
	robotsAllowed func(ctx context.Context, rawURL string) bool
	onRobotsSkip  func(ctx context.Context, nodeID, url string)
	onEnqueued    func()
	scrape        func(ctx context.Context, node model.GraphNodeDTO) error
	onSucceeded   func()
	onFailed      func()
}

// existingNodeScrapePoolConfig は既存ノード scrape プールの設定。
type existingNodeScrapePoolConfig struct {
	workerN      int
	requestDelay time.Duration
	work         []existingNodeWork
	hooks        existingNodeScrapePoolHooks
}

// runExistingNodeScrapePool は work をワーカーで並列処理する。完了順は終了順。
func runExistingNodeScrapePool(ctx context.Context, cfg existingNodeScrapePoolConfig) error {
	if len(cfg.work) == 0 {
		return nil
	}
	workerN := cfg.workerN
	if workerN < 1 {
		workerN = 1
	}

	var (
		idxMu   sync.Mutex
		nextIdx int
		stopMu  sync.Mutex
		stopErr error
	)
	shouldStop := func() bool {
		stopMu.Lock()
		defer stopMu.Unlock()
		return stopErr != nil
	}
	setStop := func(err error) {
		if err == nil {
			return
		}
		stopMu.Lock()
		if stopErr == nil {
			stopErr = err
		}
		stopMu.Unlock()
	}

	var wg sync.WaitGroup
	for i := 0; i < workerN; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				if shouldStop() {
					return
				}
				if err := cfg.hooks.waitIfPaused(ctx); err != nil {
					setStop(err)
					return
				}
				idxMu.Lock()
				if nextIdx >= len(cfg.work) {
					idxMu.Unlock()
					return
				}
				item := cfg.work[nextIdx]
				nextIdx++
				idxMu.Unlock()

				url := item.node.URLNormalized
				if cfg.hooks.robotsAllowed != nil && !cfg.hooks.robotsAllowed(ctx, url) {
					if ctx.Err() != nil {
						setStop(ctx.Err())
						return
					}
					cfg.hooks.onRobotsSkip(ctx, item.nodeID, url)
					continue
				}
				if ctx.Err() != nil {
					setStop(ctx.Err())
					return
				}

				cfg.hooks.onEnqueued()
				err := cfg.hooks.scrape(ctx, item.node)
				if err != nil {
					if ctx.Err() != nil {
						setStop(ctx.Err())
						return
					}
					cfg.hooks.onFailed()
				} else {
					cfg.hooks.onSucceeded()
				}

				if cfg.requestDelay > 0 && workerN == 1 {
					select {
					case <-ctx.Done():
						setStop(ctx.Err())
						return
					case <-time.After(cfg.requestDelay):
					}
				}
			}
		}()
	}
	wg.Wait()

	stopMu.Lock()
	defer stopMu.Unlock()
	return stopErr
}
