//go:build wireinject
// +build wireinject

package app

import (
	"context"

	"github.com/google/wire"

	"meguri/internal/domain/model"
)

//go:generate go run github.com/google/wire/cmd/wire

// Initialize は Config と context から Application と cleanup 関数を構築する。
func Initialize(ctx context.Context, cfg *model.Config) (*Application, func(), error) {
	wire.Build(
		wire.Struct(new(Application), "*"),
		ProvideRegistry,
		ProvideHost,
		ProvideKernel,
		ProvideFileWriter,
		ProvidePipeline,
		ProvideScrape,
		ProvideCrawl,
		wire.Struct(new(FileResultSink), "*"),
	)
	return nil, nil, nil
}
