package runner

import (
	"context"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

//go:generate go tool gowrap gen -p meguri/pkg/runner -i FetchLimiterPreparer -t templates/slog_debug.gotmpl -o fetch_limit_preparer_with_debug_log.go

// FetchLimiterPreparer は FetchLimiter 構築を抽象化する。
type FetchLimiterPreparer interface {
	// PrepareFetchLimiter は設定から FetchLimiter を構築し、必要ならキャリブレーションと動的監視を開始する。
	PrepareFetchLimiter(ctx context.Context, cfg *model.Config, opts *RunOptions) *fetchlimit.FetchLimiter
}

type fetchLimiterPreparerImpl struct{}

func (fetchLimiterPreparerImpl) PrepareFetchLimiter(ctx context.Context, cfg *model.Config, opts *RunOptions) *fetchlimit.FetchLimiter {
	return usecase.PrepareFetchLimiter(ctx, cfg, opts)
}

// PrepareFetchLimiter は設定から FetchLimiter を構築し、必要ならキャリブレーションと動的監視を開始する。
func PrepareFetchLimiter(ctx context.Context, cfg *model.Config, opts *RunOptions) *fetchlimit.FetchLimiter {
	return defaultFetchLimiterPreparer.PrepareFetchLimiter(ctx, cfg, opts)
}
