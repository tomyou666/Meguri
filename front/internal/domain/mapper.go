package domain

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"meguri-app/internal/model"
)

// dtoToBundle は WorkspaceDTO を永続化バンドルへ変換する。
func dtoToBundle(dto model.WorkspaceDTO) (model.WorkspaceBundle, error) {
	settingsJSON, err := settingsJSONFromRaw(dto.Settings)
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
		SettingsJSON:         settingsJSON,
		ExcludeUrlsJSON:      string(exclude),
		GraphLayoutDirection: model.StrPtr(dto.GraphLayoutDirection),
		CreatedAt:            dto.CreatedAt,
	}
	if dto.BaselineRunID != "" {
		ws.BaselineRunID = &dto.BaselineRunID
	}

	nodes := make([]model.GraphNode, len(dto.Nodes))
	for i, n := range dto.Nodes {
		ns, err := settingsJSONFromRaw(n.NodeSettings)
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
		origin := n.Origin
		if origin == "" {
			origin = "crawl"
		}
		nodes[i] = model.GraphNode{
			WorkspaceID:      dto.ID,
			ID:               n.ID,
			URLNormalized:    n.URLNormalized,
			Label:            n.Label,
			PositionX:        n.Position.X,
			PositionY:        n.Position.Y,
			UserPositioned:   up,
			NodeSettingsJSON: ns,
			CrawlExclude:     ex,
			Origin:           origin,
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
		Workspace: ws,
		Nodes:     nodes,
		Edges:     edges,
		UIState:   ui,
	}, nil
}

// bundleToDTO は永続化バンドルと lastResult メタを WorkspaceDTO にする。
func bundleToDTO(bundle *model.WorkspaceBundle, previews map[string]*model.CrawlResultMetaDTO) (model.WorkspaceDTO, error) {
	ws := bundle.Workspace
	var exclude []string
	if err := json.Unmarshal([]byte(ws.ExcludeUrlsJSON), &exclude); err != nil {
		return model.WorkspaceDTO{}, err
	}

	nodes := make([]model.GraphNodeDTO, len(bundle.Nodes))
	for i, n := range bundle.Nodes {
		var preview *model.CrawlResultMetaDTO
		if previews != nil {
			preview = previews[n.ID]
		}
		lastErr := ""
		if n.LastError != nil {
			lastErr = *n.LastError
		}
		origin := n.Origin
		if origin == "" {
			origin = "crawl"
		}
		nodes[i] = model.GraphNodeDTO{
			ID:             n.ID,
			URLNormalized:  n.URLNormalized,
			Label:          n.Label,
			Position:       model.PositionDTO{X: n.PositionX, Y: n.PositionY},
			UserPositioned: n.UserPositioned == 1,
			NodeSettings:   json.RawMessage(n.NodeSettingsJSON),
			CrawlExclude:   n.CrawlExclude == 1,
			Origin:         origin,
			Status:         model.StrVal(n.Status),
			LastError:      lastErr,
			LastResult:     preview,
		}
	}

	edges := make([]model.GraphEdgeDTO, len(bundle.Edges))
	for i, e := range bundle.Edges {
		edges[i] = model.GraphEdgeDTO{ID: e.ID, Source: e.SourceNodeID, Target: e.TargetNodeID}
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

// latestSuccessByNode はノードごとの最新成功行を返す。
//
// rows は fetched_at DESC 想定。先頭の成功行が最新。
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

// latestResultByNode はノードごとの最新結果（成功/失敗問わず）を返す。
//
// rows は fetched_at DESC 想定。先頭行が最新。
func latestResultByNode(rows []model.NodeResult) map[string]model.NodeResult {
	out := map[string]model.NodeResult{}
	for _, r := range rows {
		if _, ok := out[r.NodeID]; !ok {
			out[r.NodeID] = r
		}
	}
	return out
}

// rowsForRun は指定 run のノードごとの行を返す。同一ノードは後勝ち。
func rowsForRun(rows []model.NodeResult, runID string) map[string]model.NodeResult {
	out := map[string]model.NodeResult{}
	for _, r := range rows {
		if r.RunID == runID {
			out[r.NodeID] = r
		}
	}
	return out
}

// nodeResultMetaToDTO はメタ行を CrawlResultMetaDTO にする。
func nodeResultMetaToDTO(row model.NodeResult) model.CrawlResultMetaDTO {
	return model.CrawlResultMetaDTO{
		URL:            row.URL,
		ContentHash:    model.StrVal(row.ContentHash),
		LinksHash:      model.StrVal(row.LinksHash),
		ManuallyEdited: row.ManuallyEdited != 0,
	}
}

// nodeResultToPreview はメタ + 本文から CrawlResultDTO を作る。
func nodeResultToPreview(meta model.NodeResult, body *model.NodeResultBody) model.CrawlResultDTO {
	dto := model.CrawlResultDTO{
		URL:            meta.URL,
		ManuallyEdited: meta.ManuallyEdited != 0,
	}
	if body == nil {
		return dto
	}
	if body.Markdown != nil {
		dto.Markdown = *body.Markdown
	}
	if body.HTML != nil {
		dto.HTML = *body.HTML
	}
	if body.RawHTML != nil {
		dto.RawHTML = *body.RawHTML
	}
	if body.JSONBody != nil {
		dto.JSONBody = *body.JSONBody
	}
	if body.LinksJSON != nil && *body.LinksJSON != "" {
		_ = json.Unmarshal([]byte(*body.LinksJSON), &dto.Links)
	}
	if body.MetadataJSON != nil && *body.MetadataJSON != "" {
		_ = json.Unmarshal([]byte(*body.MetadataJSON), &dto.Metadata)
	}
	return dto
}

// bodiesByID は本文行を id キーの map にする。
func bodiesByID(bodies []model.NodeResultBody) map[string]model.NodeResultBody {
	out := map[string]model.NodeResultBody{}
	for _, b := range bodies {
		if b.ID != nil {
			out[*b.ID] = b
		}
	}
	return out
}

// linksFromBody は本文の links_json を URL 配列にする。
func linksFromBody(body *model.NodeResultBody) []string {
	if body == nil || body.LinksJSON == nil || *body.LinksJSON == "" {
		return nil
	}
	var links []string
	_ = json.Unmarshal([]byte(*body.LinksJSON), &links)
	return links
}

// strPtr は空文字なら nil、それ以外はポインタを返す。
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// genID は時刻ミリ秒と乱数からなる ID を返す。
func genID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return fmt.Sprintf("%d-%s", time.Now().UnixMilli(), hex.EncodeToString(b[:]))
}
