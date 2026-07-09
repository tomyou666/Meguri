package chromiumfetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/core"
	"meguri/internal/domain/model"
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

// TestClient_Get_SharedSessionSurvivesRequestCancel は 1 回目のリクエスト context を
// cancel しても共有ブラウザが生き、2 回目の取得が成功することを検証する。
func TestClient_Get_SharedSessionSurvivesRequestCancel(t *testing.T) {
	if _, err := resolveBrowserPath(""); err != nil {
		t.Skip("chromium browser not available: " + err.Error())
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1 id="done">done</h1></body></html>`))
	}))
	t.Cleanup(srv.Close)

	cfg := model.Default()
	cfg.Plugins.Fetcher = model.FetcherChromium
	cfg.Plugins.FetcherConfig.WaitUntil = model.WaitUntilNone
	host := core.NewHost(&cfg)
	c := &client{}
	require.NoError(t, c.Init(context.Background(), host))
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	u, err := url.Parse(srv.URL + "/")
	require.NoError(t, err)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 30*time.Second)
	res1, err := c.Get(ctx1, u, nil)
	require.NoError(t, err)
	assert.Contains(t, string(res1.Body), "done")
	cancel1()

	require.True(t, c.poolJoined)
	defaultBrowserPool.mu.Lock()
	_, ok := defaultBrowserPool.entries[c.poolKey]
	defaultBrowserPool.mu.Unlock()
	require.True(t, ok, "shared browser session must survive request cancel")

	ctx2, cancel2 := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel2)
	res2, err := c.Get(ctx2, u, nil)
	require.NoError(t, err)
	assert.Contains(t, string(res2.Body), "done")
}
