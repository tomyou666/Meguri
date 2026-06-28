package core

import (
	"context"
	"net/url"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// routingFetcher は primary fetcher に加え PDF URL 向け HTTP フォールバックを提供する。
type routingFetcher struct {
	// primary は設定上のメイン fetcher（通常 chromium）。
	primary plugin.Fetcher
	// httpFallback は PDF バイナリ取得用 HTTP fetcher。
	httpFallback plugin.Fetcher
	// primaryKind は primary 取得時の limiter 種別。
	primaryKind model.FetcherKind
	// lim は取得並列上限（nil 可）。
	lim *fetchlimit.FetchLimiter
	// cfg は PDF 有効化などルーティング判定に使う。
	cfg *model.Config
}

// newRoutingFetcher は chromium 等 primary + HTTP fallback を束ねる fetcher を構築する。
func newRoutingFetcher(
	primary plugin.Fetcher,
	httpFallback plugin.Fetcher,
	primaryKind model.FetcherKind,
	lim *fetchlimit.FetchLimiter,
	cfg *model.Config,
) plugin.Fetcher {
	if primaryKind == "" {
		primaryKind = model.FetcherChromium
	}
	return &routingFetcher{
		primary:      primary,
		httpFallback: httpFallback,
		primaryKind:  primaryKind,
		lim:          lim,
		cfg:          cfg,
	}
}

// Metadata は plugin.Plugin.Metadata の実装。
func (f *routingFetcher) Metadata() plugin.Metadata {
	return f.primary.Metadata()
}

// Init は plugin.Plugin.Init の実装（inner は Kernel 側で Init 済みのため no-op）。
func (f *routingFetcher) Init(_ context.Context, _ plugin.Host) error {
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (f *routingFetcher) Close(ctx context.Context) error {
	return nil
}

// Get は URL 種別に応じて primary または HTTP fallback へ委譲する。
func (f *routingFetcher) Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	if f.shouldRoutePDFToHTTP(u) {
		return f.getWithLimit(ctx, f.httpFallback, model.FetcherHTTP, u, headers)
	}
	return f.getWithLimit(ctx, f.primary, f.primaryKind, u, headers)
}

// shouldRoutePDFToHTTP は PDF URL を HTTP fallback へ送るべきかを返す。
func (f *routingFetcher) shouldRoutePDFToHTTP(u *url.URL) bool {
	return f.httpFallback != nil && IsPDFTarget(u, f.cfg)
}

// getWithLimit は limiter スロットを確保して inner.Get を呼ぶ。
func (f *routingFetcher) getWithLimit(
	ctx context.Context,
	inner plugin.Fetcher,
	kind model.FetcherKind,
	u *url.URL,
	headers map[string]string,
) (*model.Response, error) {
	if f.lim != nil {
		if err := f.lim.Acquire(ctx, kind); err != nil {
			return nil, err
		}
		defer f.lim.Release(kind)
	}
	return inner.Get(ctx, u, headers)
}
