package chromiumfetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// newNetworkIdleTestServer は network_idle 結合テスト用の httptest サーバーを起動する。
func newNetworkIdleTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/hang", func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch path {
		case "/network_idle_iframe_hang.html":
			serveNetworkIdleHTML(t, w, "network_idle_iframe_hang.html")
		case "/network_idle_iframe_child.html":
			serveNetworkIdleHTML(t, w, "network_idle_iframe_child.html")
		case "/network_idle_main_hang.html":
			serveNetworkIdleHTML(t, w, "network_idle_main_hang.html")
		default:
			http.NotFound(w, r)
		}
	})
	return httptest.NewServer(mux)
}

func serveNetworkIdleHTML(t *testing.T, w http.ResponseWriter, name string) {
	t.Helper()
	path := filepath.Join(testdataDir(t), "html", name)
	b, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "testdata not found: "+name+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(b)
}

func newNetworkIdleTestClient(t *testing.T, maxAge time.Duration) *client {
	t.Helper()
	cfg := model.Default()
	cfg.Plugins.Fetcher = model.FetcherChromium
	cfg.Plugins.FetcherConfig.WaitUntil = model.WaitUntilNetworkIdle
	cfg.Plugins.FetcherConfig.WaitTimeout = 30 * time.Second
	cfg.Plugins.FetcherConfig.NetworkIdleDuration = 200 * time.Millisecond
	cfg.Plugins.FetcherConfig.NetworkIdleRequestMaxAge = maxAge

	host := core.NewHost(&cfg)
	c := &client{}
	require.NoError(t, c.Init(context.Background(), host))
	return c
}

// TestClient_Get_NetworkIdle_simple は単純 HTML を network_idle で取得できることを検証する。
func TestClient_Get_NetworkIdle_simple(t *testing.T) {
	if _, err := resolveBrowserPath(""); err != nil {
		t.Skip("chromium browser not available: " + err.Error())
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body><h1 id="done">done</h1></body></html>`))
	}))
	t.Cleanup(srv.Close)

	c := newNetworkIdleTestClient(t, 10*time.Second)
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	u, err := url.Parse(srv.URL + "/")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	res, err := c.Get(ctx, u, nil)
	require.NoError(t, err)
	assert.Contains(t, string(res.Body), "done")
}

// TestClient_Get_NetworkIdle_iframeHang は iframe 内のハング通信を無視して取得できることを検証する。
func TestClient_Get_NetworkIdle_iframeHang(t *testing.T) {
	if _, err := resolveBrowserPath(""); err != nil {
		t.Skip("chromium browser not available: " + err.Error())
	}

	srv := newNetworkIdleTestServer(t)
	t.Cleanup(srv.Close)

	c := newNetworkIdleTestClient(t, 10*time.Second)
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	u, err := url.Parse(srv.URL + "/network_idle_iframe_hang.html")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	res, err := c.Get(ctx, u, nil)
	require.NoError(t, err)
	require.NotNil(t, res)

	body := string(res.Body)
	assert.Contains(t, body, "done")
}

// TestClient_Get_NetworkIdle_mainHangMaxAge はメインのハング通信があっても HTML を取得できることを検証する。
// max_age による打ち切りはユニットテストで検証する（fetch は Navigate 中に始まるため結合では経過時間を断定しない）。
func TestClient_Get_NetworkIdle_mainHangMaxAge(t *testing.T) {
	if _, err := resolveBrowserPath(""); err != nil {
		t.Skip("chromium browser not available: " + err.Error())
	}

	srv := newNetworkIdleTestServer(t)
	t.Cleanup(srv.Close)

	c := newNetworkIdleTestClient(t, 2*time.Second)
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	u, err := url.Parse(srv.URL + "/network_idle_main_hang.html")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	res, err := c.Get(ctx, u, nil)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Contains(t, string(res.Body), "done")
}
func TestClient_Get_NetworkIdle_testServerRoutes(t *testing.T) {
	srv := newNetworkIdleTestServer(t)
	t.Cleanup(srv.Close)

	resp, err := http.Get(srv.URL + "/network_idle_iframe_hang.html")
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, strings.Contains(resp.Header.Get("Content-Type"), "text/html"))
}
