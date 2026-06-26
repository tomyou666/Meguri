// Package markdown は HTML Content を Markdown に変換する P6 Transformer を提供する。
package markdown

import (
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	md "github.com/JohannesKaufmann/html-to-markdown"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterTransformer("markdown", func() plugin.Transformer { return &transformer{} })
}

// transformer は Markdown 変換用 P6 Transformer の実装。
type transformer struct {
	// host は Init で受け取る Host。
	host plugin.Host
	// conv は html-to-markdown コンバータ。
	conv *md.Converter
}

// Metadata は plugin.Transformer.Metadata の実装。
func (t *transformer) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "markdown",
		Version:     "0.1.0",
		Kind:        plugin.KindTransformer,
		Description: "HTML コンテンツを Markdown に変換する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (t *transformer) Init(_ context.Context, host plugin.Host) error {
	t.host = host
	t.conv = md.NewConverter("", true, nil)
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (t *transformer) Close(_ context.Context) error { return nil }

// Transform は Content を model.Result（Markdown 等）に変換する。
func (t *transformer) Transform(_ context.Context, c *model.Content) (*model.Result, error) {
	r := &model.Result{
		URL:      c.URL,
		Metadata: c.Metadata,
	}

	switch c.Format {
	case "html":
		doc, ok := c.DOM.(*goquery.Document)
		var htmlStr string
		if ok {
			h, err := doc.Html()
			if err != nil {
				return nil, fmt.Errorf("html serialize: %w", err)
			}
			htmlStr = h
		} else {
			htmlStr = c.Text
		}

		r.HTML = htmlStr
		mdStr, err := t.conv.ConvertString(htmlStr)
		if err != nil {
			return nil, fmt.Errorf("html->markdown: %w", err)
		}
		r.Markdown = strings.TrimSpace(mdStr)
	case "pdf":
		r.Markdown = c.Text
		r.HTML = "<pre>" + c.Text + "</pre>"
	default:
		r.Markdown = c.Text
	}

	return r, nil
}
