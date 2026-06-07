//go:build wireinject

package app

import (
	"context"

	"github.com/google/wire"
)

//go:generate go run github.com/google/wire/cmd/wire

func InitializeWire(ctx context.Context) (*Application, func(), error) {
	wire.Build(
		ProvideDBPath,
		ProvideDB,
		ProvideRepository,
		ProvideAppConfigService,
		ProvideWorkspaceService,
		ProvideResultsService,
		ProvideDiffService,
		ProvideCrawlPersistService,
		ProvideProjectFileService,
		ProvideStoreService,
		ProvideProjectService,
		ProvideApplication,
	)
	return nil, nil, nil
}
