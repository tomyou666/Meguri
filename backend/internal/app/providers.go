// Package app は Wire による composition root を提供する。
package app

import (
	"context"
	"log/slog"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
	"meguri/internal/infrastructure/storage"
	"meguri/internal/usecase"
)

// ProvideRegistry は init() 済みプラグインが登録されたレジストリを返す。
func ProvideRegistry() *core.Registry {
	return core.Default()
}

// ProvideHost はプラグインに渡す Host 実装を構築する。
func ProvideHost(cfg *model.Config) plugin.Host {
	return core.NewHost(cfg)
}

// ProvideKernel はカーネルを生成し Init する。成功時は Close 用 cleanup を返す。
func ProvideKernel(
	ctx context.Context,
	cfg *model.Config,
	host plugin.Host,
	reg *core.Registry,
) (*core.Kernel, func(), error) {
	k := core.NewKernel(cfg, host, reg)
	if err := k.Init(ctx); err != nil {
		return nil, nil, err
	}
	cleanup := func() {
		_ = k.Close(ctx)
	}
	return k, cleanup, nil
}

// ProvideFileWriter は OutputConfig とフォーマット一覧から FileWriter を構築する。
func ProvideFileWriter(cfg *model.Config) *storage.FileWriter {
	return storage.NewFileWriter(cfg.Output, cfg.Content.Formats)
}

// ProvidePipeline はカーネルから 1 URL 処理用パイプラインを構築する。
func ProvidePipeline(k *core.Kernel) *core.Pipeline {
	return core.NewPipeline(k)
}

// ProvideScrape は単一 URL スクレイプ用ユースケースを構築する。
func ProvideScrape(pipeline *core.Pipeline) *usecase.Scrape {
	return usecase.NewScrape(pipeline)
}

// FileResultSink はクロール結果を FileWriter へ書き出す Sink 実装。
type FileResultSink struct {
	// Writer は結果のファイル出力先。
	Writer *storage.FileWriter
}

// Handle は core.ResultSink として Result を FileWriter へ書き出す。
func (s *FileResultSink) Handle(r *model.Result) {
	if err := s.Writer.Write(r); err != nil {
		slog.Warn("出力書き込み失敗", "url", r.URL.String(), "err", err.Error())
	}
}

// ProvideCrawl はクロール用ユースケースを構築する。
func ProvideCrawl(sink *FileResultSink) *usecase.Crawl {
	return usecase.NewCrawl(sink.Handle)
}
