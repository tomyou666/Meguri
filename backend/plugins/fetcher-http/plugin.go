// Package httpfetch は標準 HTTP による P3 Fetcher を提供する。
package httpfetch

import (
	"context"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterFetcher(string(model.FetcherHTTP), func() plugin.Fetcher { return &client{} })
}

// client は net/http ベースの P3 Fetcher 実装。
type client struct {
	// reqCfg はタイムアウト・リトライ設定（Init で設定）。
	reqCfg model.RequestConfig
}

// Metadata は plugin.Plugin.Metadata の実装。
func (c *client) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        string(model.FetcherHTTP),
		Version:     "0.1.0",
		Kind:        plugin.KindFetcher,
		Description: "標準 HTTP クライアントによる URL 取得",
	}
}

// Init は plugin.Plugin.Init の実装。
func (c *client) Init(_ context.Context, host plugin.Host) error {
	c.reqCfg = host.RequestConfig()
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (c *client) Close(_ context.Context) error { return nil }

// Get は plugin.Fetcher.Get の実装。
func (c *client) Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	return c.get(ctx, u, headers)
}
