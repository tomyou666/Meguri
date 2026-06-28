package core

import (
	"context"
	"net/url"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// recordingFetcher は Get 呼び出し回数を記録するテスト用 fetcher。
type recordingFetcher struct {
	// name は Metadata 名。
	name string
	// calls は Get 呼び出し回数。
	calls atomic.Int32
}

func (f *recordingFetcher) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindFetcher}
}

func (f *recordingFetcher) Init(context.Context, plugin.Host) error { return nil }

func (f *recordingFetcher) Close(context.Context) error { return nil }

func (f *recordingFetcher) Get(_ context.Context, u *url.URL, _ map[string]string) (*model.Response, error) {
	f.calls.Add(1)
	return &model.Response{
		URL:         u,
		StatusCode:  200,
		ContentType: "application/octet-stream",
		Body:        []byte("ok"),
	}, nil
}

// TestRoutingFetcher は PDF URL の HTTP フォールバックルーティングを検証する。
func TestRoutingFetcher(t *testing.T) {
	t.Run("正常系: PDF URL は HTTP fallback が呼ばれる", func(t *testing.T) {
		chromium := &recordingFetcher{name: "chromium"}
		httpF := &recordingFetcher{name: "http"}
		cfg := model.Default()
		cfg.PDF.Enabled = true

		rf := newRoutingFetcher(chromium, httpF, model.FetcherChromium, nil, &cfg)
		u, err := url.Parse("https://example.com/files/report.pdf")
		require.NoError(t, err)

		_, err = rf.Get(context.Background(), u, nil)
		require.NoError(t, err)
		assert.Equal(t, int32(1), httpF.calls.Load())
		assert.Equal(t, int32(0), chromium.calls.Load())
	})

	t.Run("正常系: HTML URL は primary chromium が呼ばれる", func(t *testing.T) {
		chromium := &recordingFetcher{name: "chromium"}
		httpF := &recordingFetcher{name: "http"}
		cfg := model.Default()
		cfg.PDF.Enabled = true

		rf := newRoutingFetcher(chromium, httpF, model.FetcherChromium, nil, &cfg)
		u, err := url.Parse("https://example.com/page.html")
		require.NoError(t, err)

		_, err = rf.Get(context.Background(), u, nil)
		require.NoError(t, err)
		assert.Equal(t, int32(0), httpF.calls.Load())
		assert.Equal(t, int32(1), chromium.calls.Load())
	})

	t.Run("正常系: pdf.enabled=false なら PDF も primary へ", func(t *testing.T) {
		chromium := &recordingFetcher{name: "chromium"}
		httpF := &recordingFetcher{name: "http"}
		cfg := model.Default()
		cfg.PDF.Enabled = false

		rf := newRoutingFetcher(chromium, httpF, model.FetcherChromium, nil, &cfg)
		u, err := url.Parse("https://example.com/files/report.pdf")
		require.NoError(t, err)

		_, err = rf.Get(context.Background(), u, nil)
		require.NoError(t, err)
		assert.Equal(t, int32(0), httpF.calls.Load())
		assert.Equal(t, int32(1), chromium.calls.Load())
	})
}
