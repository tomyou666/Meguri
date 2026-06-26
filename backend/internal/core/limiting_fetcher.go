package core

import (
	"context"
	"net/url"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// limitingFetcher は inner Fetcher の Get 前後で取得並列上限を適用するデコレータ。
type limitingFetcher struct {
	inner plugin.Fetcher
	lim   *fetchlimit.FetchLimiter
	kind  model.FetcherKind
}

// newLimitingFetcher は limitingFetcher を構築する。
func newLimitingFetcher(inner plugin.Fetcher, lim *fetchlimit.FetchLimiter, kind model.FetcherKind) plugin.Fetcher {
	if kind == "" {
		kind = model.FetcherHTTP
	}
	return &limitingFetcher{inner: inner, lim: lim, kind: kind}
}

// Metadata は plugin.Plugin.Metadata の実装。
func (f *limitingFetcher) Metadata() plugin.Metadata {
	return f.inner.Metadata()
}

// Init は plugin.Plugin.Init の実装（inner は既に Init 済みのため no-op）。
func (f *limitingFetcher) Init(_ context.Context, _ plugin.Host) error {
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (f *limitingFetcher) Close(ctx context.Context) error {
	return f.inner.Close(ctx)
}

// Get は取得スロットを確保してから inner.Fetcher.Get を呼ぶ。
func (f *limitingFetcher) Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	if err := f.lim.Acquire(ctx, f.kind); err != nil {
		return nil, err
	}
	defer f.lim.Release(f.kind)
	return f.inner.Get(ctx, u, headers)
}
