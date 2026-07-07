package chromiumfetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNetworkIdleWatch_isIdle は接続ゼロ継続時間の判定を検証する。
func TestNetworkIdleWatch_isIdle(t *testing.T) {
	t.Run("正常系: 接続ゼロが idleDuration 続けば true", func(t *testing.T) {
		w := newNetworkIdleWatch(100 * time.Millisecond)
		w.mu.Lock()
		w.inflight = 0
		w.idleAt = time.Now().Add(-150 * time.Millisecond)
		w.mu.Unlock()

		assert.True(t, w.isIdle())
	})

	t.Run("正常系: 接続が残っていれば false", func(t *testing.T) {
		w := newNetworkIdleWatch(100 * time.Millisecond)
		w.mu.Lock()
		w.inflight = 1
		w.idleAt = time.Now().Add(-200 * time.Millisecond)
		w.mu.Unlock()

		assert.False(t, w.isIdle())
	})
}
