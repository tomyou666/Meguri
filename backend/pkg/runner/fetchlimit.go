package runner

import (
	"context"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// PrepareFetchLimiter は設定から FetchLimiter を構築し、必要ならキャリブレーションと動的監視を開始する。
func PrepareFetchLimiter(ctx context.Context, cfg *model.Config, opts *RunOptions) *fetchlimit.FetchLimiter {
	return usecase.PrepareFetchLimiter(ctx, cfg, opts)
}
