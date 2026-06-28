package core

import (
	"context"
	"fmt"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// initFetcherWithRouting は primary fetcher を limiter / PDF ルーティング付きで k.fetcher に設定する。
func (k *Kernel) initFetcherWithRouting(
	ctx context.Context,
	primary plugin.Fetcher,
	kind model.FetcherKind,
	rollback func(error) error,
) error {
	inner := plugin.Fetcher(primary)
	if kind == model.FetcherChromium {
		httpF, err := k.reg.NewFetcher(string(model.FetcherHTTP))
		if err != nil {
			return rollback(err)
		}
		if err := httpF.Init(ctx, k.host); err != nil {
			return rollback(fmt.Errorf("init fetcher http: %w", err))
		}
		k.initialized = append(k.initialized, httpF)
		inner = newRoutingFetcher(primary, httpF, kind, k.fetchLimiter, k.cfg)
	}

	if k.fetchLimiter != nil && kind != model.FetcherChromium {
		inner = newLimitingFetcher(inner, k.fetchLimiter, kind)
	}
	k.fetcher = inner
	return nil
}
