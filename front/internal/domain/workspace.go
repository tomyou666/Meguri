package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"scraperbot-front/internal/infrastructure/persistence"
	"scraperbot-front/internal/model"
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

// SaveDomainSettings はドメイン設定を更新する。
func (s *WorkspaceService) SaveDomainSettings(ctx context.Context, workspaceID, host string, settings json.RawMessage) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, workspaceID)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	found := false
	for i, d := range bundle.DomainSettings {
		if d.Host == host {
			cur, err := unmarshalConfigMap(d.SettingsJSON)
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
			bundle.DomainSettings[i].SettingsJSON = string(merged)
			found = true
			break
		}
	}
	if !found {
		settingsJSON, err := settingsJSONFromRaw(settings)
		if err != nil {
			return err
		}
		bundle.DomainSettings = append(bundle.DomainSettings, model.DomainSetting{
			WorkspaceID:  workspaceID,
			Host:         host,
			SettingsJSON: settingsJSON,
		})
	}
	return s.repo.SaveWorkspaceBundle(ctx, *bundle)
}

// SaveNodeSettings はノード設定を更新する。
func (s *WorkspaceService) SaveNodeSettings(ctx context.Context, workspaceID, nodeID string, settings json.RawMessage) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, workspaceID)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	for i, n := range bundle.Nodes {
		if n.ID == nodeID {
			cur, err := unmarshalConfigMap(n.NodeSettingsJSON)
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
			bundle.Nodes[i].NodeSettingsJSON = string(merged)
			return s.repo.SaveWorkspaceBundle(ctx, *bundle)
		}
	}
	return fmt.Errorf("node not found")
}

// Duplicate は WS を複製する。
func (s *WorkspaceService) Duplicate(ctx context.Context, id string) (*model.WorkspaceDTO, error) {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, id)
	if err != nil || bundle == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	wsID := genID()
	idMap := map[string]string{}
	for _, n := range bundle.Nodes {
		idMap[n.ID] = genID()
	}
	bundle.Workspace.ID = model.StrPtr(wsID)
	bundle.Workspace.Name = bundle.Workspace.Name + " (copy)"
	bundle.Workspace.BaselineRunID = nil
	bundle.Workspace.CreatedAt = time.Now().UTC().Format(time.RFC3339)
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
	for i := range bundle.DomainSettings {
		bundle.DomainSettings[i].WorkspaceID = wsID
	}
	if bundle.UIState != nil {
		bundle.UIState.WorkspaceID = model.StrPtr(wsID)
	}
	if err := s.repo.SaveWorkspaceBundle(ctx, *bundle); err != nil {
		return nil, err
	}
	return s.Load(ctx, wsID)
}

// ImportBundle は新規 ID で WS をインポートする。
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
	for i := range bundle.DomainSettings {
		bundle.DomainSettings[i].WorkspaceID = wsID
	}
	if bundle.UIState != nil {
		bundle.UIState.WorkspaceID = model.StrPtr(wsID)
	}
	if err := s.repo.SaveWorkspaceBundle(ctx, bundle); err != nil {
		return "", err
	}
	return wsID, nil
}

// ExportBundle はエクスポート用バンドルを返す（baseline / results なし）。
func (s *WorkspaceService) ExportBundle(ctx context.Context, id string) (*model.WorkspaceBundle, error) {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, id)
	if err != nil || bundle == nil {
		return nil, fmt.Errorf("workspace not found")
	}
	bundle.Workspace.BaselineRunID = nil
	return bundle, nil
}
