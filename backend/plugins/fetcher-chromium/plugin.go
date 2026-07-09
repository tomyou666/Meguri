// Package chromiumfetch は chromedp による P3 Fetcher を提供する。
package chromiumfetch

import (
	"context"
	"net/url"
	"strconv"
	"sync"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// init は chromium Fetcher をコアへ登録する。
func init() {
	core.RegisterFetcher(string(model.FetcherChromium), func() plugin.Fetcher { return &client{} })
}

// client は chromedp ベースの P3 Fetcher 実装。
type client struct {
	// reqCfg はタイムアウト・リトライ設定。
	reqCfg model.RequestConfig
	// fetcherCfg はブラウザ実行・待機に関する設定。
	fetcherCfg model.FetcherConfig
	// stealthCfg は chromium ステルス設定。
	stealthCfg model.ChromiumStealthConfig
	// browserPath は解決済みブラウザ実行ファイルパス。
	browserPath string
	// pdfCfg は IsPDFTarget 判定用の PDF 設定スナップショット。
	pdfCfg *model.Config
	// poolMu は poolKey / poolJoined の読み書きを保護する。
	poolMu sync.Mutex
	// poolKey は参加中のブラウザセッションを識別するキー。
	poolKey sessionKey
	// poolJoined は defaultBrowserPool への参加済みかどうか。
	poolJoined bool
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
	c.stealthCfg = host.StealthConfig().Chromium
	path, err := resolveBrowserPath(c.fetcherCfg.BrowserPath)
	if err != nil {
		return err
	}
	c.browserPath = path
	c.pdfCfg = pdfConfigFromHost(host)
	return nil
}

// pdfConfigFromHost は Host から PDF 有効化設定を読み取る。
func pdfConfigFromHost(host plugin.Host) *model.Config {
	enabled := true
	if v, ok := host.Config("pdf.enabled"); ok {
		if parsed, err := strconv.ParseBool(v); err == nil {
			enabled = parsed
		}
	}
	return &model.Config{PDF: model.PDFConfig{Enabled: enabled}}
}

// Close は plugin.Plugin.Close の実装。
func (c *client) Close(_ context.Context) error {
	c.leavePool()
	return nil
}

// Get は plugin.Fetcher.Get の実装。
func (c *client) Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	return c.get(ctx, u, headers)
}
