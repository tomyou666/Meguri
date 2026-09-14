package wails_service

import (
	"context"
	"fmt"

	"meguri-app/internal/domain"
	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:generate go tool gowrap gen -p meguri-app/internal/usecase/wails_service -i ProjectServiceAPI -t templates/slog_debug_slow.gotmpl -o project_service_with_debug_log.go -v ServiceName=ProjectService -v DecoratorName=ProjectServiceWithDebugLog

// ProjectServiceAPI は ProjectService の公開メソッド面。
type ProjectServiceAPI interface {
	// SetApp は Wails App を後から注入する（ダイアログ用）。
	SetApp(app *application.App)
	// OpenScrb は .scrb を開き新規 WS としてインポートする。
	OpenScrb() (model.OpenScrbResponse, error)
	// SaveScrb はアクティブ WS を .crawlproj に保存する。
	SaveScrb(workspaceID string, includeResults bool) error
}

// ProjectService は .scrb 入出力 Wails RPC（debug ログ付き殻）。
type ProjectService struct {
	ProjectServiceWithDebugLog
}

// projectService は ProjectService の実装本体。
type projectService struct {
	// projects はプロジェクトファイル入出力サービス。
	projects *domain.ProjectFileService
	// workspaces はワークスペースサービス。
	workspaces *domain.WorkspaceService
	// app は Wails アプリ（ダイアログ用）。
	app *application.App
}

// NewProjectService は ProjectService を構築する。
func NewProjectService(projects *domain.ProjectFileService, workspaces *domain.WorkspaceService) *ProjectService {
	impl := &projectService{projects: projects, workspaces: workspaces}
	return &ProjectService{ProjectServiceWithDebugLog: NewProjectServiceWithDebugLog(impl)}
}

// SetApp は Wails App を後から注入する（ダイアログ用）。
func (s *projectService) SetApp(app *application.App) {
	s.app = app
}

// OpenScrb は .scrb を開き新規 WS としてインポートする。
func (s *projectService) OpenScrb() (model.OpenScrbResponse, error) {
	if s.app == nil {
		return model.OpenScrbResponse{}, fmt.Errorf("app not initialized")
	}
	path, err := s.app.Dialog.OpenFile().
		SetTitle("Open Meguri Project").
		AddFilter("Meguri Project", "*.crawlproj").
		AddFilter("Legacy Project", "*.scrb").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return model.OpenScrbResponse{}, err
	}
	id, err := s.projects.ImportFromPath(s.ctx(), path)
	if err != nil {
		return model.OpenScrbResponse{}, err
	}
	return model.OpenScrbResponse{WorkspaceID: id}, nil
}

// SaveScrb はアクティブ WS を .crawlproj に保存する。
//
// includeResults が true のとき最新成功の node_results も ZIP に含める。
func (s *projectService) SaveScrb(workspaceID string, includeResults bool) error {
	if s.app == nil {
		return fmt.Errorf("app not initialized")
	}
	ws, err := s.workspaces.Load(s.ctx(), workspaceID)
	if err != nil || ws == nil {
		return fmt.Errorf("workspace not found")
	}
	defaultName := domain.BundleName(ws)
	path, err := s.app.Dialog.SaveFile().
		SetMessage("Save Meguri Project").
		SetFilename(defaultName).
		AddFilter("Meguri Project", "*.crawlproj").
		AddFilter("Legacy Project", "*.scrb").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return err
	}
	return s.projects.ExportToPath(s.ctx(), workspaceID, path, includeResults)
}

func (s *projectService) ctx() context.Context { return context.Background() }
