// Package excludeselectors は content.exclude_selectors にマッチする要素を DOM から除去する P7 Filter を提供する。
package excludeselectors

import (
	"context"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	pluginpkg "meguri/internal/domain/plugin"
)

func init() {
	core.RegisterFilter("exclude_selectors", func() pluginpkg.Filter { return &filter{} })
}

// filter は CSS セレクタ除外用 P7 Filter の実装。
type filter struct {
	// host は Init で受け取る Host。
	host pluginpkg.Host
}

// Metadata は plugin.Filter.Metadata の実装。
func (f *filter) Metadata() pluginpkg.Metadata {
	return pluginpkg.Metadata{
		Name:        "exclude_selectors",
		Version:     "0.1.0",
		Kind:        pluginpkg.KindFilter,
		Description: "content.exclude_selectors にマッチする要素を DOM から除去する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (f *filter) Init(_ context.Context, host pluginpkg.Host) error {
	f.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (f *filter) Close(_ context.Context) error { return nil }

// Filter は content.exclude_selectors にマッチする要素を DOM から除去する。
func (f *filter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	if f.host == nil {
		return c, nil
	}
	sels := f.host.ContentConfig().ExcludeSelectors
	if len(sels) == 0 {
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
	for _, raw := range sels {
		sel := strings.TrimSpace(raw)
		if sel == "" {
			continue
		}
		matched := doc.Find(sel)
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
