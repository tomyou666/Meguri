package chromiumfetch

import (
	"testing"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/stretchr/testify/assert"
)

// TestNetworkIdleWatch_pollIdle は接続ゼロ継続時間の判定を検証する。
func TestNetworkIdleWatch_pollIdle(t *testing.T) {
	t.Run("正常系: 接続ゼロが idleDuration 続けば true", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mu.Lock()
		w.idleAt = time.Now().Add(-150 * time.Millisecond)
		w.mu.Unlock()

		assert.True(t, w.pollIdle())
	})

	t.Run("正常系: 接続が残っていれば false", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mu.Lock()
		w.tracked["req"] = time.Now()
		w.idleAt = time.Now().Add(-200 * time.Millisecond)
		w.mu.Unlock()

		assert.False(t, w.pollIdle())
	})
}

// TestNetworkIdleWatch_shouldTrackRequest は除外ルールを検証する。
func TestNetworkIdleWatch_shouldTrackRequest(t *testing.T) {
	t.Run("正常系: 追跡開始後のメインフレームリクエストは追跡する", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID: "1",
			Type:      network.ResourceTypeScript,
			FrameID:   "main",
			Request:   &network.Request{URL: "https://example.com/app.js"},
		}

		assert.True(t, w.shouldTrackRequest(ev))
	})

	t.Run("正常系: 追跡開始前は追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		ev := &network.EventRequestWillBeSent{
			RequestID: "1",
			Type:      network.ResourceTypeScript,
			FrameID:   "main",
			Request:   &network.Request{URL: "https://example.com/app.js"},
		}

		assert.False(t, w.shouldTrackRequest(ev))
	})

	t.Run("正常系: リダイレクト継続は追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID:        "1",
			Type:             network.ResourceTypeDocument,
			FrameID:          "main",
			RedirectResponse: &network.Response{},
		}

		assert.False(t, w.shouldTrackRequest(ev))
	})

	t.Run("正常系: サブフレームは追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID: "iframe-xhr",
			Type:      network.ResourceTypeXHR,
			FrameID:   "child",
		}

		assert.False(t, w.shouldTrackRequest(ev))
	})

	t.Run("正常系: FrameID が空なら追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID: "empty-frame",
			Type:      network.ResourceTypeScript,
		}

		assert.False(t, w.shouldTrackRequest(ev))
	})

	t.Run("正常系: mainFrameID 未設定なら追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID: "doc",
			Type:      network.ResourceTypeDocument,
			FrameID:   "main",
		}

		assert.False(t, w.shouldTrackRequest(ev))
		assert.Equal(t, cdp.FrameID(""), w.mainFrameID)
	})

	t.Run("正常系: WebSocket は追跡しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true
		ev := &network.EventRequestWillBeSent{
			RequestID: "ws",
			Type:      network.ResourceTypeWebSocket,
			FrameID:   "main",
		}

		assert.False(t, w.shouldTrackRequest(ev))
	})
}

// TestNetworkIdleWatch_onEvent_subframe は iframe 配下が inflight を汚さないことを検証する。
func TestNetworkIdleWatch_onEvent_subframe(t *testing.T) {
	t.Run("正常系: サブフレームの Document / XHR は inflight に加算しない", func(t *testing.T) {
		w := newNetworkIdleWatch(100*time.Millisecond, 10*time.Second)
		w.mainFrameID = "main"
		w.trackingActive = true

		w.onEvent(&network.EventRequestWillBeSent{
			RequestID: "main-doc",
			Type:      network.ResourceTypeDocument,
			FrameID:   "main",
			Request:   &network.Request{URL: "https://example.com/"},
		})
		w.onEvent(&network.EventLoadingFinished{RequestID: "main-doc"})

		w.onEvent(&network.EventRequestWillBeSent{
			RequestID: "iframe-doc",
			Type:      network.ResourceTypeDocument,
			FrameID:   "child",
			Request:   &network.Request{URL: "https://example.com/iframe"},
		})
		w.onEvent(&network.EventRequestWillBeSent{
			RequestID: "iframe-xhr",
			Type:      network.ResourceTypeXHR,
			FrameID:   "child",
			Request:   &network.Request{URL: "https://example.com/hang"},
		})

		w.mu.Lock()
		defer w.mu.Unlock()
		assert.Equal(t, 0, w.inflightCount())
		assert.Empty(t, w.tracked)
	})
}

// TestNetworkIdleWatch_evictStaleRequests は requestMaxAge 超過で追跡を外すことを検証する。
func TestNetworkIdleWatch_evictStaleRequests(t *testing.T) {
	t.Run("正常系: max_age 超過リクエストを外して idle 判定できる", func(t *testing.T) {
		w := newNetworkIdleWatch(50*time.Millisecond, 100*time.Millisecond)
		now := time.Now()

		w.tracked["stuck"] = now.Add(-200 * time.Millisecond)

		w.evictStaleRequests(now)

		assert.Equal(t, 0, w.inflightCount())
		assert.Empty(t, w.tracked)
		assert.False(t, w.idleAt.IsZero())
	})
}
