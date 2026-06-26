package wails_service

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"scraperbot-front/internal/domain"
	"scraperbot-front/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// StoreService は Wails 公開 Store RPC。
type StoreService struct {
	app           *application.App
	appConfig     *domain.AppConfigService
	workspaces    *domain.WorkspaceService
	results       *domain.ResultsService
	diff          *domain.DiffService
	crawlPersist  *domain.CrawlPersistService
	nodeResultWin *NodeResultWindowManager
	exportWin     *ExportWindowManager
	nodeDiffWin   *NodeDiffWindowManager
}

// NewStoreService は StoreService を構築する。
func NewStoreService(
	appConfig *domain.AppConfigService,
	workspaces *domain.WorkspaceService,
	results *domain.ResultsService,
	diff *domain.DiffService,
	crawlPersist *domain.CrawlPersistService,
) *StoreService {
	return &StoreService{
		appConfig:    appConfig,
		workspaces:   workspaces,
		results:      results,
		diff:         diff,
		crawlPersist: crawlPersist,
	}
}

// SetApp は Wails App を後から注入する（最大化・エクスポートウィンドウ用）。
func (s *StoreService) SetApp(app *application.App) {
	s.app = app
	s.nodeResultWin = NewNodeResultWindowManager(app)
	s.exportWin = NewExportWindowManager(app)
	s.nodeDiffWin = NewNodeDiffWindowManager(app)
}

// WireMainWindow はメインウィンドウ終了時のプレビュー連動を登録する。
func WireMainWindow(s *StoreService, w application.Window) {
	if s.nodeResultWin != nil {
		s.nodeResultWin.SetMainWindow(w)
	}
	if s.exportWin != nil {
		s.exportWin.SetMainWindow(w)
	}
	if s.nodeDiffWin != nil {
		s.nodeDiffWin.SetMainWindow(w)
	}
}

func (s *StoreService) ctx() context.Context { return context.Background() }

// GetAppDefaults はアプリ既定設定を返す。
func (s *StoreService) GetAppDefaults() (json.RawMessage, error) {
	return s.appConfig.GetDefaults(s.ctx())
}

// SetAppDefaults はアプリ既定設定を設定する。
func (s *StoreService) SetAppDefaults(config json.RawMessage) error {
	return s.appConfig.SaveDefaults(s.ctx(), config)
}

// SaveAppDefaults はアプリ既定設定を保存する。
func (s *StoreService) SaveAppDefaults(config json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.appConfig.SaveDefaults(s.ctx(), config); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "app"}, nil
}

// ListWorkspaces は WS 一覧を返す。
func (s *StoreService) ListWorkspaces() ([]model.WorkspaceListItemDTO, error) {
	return s.workspaces.List(s.ctx())
}

// LoadWorkspace は WS を読み込む。
func (s *StoreService) LoadWorkspace(id string) (*model.WorkspaceDTO, error) {
	return s.workspaces.Load(s.ctx(), id)
}

// SaveWorkspace は WS を保存する。
func (s *StoreService) SaveWorkspace(ws model.WorkspaceDTO) error {
	return s.workspaces.Save(s.ctx(), ws)
}

// SaveWorkspaceSettings は WS 設定を保存する。
func (s *StoreService) SaveWorkspaceSettings(workspaceID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.workspaces.SaveWorkspaceSettings(s.ctx(), workspaceID, settings); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "workspace"}, nil
}

// SaveNodeSettings はノード設定を保存する。
func (s *StoreService) SaveNodeSettings(workspaceID, nodeID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.workspaces.SaveNodeSettings(s.ctx(), workspaceID, nodeID, settings); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "node"}, nil
}

// DeleteWorkspace は WS を削除する。
func (s *StoreService) DeleteWorkspace(id string) error {
	return s.workspaces.Delete(s.ctx(), id)
}

// DuplicateWorkspace は WS を複製する。
//
// name は複製先 WS 名。
// 空文字の場合はコピー元の名前を使用する。
func (s *StoreService) DuplicateWorkspace(id, name string) (*model.WorkspaceDTO, error) {
	return s.workspaces.Duplicate(s.ctx(), id, name)
}

// GetNodeResult はノード結果を返す。
func (s *StoreService) GetNodeResult(workspaceID, nodeID string) (*model.CrawlResultDTO, error) {
	return s.results.GetNodeResult(s.ctx(), workspaceID, nodeID)
}

// GetNodeResults は複数ノード結果を返す。
func (s *StoreService) GetNodeResults(workspaceID string, nodeIDs []string) ([]model.CrawlResultDTO, error) {
	return s.results.GetNodeResults(s.ctx(), workspaceID, nodeIDs)
}

const topicNodeResultUpdated = "node-result:updated"

// UpdateNodeResult はノード結果の手動編集を保存する。
func (s *StoreService) UpdateNodeResult(req model.UpdateNodeResultRequest) (*model.CrawlResultDTO, error) {
	dto, err := s.results.UpdateNodeResult(s.ctx(), req)
	if err != nil {
		return nil, err
	}
	if s.app != nil && dto != nil {
		s.app.Event.Emit(topicNodeResultUpdated, model.NodeResultUpdatedEvent{
			WorkspaceID: req.WorkspaceID,
			NodeID:      req.NodeID,
			Result:      *dto,
		})
	}
	return dto, err
}

// ShowMaximizedNodeResult は別 WebviewWindow でノード結果を拡大表示する。
func (s *StoreService) ShowMaximizedNodeResult(req model.MaximizedNodeResultRequest) error {
	if s.nodeResultWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.nodeResultWin.Show(req)
}

// GetMaximizedNodeResult は最大化ウィンドウ用の直近スナップショットを返す。
func (s *StoreService) GetMaximizedNodeResult() (model.MaximizedNodeResultRequest, error) {
	if s.nodeResultWin == nil {
		return model.MaximizedNodeResultRequest{}, fmt.Errorf("app not initialized")
	}
	return s.nodeResultWin.GetSnapshot()
}

// ShowExportWindow は別 WebviewWindow でエクスポート画面を表示する。
func (s *StoreService) ShowExportWindow(req model.ExportSessionRequest) error {
	if s.exportWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.exportWin.Show(req)
}

// GetExportSession はエクスポートウィンドウ用の直近スナップショットを返す。
func (s *StoreService) GetExportSession() (model.ExportSessionRequest, error) {
	if s.exportWin == nil {
		return model.ExportSessionRequest{}, fmt.Errorf("app not initialized")
	}
	return s.exportWin.GetSnapshot()
}

// SaveExportFile はエクスポート本文をファイルに保存する。
//
// defaultExt はダイアログの既定拡張子（"md" または "html"）。
func (s *StoreService) SaveExportFile(content string, defaultExt string) error {
	if s.app == nil {
		return fmt.Errorf("app not initialized")
	}
	ext := strings.TrimPrefix(strings.ToLower(defaultExt), ".")
	if ext == "" {
		ext = "md"
	}
	filterName := "Markdown"
	filterPattern := "*.md"
	defaultName := "export.md"
	if ext == "html" {
		filterName = "HTML"
		filterPattern = "*.html"
		defaultName = "export.html"
	}
	path, err := s.app.Dialog.SaveFile().
		SetMessage("Save export").
		SetFilename(defaultName).
		AddFilter(filterName, filterPattern).
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return err
	}
	if filepath.Ext(path) == "" {
		path += "." + ext
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

// SaveExportZip は複数ファイルを ZIP にまとめて保存する。
//
// defaultExt はダイアログ表示用のヒント（"md" または "html"）。
// ZIP 内のファイル名は entries の Name をそのまま使う。
func (s *StoreService) SaveExportZip(entries []model.ExportZipEntryDTO, defaultExt string) error {
	_ = defaultExt
	if s.app == nil {
		return fmt.Errorf("app not initialized")
	}
	if len(entries) == 0 {
		return fmt.Errorf("no export entries")
	}
	path, err := s.app.Dialog.SaveFile().
		SetMessage("Save export ZIP").
		SetFilename("export.zip").
		AddFilter("ZIP archive", "*.zip").
		AddFilter("All Files", "*.*").
		PromptForSingleSelection()
	if err != nil || path == "" {
		return err
	}
	if filepath.Ext(path) == "" {
		path += ".zip"
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for _, entry := range entries {
		if entry.Name == "" {
			continue
		}
		hdr := &zip.FileHeader{
			Name:   entry.Name,
			Method: zip.Deflate,
		}
		writer, err := w.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := writer.Write([]byte(entry.Content)); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return f.Close()
}

// MergeResults は結果をマージする。
func (s *StoreService) MergeResults(workspaceID string, nodeIDs []string, formats []string) (model.MergeResultsResponseDTO, error) {
	var ids []string
	if len(nodeIDs) > 0 {
		ids = nodeIDs
	}
	return s.results.MergeResults(s.ctx(), workspaceID, ids, formats)
}

// SaveResults は baseline 用に結果を保存する。
func (s *StoreService) SaveResults(workspaceID string, nodeIDs []string) error {
	return s.results.SaveResults(s.ctx(), workspaceID, nodeIDs)
}

// DeleteResults は最新結果を削除する。
func (s *StoreService) DeleteResults(workspaceID string, nodeIDs []string) error {
	return s.results.DeleteResults(s.ctx(), workspaceID, nodeIDs)
}

// SaveResultsSnapshot は baseline snapshot を保存する。
func (s *StoreService) SaveResultsSnapshot(workspaceID, runID string) (string, error) {
	return s.results.SaveResultsSnapshot(s.ctx(), workspaceID, runID)
}

// GetWorkspaceDiff は WS 差分を返す。
func (s *StoreService) GetWorkspaceDiff(workspaceID string) (model.WorkspaceDiffDTO, error) {
	return s.diff.GetWorkspaceDiff(s.ctx(), workspaceID)
}

// GetNodeDiffDetail は単一ノードの差分詳細を返す。
func (s *StoreService) GetNodeDiffDetail(workspaceID, nodeID string) (model.NodeDiffDetailDTO, error) {
	return s.diff.GetNodeDiffDetail(s.ctx(), workspaceID, nodeID)
}

// ShowNodeDiffWindow は別 WebviewWindow でノード差分を表示する。
func (s *StoreService) ShowNodeDiffWindow(req model.NodeDiffViewerRequest) error {
	if s.nodeDiffWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.nodeDiffWin.Show(req)
}

// GetNodeDiffViewerSession は差分ビューアウィンドウ用の直近スナップショットを返す。
func (s *StoreService) GetNodeDiffViewerSession() (model.NodeDiffViewerRequest, error) {
	if s.nodeDiffWin == nil {
		return model.NodeDiffViewerRequest{}, fmt.Errorf("app not initialized")
	}
	return s.nodeDiffWin.GetSnapshot()
}

// BeginCrawlRun は crawl run を開始する。
func (s *StoreService) BeginCrawlRun(req model.BeginCrawlRunRequest) error {
	return s.crawlPersist.BeginCrawlRun(s.ctx(), req)
}

// FinishCrawlRun は crawl run を終了する。
func (s *StoreService) FinishCrawlRun(req model.FinishCrawlRunRequest) error {
	return s.crawlPersist.FinishCrawlRun(s.ctx(), req)
}

// AppendNodeResult はノード結果行を追加する。
func (s *StoreService) AppendNodeResult(req model.AppendNodeResultRequest) error {
	return s.results.AppendNodeResultRow(s.ctx(), req)
}

// PatchGraphNodeStatus はノード status を更新する。
func (s *StoreService) PatchGraphNodeStatus(req model.PatchGraphNodeStatusRequest) error {
	return s.crawlPersist.PatchGraphNodeStatus(s.ctx(), req)
}

// PatchGraphNodePositions はノード座標を部分更新する。
func (s *StoreService) PatchGraphNodePositions(req model.PatchGraphNodePositionsRequest) error {
	return s.workspaces.PatchGraphNodePositions(s.ctx(), req)
}

// UpsertDiscoveredGraph は crawl 中に発見したノードとエッジを永続化する。
func (s *StoreService) UpsertDiscoveredGraph(req model.UpsertDiscoveredGraphRequest) error {
	return s.crawlPersist.UpsertDiscoveredGraph(s.ctx(), req)
}

// Bootstrap は起動時 DB 初期化。
func (s *StoreService) Bootstrap() error {
	if err := s.appConfig.Bootstrap(s.ctx()); err != nil {
		return fmt.Errorf("bootstrap app config: %w", err)
	}
	return nil
}
