package domain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
)

// WorkspaceService はワークスペース CRUD。
type WorkspaceService struct {
	repo persistence.Repository
}

// NewWorkspaceService は WorkspaceService を構築する。
func NewWorkspaceService(repo persistence.Repository) *WorkspaceService {
	return &WorkspaceService{repo: repo}
}

// List は WS 一覧を返す。
func (s *WorkspaceService) List(ctx context.Context) ([]model.WorkspaceListItemDTO, error) {
	items, err := s.repo.ListWorkspaces(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]model.WorkspaceListItemDTO, len(items))
	for i, it := range items {
		out[i] = model.WorkspaceListItemDTO(it)
	}
	return out, nil
}

// Load は WS DTO を返す。
func (s *WorkspaceService) Load(ctx context.Context, id string) (*model.WorkspaceDTO, error) {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, id)
	if err != nil || bundle == nil {
		return nil, err
	}
	rows, err := s.repo.GetNodeResults(ctx, id)
	if err != nil {
		return nil, err
	}
	previews := map[string]*model.CrawlResultDTO{}
	for nodeID, row := range latestSuccessByNode(rows) {
		p := nodeResultToPreview(row)
		previews[nodeID] = &p
	}
	dto, err := bundleToDTO(bundle, previews)
	if err != nil {
		return nil, err
	}
	return &dto, nil
}

// Save は WS DTO を保存する。
func (s *WorkspaceService) Save(ctx context.Context, dto model.WorkspaceDTO) error {
	bundle, err := dtoToBundle(dto)
	if err != nil {
		return err
	}
	return s.repo.SaveWorkspaceBundle(ctx, bundle)
}

// SaveWorkspaceSettings は WS 設定を部分更新する。
func (s *WorkspaceService) SaveWorkspaceSettings(ctx context.Context, workspaceID string, settings json.RawMessage) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, workspaceID)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	cur, err := unmarshalConfigMap(bundle.Workspace.SettingsJSON)
	if err != nil {
		return err
	}
	patch, err := unmarshalConfigMap(string(settings))
	if err != nil {
		return err
	}
	for k, v := range patch {
		cur[k] = v
	}
	merged, err := json.Marshal(cur)
	if err != nil {
		return err
	}
	bundle.Workspace.SettingsJSON = string(merged)
	return s.repo.SaveWorkspaceBundle(ctx, *bundle)
}

// SaveNodeSettings はノード設定を置き換える。
//
// settings は PartialConfig JSON。空 / null は "{}" として保存する。
// 既存キーとのマージは行わない。
func (s *WorkspaceService) SaveNodeSettings(ctx context.Context, workspaceID, nodeID string, settings json.RawMessage) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, workspaceID)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	for i, n := range bundle.Nodes {
		if n.ID == nodeID {
			ns, err := settingsJSONFromRaw(settings)
			if err != nil {
				return err
			}
			if _, err := unmarshalConfigMap(ns); err != nil {
				return err
			}
			bundle.Nodes[i].NodeSettingsJSON = ns
			return s.repo.SaveWorkspaceBundle(ctx, *bundle)
		}
	}
	return fmt.Errorf("node not found")
}

// PatchGraphNodePositions はノード座標を部分更新する。
func (s *WorkspaceService) PatchGraphNodePositions(ctx context.Context, req model.PatchGraphNodePositionsRequest) error {
	return s.repo.PatchGraphNodePositions(ctx, req.WorkspaceID, req.Updates)
}

// Delete は WS を削除する。
func (s *WorkspaceService) Delete(ctx context.Context, id string) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, id)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	return s.repo.DeleteWorkspace(ctx, id)
}

// Duplicate は WS を複製する。
//
// req.Mode は複製範囲を表す。
// "full": 設定・ノード・エッジ・UIState をコピーする（結果はコピーしない。baseline はクリア）。
// "settings": settings と exclude_urls のみコピーし、req.SeedURL からシードノードを新規作成する。
//
// req.Name は複製先 WS 名。空文字の場合はコピー元の名前を使用する。
func (s *WorkspaceService) Duplicate(ctx context.Context, req model.DuplicateWorkspaceRequest) (*model.WorkspaceDTO, error) {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, req.ID)
	if err != nil || bundle == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	wsID := genID()
	copyName := req.Name
	if copyName == "" {
		copyName = bundle.Workspace.Name
	}
	bundle.Workspace.ID = model.StrPtr(wsID)
	bundle.Workspace.Name = copyName
	bundle.Workspace.BaselineRunID = nil
	bundle.Workspace.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	switch req.Mode {
	case "full":
		idMap := map[string]string{}
		for _, n := range bundle.Nodes {
			idMap[n.ID] = genID()
		}
		for i := range bundle.Nodes {
			old := bundle.Nodes[i].ID
			bundle.Nodes[i].WorkspaceID = wsID
			bundle.Nodes[i].ID = idMap[old]
			bundle.Nodes[i].Status = model.StrPtr("idle")
			bundle.Nodes[i].LastError = nil
		}
		for i := range bundle.Edges {
			bundle.Edges[i].WorkspaceID = wsID
			bundle.Edges[i].ID = fmt.Sprintf("e-%s-%s", idMap[bundle.Edges[i].SourceNodeID], idMap[bundle.Edges[i].TargetNodeID])
			bundle.Edges[i].SourceNodeID = idMap[bundle.Edges[i].SourceNodeID]
			bundle.Edges[i].TargetNodeID = idMap[bundle.Edges[i].TargetNodeID]
		}
		if bundle.UIState != nil {
			bundle.UIState.WorkspaceID = model.StrPtr(wsID)
		}
	case "settings":
		seedURL, err := NormalizeCrawlURL(req.SeedURL)
		if err != nil {
			return nil, fmt.Errorf("invalid seed URL: %w", err)
		}
		parsed, err := url.Parse(seedURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return nil, fmt.Errorf("invalid seed URL")
		}
		bundle.Workspace.SeedURL = seedURL
		nodeID := genID()
		bundle.Nodes = []model.GraphNode{
			{
				WorkspaceID:      wsID,
				ID:               nodeID,
				URLNormalized:    seedURL,
				Label:            seedURL,
				PositionX:        250,
				PositionY:        200,
				NodeSettingsJSON: `{}`,
				Origin:           "crawl",
				Status:           model.StrPtr("idle"),
			},
		}
		bundle.Edges = nil
		bundle.UIState = &model.GraphUIState{
			WorkspaceID:          model.StrPtr(wsID),
			CollapsedNodeIdsJSON: `{"collapsed":[],"expandedDetail":[]}`,
		}
	default:
		return nil, fmt.Errorf("invalid duplicate mode: %q", req.Mode)
	}

	if err := s.repo.SaveWorkspaceBundle(ctx, *bundle); err != nil {
		return nil, err
	}
	return s.Load(ctx, wsID)
}

// ImportBundle は新規 ID で WS をインポートする。
//
// bundle.Results がある場合は合成 crawl_run（status=completed）1本を作り、
// node_id を remap して結果行を挿入する。
// グラフ保存後に run／結果の挿入が失敗した場合は補償として WS を削除する。
// 補償削除も失敗したときは errors.Join で元エラーと削除エラーを返す。
func (s *WorkspaceService) ImportBundle(ctx context.Context, bundle model.WorkspaceBundle) (string, error) {
	wsID := genID()
	bundle.Workspace.ID = model.StrPtr(wsID)
	bundle.Workspace.BaselineRunID = nil
	bundle.Workspace.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	idMap := map[string]string{}
	for _, n := range bundle.Nodes {
		idMap[n.ID] = genID()
	}
	for i := range bundle.Nodes {
		old := bundle.Nodes[i].ID
		bundle.Nodes[i].WorkspaceID = wsID
		bundle.Nodes[i].ID = idMap[old]
	}
	for i := range bundle.Edges {
		bundle.Edges[i].WorkspaceID = wsID
		bundle.Edges[i].SourceNodeID = idMap[bundle.Edges[i].SourceNodeID]
		bundle.Edges[i].TargetNodeID = idMap[bundle.Edges[i].TargetNodeID]
		bundle.Edges[i].ID = fmt.Sprintf("e-%s-%s", bundle.Edges[i].SourceNodeID, bundle.Edges[i].TargetNodeID)
	}
	if bundle.UIState != nil {
		bundle.UIState.WorkspaceID = model.StrPtr(wsID)
	}
	results := bundle.Results
	bundle.Results = nil
	if err := s.repo.SaveWorkspaceBundle(ctx, bundle); err != nil {
		return "", err
	}
	if len(results) == 0 {
		return wsID, nil
	}
	if err := s.importBundleResults(ctx, wsID, idMap, results); err != nil {
		if delErr := s.repo.DeleteWorkspace(ctx, wsID); delErr != nil {
			return "", errors.Join(err, delErr)
		}
		return "", err
	}
	return wsID, nil
}

// importBundleResults は合成 crawl_run と結果行を挿入する。
func (s *WorkspaceService) importBundleResults(
	ctx context.Context,
	wsID string,
	idMap map[string]string,
	results []model.NodeResult,
) error {
	runID := genID()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.repo.BeginCrawlRun(ctx, model.CrawlRun{
		ID: model.StrPtr(runID), WorkspaceID: wsID, Mode: 1,
		Status: model.StrPtr("completed"), StartedAt: now, FinishedAt: &now,
	}); err != nil {
		return err
	}
	for _, source := range results {
		newNodeID, ok := idMap[source.NodeID]
		if !ok {
			continue
		}
		row := source
		row.ID = model.StrPtr(genID())
		row.RunID = runID
		row.WorkspaceID = wsID
		row.NodeID = newNodeID
		if err := s.repo.AppendNodeResult(ctx, row); err != nil {
			return err
		}
	}
	return nil
}

// ExportBundle はエクスポート用バンドルを返す（baseline なし）。
//
// includeResults が true のとき、ノードごとの最新成功結果を bundle.Results に載せる。
// false のときは結果を含めない（従来どおり）。
func (s *WorkspaceService) ExportBundle(ctx context.Context, id string, includeResults bool) (*model.WorkspaceBundle, error) {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, id)
	if err != nil || bundle == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	bundle.Workspace.BaselineRunID = nil
	if !includeResults {
		return bundle, nil
	}
	rows, err := s.repo.GetNodeResults(ctx, id)
	if err != nil {
		return nil, err
	}
	byNode := latestSuccessByNode(rows)
	if len(byNode) == 0 {
		return bundle, nil
	}
	results := make([]model.NodeResult, 0, len(byNode))
	for _, row := range byNode {
		results = append(results, row)
	}
	bundle.Results = results
	return bundle, nil
}
