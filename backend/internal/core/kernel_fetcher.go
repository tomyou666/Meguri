package core

import (
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// initFetcherWithLimiter は primary fetcher を取得並列上限付きで k.fetcher に設定する。
func (k *Kernel) initFetcherWithLimiter(primary plugin.Fetcher, kind model.FetcherKind) {
	inner := plugin.Fetcher(primary)
	if k.fetchLimiter != nil {
		inner = newLimitingFetcher(inner, k.fetchLimiter, kind)
	}
	k.fetcher = inner
}
