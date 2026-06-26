package app

import (
	"context"
	"fmt"

	"meguri-app/internal/domain"
	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/usecase/wails_service"

	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
)

// ProvideDBPath は SQLite ファイルパスを返す。
func ProvideDBPath() (string, error) {
	return ResolveDBPath()
}

// ProvideMigrations はマイグレーションを適用する。
func ProvideMigrations(dbPath string) error {
	return RunMigrations(dbPath)
}

// ProvideDB は GORM DB を開く（マイグレーション適用後）。
func ProvideDB(dbPath string) (*gorm.DB, func(), error) {
	if err := RunMigrations(dbPath); err != nil {
		return nil, nil, err
	}
	db, err := gorm.Open(sqlite.Open(SQLiteDSN(dbPath)), &gorm.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("open db: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = sqlDB.Close() }
	return db, cleanup, nil
}

// ProvideRepository は persistence.Repository を返す。
func ProvideRepository(db *gorm.DB) persistence.Repository {
	return persistence.NewStore(db)
}

// ProvideAppConfigService は AppConfigService を返す。
func ProvideAppConfigService(repo persistence.Repository) *domain.AppConfigService {
	return domain.NewAppConfigService(repo)
}

// ProvideWorkspaceService は WorkspaceService を返す。
func ProvideWorkspaceService(repo persistence.Repository) *domain.WorkspaceService {
	return domain.NewWorkspaceService(repo)
}

// ProvideResultsService は ResultsService を返す。
func ProvideResultsService(repo persistence.Repository, ws *domain.WorkspaceService) *domain.ResultsService {
	return domain.NewResultsService(repo, ws)
}

// ProvideDiffService は DiffService を返す。
func ProvideDiffService(repo persistence.Repository, ws *domain.WorkspaceService) *domain.DiffService {
	return domain.NewDiffService(repo, ws)
}

// ProvideCrawlPersistService は CrawlPersistService を返す。
func ProvideCrawlPersistService(repo persistence.Repository) *domain.CrawlPersistService {
	return domain.NewCrawlPersistService(repo)
}

// ProvideProjectFileService は ProjectFileService を返す。
func ProvideProjectFileService(ws *domain.WorkspaceService) *domain.ProjectFileService {
	return domain.NewProjectFileService(ws)
}

// ProvideStoreService は Wails StoreService を返す。
func ProvideStoreService(
	appConfig *domain.AppConfigService,
	workspaces *domain.WorkspaceService,
	results *domain.ResultsService,
	diff *domain.DiffService,
	crawlPersist *domain.CrawlPersistService,
) *wails_service.StoreService {
	return wails_service.NewStoreService(appConfig, workspaces, results, diff, crawlPersist)
}

// ProvideProjectService は Wails ProjectService を返す。
func ProvideProjectService(projects *domain.ProjectFileService, workspaces *domain.WorkspaceService) *wails_service.ProjectService {
	return wails_service.NewProjectService(projects, workspaces)
}

// ProvideScraperService は ScraperService を返す。
func ProvideScraperService(persist *domain.CrawlPersistService) *wails_service.ScraperService {
	return wails_service.NewScraperService(persist)
}

// ProvideApplication は Application を組み立てる。
func ProvideApplication(
	store *wails_service.StoreService,
	project *wails_service.ProjectService,
	scraper *wails_service.ScraperService,
	db *gorm.DB,
) (*Application, func(), error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = sqlDB.Close() }
	app := &Application{
		StoreService:   store,
		ProjectService: project,
		ScraperService: scraper,
		cleanup:        cleanup,
	}
	return app, cleanup, nil
}

// Initialize はアプリケーション依存を組み立てる。
func Initialize(ctx context.Context) (*Application, func(), error) {
	dbPath, err := ProvideDBPath()
	if err != nil {
		return nil, nil, err
	}
	db, _, err := ProvideDB(dbPath)
	if err != nil {
		return nil, nil, err
	}
	repo := ProvideRepository(db)
	appConfig := ProvideAppConfigService(repo)
	workspaces := ProvideWorkspaceService(repo)
	results := ProvideResultsService(repo, workspaces)
	diff := ProvideDiffService(repo, workspaces)
	crawlPersist := ProvideCrawlPersistService(repo)
	projects := ProvideProjectFileService(workspaces)
	store := ProvideStoreService(appConfig, workspaces, results, diff, crawlPersist)
	project := ProvideProjectService(projects, workspaces)
	scraper := ProvideScraperService(crawlPersist)
	app, cleanup, err := ProvideApplication(store, project, scraper, db)
	if err != nil {
		cleanup()
		return nil, nil, err
	}
	if err := store.Bootstrap(); err != nil {
		cleanup()
		return nil, nil, err
	}
	return app, cleanup, nil
}
