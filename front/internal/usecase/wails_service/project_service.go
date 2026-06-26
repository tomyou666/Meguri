package wails_service

import (
	"context"
	"fmt"

	"meguri-app/internal/domain"
	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// ProjectService は .scrb 入出力 Wails RPC。
type ProjectService struct {
	projects   *domain.ProjectFileService
	workspaces *domain.WorkspaceService
	app        *application.App
}

// NewProjectService は ProjectService を構築する。
func NewProjectService(projects *domain.ProjectFileService, workspaces *domain.WorkspaceService) *ProjectService {
	return &ProjectService{projects: projects, workspaces: workspaces}
}

// SetApp は Wails App を後から注入する（ダイアログ用）。
func (s *ProjectService) SetApp(app *application.App) {
	s.app = app
}

// OpenScrb は .scrb を開き新規 WS としてインポートする。
func (s *ProjectService) OpenScrb() (model.OpenScrbResponse, error) {
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

// SaveScrb はアクティブ WS を .scrb に保存する。
func (s *ProjectService) SaveScrb(workspaceID string) error {
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
	return s.projects.ExportToPath(s.ctx(), workspaceID, path)
}

func (s *ProjectService) ctx() context.Context { return context.Background() }
