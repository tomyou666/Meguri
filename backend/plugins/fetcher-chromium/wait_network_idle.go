package chromiumfetch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

// networkIdleWatch は Navigate 前後でネットワーク静止を監視する。
type networkIdleWatch struct {
	idleDuration   time.Duration
	requestMaxAge  time.Duration
	mu             sync.Mutex
	idleAt         time.Time
	mainFrameID    cdp.FrameID
	tracked        map[network.RequestID]time.Time
	trackingActive bool
}

// newNetworkIdleWatch は network_idle 待機用の監視器を構築する。
func newNetworkIdleWatch(idleDuration, requestMaxAge time.Duration) *networkIdleWatch {
	if idleDuration <= 0 {
		idleDuration = 500 * time.Millisecond
	}
	if requestMaxAge <= 0 {
		requestMaxAge = 10 * time.Second
	}
	return &networkIdleWatch{
		idleDuration:  idleDuration,
		requestMaxAge: requestMaxAge,
		tracked:       make(map[network.RequestID]time.Time),
	}
}

func (w *networkIdleWatch) inflightCount() int {
	return len(w.tracked)
}

// shouldTrackRequest は network_idle のカウント対象かを返す。
//
// 除外するリクエスト:
// - 追跡開始前のイベント
// - リダイレクト継続（同一 requestId の再送。LoadingFinished は最終段のみ）
// - 既に追跡中の requestId
// - メインフレーム以外（iframe 配下を含む）
// - FrameID が空のリクエスト
// - WebSocket / EventSource（長寿命接続）
func (w *networkIdleWatch) shouldTrackRequest(ev *network.EventRequestWillBeSent) bool {
	if !w.trackingActive {
		return false
	}
	if ev.RedirectResponse != nil {
		return false
	}
	if _, ok := w.tracked[ev.RequestID]; ok {
		return false
	}
	if ev.Type == network.ResourceTypeWebSocket || ev.Type == network.ResourceTypeEventSource {
		return false
	}
	if w.mainFrameID == "" || ev.FrameID == "" || ev.FrameID != w.mainFrameID {
		return false
	}
	return true
}

func (w *networkIdleWatch) onRequestStart(ev *network.EventRequestWillBeSent) {
	if !w.shouldTrackRequest(ev) {
		return
	}
	w.tracked[ev.RequestID] = time.Now()
	w.idleAt = time.Time{}
}

func (w *networkIdleWatch) onRequestEnd(requestID network.RequestID) {
	if _, ok := w.tracked[requestID]; !ok {
		return
	}
	delete(w.tracked, requestID)
	if w.inflightCount() == 0 {
		w.idleAt = time.Now()
	}
}

// evictStaleRequests は requestMaxAge を超えた追跡中リクエストを諦めて外す。
func (w *networkIdleWatch) evictStaleRequests(now time.Time) {
	evicted := false
	for id, startedAt := range w.tracked {
		if now.Sub(startedAt) < w.requestMaxAge {
			continue
		}
		delete(w.tracked, id)
		evicted = true
	}
	if evicted && w.inflightCount() == 0 {
		w.idleAt = now
	}
}

func (w *networkIdleWatch) onEvent(ev any) {
	w.mu.Lock()
	defer w.mu.Unlock()

	switch ev := ev.(type) {
	case *network.EventRequestWillBeSent:
		w.onRequestStart(ev)
	case *network.EventLoadingFinished:
		w.onRequestEnd(ev.RequestID)
	case *network.EventLoadingFailed:
		w.onRequestEnd(ev.RequestID)
	}
}

// begin は Navigate 前に network 監視を開始するアクションを返す。
func (w *networkIdleWatch) begin() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		chromedp.ListenTarget(ctx, w.onEvent)
		return network.Enable().Do(ctx)
	})
}

// startTracking は Navigate 後にメインフレーム ID を確定し、追跡を開始する。
func (w *networkIdleWatch) startTracking(ctx context.Context) error {
	tree, err := page.GetFrameTree().Do(ctx)
	if err != nil {
		return fmt.Errorf("network_idle: get frame tree: %w", err)
	}
	if tree == nil || tree.Frame == nil {
		return errors.New("network_idle: main frame not found")
	}
	w.mu.Lock()
	w.mainFrameID = tree.Frame.ID
	w.tracked = make(map[network.RequestID]time.Time)
	w.idleAt = time.Time{}
	w.trackingActive = true
	if w.inflightCount() == 0 {
		w.idleAt = time.Now()
	}
	w.mu.Unlock()
	return nil
}

// pollIdle は現在の追跡状態が静止条件を満たすかを返す。
func (w *networkIdleWatch) pollIdle() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evictStaleRequests(time.Now())
	if w.inflightCount() == 0 {
		if w.idleAt.IsZero() {
			w.idleAt = time.Now()
			return false
		}
		return time.Since(w.idleAt) >= w.idleDuration
	}
	return false
}

// wait は Navigate 後にネットワーク静止を待つアクションを返す。
func (w *networkIdleWatch) wait() chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := w.startTracking(ctx); err != nil {
			return err
		}

		for {
			if w.pollIdle() {
				return nil
			}
			if err := chromedp.Sleep(50 * time.Millisecond).Do(ctx); err != nil {
				return err
			}
		}
	})
}
