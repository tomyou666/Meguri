package domain

import (
	"context"
	"encoding/json"
	"sort"

	"scraperbot-front/internal/infrastructure/persistence"
	"scraperbot-front/internal/model"
)

// DiffService は WS 差分計算。
type DiffService struct {
	repo persistence.Repository
	ws   *WorkspaceService
}

// NewDiffService は DiffService を構築する。
func NewDiffService(repo persistence.Repository, ws *WorkspaceService) *DiffService {
	return &DiffService{repo: repo, ws: ws}
}

// GetWorkspaceDiff は baseline vs current の差分を返す。
func (s *DiffService) GetWorkspaceDiff(ctx context.Context, workspaceID string) (model.WorkspaceDiffDTO, error) {
	dto, err := s.ws.Load(ctx, workspaceID)
	if err != nil || dto == nil {
		return model.WorkspaceDiffDTO{WorkspaceID: workspaceID}, err
	}
	out := model.WorkspaceDiffDTO{
		WorkspaceID:   workspaceID,
		BaselineRunID: dto.BaselineRunID,
	}
	if dto.BaselineRunID == "" {
		return out, nil
	}
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return out, err
	}
	baseline := rowsForRun(rows, dto.BaselineRunID)
	current := latestSuccessByNode(rows)
	latest := latestResultByNode(rows)

	for _, node := range dto.Nodes {
		var kinds []string
		base := baseline[node.ID]
		cur := current[node.ID]

		baseHash := ""
		if base.ContentHash != nil {
			baseHash = *base.ContentHash
		}
		curHash := ""
		if cur.ContentHash != nil {
			curHash = *cur.ContentHash
		}
		if baseHash != curHash {
			kinds = append(kinds, "content")
			out.Summary.Content++
		}

		baseLinks := linksFromRow(base)
		curLinks := linksFromRow(cur)
		if canonicalLinks(baseLinks) != canonicalLinks(curLinks) {
			kinds = append(kinds, "links")
			out.Summary.Links++
		}

		baseFetch := fetchState(base)
		curFetch := fetchState(latest[node.ID], node.Status)
		if baseFetch != curFetch {
			kinds = append(kinds, "fetch")
			out.Summary.Fetch++
		}
		if len(kinds) > 0 {
			out.Nodes = append(out.Nodes, model.NodeDiffDTO{
				NodeID: node.ID,
				URL:    node.URLNormalized,
				Kinds:  kinds,
			})
		}
	}
	out.HasDiff = len(out.Nodes) > 0
	return out, nil
}

// GetNodeDiffDetail は単一ノードの baseline vs current 差分詳細を返す。
func (s *DiffService) GetNodeDiffDetail(ctx context.Context, workspaceID, nodeID string) (model.NodeDiffDetailDTO, error) {
	dto, err := s.ws.Load(ctx, workspaceID)
	if err != nil || dto == nil {
		return model.NodeDiffDetailDTO{NodeID: nodeID}, err
	}
	out := model.NodeDiffDetailDTO{NodeID: nodeID}
	var nodeStatus string
	for _, node := range dto.Nodes {
		if node.ID == nodeID {
			out.URL = node.URLNormalized
			nodeStatus = node.Status
			break
		}
	}
	if dto.BaselineRunID == "" {
		return out, nil
	}
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return out, err
	}
	baseline := rowsForRun(rows, dto.BaselineRunID)
	current := latestSuccessByNode(rows)
	latest := latestResultByNode(rows)
	base := baseline[nodeID]
	cur := current[nodeID]

	var kinds []string
	baseHash := ""
	if base.ContentHash != nil {
		baseHash = *base.ContentHash
	}
	curHash := ""
	if cur.ContentHash != nil {
		curHash = *cur.ContentHash
	}
	if baseHash != curHash {
		kinds = append(kinds, "content")
		out.Content = &model.DiffPairDTO{
			Old: model.StrVal(base.Markdown),
			New: model.StrVal(cur.Markdown),
		}
	}

	baseLinks := linksFromRow(base)
	curLinks := linksFromRow(cur)
	if canonicalLinks(baseLinks) != canonicalLinks(curLinks) {
		kinds = append(kinds, "links")
		out.Links = &model.DiffPairDTO{
			Old: prettyLinksJSON(baseLinks),
			New: prettyLinksJSON(curLinks),
		}
	}

	baseFetch := fetchState(base)
	curFetch := fetchState(latest[nodeID], nodeStatus)
	if baseFetch != curFetch {
		kinds = append(kinds, "fetch")
		out.Fetch = &model.DiffPairDTO{Old: baseFetch, New: curFetch}
	}
	out.Kinds = sortDiffKinds(kinds)
	return out, nil
}

func sortDiffKinds(kinds []string) []string {
	order := map[string]int{"content": 0, "links": 1, "fetch": 2}
	cp := append([]string(nil), kinds...)
	sort.Slice(cp, func(i, j int) bool {
		return order[cp[i]] < order[cp[j]]
	})
	return cp
}

func prettyLinksJSON(links []string) string {
	if len(links) == 0 {
		return "[]"
	}
	cp := append([]string(nil), links...)
	sort.Strings(cp)
	b, _ := json.MarshalIndent(cp, "", "  ")
	return string(b)
}

func linksFromRow(r model.NodeResult) []string {
	if r.LinksJSON == nil || *r.LinksJSON == "" {
		return nil
	}
	var links []string
	_ = json.Unmarshal([]byte(*r.LinksJSON), &links)
	return links
}

func canonicalLinks(links []string) string {
	if len(links) == 0 {
		return "[]"
	}
	cp := append([]string(nil), links...)
	sort.Strings(cp)
	b, _ := json.Marshal(cp)
	return string(b)
}

func fetchState(r model.NodeResult, nodeStatus ...string) string {
	if model.StrVal(r.ID) != "" {
		if r.Error != nil && *r.Error != "" {
			return "error"
		}
		return "success"
	}
	if len(nodeStatus) > 0 && nodeStatus[0] == "skipped" {
		return "skipped"
	}
	return "none"
}
