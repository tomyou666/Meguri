package core

import (
	"context"
	"errors"
	"fmt"

	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// Kernel はプラグインのライフサイクル制御とパイプラインの依存提供を担う。
type Kernel struct {
	// cfg は実行設定。
	cfg *model.Config
	// host はプラグインへ渡す依存。
	host plugin.Host
	// reg はプラグインファクトリのレジストリ。
	reg *Registry

	// preprocessors は Init 済み PreProcessor チェーン。
	preprocessors []plugin.PreProcessor
	// parsers は Init 済み Parser 一覧。
	parsers []plugin.Parser
	// transformer は Init 済み Transformer。
	transformer plugin.Transformer
	// filters は Init 済み Filter チェーン。
	filters []plugin.Filter
	// linkExtractor は Init 済み LinkExtractor。
	linkExtractor plugin.LinkExtractor
	// fetcher は Init 済み P3 Fetcher（limitingFetcher で包まれる場合あり）。
	fetcher plugin.Fetcher

	// fetchLimiter は取得並列上限（nil 可）。
	fetchLimiter *fetchlimit.FetchLimiter

	// initialized は Init 成功順のプラグイン（Close 用）。
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

// SetFetchLimiter は Init 前に取得並列上限を設定する。
func (k *Kernel) SetFetchLimiter(l *fetchlimit.FetchLimiter) {
	k.fetchLimiter = l
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

	fetcherName := string(k.cfg.Plugins.Fetcher)
	if fetcherName == "" {
		fetcherName = string(model.FetcherHTTP)
	}
	f, err := k.reg.NewFetcher(fetcherName)
	if err != nil {
		return rollback(err)
	}
	if err := f.Init(ctx, k.host); err != nil {
		return rollback(fmt.Errorf("init fetcher %s: %w", fetcherName, err))
	}
	k.initialized = append(k.initialized, f)

	kind := k.cfg.Plugins.Fetcher
	if kind == "" {
		kind = model.FetcherHTTP
	}

	k.initFetcherWithLimiter(f, kind)

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

// Config はカーネルが保持する実行設定を返す。
func (k *Kernel) Config() *model.Config { return k.cfg }

// Host はプラグインに渡した Host 実装を返す。
func (k *Kernel) Host() plugin.Host { return k.host }

// PreProcessors は Init 済みの PreProcessor 一覧を返す。
func (k *Kernel) PreProcessors() []plugin.PreProcessor { return k.preprocessors }

// Parsers は Init 済みの Parser 一覧を返す。
func (k *Kernel) Parsers() []plugin.Parser { return k.parsers }

// Transformer は Init 済みの Transformer を返す。
func (k *Kernel) Transformer() plugin.Transformer { return k.transformer }

// Filters は Init 済みの Filter 一覧を返す。
func (k *Kernel) Filters() []plugin.Filter { return k.filters }

// LinkExtractor は Init 済みの LinkExtractor を返す。
func (k *Kernel) LinkExtractor() plugin.LinkExtractor { return k.linkExtractor }

// Fetcher は Init 済みの P3 Fetcher を返す。
func (k *Kernel) Fetcher() plugin.Fetcher { return k.fetcher }
