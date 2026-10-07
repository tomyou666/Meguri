package runner

import (
	"context"

	"meguri/internal/usecase"
)

// RobotsAllowance はジョブ単位の robots.txt 許可判定（usecase.RobotsAllowance のエイリアス）。
type RobotsAllowance = usecase.RobotsAllowance

// NewRobotsAllowance は run 設定とジョブ FetchLimiter から許可判定を構築する。
func NewRobotsAllowance(ctx context.Context, cfg *Config, opts *RunOptions) (*RobotsAllowance, error) {
	return usecase.NewRobotsAllowance(ctx, cfg, opts)
}
