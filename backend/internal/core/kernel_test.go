package core_test

import (
	"context"
	"errors"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

// --- Init/Close 動作確認用フェイク ---

type call struct {
	name  string
	event string // "init" / "close"
}

type recorder struct {
	calls *[]call
}

type ppFake struct {
	name      string
	rec       *recorder
	failInit  bool
	failClose bool
}

func (p *ppFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: p.name, Kind: plugin.KindPreProcessor}
}
func (p *ppFake) Init(_ context.Context, _ plugin.Host) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "init"})
	if p.failInit {
		return errors.New("init failed")
	}
	return nil
}
func (p *ppFake) Close(_ context.Context) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "close"})
	if p.failClose {
		return errors.New("close failed")
	}
	return nil
}
func (p *ppFake) PreProcess(_ context.Context, _ *model.Request) error { return nil }

type parserFake struct {
	name string
	rec  *recorder
}

func (p *parserFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: p.name, Kind: plugin.KindParser}
}
func (p *parserFake) Init(context.Context, plugin.Host) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "init"})
	return nil
}
func (p *parserFake) Close(context.Context) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "close"})
	return nil
}
func (p *parserFake) CanParse(*model.Response) bool { return true }
func (p *parserFake) Parse(context.Context, *model.Response) (*model.Content, error) {
	return &model.Content{Format: "html"}, nil
}

type trFake struct {
	name string
	rec  *recorder
}

func (t *trFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: t.name, Kind: plugin.KindTransformer}
}
func (t *trFake) Init(context.Context, plugin.Host) error {
	*t.rec.calls = append(*t.rec.calls, call{t.name, "init"})
	return nil
}
func (t *trFake) Close(context.Context) error {
	*t.rec.calls = append(*t.rec.calls, call{t.name, "close"})
	return nil
}
func (t *trFake) Transform(context.Context, *model.Content) (*model.Result, error) {
	return &model.Result{}, nil
}

type fltFake struct {
	name string
	rec  *recorder
}

func (f *fltFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindFilter}
}
func (f *fltFake) Init(context.Context, plugin.Host) error {
	*f.rec.calls = append(*f.rec.calls, call{f.name, "init"})
	return nil
}
func (f *fltFake) Close(context.Context) error {
	*f.rec.calls = append(*f.rec.calls, call{f.name, "close"})
	return nil
}
func (f *fltFake) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	return c, nil
}

type fetcherFake struct {
	name string
	rec  *recorder
}

func (f *fetcherFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: f.name, Kind: plugin.KindFetcher}
}
func (f *fetcherFake) Init(context.Context, plugin.Host) error {
	*f.rec.calls = append(*f.rec.calls, call{f.name, "init"})
	return nil
}
func (f *fetcherFake) Close(context.Context) error {
	*f.rec.calls = append(*f.rec.calls, call{f.name, "close"})
	return nil
}
func (f *fetcherFake) Get(context.Context, *url.URL, map[string]string) (*model.Response, error) {
	return &model.Response{}, nil
}

type leFake struct {
	name string
	rec  *recorder
}

func (l *leFake) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: l.name, Kind: plugin.KindLinkExtractor}
}
func (l *leFake) Init(context.Context, plugin.Host) error {
	*l.rec.calls = append(*l.rec.calls, call{l.name, "init"})
	return nil
}
func (l *leFake) Close(context.Context) error {
	*l.rec.calls = append(*l.rec.calls, call{l.name, "close"})
	return nil
}
func (l *leFake) Extract(context.Context, *model.Content, *url.URL) ([]*url.URL, error) {
	return nil, nil
}

func newRecorder() *recorder { return &recorder{calls: &[]call{}} }

// TestKernel_Lifecycle はプラグインの Init/Close 順序と失敗時のロールバックを検証する。
func TestKernel_Lifecycle(t *testing.T) {
	t.Run("正常系: 設定で指定したプラグインが順番にInitされ、逆順でCloseされる", func(t *testing.T) {
		reg := core.NewRegistry()
		rec := newRecorder()

		registerAll(reg, rec, false, "")

		cfg := newCfgForKernel()
		k := core.NewKernel(cfg, nil, reg)

		assert.NoError(t, k.Init(context.Background()))
		assert.NoError(t, k.Close(context.Background()))

		names := orderOf(rec, "init")
		assert.Equal(t,
			[]string{"pp1", "http", "ps1", "tr1", "ft1", "le1"},
			names,
			"設計通りの初期化順 (P2→P3→P5→P6→P7→P8)")

		closeNames := orderOf(rec, "close")
		assert.Equal(t,
			[]string{"le1", "ft1", "tr1", "ps1", "http", "pp1"},
			closeNames,
			"初期化と逆順でCloseされる")
	})

	t.Run("異常系: 途中のInit失敗時、それまで成功したものを逆順でCloseする", func(t *testing.T) {
		reg := core.NewRegistry()
		rec := newRecorder()

		// Filter で Init を失敗させる
		registerAll(reg, rec, true, "ft1")

		cfg := newCfgForKernel()
		k := core.NewKernel(cfg, nil, reg)

		err := k.Init(context.Background())

		assert.Error(t, err, "Init は失敗するはず")
		assert.Contains(t, err.Error(), "filter ft1")

		// ロールバックで Close されたものは ft1 より前の (tr1, ps1, pp1)
		closeNames := orderOf(rec, "close")
		assert.Equal(t,
			[]string{"tr1", "ps1", "http", "pp1"},
			closeNames,
			"Init成功済みのプラグインだけが逆順でCloseされる")
	})

	t.Run("異常系: 設定で指定したプラグインがレジストリ未登録だとエラー", func(t *testing.T) {
		reg := core.NewRegistry()
		// 何も登録しない

		cfg := newCfgForKernel()
		k := core.NewKernel(cfg, nil, reg)

		err := k.Init(context.Background())

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func registerAll(reg *core.Registry, rec *recorder, failFilterInit bool, _ string) {
	// 名前は固定 (pp1, ps1, tr1, ft1, le1)
	// テストでは Filter Init を失敗させて挙動確認したい。
	pp1 := &ppFake{name: "pp1", rec: rec}
	ps1 := &parserFake{name: "ps1", rec: rec}
	tr1 := &trFake{name: "tr1", rec: rec}
	ft1 := &fltFake{name: "ft1", rec: rec}
	le1 := &leFake{name: "le1", rec: rec}

	regPreProcessor(reg, "pp1", pp1)
	core.RegisterFetcherTo(reg, "http", func() plugin.Fetcher { return &fetcherFake{name: "http", rec: rec} })
	regParser(reg, "ps1", ps1)
	regTransformer(reg, "tr1", tr1)
	if failFilterInit {
		// Init で失敗するフェイクに差し替え
		bad := &ppFakeAsFilter{name: "ft1", rec: rec, fail: true}
		regFilter(reg, "ft1", bad)
	} else {
		regFilter(reg, "ft1", ft1)
	}
	regLinkExtractor(reg, "le1", le1)
}

// ppFakeAsFilter は Filter.Init で失敗するフェイク。
type ppFakeAsFilter struct {
	name string
	rec  *recorder
	fail bool
}

func (p *ppFakeAsFilter) Metadata() plugin.Metadata {
	return plugin.Metadata{Name: p.name, Kind: plugin.KindFilter}
}
func (p *ppFakeAsFilter) Init(context.Context, plugin.Host) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "init"})
	if p.fail {
		return errors.New("filter init failed")
	}
	return nil
}
func (p *ppFakeAsFilter) Close(context.Context) error {
	*p.rec.calls = append(*p.rec.calls, call{p.name, "close"})
	return nil
}
func (p *ppFakeAsFilter) Filter(_ context.Context, c *model.Content) (*model.Content, error) {
	return c, nil
}

// 登録ヘルパ（リフレクションを避け、ファクトリでラップ）

func regPreProcessor(reg *core.Registry, name string, p plugin.PreProcessor) {
	core.RegisterPreProcessorTo(reg, name, func() plugin.PreProcessor { return p })
}
func regParser(reg *core.Registry, name string, p plugin.Parser) {
	core.RegisterParserTo(reg, name, func() plugin.Parser { return p })
}
func regTransformer(reg *core.Registry, name string, p plugin.Transformer) {
	core.RegisterTransformerTo(reg, name, func() plugin.Transformer { return p })
}
func regFilter(reg *core.Registry, name string, p plugin.Filter) {
	core.RegisterFilterTo(reg, name, func() plugin.Filter { return p })
}
func regLinkExtractor(reg *core.Registry, name string, p plugin.LinkExtractor) {
	core.RegisterLinkExtractorTo(reg, name, func() plugin.LinkExtractor { return p })
}

func newCfgForKernel() *model.Config {
	c := model.Default()
	c.Targets = []string{"https://example.com/"}
	c.Plugins = model.PluginSelection{
		Fetcher:       model.FetcherHTTP,
		PreProcessors: []string{"pp1"},
		Parsers:       []string{"ps1"},
		Transformer:   "tr1",
		Filters:       []string{"ft1"},
		LinkExtractor: "le1",
	}
	return &c
}

func orderOf(rec *recorder, event string) []string {
	var out []string
	for _, c := range *rec.calls {
		if c.event == event {
			out = append(out, c.name)
		}
	}
	return out
}
