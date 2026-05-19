package core

import (
	"fmt"
	"sort"
	"sync"

	"scraperbot/internal/domain/plugin"
)

// Registry は種類別にプラグインファクトリを保持するレジストリ。
// プラグインは init() から RegisterXxx を呼ぶことで自己登録する。
type Registry struct {
	mu             sync.RWMutex
	preprocessors  map[string]func() plugin.PreProcessor
	parsers        map[string]func() plugin.Parser
	transformers   map[string]func() plugin.Transformer
	filters        map[string]func() plugin.Filter
	linkextractors map[string]func() plugin.LinkExtractor
}

func newRegistry() *Registry {
	return &Registry{
		preprocessors:  map[string]func() plugin.PreProcessor{},
		parsers:        map[string]func() plugin.Parser{},
		transformers:   map[string]func() plugin.Transformer{},
		filters:        map[string]func() plugin.Filter{},
		linkextractors: map[string]func() plugin.LinkExtractor{},
	}
}

var defaultRegistry = newRegistry()

// Default はプロセス共有のデフォルトレジストリを返す。
func Default() *Registry { return defaultRegistry }

// NewRegistry はテストで分離したい場合に使う独立レジストリを返す。
func NewRegistry() *Registry { return newRegistry() }

// RegisterPreProcessor は P2 プラグインを登録する。同名は panic。
func RegisterPreProcessor(name string, f func() plugin.PreProcessor) {
	defaultRegistry.registerPreProcessor(name, f)
}

// RegisterParser は P5 プラグインを登録する。
func RegisterParser(name string, f func() plugin.Parser) {
	defaultRegistry.registerParser(name, f)
}

// RegisterTransformer は P6 プラグインを登録する。
func RegisterTransformer(name string, f func() plugin.Transformer) {
	defaultRegistry.registerTransformer(name, f)
}

// RegisterFilter は P7 プラグインを登録する。
func RegisterFilter(name string, f func() plugin.Filter) {
	defaultRegistry.registerFilter(name, f)
}

// RegisterLinkExtractor は P8 プラグインを登録する。
func RegisterLinkExtractor(name string, f func() plugin.LinkExtractor) {
	defaultRegistry.registerLinkExtractor(name, f)
}

// RegisterPreProcessorTo は任意のレジストリへ登録するための公開ヘルパ。
// テストや独立レジストリでの登録を可能にする。
func RegisterPreProcessorTo(r *Registry, name string, f func() plugin.PreProcessor) {
	r.registerPreProcessor(name, f)
}
func RegisterParserTo(r *Registry, name string, f func() plugin.Parser) {
	r.registerParser(name, f)
}
func RegisterTransformerTo(r *Registry, name string, f func() plugin.Transformer) {
	r.registerTransformer(name, f)
}
func RegisterFilterTo(r *Registry, name string, f func() plugin.Filter) {
	r.registerFilter(name, f)
}
func RegisterLinkExtractorTo(r *Registry, name string, f func() plugin.LinkExtractor) {
	r.registerLinkExtractor(name, f)
}

func (r *Registry) registerPreProcessor(name string, f func() plugin.PreProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.preprocessors[name]; dup {
		panic(fmt.Sprintf("preprocessor already registered: %s", name))
	}
	r.preprocessors[name] = f
}

func (r *Registry) registerParser(name string, f func() plugin.Parser) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.parsers[name]; dup {
		panic(fmt.Sprintf("parser already registered: %s", name))
	}
	r.parsers[name] = f
}

func (r *Registry) registerTransformer(name string, f func() plugin.Transformer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.transformers[name]; dup {
		panic(fmt.Sprintf("transformer already registered: %s", name))
	}
	r.transformers[name] = f
}

func (r *Registry) registerFilter(name string, f func() plugin.Filter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.filters[name]; dup {
		panic(fmt.Sprintf("filter already registered: %s", name))
	}
	r.filters[name] = f
}

func (r *Registry) registerLinkExtractor(name string, f func() plugin.LinkExtractor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, dup := r.linkextractors[name]; dup {
		panic(fmt.Sprintf("link_extractor already registered: %s", name))
	}
	r.linkextractors[name] = f
}

// 以下は登録名から新しいインスタンスを生成するファクトリ呼び出し群。

func (r *Registry) NewPreProcessor(name string) (plugin.PreProcessor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.preprocessors[name]
	if !ok {
		return nil, fmt.Errorf("preprocessor not found: %s", name)
	}
	return f(), nil
}

func (r *Registry) NewParser(name string) (plugin.Parser, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.parsers[name]
	if !ok {
		return nil, fmt.Errorf("parser not found: %s", name)
	}
	return f(), nil
}

func (r *Registry) NewTransformer(name string) (plugin.Transformer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.transformers[name]
	if !ok {
		return nil, fmt.Errorf("transformer not found: %s", name)
	}
	return f(), nil
}

func (r *Registry) NewFilter(name string) (plugin.Filter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.filters[name]
	if !ok {
		return nil, fmt.Errorf("filter not found: %s", name)
	}
	return f(), nil
}

func (r *Registry) NewLinkExtractor(name string) (plugin.LinkExtractor, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.linkextractors[name]
	if !ok {
		return nil, fmt.Errorf("link_extractor not found: %s", name)
	}
	return f(), nil
}

// Has は指定された Kind と name が登録されているかを返す。
func (r *Registry) Has(kind plugin.Kind, name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	switch kind {
	case plugin.KindPreProcessor:
		_, ok := r.preprocessors[name]
		return ok
	case plugin.KindParser:
		_, ok := r.parsers[name]
		return ok
	case plugin.KindTransformer:
		_, ok := r.transformers[name]
		return ok
	case plugin.KindFilter:
		_, ok := r.filters[name]
		return ok
	case plugin.KindLinkExtractor:
		_, ok := r.linkextractors[name]
		return ok
	}
	return false
}

// Names は登録されている特定 Kind の名前一覧をソート順で返す（テスト・デバッグ用）。
func (r *Registry) Names(kind plugin.Kind) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var src map[string]struct{}
	switch kind {
	case plugin.KindPreProcessor:
		src = keysOf(r.preprocessors)
	case plugin.KindParser:
		src = keysOf(r.parsers)
	case plugin.KindTransformer:
		src = keysOf(r.transformers)
	case plugin.KindFilter:
		src = keysOf(r.filters)
	case plugin.KindLinkExtractor:
		src = keysOf(r.linkextractors)
	default:
		return nil
	}
	out := make([]string, 0, len(src))
	for k := range src {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func keysOf[V any](m map[string]V) map[string]struct{} {
	out := make(map[string]struct{}, len(m))
	for k := range m {
		out[k] = struct{}{}
	}
	return out
}

// テスト用にレジストリを差し替える小道具。
// 既存実装に副作用がないように、関数呼び出し側で defer restore する想定。
func swapDefaultRegistry(r *Registry) (restore func()) {
	old := defaultRegistry
	defaultRegistry = r
	return func() { defaultRegistry = old }
}
