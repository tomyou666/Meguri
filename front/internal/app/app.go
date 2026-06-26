package app

import (
	"meguri-app/internal/usecase/wails_service"
)

// Application は Wails 登録用ルート struct。
type Application struct {
	StoreService   *wails_service.StoreService
	ProjectService *wails_service.ProjectService
	ScraperService *wails_service.ScraperService
	cleanup        func()
}

// Cleanup は DB 等の後処理を実行する。
func (a *Application) Cleanup() {
	if a.cleanup != nil {
		a.cleanup()
	}
}
