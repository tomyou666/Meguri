// Package includetags は content.include_tags に指定された HTML タグだけを残す P7 Filter を提供する。
package includetags

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	pluginpkg "meguri/internal/domain/plugin"
)

func init() {
	core.RegisterFilter("include_tags", func() pluginpkg.Filter { return &filter{} })
}

// filter はタグ名許可用 P7 Filter の実装。
type filter struct {
	// host は Init で受け取る Host。
	host pluginpkg.Host
}

// Metadata は plugin.Filter.Metadata の実装。
func (f *filter) Metadata() pluginpkg.Metadata {
	return pluginpkg.Metadata{
		Name:        "include_tags",
		Version:     "0.1.0",
		Kind:        pluginpkg.KindFilter,
		Description: "content.include_tags に指定された HTML タグだけを残す",
	}
}

// Init は plugin.Plugin.Init の実装。
func (f *filter) Init(_ context.Context, host pluginpkg.Host) error {
	f.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (f *filter) Close(_ context.Context) error { return nil }

// Filter は content.include_tags に一致するタグの要素だけを残す。
// 一致ゼロの場合は空の DOM にする。
func (f *filter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	if f.host == nil {
		return c, nil
	}
	tags := f.host.ContentConfig().IncludeTags
	if len(tags) == 0 {
		return c, nil
	}
	if c.Format != "html" {
		return c, nil
	}
	doc, ok := c.DOM.(*goquery.Document)
	if !ok {
		return c, nil
	}

	var parts []string
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if !isTagName(tag) {
			continue
		}
		parts = append(parts, tag)
	}
	if len(parts) == 0 {
		return c, nil
	}

	sel := strings.Join(parts, ",")
	sub := doc.Find(sel)

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

// isTagName は s が HTML タグ名として扱える形（英字始まりの英数字とハイフン）かを返す。
// CSS セレクタ片が紛れ込んで意図しない要素を残さないよう、タグ名以外は弾く。
func isTagName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case i > 0 && (r >= '0' && r <= '9' || r == '-'):
		default:
			return false
		}
	}
	return true
}

// cloneNode は n とその子孫を深いコピーする。
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
