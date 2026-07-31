package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// TestFilterNames は content / only_main_content から Filter チェーンが固定順で組み立てられることを検証する。
func TestFilterNames(t *testing.T) {
	reg := NewRegistry()
	for _, name := range []string{
		selectorFilter,
		maincontentFilter,
		excludeSelectorsFilter,
		includeTagsFilter,
		excludeTagsFilter,
		"custom",
	} {
		n := name
		RegisterFilterTo(reg, n, func() plugin.Filter { return &stubFilter{name: n} })
	}

	t.Run("正常系: only_main_content と content 由来を固定順で補完する", func(t *testing.T) {
		cfg := model.Default()
		cfg.Content.OnlyMainContent = true
		cfg.Content.Selector = ""
		cfg.Content.ExcludeSelectors = []string{".ad"}
		cfg.Content.IncludeTags = []string{"article"}
		cfg.Content.ExcludeTags = []string{"script", "dialog"}
		cfg.Plugins.Filters = []string{maincontentFilter, "custom"}

		k := NewKernel(&cfg, nil, reg)
		got := k.filterNames()

		assert.Equal(t, []string{
			maincontentFilter,
			excludeSelectorsFilter,
			includeTagsFilter,
			excludeTagsFilter,
			"custom",
		}, got)
	})

	t.Run("正常系: selector 指定時は maincontent を入れない", func(t *testing.T) {
		cfg := model.Default()
		cfg.Content.OnlyMainContent = true
		cfg.Content.Selector = "main"
		cfg.Content.ExcludeTags = []string{"script"}
		cfg.Plugins.Filters = []string{maincontentFilter}

		k := NewKernel(&cfg, nil, reg)
		got := k.filterNames()

		assert.Equal(t, []string{selectorFilter, excludeTagsFilter}, got)
		assert.NotContains(t, got, maincontentFilter)
	})

	t.Run("正常系: only_main_content false なら plugins.filters の maincontent も外す", func(t *testing.T) {
		cfg := model.Default()
		cfg.Content.OnlyMainContent = false
		cfg.Content.ExcludeTags = nil
		cfg.Plugins.Filters = []string{maincontentFilter, "custom"}

		k := NewKernel(&cfg, nil, reg)
		got := k.filterNames()

		assert.Equal(t, []string{"custom"}, got)
	})

	t.Run("正常系: managed 名は plugins.filters の位置を無視して固定順に寄せる", func(t *testing.T) {
		cfg := model.Default()
		cfg.Content.OnlyMainContent = true
		cfg.Content.Selector = "article"
		cfg.Content.ExcludeSelectors = []string{"#x"}
		cfg.Content.IncludeTags = []string{"p"}
		cfg.Content.ExcludeTags = []string{"script"}
		cfg.Plugins.Filters = []string{excludeTagsFilter, includeTagsFilter, selectorFilter}

		k := NewKernel(&cfg, nil, reg)
		got := k.filterNames()

		assert.Equal(t, []string{
			selectorFilter,
			excludeSelectorsFilter,
			includeTagsFilter,
			excludeTagsFilter,
		}, got)
	})
}

// stubFilter は filterNames テスト用の最小 Filter。
type stubFilter struct {
	name string
}

func (f *stubFilter) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindFilter}
}
func (f *stubFilter) Init(context.Context, plugin.Host) error { return nil }
func (f *stubFilter) Close(context.Context) error             { return nil }
func (f *stubFilter) Filter(context.Context, *model.Content) (*model.Content, error) {
	return nil, nil
}
