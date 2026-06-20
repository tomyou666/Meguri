// Package rawhtml は生 HTML を Result に出力する P6 Transformer を提供する。
package rawhtml

import (
	"context"
	"log/slog"
	"net/url"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterTransformer("raw_html", func() plugin.Transformer { return &transformer{} })
}

// transformer は生 HTML 出力用 P6 Transformer の実装。
type transformer struct {
	// host は Init で受け取る Host。
	host plugin.Host
}

// Metadata は plugin.Transformer.Metadata の実装。
func (t *transformer) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "raw_html",
		Version:     "0.1.0",
		Kind:        plugin.KindTransformer,
		Description: "HTTP レスポンス本文の生 HTML を出力する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (t *transformer) Init(_ context.Context, host plugin.Host) error {
	t.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (t *transformer) Close(_ context.Context) error { return nil }

// Transform は Content の RawHTML を model.Result にコピーする。
func (t *transformer) Transform(_ context.Context, c *model.Content) (*model.Result, error) {
	r := &model.Result{
		URL:      c.URL,
		Metadata: c.Metadata,
	}

	switch c.Format {
	case "html":
		r.RawHTML = c.RawHTML
	case "pdf":
		slog.Warn("raw_html is not supported for PDF", "url", urlString(c.URL))
		r.RawHTML = ""
	default:
		r.RawHTML = ""
	}

	return r, nil
}

// urlString は nil 安全に URL 文字列を返す。
func urlString(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
