package wails_service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri-app/internal/model"
	"meguri/pkg/runner"
)

func TestExistingNodeWorkerCount(t *testing.T) {
	t.Run("未設定は 4", func(t *testing.T) {
		cfg := &runner.Config{}
		cfg.Crawl.MaxConcurrency = 0
		assert.Equal(t, 4, existingNodeWorkerCount(cfg))
	})

	t.Run("request_delay が正なら 1", func(t *testing.T) {
		cfg := &runner.Config{}
		cfg.Crawl.MaxConcurrency = 8
		cfg.Crawl.RequestDelay = time.Second
		assert.Equal(t, 1, existingNodeWorkerCount(cfg))
	})

	t.Run("65 以上は 64", func(t *testing.T) {
		cfg := &runner.Config{}
		cfg.Crawl.MaxConcurrency = 100
		assert.Equal(t, 64, existingNodeWorkerCount(cfg))
	})
}

func TestRunExistingNodeScrapePool(t *testing.T) {
	makeWork := func(n int) []existingNodeWork {
		out := make([]existingNodeWork, n)
		for i := range out {
			id := fmt.Sprintf("n%d", i)
			out[i] = existingNodeWork{
				nodeID: id,
				node:   model.GraphNodeDTO{ID: id, URLNormalized: "https://example.com/" + id},
			}
		}
		return out
	}

	t.Run("正常系: max_concurrency まで並列に scrape する", func(t *testing.T) {
		var (
			inFlight atomic.Int64
			maxSeen  atomic.Int64
			done     atomic.Int64
		)
		const workers = 3
		const jobs = 6
		work := makeWork(jobs)

		err := runExistingNodeScrapePool(t.Context(), existingNodeScrapePoolConfig{
			workerN: workers,
			work:    work,
			hooks: existingNodeScrapePoolHooks{
				waitIfPaused:  func(context.Context) error { return nil },
				robotsAllowed: func(context.Context, string) bool { return true },
				onRobotsSkip:  func(context.Context, string, string) {},
				onEnqueued:    func() {},
				scrape: func(ctx context.Context, _ model.GraphNodeDTO) error {
					cur := inFlight.Add(1)
					for {
						old := maxSeen.Load()
						if cur <= old || maxSeen.CompareAndSwap(old, cur) {
							break
						}
					}
					time.Sleep(30 * time.Millisecond)
					inFlight.Add(-1)
					done.Add(1)
					return nil
				},
				onSucceeded: func() {},
				onFailed:    func() {},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(jobs), done.Load())
		assert.GreaterOrEqual(t, maxSeen.Load(), int64(2))
		assert.LessOrEqual(t, maxSeen.Load(), int64(workers))
	})

	t.Run("正常系: request_delay は取得後のみ worker 1", func(t *testing.T) {
		var scrapeTimes []time.Time
		var mu sync.Mutex
		work := makeWork(2)

		start := time.Now()
		err := runExistingNodeScrapePool(t.Context(), existingNodeScrapePoolConfig{
			workerN:      1,
			requestDelay: 80 * time.Millisecond,
			work:         work,
			hooks: existingNodeScrapePoolHooks{
				waitIfPaused:  func(context.Context) error { return nil },
				robotsAllowed: func(context.Context, string) bool { return true },
				onRobotsSkip:  func(context.Context, string, string) {},
				onEnqueued:    func() {},
				scrape: func(context.Context, model.GraphNodeDTO) error {
					mu.Lock()
					scrapeTimes = append(scrapeTimes, time.Now())
					mu.Unlock()
					return nil
				},
				onSucceeded: func() {},
				onFailed:    func() {},
			},
		})
		require.NoError(t, err)
		require.Len(t, scrapeTimes, 2)
		gap := scrapeTimes[1].Sub(scrapeTimes[0])
		assert.GreaterOrEqual(t, gap, 70*time.Millisecond)
		assert.GreaterOrEqual(t, time.Since(start), 70*time.Millisecond)
	})

	t.Run("正常系: robots スキップは enqueued せず scrape もしない", func(t *testing.T) {
		var enqueued atomic.Int64
		var skipped atomic.Int64
		var scraped atomic.Int64
		work := []existingNodeWork{
			{nodeID: "a", node: model.GraphNodeDTO{ID: "a", URLNormalized: "https://example.com/deny"}},
			{nodeID: "b", node: model.GraphNodeDTO{ID: "b", URLNormalized: "https://example.com/ok"}},
		}

		err := runExistingNodeScrapePool(t.Context(), existingNodeScrapePoolConfig{
			workerN:      1,
			requestDelay: 200 * time.Millisecond,
			work:         work,
			hooks: existingNodeScrapePoolHooks{
				waitIfPaused: func(context.Context) error { return nil },
				robotsAllowed: func(_ context.Context, rawURL string) bool {
					return rawURL != "https://example.com/deny"
				},
				onRobotsSkip: func(context.Context, string, string) { skipped.Add(1) },
				onEnqueued:   func() { enqueued.Add(1) },
				scrape: func(context.Context, model.GraphNodeDTO) error {
					scraped.Add(1)
					return nil
				},
				onSucceeded: func() {},
				onFailed:    func() {},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), enqueued.Load())
		assert.Equal(t, int64(1), skipped.Load())
		assert.Equal(t, int64(1), scraped.Load())
	})

	t.Run("正常系: キャンセル時は ctx.Err を返し failed は加算しない", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		var failed atomic.Int64
		block := make(chan struct{})
		work := makeWork(1)

		go func() {
			time.Sleep(20 * time.Millisecond)
			cancel()
		}()

		err := runExistingNodeScrapePool(ctx, existingNodeScrapePoolConfig{
			workerN: 1,
			work:    work,
			hooks: existingNodeScrapePoolHooks{
				waitIfPaused:  func(context.Context) error { return nil },
				robotsAllowed: func(context.Context, string) bool { return true },
				onRobotsSkip:  func(context.Context, string, string) {},
				onEnqueued:    func() {},
				scrape: func(ctx context.Context, _ model.GraphNodeDTO) error {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-block:
						return nil
					}
				},
				onSucceeded: func() {},
				onFailed:    func() { failed.Add(1) },
			},
		})
		require.ErrorIs(t, err, context.Canceled)
		assert.Equal(t, int64(0), failed.Load())
	})

	t.Run("正常系: 一時停止中は次のノードに進まない", func(t *testing.T) {
		pause := runner.NewPauseController()
		pause.Pause()
		var scraped atomic.Int64
		work := makeWork(2)
		done := make(chan struct{})

		go func() {
			_ = runExistingNodeScrapePool(t.Context(), existingNodeScrapePoolConfig{
				workerN: 1,
				work:    work,
				hooks: existingNodeScrapePoolHooks{
					waitIfPaused:  pause.WaitIfPaused,
					robotsAllowed: func(context.Context, string) bool { return true },
					onRobotsSkip:  func(context.Context, string, string) {},
					onEnqueued:    func() {},
					scrape: func(context.Context, model.GraphNodeDTO) error {
						scraped.Add(1)
						return nil
					},
					onSucceeded: func() {},
					onFailed:    func() {},
				},
			})
			close(done)
		}()

		time.Sleep(50 * time.Millisecond)
		assert.Equal(t, int64(0), scraped.Load())
		pause.Resume()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Fatal("pool did not finish after resume")
		}
		assert.Equal(t, int64(2), scraped.Load())
	})
}
