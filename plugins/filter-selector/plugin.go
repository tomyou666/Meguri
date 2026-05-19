// Package selector は content.selector を適用して HTML を絞り込む P7 Filter を提供する。
package selector

import (
	"context"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	pluginpkg "scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterFilter("selector", func() pluginpkg.Filter { return &filter{} })
}

type filter struct {
	host pluginpkg.Host
}

func (f *filter) Metadata() pluginpkg.Metadata {
	return pluginpkg.Metadata{
		Name:        "selector",
		Version:     "0.1.0",
		Kind:        pluginpkg.KindFilter,
		Description: "content.selector で指定された範囲だけを残す",
	}
}

func (f *filter) Init(_ context.Context, host pluginpkg.Host) error {
	f.host = host
	return nil
}

func (f *filter) Close(_ context.Context) error { return nil }

func (f *filter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	sel, ok := f.host.Config("content.selector")
	if !ok || sel == "" {
		return c, nil
	}
	if c.Format != "html" {
		return c, nil
	}
	doc, ok := c.DOM.(*goquery.Document)
	if !ok {
		return c, nil
	}

	sub := doc.Find(sel)
	if sub.Length() == 0 {
		return c, nil
	}

	// 選択範囲だけを新しい root にする。
	root := &html.Node{Type: html.ElementNode, Data: "div"}
	sub.Each(func(_ int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			root.AppendChild(cloneNode(n))
		}
	})
	newDoc := goquery.NewDocumentFromNode(root)
	c.DOM = newDoc
	c.Text = newDoc.Text()

	return c, nil
}

func cloneNode(n *html.Node) *html.Node {
	cp := &html.Node{
		Type:      n.Type,
		DataAtom:  n.DataAtom,
		Data:      n.Data,
		Namespace: n.Namespace,
	}
	cp.Attr = append(cp.Attr, n.Attr...)
	for child := n.FirstChild; child != nil; child = child.NextSibling {
		cp.AppendChild(cloneNode(child))
	}
	return cp
}
