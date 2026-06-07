package domain

import (
	"context"
	"os"

	"scraperbot-front/internal/infrastructure/scrb"
	"scraperbot-front/internal/model"
)

// ProjectFileService は .scrb 入出力。
type ProjectFileService struct {
	ws *WorkspaceService
}

// NewProjectFileService は ProjectFileService を構築する。
func NewProjectFileService(ws *WorkspaceService) *ProjectFileService {
	return &ProjectFileService{ws: ws}
}

// ImportFromPath は .scrb を読み込み新規 WS として保存する。
func (s *ProjectFileService) ImportFromPath(ctx context.Context, path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	bundle, err := scrb.Import(data)
	if err != nil {
		return "", err
	}
	return s.ws.ImportBundle(ctx, bundle)
}

// ExportToPath は WS を .scrb に書き出す。
func (s *ProjectFileService) ExportToPath(ctx context.Context, workspaceID, path string) error {
	bundle, err := s.ws.ExportBundle(ctx, workspaceID)
	if err != nil {
		return err
	}
	data, err := scrb.Export(*bundle)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ExportBytes は WS を .scrb バイト列にする。
func (s *ProjectFileService) ExportBytes(ctx context.Context, workspaceID string) ([]byte, error) {
	bundle, err := s.ws.ExportBundle(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	return scrb.Export(*bundle)
}

// ImportBytes は .scrb バイト列からインポートする。
func (s *ProjectFileService) ImportBytes(ctx context.Context, data []byte) (string, error) {
	bundle, err := scrb.Import(data)
	if err != nil {
		return "", err
	}
	return s.ws.ImportBundle(ctx, bundle)
}

// BundleName はエクスポートファイル名の提案。
func BundleName(ws *model.WorkspaceDTO) string {
	if ws == nil {
		return "workspace.scrb"
	}
	name := ws.Name
	if name == "" {
		name = "workspace"
	}
	return name + ".scrb"
}
