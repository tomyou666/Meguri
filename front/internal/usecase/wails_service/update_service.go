package wails_service

import (
	"context"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// UpdateCheckResult は手動更新確認の結果。
type UpdateCheckResult struct {
	// Status は結果種別（up_to_date / update_ready）。
	Status string `json:"status"`
	// Version は対象リリースのバージョン（update_ready のとき）。
	Version string `json:"version,omitempty"`
}

// UpdateService は Wails updater の手動確認 RPC。
type UpdateService struct {
	app *application.App
}

// NewUpdateService は UpdateService を構築する。
func NewUpdateService() *UpdateService {
	return &UpdateService{}
}

// SetApp は Wails App を後から注入する。
func (s *UpdateService) SetApp(app *application.App) {
	s.app = app
}

// CheckAndInstall は更新を確認し、利用可能ならダウンロードしてステージする。
//
// WindowNone 構成では Wails 標準 UI を開かず、結果を UpdateCheckResult で返す。
func (s *UpdateService) CheckAndInstall() (UpdateCheckResult, error) {
	if s.app == nil || s.app.Updater == nil {
		return UpdateCheckResult{}, ErrUpdaterUnavailable
	}
	ctx := context.Background()
	rel, err := s.app.Updater.Check(ctx)
	if err != nil {
		return UpdateCheckResult{}, err
	}
	if rel == nil {
		return UpdateCheckResult{Status: "up_to_date"}, nil
	}
	if err := s.app.Updater.DownloadAndInstall(ctx); err != nil {
		return UpdateCheckResult{}, err
	}
	return UpdateCheckResult{
		Status:  "update_ready",
		Version: rel.Version,
	}, nil
}
