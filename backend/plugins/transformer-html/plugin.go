// Package htmlfmt は HTML Content を整形 HTML に変換する P6 Transformer を提供する。
package htmlfmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/yosssi/gohtml"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterTransformer("html", func() plugin.Transformer { return &transformer{} })
}

// transformer は HTML 整形用 P6 Transformer の実装。
type transformer struct {
	// host は Init で受け取る Host。
	host plugin.Host
}

// Metadata は plugin.Transformer.Metadata の実装。
func (t *transformer) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "html",
		Version:     "0.1.0",
		Kind:        plugin.KindTransformer,
		Description: "HTML コンテンツを整形 HTML に変換する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (t *transformer) Init(_ context.Context, host plugin.Host) error {
	t.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (t *transformer) Close(_ context.Context) error { return nil }

// Transform は Content を model.Result（整形 HTML）に変換する。
func (t *transformer) Transform(_ context.Context, c *model.Content) (*model.Result, error) {
	r := &model.Result{
		URL:      c.URL,
		Metadata: c.Metadata,
	}

	switch c.Format {
	case "html":
		htmlStr, err := serializeHTML(c)
		if err != nil {
			return nil, err
		}
		r.HTML = strings.TrimSpace(gohtml.Format(htmlStr))
	case "pdf":
		r.HTML = strings.TrimSpace(gohtml.Format("<section><pre>" + c.Text + "</pre></section>"))
	default:
		r.HTML = ""
	}

	return r, nil
}

// serializeHTML は Content から HTML 文字列を取り出す。
func serializeHTML(c *model.Content) (string, error) {
	doc, ok := c.DOM.(*goquery.Document)
	if ok {
		h, err := doc.Html()
		if err != nil {
			return "", fmt.Errorf("html serialize: %w", err)
		}
		return h, nil
	}
	return c.Text, nil
}
