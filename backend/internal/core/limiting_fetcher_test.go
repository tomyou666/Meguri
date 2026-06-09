package core_test

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/core"
	"scraperbot/internal/core/fetchlimit"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

type slowFetcher struct {
	delay    time.Duration
	inflight atomic.Int32
	maxSeen  atomic.Int32
}

func (f *slowFetcher) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: "slow", Kind: plugin.KindFetcher}
}

func (f *slowFetcher) Init(context.Context, plugin.Host) error { return nil }

func (f *slowFetcher) Close(context.Context) error { return nil }

func (f *slowFetcher) Get(ctx context.Context, u *url.URL, _ map[string]string) (*model.Response, error) {
	cur := f.inflight.Add(1)
	for {
		prev := f.maxSeen.Load()
		if cur <= prev || f.maxSeen.CompareAndSwap(prev, cur) {
			break
		}
	}
	defer f.inflight.Add(-1)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(f.delay):
	}
	return &model.Response{URL: u, StatusCode: 200, Body: []byte("ok")}, nil
}

// TestLimitingFetcher は同時 Get 数が上限を超えないことを検証する。
func TestLimitingFetcher(t *testing.T) {
	inner := &slowFetcher{delay: 80 * time.Millisecond}
	core.RegisterFetcher("slow", func() plugin.Fetcher { return inner })

	lim := fetchlimit.NewFromConfig(model.FetchLimitsConfig{HTTPMaxInflight: 2})

	cfg := model.Default()
	cfg.Plugins.Fetcher = model.FetcherKind("slow")
	host := core.NewHost(&cfg)
	k := core.NewKernel(&cfg, host, core.Default())
	k.SetFetchLimiter(lim)
	require.NoError(t, k.Init(context.Background()))
	defer func() { _ = k.Close(context.Background()) }()

	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := url.Parse("https://example.com")
			require.NoError(t, err)
			_, err = k.Fetcher().Get(context.Background(), u, nil)
			require.NoError(t, err)
		}()
	}
	wg.Wait()
	assert.LessOrEqual(t, int(inner.maxSeen.Load()), 2)
}
