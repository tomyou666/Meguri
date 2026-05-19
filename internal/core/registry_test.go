package core

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

// --- レジストリ単体テスト用の最小フェイクプラグイン群 ---

type fakePreProc struct{ name string }

func (f *fakePreProc) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindPreProcessor}
}
func (f *fakePreProc) Init(context.Context, plugin.Host) error          { return nil }
func (f *fakePreProc) Close(context.Context) error                      { return nil }
func (f *fakePreProc) PreProcess(context.Context, *model.Request) error { return nil }

type fakeParser struct{ name string }

func (f *fakeParser) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindParser}
}
func (f *fakeParser) Init(context.Context, plugin.Host) error { return nil }
func (f *fakeParser) Close(context.Context) error             { return nil }
func (f *fakeParser) CanParse(*model.Response) bool           { return true }
func (f *fakeParser) Parse(context.Context, *model.Response) (*model.Content, error) {
	return &model.Content{Format: "html"}, nil
}

type fakeTransformer struct{ name string }

func (f *fakeTransformer) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindTransformer}
}
func (f *fakeTransformer) Init(context.Context, plugin.Host) error { return nil }
func (f *fakeTransformer) Close(context.Context) error             { return nil }
func (f *fakeTransformer) Transform(context.Context, *model.Content) (*model.Result, error) {
	return &model.Result{}, nil
}

type fakeFilter struct{ name string }

func (f *fakeFilter) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindFilter}
}
func (f *fakeFilter) Init(context.Context, plugin.Host) error { return nil }
func (f *fakeFilter) Close(context.Context) error             { return nil }
func (f *fakeFilter) Filter(context.Context, *model.Content) (*model.Content, error) {
	return &model.Content{}, nil
}

type fakeLinkExtractor struct{ name string }

func (f *fakeLinkExtractor) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindLinkExtractor}
}
func (f *fakeLinkExtractor) Init(context.Context, plugin.Host) error { return nil }
func (f *fakeLinkExtractor) Close(context.Context) error             { return nil }
func (f *fakeLinkExtractor) Extract(context.Context, *model.Content, *url.URL) ([]*url.URL, error) {
	return nil, nil
}

// --- テスト本体 ---

func TestRegistry(t *testing.T) {
	t.Run("正常系: 各Kindのプラグインを登録して取得できる", func(t *testing.T) {
		reg := NewRegistry()
		reg.registerPreProcessor("pp", func() plugin.PreProcessor { return &fakePreProc{name: "pp"} })
		reg.registerParser("ps", func() plugin.Parser { return &fakeParser{name: "ps"} })
		reg.registerTransformer("tr", func() plugin.Transformer { return &fakeTransformer{name: "tr"} })
		reg.registerFilter("ft", func() plugin.Filter { return &fakeFilter{name: "ft"} })
		reg.registerLinkExtractor("le", func() plugin.LinkExtractor { return &fakeLinkExtractor{name: "le"} })

		pp, err := reg.NewPreProcessor("pp")
		assert.NoError(t, err)
		assert.NotNil(t, pp)
		assert.Equal(t, "pp", pp.Metadata().Name)

		ps, err := reg.NewParser("ps")
		assert.NoError(t, err)
		assert.NotNil(t, ps)

		tr, err := reg.NewTransformer("tr")
		assert.NoError(t, err)
		assert.NotNil(t, tr)

		ft, err := reg.NewFilter("ft")
		assert.NoError(t, err)
		assert.NotNil(t, ft)

		le, err := reg.NewLinkExtractor("le")
		assert.NoError(t, err)
		assert.NotNil(t, le)
	})

	t.Run("異常系: 未登録名を取得するとエラー", func(t *testing.T) {
		reg := NewRegistry()

		_, err := reg.NewParser("missing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parser not found")
	})

	t.Run("異常系: 同名で2回登録するとpanic", func(t *testing.T) {
		reg := NewRegistry()
		reg.registerParser("dup", func() plugin.Parser { return &fakeParser{name: "dup"} })

		assert.Panics(t, func() {
			reg.registerParser("dup", func() plugin.Parser { return &fakeParser{name: "dup"} })
		})
	})

	t.Run("正常系: ファクトリは毎回新しいインスタンスを返す", func(t *testing.T) {
		reg := NewRegistry()
		reg.registerParser("x", func() plugin.Parser { return &fakeParser{name: "x"} })

		a, _ := reg.NewParser("x")
		b, _ := reg.NewParser("x")

		assert.NotSame(t, a, b, "別インスタンスであるべき")
	})

	t.Run("正常系: Has は登録の有無を Kind 単位で正しく判定する", func(t *testing.T) {
		reg := NewRegistry()
		reg.registerFilter("f", func() plugin.Filter { return &fakeFilter{name: "f"} })

		assert.True(t, reg.Has(plugin.KindFilter, "f"))
		assert.False(t, reg.Has(plugin.KindFilter, "missing"))
		assert.False(t, reg.Has(plugin.KindParser, "f"), "他のKindでは見えてはいけない")
	})

	t.Run("正常系: Names はソート済みリストを返す", func(t *testing.T) {
		reg := NewRegistry()
		reg.registerFilter("bbb", func() plugin.Filter { return &fakeFilter{name: "bbb"} })
		reg.registerFilter("aaa", func() plugin.Filter { return &fakeFilter{name: "aaa"} })

		got := reg.Names(plugin.KindFilter)

		assert.Equal(t, []string{"aaa", "bbb"}, got)
	})

	t.Run("正常系: デフォルトレジストリの差し替えはdeferでロールバックされる", func(t *testing.T) {
		original := Default()
		tmp := NewRegistry()
		restore := swapDefaultRegistry(tmp)

		assert.Same(t, tmp, Default())
		restore()
		assert.Same(t, original, Default())
	})
}
