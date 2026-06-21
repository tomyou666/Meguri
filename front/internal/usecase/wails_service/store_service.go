package wails_service

import (
	"context"
	"encoding/json"
	"fmt"

	"scraperbot-front/internal/domain"
	"scraperbot-front/internal/model"
)

// StoreService は Wails 公開 Store RPC。
type StoreService struct {
	appConfig    *domain.AppConfigService
	workspaces   *domain.WorkspaceService
	results      *domain.ResultsService
	diff         *domain.DiffService
	crawlPersist *domain.CrawlPersistService
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
