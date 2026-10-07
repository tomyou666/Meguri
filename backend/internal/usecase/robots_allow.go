package usecase

import (
	"context"
	"fmt"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/infrastructure/robots"
)

// RobotsAllowance はジョブ単位の robots.txt 許可判定（BFS の robots.Cache と同 semantics）。
type RobotsAllowance struct {
	cfg   *model.Config
	cache *robots.Cache
	k     *core.Kernel
}

// NewRobotsAllowance は run 設定とジョブ FetchLimiter から許可判定を構築する。
// cfg が nil、または respect_robots_txt が false のときは nil を返す（Kernel は作らない）。
func NewRobotsAllowance(ctx context.Context, cfg *model.Config, opts *RunOptions) (*RobotsAllowance, error) {
	if cfg == nil || !cfg.Crawl.RespectRobotsTxt {
		return nil, nil
	}
	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
	if opts != nil && opts.FetchLimiter != nil {
		k.SetFetchLimiter(opts.FetchLimiter)
	}
	if err := k.Init(ctx); err != nil {
		return nil, fmt.Errorf("robots allowance kernel init: %w", err)
	}
	return &RobotsAllowance{
		cfg:   cfg,
		cache: robots.NewCache(k.Fetcher()),
		k:     k,
	}, nil
}

// Allowed は URL が robots 上で許可されるかを返す。取得・パース失敗は許可扱い。
func (a *RobotsAllowance) Allowed(ctx context.Context, rawURL string) bool {
	if a == nil {
		return true
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	ua := a.cfg.Plugins.Stealth.HTTP.EffectiveUserAgent()
	return a.cache.Allowed(ctx, u, ua)
}

// Close は robots 判定用 Kernel を解放する。
func (a *RobotsAllowance) Close(ctx context.Context) {
	if a == nil || a.k == nil {
		return
	}
	_ = a.k.Close(ctx)
}
