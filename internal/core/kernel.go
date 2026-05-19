package core

import (
	"context"
	"errors"
	"fmt"

	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

// Kernel はプラグインのライフサイクル制御とパイプラインの依存提供を担う。
type Kernel struct {
	cfg  *model.Config
	host plugin.Host
	reg  *Registry

	preprocessors []plugin.PreProcessor
	parsers       []plugin.Parser
	transformer   plugin.Transformer
	filters       []plugin.Filter
	linkExtractor plugin.LinkExtractor

	initialized []plugin.Plugin
}

// NewKernel は与えられた設定とホストを保持するカーネルを返す。
// レジストリ未指定（nil）の場合は Default() を使う。
func NewKernel(cfg *model.Config, host plugin.Host, reg *Registry) *Kernel {
	if reg == nil {
		reg = Default()
	}
	return &Kernel{cfg: cfg, host: host, reg: reg}
}

// Init は設定で指定された名前のプラグインをレジストリから生成して Init する。
// 途中の失敗は致命扱いとし、それまでに成功したプラグインを逆順で Close してロールバックする。
func (k *Kernel) Init(ctx context.Context) error {
	rollback := func(initErr error) error {
		var errs []error
		for i := len(k.initialized) - 1; i >= 0; i-- {
			if err := k.initialized[i].Close(ctx); err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 0 {
			return initErr
		}
		return errors.Join(append([]error{initErr}, errs...)...)
	}

	for _, name := range k.cfg.Plugins.PreProcessors {
		p, err := k.reg.NewPreProcessor(name)
		if err != nil {
			return rollback(err)
		}
		if err := p.Init(ctx, k.host); err != nil {
			return rollback(fmt.Errorf("init preprocessor %s: %w", name, err))
		}
		k.preprocessors = append(k.preprocessors, p)
		k.initialized = append(k.initialized, p)
	}

	for _, name := range k.cfg.Plugins.Parsers {
		p, err := k.reg.NewParser(name)
		if err != nil {
			return rollback(err)
		}
		if err := p.Init(ctx, k.host); err != nil {
			return rollback(fmt.Errorf("init parser %s: %w", name, err))
		}
		k.parsers = append(k.parsers, p)
		k.initialized = append(k.initialized, p)
	}

	t, err := k.reg.NewTransformer(k.cfg.Plugins.Transformer)
	if err != nil {
		return rollback(err)
	}
	if err := t.Init(ctx, k.host); err != nil {
		return rollback(fmt.Errorf("init transformer %s: %w", k.cfg.Plugins.Transformer, err))
	}
	k.transformer = t
	k.initialized = append(k.initialized, t)

	for _, name := range k.cfg.Plugins.Filters {
		p, err := k.reg.NewFilter(name)
		if err != nil {
			return rollback(err)
		}
		if err := p.Init(ctx, k.host); err != nil {
			return rollback(fmt.Errorf("init filter %s: %w", name, err))
		}
		k.filters = append(k.filters, p)
		k.initialized = append(k.initialized, p)
	}

	le, err := k.reg.NewLinkExtractor(k.cfg.Plugins.LinkExtractor)
	if err != nil {
		return rollback(err)
	}
	if err := le.Init(ctx, k.host); err != nil {
		return rollback(fmt.Errorf("init link_extractor %s: %w", k.cfg.Plugins.LinkExtractor, err))
	}
	k.linkExtractor = le
	k.initialized = append(k.initialized, le)

	return nil
}

// Close は初期化済みプラグインを登録の逆順で Close する。
// 個別の Close エラーは集約して返すが、起動失敗とは扱わない。
func (k *Kernel) Close(ctx context.Context) error {
	var errs []error
	for i := len(k.initialized) - 1; i >= 0; i-- {
		if err := k.initialized[i].Close(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// --- パイプラインから参照するアクセサ ---

func (k *Kernel) Config() *model.Config                { return k.cfg }
func (k *Kernel) Host() plugin.Host                    { return k.host }
func (k *Kernel) PreProcessors() []plugin.PreProcessor { return k.preprocessors }
func (k *Kernel) Parsers() []plugin.Parser             { return k.parsers }
func (k *Kernel) Transformer() plugin.Transformer      { return k.transformer }
func (k *Kernel) Filters() []plugin.Filter             { return k.filters }
func (k *Kernel) LinkExtractor() plugin.LinkExtractor  { return k.linkExtractor }
