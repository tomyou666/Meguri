package chromiumfetch

import (
	"context"
	"sync"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// networkIdleWatch は Navigate 前後でネットワーク静止を監視する。
type networkIdleWatch struct {
	idleDuration time.Duration
	mu           sync.Mutex
	inflight     int
	idleAt       time.Time
	ready        chan struct{}
}

// newNetworkIdleWatch は network_idle 待機用の監視器を構築する。
func newNetworkIdleWatch(idleDuration time.Duration) *networkIdleWatch {
	if idleDuration <= 0 {
		idleDuration = 500 * time.Millisecond
	}
	return &networkIdleWatch{
		idleDuration: idleDuration,
		ready:        make(chan struct{}, 1),
	}
}

func (w *networkIdleWatch) markReady() {
	select {
	case w.ready <- struct{}{}:
	default:
	}
}

func (w *networkIdleWatch) onEvent(ev any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch ev := ev.(type) {
	case *network.EventRequestWillBeSent:
		w.inflight++
		w.idleAt = time.Time{}
	case *network.EventLoadingFinished, *network.EventLoadingFailed:
		_ = ev
		if w.inflight > 0 {
			w.inflight--
		}
		if w.inflight == 0 {
			w.idleAt = time.Now()
			w.markReady()
		}
	}
}

func (w *networkIdleWatch) isIdle() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.inflight == 0 && !w.idleAt.IsZero() && time.Since(w.idleAt) >= w.idleDuration
}

// begin は Navigate 前に network 監視を開始するアクションを返す。
func (w *networkIdleWatch) begin() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		chromedp.ListenTarget(ctx, w.onEvent)
		return network.Enable().Do(ctx)
	})
}

// wait は Navigate 後にネットワーク静止を待つアクションを返す。
func (w *networkIdleWatch) wait() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			if w.isIdle() {
				return nil
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
				w.mu.Lock()
				if w.inflight == 0 {
					if w.idleAt.IsZero() {
						w.idleAt = time.Now()
					} else if time.Since(w.idleAt) >= w.idleDuration {
						w.mu.Unlock()
						return nil
					}
				}
				w.mu.Unlock()
			case <-w.ready:
			}
		}
	})
}
