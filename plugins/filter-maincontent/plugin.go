// Package maincontent はヘッダ・フッタ・ナビ・script/style/noscript などのノイズ要素を
// HTML から除去する P7 Filter を提供する。
package maincontent

import (
	"context"

	"github.com/PuerkitoBio/goquery"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterFilter("maincontent", func() plugin.Filter { return &filter{} })
}

type filter struct {
	host plugin.Host
}

func (f *filter) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "maincontent",
		Version:     "0.1.0",
		Kind:        plugin.KindFilter,
		Description: "ヘッダー・フッター・ナビ・script/style/noscript を除去する",
	}
}

func (f *filter) Init(_ context.Context, host plugin.Host) error {
	f.host = host
	return nil
}

func (f *filter) Close(_ context.Context) error { return nil }

func (f *filter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	if c.Format != "html" {
		return c, nil
	}
	doc, ok := c.DOM.(*goquery.Document)
	if !ok {
		return c, nil
	}

	doc.Find("header, footer, nav, aside, script, style, noscript").Remove()

	if main := doc.Find("main, article").First(); main.Length() > 0 {
		c.Text = main.Text()
	}
	return c, nil
}
