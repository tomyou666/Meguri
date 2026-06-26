// Package chromiumfetch は chromedp による P3 Fetcher を提供する。
package chromiumfetch

import (
	"context"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterFetcher(string(model.FetcherChromium), func() plugin.Fetcher { return &client{} })
}

// client は chromedp ベースの P3 Fetcher 実装。
type client struct {
	// reqCfg はタイムアウト・リトライ設定。
	reqCfg model.RequestConfig
	// fetcherCfg はブラウザ実行・待機に関する設定。
	fetcherCfg model.FetcherConfig
	// browserPath は解決済みブラウザ実行ファイルパス。
	browserPath string
}

// Metadata は plugin.Plugin.Metadata の実装。
func (c *client) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        string(model.FetcherChromium),
		Version:     "0.1.0",
		Kind:        plugin.KindFetcher,
		Description: "chromedp によるヘッドレスブラウザ URL 取得",
	}
}

// Init は plugin.Plugin.Init の実装。
func (c *client) Init(_ context.Context, host plugin.Host) error {
	c.reqCfg = host.RequestConfig()
	c.fetcherCfg = host.FetcherConfig()
	path, err := resolveBrowserPath(c.fetcherCfg.BrowserPath)
	if err != nil {
		return err
	}
	c.browserPath = path
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (c *client) Close(_ context.Context) error { return nil }

// Get は plugin.Fetcher.Get の実装。
func (c *client) Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	return c.get(ctx, u, headers)
}
