package chromiumfetch

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCloseAllBrowserSessions はプール内セッションの強制終了を検証する。
func TestCloseAllBrowserSessions(t *testing.T) {
	t.Run("正常系: 空プールでも panic しない", func(t *testing.T) {
		CloseAllBrowserSessions()
		defaultBrowserPool.mu.Lock()
		assert.Empty(t, defaultBrowserPool.entries)
		defaultBrowserPool.mu.Unlock()
	})

	t.Run("正常系: 登録済みセッションを全削除する", func(t *testing.T) {
		key := sessionKey("test|true|ua")
		defaultBrowserPool.mu.Lock()
		defaultBrowserPool.entries[key] = &browserSession{key: key, clients: 3}
		defaultBrowserPool.mu.Unlock()

		CloseAllBrowserSessions()

		defaultBrowserPool.mu.Lock()
		assert.Empty(t, defaultBrowserPool.entries)
		defaultBrowserPool.mu.Unlock()
	})
}
