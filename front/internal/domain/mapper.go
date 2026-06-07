package domain

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"scraperbot-front/internal/model"
)

func dtoToBundle(dto model.WorkspaceDTO) (model.WorkspaceBundle, error) {
	settings, err := json.Marshal(dto.Settings)
	if err != nil {
		return model.WorkspaceBundle{}, err
	}
	exclude, err := json.Marshal(dto.ExcludeURLs)
	if err != nil {
		return model.WorkspaceBundle{}, err
	}
	ws := model.Workspace{
		ID:                   model.StrPtr(dto.ID),
		Name:                 dto.Name,
		SeedURL:              dto.SeedURL,
		SettingsJSON:         string(settings),
		ExcludeUrlsJSON:      string(exclude),
		GraphLayoutDirection: model.StrPtr(dto.GraphLayoutDirection),
		CreatedAt:            dto.CreatedAt,
	}
	if dto.BaselineRunID != "" {
		ws.BaselineRunID = &dto.BaselineRunID
	}

	nodes := make([]model.GraphNode, len(dto.Nodes))
	for i, n := range dto.Nodes {
		ns, err := json.Marshal(n.NodeSettings)
		if err != nil {
			return model.WorkspaceBundle{}, err
		}
		var lastErr *string
		if n.LastError != "" {
			lastErr = &n.LastError
		}
		up := int32(0)
		if n.UserPositioned {
			up = 1
		}
		ex := int32(0)
		if n.CrawlExclude {
			ex = 1
		}
		nodes[i] = model.GraphNode{
			WorkspaceID:      dto.ID,
			ID:               n.ID,
			URLNormalized:    n.URLNormalized,
			Label:            n.Label,
			PositionX:        n.Position.X,
			PositionY:        n.Position.Y,
			UserPositioned:   up,
			NodeSettingsJSON: string(ns),
			CrawlExclude:     ex,
			Status:           model.StrPtr(n.Status),
			LastError:        lastErr,
		}
	}

	edges := make([]model.GraphEdge, len(dto.Edges))
	for i, e := range dto.Edges {
		edges[i] = model.GraphEdge{
			WorkspaceID:  dto.ID,
			ID:           e.ID,
			SourceNodeID: e.Source,
			TargetNodeID: e.Target,
		}
	}

	domains := make([]model.DomainSetting, 0, len(dto.DomainSettings))
	for host, raw := range dto.DomainSettings {
		domains = append(domains, model.DomainSetting{
			WorkspaceID:  dto.ID,
			Host:         host,
			SettingsJSON: string(raw),
		})
	}

	uiJSON, err := json.Marshal(map[string][]string{
		"collapsed":      dto.CollapsedNodeIDs,
		"expandedDetail": dto.ExpandedDetailNodeIDs,
	})
	if err != nil {
		return model.WorkspaceBundle{}, err
	}
	ui := &model.GraphUIState{
		WorkspaceID:          model.StrPtr(dto.ID),
		CollapsedNodeIdsJSON: string(uiJSON),
	}

	return model.WorkspaceBundle{
		Workspace:      ws,
		Nodes:          nodes,
		Edges:          edges,
		DomainSettings: domains,
		UIState:        ui,
	}, nil
}

func bundleToDTO(bundle *model.WorkspaceBundle, previews map[string]*model.CrawlResultDTO) (model.WorkspaceDTO, error) {
	ws := bundle.Workspace
	var exclude []string
	if err := json.Unmarshal([]byte(ws.ExcludeUrlsJSON), &exclude); err != nil {
		return model.WorkspaceDTO{}, err
	}

	nodes := make([]model.GraphNodeDTO, len(bundle.Nodes))
	for i, n := range bundle.Nodes {
		var preview *model.CrawlResultDTO
		if previews != nil {
			preview = previews[n.ID]
		}
		lastErr := ""
		if n.LastError != nil {
			lastErr = *n.LastError
		}
		nodes[i] = model.GraphNodeDTO{
			ID:             n.ID,
			URLNormalized:  n.URLNormalized,
			Label:          n.Label,
			Position:       model.PositionDTO{X: n.PositionX, Y: n.PositionY},
			UserPositioned: n.UserPositioned == 1,
			NodeSettings:   json.RawMessage(n.NodeSettingsJSON),
			CrawlExclude:   n.CrawlExclude == 1,
			Status:         model.StrVal(n.Status),
			LastError:      lastErr,
			LastResult:     preview,
		}
	}

	edges := make([]model.GraphEdgeDTO, len(bundle.Edges))
	for i, e := range bundle.Edges {
		edges[i] = model.GraphEdgeDTO{ID: e.ID, Source: e.SourceNodeID, Target: e.TargetNodeID}
	}

	domainMap := make(map[string]json.RawMessage, len(bundle.DomainSettings))
	for _, d := range bundle.DomainSettings {
		domainMap[d.Host] = json.RawMessage(d.SettingsJSON)
	}

	dto := model.WorkspaceDTO{
		ID:                   model.StrVal(ws.ID),
		Name:                 ws.Name,
		SeedURL:              ws.SeedURL,
		Settings:             json.RawMessage(ws.SettingsJSON),
		ExcludeURLs:          exclude,
		Nodes:                nodes,
		Edges:                edges,
		GraphLayoutDirection: model.StrVal(ws.GraphLayoutDirection),
		DomainSettings:       domainMap,
		CreatedAt:            ws.CreatedAt,
	}
	if ws.BaselineRunID != nil {
		dto.BaselineRunID = *ws.BaselineRunID
	}
	if bundle.UIState != nil {
		var ui struct {
			Collapsed      []string `json:"collapsed"`
			ExpandedDetail []string `json:"expandedDetail"`
		}
		if err := json.Unmarshal([]byte(bundle.UIState.CollapsedNodeIdsJSON), &ui); err == nil {
			dto.CollapsedNodeIDs = ui.Collapsed
			dto.ExpandedDetailNodeIDs = ui.ExpandedDetail
		}
	}
	return dto, nil
}

func latestSuccessByNode(rows []model.NodeResult) map[string]model.NodeResult {
	out := map[string]model.NodeResult{}
	for _, r := range rows {
		if r.Error != nil && *r.Error != "" {
			continue
		}
		if _, ok := out[r.NodeID]; !ok {
			out[r.NodeID] = r
		}
	}
	return out
}

func rowsForRun(rows []model.NodeResult, runID string) map[string]model.NodeResult {
	out := map[string]model.NodeResult{}
	for _, r := range rows {
		if r.RunID == runID {
			out[r.NodeID] = r
		}
	}
	return out
}

func nodeResultToPreview(row model.NodeResult) model.CrawlResultDTO {
	dto := model.CrawlResultDTO{URL: row.URL}
	if row.Markdown != nil {
		dto.Markdown = *row.Markdown
	}
	if row.LinksJSON != nil && *row.LinksJSON != "" {
		_ = json.Unmarshal([]byte(*row.LinksJSON), &dto.Links)
	}
	if row.MetadataJSON != nil && *row.MetadataJSON != "" {
		_ = json.Unmarshal([]byte(*row.MetadataJSON), &dto.Metadata)
	}
	return dto
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func genID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), hex.EncodeToString(b[:]))
}
