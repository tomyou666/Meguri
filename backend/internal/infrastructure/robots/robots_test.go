package robots

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// countingFetcher は Get 呼び出し回数と URL を記録するテスト用 Fetcher。
type countingFetcher struct {
	calls atomic.Int64
	delay time.Duration
	mu    sync.Mutex
	urls  []string
	body  []byte
}

func (f *countingFetcher) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: "counting", Kind: plugin.KindFetcher}
}
func (f *countingFetcher) Init(context.Context, plugin.Host) error { return nil }
func (f *countingFetcher) Close(context.Context) error             { return nil }

func (f *countingFetcher) Get(_ context.Context, u *url.URL, _ map[string]string) (*model.Response, error) {
	f.calls.Add(1)
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	f.mu.Lock()
	f.urls = append(f.urls, u.String())
	f.mu.Unlock()
	body := f.body
	if body == nil {
		body = []byte("User-agent: *\nAllow: /\n")
	}
	return &model.Response{StatusCode: 200, Body: body, URL: u}, nil
}

// errFetcher は常に指定エラーを返すテスト用 Fetcher。
type errFetcher struct {
	err   error
	calls atomic.Int64
}

func (f *errFetcher) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: "err", Kind: plugin.KindFetcher}
}
func (f *errFetcher) Init(context.Context, plugin.Host) error { return nil }
func (f *errFetcher) Close(context.Context) error             { return nil }

func (f *errFetcher) Get(context.Context, *url.URL, map[string]string) (*model.Response, error) {
	f.calls.Add(1)
	return nil, f.err
}

// TestCache はホスト単位キャッシュ・singleflight・http/https 共有を検証する。
func TestCache(t *testing.T) {
	t.Run("正常系: 同一ホスト並行 Allowed で fetch が 1 回にまとまる", func(t *testing.T) {
		// delay で単発 Get を引き延ばし、複数 goroutine がミスを共有する前提。
		f := &countingFetcher{delay: 50 * time.Millisecond}
		c := NewCache(f)
		u, err := url.Parse("https://example.com/page")
		require.NoError(t, err)

		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				assert.True(t, c.Allowed(context.Background(), u, "TestBot"))
			}()
		}
		wg.Wait()
		assert.Equal(t, int64(1), f.calls.Load())
	})

	t.Run("正常系: http と https でキャッシュを共有する", func(t *testing.T) {
		f := &countingFetcher{}
		c := NewCache(f)
		httpURL, err := url.Parse("http://example.com/a")
		require.NoError(t, err)
		httpsURL, err := url.Parse("https://example.com/b")
		require.NoError(t, err)

		assert.True(t, c.Allowed(context.Background(), httpURL, "TestBot"))
		assert.True(t, c.Allowed(context.Background(), httpsURL, "TestBot"))
		assert.Equal(t, int64(1), f.calls.Load(), "host キー共有で 2 回目は fetch しない")
		f.mu.Lock()
		require.Len(t, f.urls, 1)
		assert.Equal(t, "http://example.com/robots.txt", f.urls[0], "先勝ち scheme で取得")
		f.mu.Unlock()
	})

	t.Run("正常系: 異なるホストは別 fetch", func(t *testing.T) {
		f := &countingFetcher{}
		c := NewCache(f)
		a, err := url.Parse("https://a.example.com/")
		require.NoError(t, err)
		b, err := url.Parse("https://b.example.com/")
		require.NoError(t, err)

		assert.True(t, c.Allowed(context.Background(), a, "TestBot"))
		assert.True(t, c.Allowed(context.Background(), b, "TestBot"))
		assert.Equal(t, int64(2), f.calls.Load())
	})

	t.Run("正常系: Disallow パスは Allowed が false", func(t *testing.T) {
		f := &countingFetcher{body: []byte("User-agent: *\nDisallow: /private\n")}
		c := NewCache(f)
		denied, err := url.Parse("https://example.com/private")
		require.NoError(t, err)
		allowed, err := url.Parse("https://example.com/public")
		require.NoError(t, err)

		assert.False(t, c.Allowed(context.Background(), denied, "TestBot"))
		assert.True(t, c.Allowed(context.Background(), allowed, "TestBot"))
		assert.Equal(t, int64(1), f.calls.Load())
	})

	t.Run("正常系: context.Canceled ではキャッシュせず再取得する", func(t *testing.T) {
		f := &errFetcher{err: context.Canceled}
		c := NewCache(f)
		u, err := url.Parse("https://example.com/page")
		require.NoError(t, err)

		assert.True(t, c.Allowed(context.Background(), u, "TestBot"), "キャンセル時も許可扱い")
		assert.True(t, c.Allowed(context.Background(), u, "TestBot"))
		assert.Equal(t, int64(2), f.calls.Load(), "Canceled は hosts に書かない")
	})

	t.Run("正常系: 通常の fetch 失敗は許可扱いでキャッシュする", func(t *testing.T) {
		f := &errFetcher{err: assert.AnError}
		c := NewCache(f)
		u, err := url.Parse("https://example.com/page")
		require.NoError(t, err)

		assert.True(t, c.Allowed(context.Background(), u, "TestBot"))
		assert.True(t, c.Allowed(context.Background(), u, "TestBot"))
		assert.Equal(t, int64(1), f.calls.Load())
	})
}
