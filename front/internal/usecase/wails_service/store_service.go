package wails_service

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"meguri-app/internal/domain"
	"meguri-app/internal/model"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:generate go tool gowrap gen -p meguri-app/internal/usecase/wails_service -i StoreServiceAPI -t templates/slog_debug_slow.gotmpl -o store_service_with_debug_log.go -v ServiceName=StoreService -v DecoratorName=StoreServiceWithDebugLog

// StoreServiceAPI は StoreService の公開メソッド面。
type StoreServiceAPI interface {
	// SetApp は Wails App を後から注入する（最大化・エクスポートウィンドウ用）。
	SetApp(app *application.App)
	// GetAppDefaults はアプリ既定設定を返す。
	GetAppDefaults() (json.RawMessage, error)
	// SetAppDefaults はアプリ既定設定を設定する。
	SetAppDefaults(config json.RawMessage) error
	// SaveAppDefaults はアプリ既定設定を保存する。
	SaveAppDefaults(config json.RawMessage) (model.SaveSettingsResponseDTO, error)
	// ListWorkspaces は WS 一覧を返す。
	ListWorkspaces() ([]model.WorkspaceListItemDTO, error)
	// LoadWorkspace は WS を読み込む。
	LoadWorkspace(id string) (*model.WorkspaceDTO, error)
	// SaveWorkspace は WS を保存する。
	SaveWorkspace(ws model.WorkspaceDTO) error
	// SaveWorkspaceSettings は WS 設定を保存する。
	SaveWorkspaceSettings(workspaceID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error)
	// SaveNodeSettings はノード設定を置き換えて保存する。
	SaveNodeSettings(workspaceID, nodeID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error)
	// DeleteWorkspace は WS を削除する。
	DeleteWorkspace(id string) error
	// DuplicateWorkspace は WS を複製する。
	DuplicateWorkspace(req model.DuplicateWorkspaceRequest) (*model.WorkspaceDTO, error)
	// GetNodeResult はノード結果を返す。
	GetNodeResult(workspaceID, nodeID string) (*model.CrawlResultDTO, error)
	// GetNodeResults は複数ノード結果を返す。
	GetNodeResults(workspaceID string, nodeIDs []string) ([]model.CrawlResultDTO, error)
	// UpdateNodeResult はノード結果の手動編集を保存する。
	UpdateNodeResult(req model.UpdateNodeResultRequest) (*model.CrawlResultDTO, error)
	// ShowMaximizedNodeResult は別 WebviewWindow でノード結果を拡大表示する。
	ShowMaximizedNodeResult(req model.MaximizedNodeResultRequest) error
	// GetMaximizedNodeResult は最大化ウィンドウ用の直近スナップショットを返す。
	GetMaximizedNodeResult() (model.MaximizedNodeResultRequest, error)
	// ShowExportWindow は別 WebviewWindow でエクスポート画面を表示する。
	ShowExportWindow(req model.ExportSessionRequest) error
	// GetExportSession はエクスポートウィンドウ用の直近スナップショットを返す。
	GetExportSession() (model.ExportSessionRequest, error)
	// SaveExportFile はエクスポート本文をファイルに保存する。
	SaveExportFile(content string, defaultExt string) error
	// SaveExportZip は複数ファイルを ZIP にまとめて保存する。
	SaveExportZip(entries []model.ExportZipEntryDTO, defaultExt string) error
	// MergeResults は結果をマージする。
	MergeResults(workspaceID string, nodeIDs []string, formats []string) (model.MergeResultsResponseDTO, error)
	// SaveResults は baseline 用に結果を保存する。
	SaveResults(workspaceID string, nodeIDs []string) error
	// DeleteResults は最新結果を削除する。
	DeleteResults(workspaceID string, nodeIDs []string) error
	// SaveResultsSnapshot は baseline snapshot を保存する。
	SaveResultsSnapshot(workspaceID, runID string) (string, error)
	// GetWorkspaceDiff は WS 差分を返す。
	GetWorkspaceDiff(workspaceID string) (model.WorkspaceDiffDTO, error)
	// GetNodeDiffDetail は単一ノードの差分詳細を返す。
	GetNodeDiffDetail(workspaceID, nodeID string) (model.NodeDiffDetailDTO, error)
	// ShowNodeDiffWindow は別 WebviewWindow でノード差分を表示する。
	ShowNodeDiffWindow(req model.NodeDiffViewerRequest) error
	// GetNodeDiffViewerSession は差分ビューアウィンドウ用の直近スナップショットを返す。
	GetNodeDiffViewerSession() (model.NodeDiffViewerRequest, error)
	// BeginCrawlRun は crawl run を開始する。
	BeginCrawlRun(req model.BeginCrawlRunRequest) error
	// FinishCrawlRun は crawl run を終了する。
	FinishCrawlRun(req model.FinishCrawlRunRequest) error
	// AppendNodeResult はノード結果行を追加する。
	AppendNodeResult(req model.AppendNodeResultRequest) error
	// PatchGraphNodeStatus はノード status を更新する。
	PatchGraphNodeStatus(req model.PatchGraphNodeStatusRequest) error
	// GetGraphNodeStatuses は指定ノードの status と lastError を返す。
	GetGraphNodeStatuses(workspaceID string, nodeIDs []string) ([]model.GraphNodeStatusDTO, error)
	// PatchGraphNodePositions はノード座標を部分更新する。
	PatchGraphNodePositions(req model.PatchGraphNodePositionsRequest) error
	// UpsertDiscoveredGraph は crawl 中に発見したノードとエッジを永続化する。
	UpsertDiscoveredGraph(req model.UpsertDiscoveredGraphRequest) error
	// Bootstrap は起動時 DB 初期化。
	Bootstrap() error
}

// StoreService は Wails 公開 Store RPC（debug ログ付き殻）。
type StoreService struct {
	StoreServiceWithDebugLog
}

// storeService は StoreService の実装本体。
type storeService struct {
	// app は Wails アプリ（イベント・ダイアログ用）。
	app *application.App
	// appConfig はアプリ既定設定サービス。
	appConfig *domain.AppConfigService
	// workspaces はワークスペースサービス。
	workspaces *domain.WorkspaceService
	// results はクロール結果サービス。
	results *domain.ResultsService
	// diff は差分サービス。
	diff *domain.DiffService
	// crawlPersist はクロール永続化サービス。
	crawlPersist *domain.CrawlPersistService
	// nodeResultWin はノード結果最大化ウィンドウ。
	nodeResultWin *NodeResultWindowManager
	// exportWin はエクスポートウィンドウ。
	exportWin *ExportWindowManager
	// nodeDiffWin はノード差分ウィンドウ。
	nodeDiffWin *NodeDiffWindowManager
}

// NewStoreService は StoreService を構築する。
func NewStoreService(
	appConfig *domain.AppConfigService,
	workspaces *domain.WorkspaceService,
	results *domain.ResultsService,
	diff *domain.DiffService,
	crawlPersist *domain.CrawlPersistService,
) *StoreService {
	impl := &storeService{
		appConfig:    appConfig,
		workspaces:   workspaces,
		results:      results,
		diff:         diff,
		crawlPersist: crawlPersist,
	}
	return &StoreService{StoreServiceWithDebugLog: NewStoreServiceWithDebugLog(impl)}
}

// SetApp は Wails App を後から注入する（最大化・エクスポートウィンドウ用）。
func (s *storeService) SetApp(app *application.App) {
	s.app = app
	s.nodeResultWin = NewNodeResultWindowManager(app)
	s.exportWin = NewExportWindowManager(app)
	s.nodeDiffWin = NewNodeDiffWindowManager(app)
}

// WireMainWindow はメインウィンドウ終了時のプレビュー連動を登録する。
func WireMainWindow(s *StoreService, w application.Window) {
	impl, ok := s._base.(*storeService)
	if !ok {
		return
	}
	if impl.nodeResultWin != nil {
		impl.nodeResultWin.SetMainWindow(w)
	}
	if impl.exportWin != nil {
		impl.exportWin.SetMainWindow(w)
	}
	if impl.nodeDiffWin != nil {
		impl.nodeDiffWin.SetMainWindow(w)
	}
}

func (s *storeService) ctx() context.Context { return context.Background() }

// GetAppDefaults はアプリ既定設定を返す。
func (s *storeService) GetAppDefaults() (json.RawMessage, error) {
	return s.appConfig.GetDefaults(s.ctx())
}

// SetAppDefaults はアプリ既定設定を設定する。
func (s *storeService) SetAppDefaults(config json.RawMessage) error {
	return s.appConfig.SaveDefaults(s.ctx(), config)
}

// SaveAppDefaults はアプリ既定設定を保存する。
func (s *storeService) SaveAppDefaults(config json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.appConfig.SaveDefaults(s.ctx(), config); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "app"}, nil
}

// ListWorkspaces は WS 一覧を返す。
func (s *storeService) ListWorkspaces() ([]model.WorkspaceListItemDTO, error) {
	return s.workspaces.List(s.ctx())
}

// LoadWorkspace は WS を読み込む。
func (s *storeService) LoadWorkspace(id string) (*model.WorkspaceDTO, error) {
	return s.workspaces.Load(s.ctx(), id)
}

// SaveWorkspace は WS を保存する。
func (s *storeService) SaveWorkspace(ws model.WorkspaceDTO) error {
	return s.workspaces.Save(s.ctx(), ws)
}

// SaveWorkspaceSettings は WS 設定を保存する。
func (s *storeService) SaveWorkspaceSettings(workspaceID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.workspaces.SaveWorkspaceSettings(s.ctx(), workspaceID, settings); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "workspace"}, nil
}

// SaveNodeSettings はノード設定を置き換えて保存する。
func (s *storeService) SaveNodeSettings(workspaceID, nodeID string, settings json.RawMessage) (model.SaveSettingsResponseDTO, error) {
	if err := s.workspaces.SaveNodeSettings(s.ctx(), workspaceID, nodeID, settings); err != nil {
		return model.SaveSettingsResponseDTO{}, err
	}
	return model.SaveSettingsResponseDTO{OK: true, Scope: "node"}, nil
}

// DeleteWorkspace は WS を削除する。
func (s *storeService) DeleteWorkspace(id string) error {
	return s.workspaces.Delete(s.ctx(), id)
}

// DuplicateWorkspace は WS を複製する。
//
// req.Mode は複製範囲を表す。
// "full": 設定・ノード・エッジ・UIState をコピーする。
// "settings": 設定と除外 URL のみコピーし、起点 URL からシードノードを新規作成する。
func (s *storeService) DuplicateWorkspace(req model.DuplicateWorkspaceRequest) (*model.WorkspaceDTO, error) {
	return s.workspaces.Duplicate(s.ctx(), req)
}

// GetNodeResult はノード結果を返す。
func (s *storeService) GetNodeResult(workspaceID, nodeID string) (*model.CrawlResultDTO, error) {
	return s.results.GetNodeResult(s.ctx(), workspaceID, nodeID)
}

// GetNodeResults は複数ノード結果を返す。
func (s *storeService) GetNodeResults(workspaceID string, nodeIDs []string) ([]model.CrawlResultDTO, error) {
	return s.results.GetNodeResults(s.ctx(), workspaceID, nodeIDs)
}

const topicNodeResultUpdated = "node-result:updated"

// UpdateNodeResult はノード結果の手動編集を保存する。
func (s *storeService) UpdateNodeResult(req model.UpdateNodeResultRequest) (*model.CrawlResultDTO, error) {
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
func (s *storeService) ShowMaximizedNodeResult(req model.MaximizedNodeResultRequest) error {
	if s.nodeResultWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.nodeResultWin.Show(req)
}

// GetMaximizedNodeResult は最大化ウィンドウ用の直近スナップショットを返す。
func (s *storeService) GetMaximizedNodeResult() (model.MaximizedNodeResultRequest, error) {
	if s.nodeResultWin == nil {
		return model.MaximizedNodeResultRequest{}, fmt.Errorf("app not initialized")
	}
	return s.nodeResultWin.GetSnapshot()
}

// ShowExportWindow は別 WebviewWindow でエクスポート画面を表示する。
func (s *storeService) ShowExportWindow(req model.ExportSessionRequest) error {
	if s.exportWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.exportWin.Show(req)
}

// GetExportSession はエクスポートウィンドウ用の直近スナップショットを返す。
func (s *storeService) GetExportSession() (model.ExportSessionRequest, error) {
	if s.exportWin == nil {
		return model.ExportSessionRequest{}, fmt.Errorf("app not initialized")
	}
	return s.exportWin.GetSnapshot()
}

// SaveExportFile はエクスポート本文をファイルに保存する。
//
// defaultExt はダイアログの既定拡張子（"md" または "html"）。
func (s *storeService) SaveExportFile(content string, defaultExt string) error {
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
func (s *storeService) SaveExportZip(entries []model.ExportZipEntryDTO, defaultExt string) error {
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
func (s *storeService) MergeResults(workspaceID string, nodeIDs []string, formats []string) (model.MergeResultsResponseDTO, error) {
	var ids []string
	if len(nodeIDs) > 0 {
		ids = nodeIDs
	}
	return s.results.MergeResults(s.ctx(), workspaceID, ids, formats)
}

// SaveResults は baseline 用に結果を保存する。
func (s *storeService) SaveResults(workspaceID string, nodeIDs []string) error {
	return s.results.SaveResults(s.ctx(), workspaceID, nodeIDs)
}

// DeleteResults は最新結果を削除する。
func (s *storeService) DeleteResults(workspaceID string, nodeIDs []string) error {
	return s.results.DeleteResults(s.ctx(), workspaceID, nodeIDs)
}

// SaveResultsSnapshot は baseline snapshot を保存する。
func (s *storeService) SaveResultsSnapshot(workspaceID, runID string) (string, error) {
	return s.results.SaveResultsSnapshot(s.ctx(), workspaceID, runID)
}

// GetWorkspaceDiff は WS 差分を返す。
func (s *storeService) GetWorkspaceDiff(workspaceID string) (model.WorkspaceDiffDTO, error) {
	return s.diff.GetWorkspaceDiff(s.ctx(), workspaceID)
}

// GetNodeDiffDetail は単一ノードの差分詳細を返す。
func (s *storeService) GetNodeDiffDetail(workspaceID, nodeID string) (model.NodeDiffDetailDTO, error) {
	return s.diff.GetNodeDiffDetail(s.ctx(), workspaceID, nodeID)
}

// ShowNodeDiffWindow は別 WebviewWindow でノード差分を表示する。
func (s *storeService) ShowNodeDiffWindow(req model.NodeDiffViewerRequest) error {
	if s.nodeDiffWin == nil {
		return fmt.Errorf("app not initialized")
	}
	return s.nodeDiffWin.Show(req)
}

// GetNodeDiffViewerSession は差分ビューアウィンドウ用の直近スナップショットを返す。
func (s *storeService) GetNodeDiffViewerSession() (model.NodeDiffViewerRequest, error) {
	if s.nodeDiffWin == nil {
		return model.NodeDiffViewerRequest{}, fmt.Errorf("app not initialized")
	}
	return s.nodeDiffWin.GetSnapshot()
}

// BeginCrawlRun は crawl run を開始する。
func (s *storeService) BeginCrawlRun(req model.BeginCrawlRunRequest) error {
	return s.crawlPersist.BeginCrawlRun(s.ctx(), req)
}

// FinishCrawlRun は crawl run を終了する。
func (s *storeService) FinishCrawlRun(req model.FinishCrawlRunRequest) error {
	return s.crawlPersist.FinishCrawlRun(s.ctx(), req)
}

// AppendNodeResult はノード結果行を追加する。
func (s *storeService) AppendNodeResult(req model.AppendNodeResultRequest) error {
	return s.crawlPersist.AppendNodeResult(s.ctx(), req)
}

// PatchGraphNodeStatus はノード status を更新する。
func (s *storeService) PatchGraphNodeStatus(req model.PatchGraphNodeStatusRequest) error {
	return s.crawlPersist.PatchGraphNodeStatus(s.ctx(), req)
}

// GetGraphNodeStatuses は指定ノードの status と lastError を返す。
func (s *storeService) GetGraphNodeStatuses(workspaceID string, nodeIDs []string) ([]model.GraphNodeStatusDTO, error) {
	return s.crawlPersist.GetGraphNodeStatuses(s.ctx(), workspaceID, nodeIDs)
}

// PatchGraphNodePositions はノード座標を部分更新する。
func (s *storeService) PatchGraphNodePositions(req model.PatchGraphNodePositionsRequest) error {
	return s.workspaces.PatchGraphNodePositions(s.ctx(), req)
}

// UpsertDiscoveredGraph は crawl 中に発見したノードとエッジを永続化する。
func (s *storeService) UpsertDiscoveredGraph(req model.UpsertDiscoveredGraphRequest) error {
	return s.crawlPersist.UpsertDiscoveredGraph(s.ctx(), req)
}

// Bootstrap は起動時 DB 初期化。
func (s *storeService) Bootstrap() error {
	if err := s.appConfig.Bootstrap(s.ctx()); err != nil {
		return fmt.Errorf("bootstrap app config: %w", err)
	}
	return nil
}
