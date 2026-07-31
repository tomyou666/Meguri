// Package excludetags は content.exclude_tags に指定された HTML タグを DOM から除去する P7 Filter を提供する。
package excludetags

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	pluginpkg "meguri/internal/domain/plugin"
)

func init() {
	core.RegisterFilter("exclude_tags", func() pluginpkg.Filter { return &filter{} })
}

// filter はタグ名除外用 P7 Filter の実装。
type filter struct {
	// host は Init で受け取る Host。
	host pluginpkg.Host
}

// Metadata は plugin.Filter.Metadata の実装。
func (f *filter) Metadata() pluginpkg.Metadata {
	return pluginpkg.Metadata{
		Name:        "exclude_tags",
		Version:     "0.1.0",
		Kind:        pluginpkg.KindFilter,
		Description: "content.exclude_tags に指定された HTML タグを DOM から除去する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (f *filter) Init(_ context.Context, host pluginpkg.Host) error {
	f.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (f *filter) Close(_ context.Context) error { return nil }

// Filter は content.exclude_tags に一致するタグの要素を DOM から除去する。
func (f *filter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	if f.host == nil {
		return c, nil
	}
	tags := f.host.ContentConfig().ExcludeTags
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

	removed := false
	for _, raw := range tags {
		tag := strings.ToLower(strings.TrimSpace(raw))
		if !isTagName(tag) {
			continue
		}
		matched := doc.Find(tag)
		if matched.Length() == 0 {
			continue
		}
		matched.Remove()
		removed = true
	}
	if removed {
		c.Text = doc.Text()
	}
	return c, nil
}

// isTagName は s が HTML タグ名として扱える形（英字始まりの英数字とハイフン）かを返す。
// CSS セレクタ片が紛れ込んで意図しない要素を除去しないよう、タグ名以外は弾く。
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
