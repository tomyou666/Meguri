package usecase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"

	_ "meguri/plugins/fetcher-http"
)

func testRobotsAllowConfig(t *testing.T, respect bool) *model.Config {
	t.Helper()
	layer := testRobotsConfigLayer(t)
	merged, err := usecase.MergeUIConfigLayers(layer)
	require.NoError(t, err)
	cfg, err := usecase.ParseUIConfig(merged)
	require.NoError(t, err)
	cfg.Crawl.RespectRobotsTxt = respect
	return cfg
}

// TestRobotsAllowance は mode 3/4 向けジョブ単位 Allowed 判定を検証する。
func TestRobotsAllowance(t *testing.T) {
	t.Run("正常系: respect_robots_txt が false なら nil", func(t *testing.T) {
		cfg := testRobotsAllowConfig(t, false)
		allow, err := usecase.NewRobotsAllowance(context.Background(), cfg, nil)
		require.NoError(t, err)
		assert.Nil(t, allow)
	})

	t.Run("正常系: Disallow は false", func(t *testing.T) {
		var robotsGets atomic.Int64
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				robotsGets.Add(1)
				_, _ = w.Write([]byte("User-agent: *\nDisallow: /private\n"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		cfg := testRobotsAllowConfig(t, true)
		allow, err := usecase.NewRobotsAllowance(context.Background(), cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, allow)
		defer allow.Close(context.Background())

		assert.False(t, allow.Allowed(context.Background(), srv.URL+"/private/page"))
		assert.True(t, allow.Allowed(context.Background(), srv.URL+"/public/page"))
		assert.Equal(t, int64(1), robotsGets.Load(), "同一ホストは robots.txt を 1 回だけ取得")
	})

	t.Run("正常系: robots 5xx は許可", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		cfg := testRobotsAllowConfig(t, true)
		allow, err := usecase.NewRobotsAllowance(context.Background(), cfg, nil)
		require.NoError(t, err)
		require.NotNil(t, allow)
		defer allow.Close(context.Background())

		assert.True(t, allow.Allowed(context.Background(), srv.URL+"/any"))
	})

	t.Run("正常系: ジョブ FetchLimiter を Kernel に載せる", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/robots.txt" {
				_, _ = w.Write([]byte("User-agent: *\nAllow: /\n"))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		cfg := testRobotsAllowConfig(t, true)
		cfg.Crawl.FetchLimits.HTTPMaxInflight = 1
		lim := usecase.PrepareFetchLimiter(context.Background(), cfg, &usecase.RunOptions{})
		opts := &usecase.RunOptions{FetchLimiter: lim}
		defer lim.Close()

		allow, err := usecase.NewRobotsAllowance(context.Background(), cfg, opts)
		require.NoError(t, err)
		require.NotNil(t, allow)
		defer allow.Close(context.Background())

		assert.True(t, allow.Allowed(context.Background(), srv.URL+"/page"))
	})

	t.Run("正常系: cfg nil は nil", func(t *testing.T) {
		allow, err := usecase.NewRobotsAllowance(context.Background(), nil, nil)
		require.NoError(t, err)
		assert.Nil(t, allow)
	})
}
