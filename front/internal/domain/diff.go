package domain

import (
	"context"
	"encoding/json"
	"sort"

	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
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
	rows, err := s.repo.GetNodeResultMetas(ctx, workspaceID)
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

		baseHash := model.StrVal(base.ContentHash)
		curHash := model.StrVal(cur.ContentHash)
		if baseHash != curHash {
			kinds = append(kinds, "content")
			out.Summary.Content++
		}

		baseLinksHash := model.StrVal(base.LinksHash)
		curLinksHash := model.StrVal(cur.LinksHash)
		if baseLinksHash != curLinksHash {
			kinds = append(kinds, "links")
			out.Summary.Links++
		}

		baseFetch := fetchState(base)
		curFetch := fetchState(latest[node.ID])
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
	for _, node := range dto.Nodes {
		if node.ID == nodeID {
			out.URL = node.URLNormalized
			break
		}
	}
	if dto.BaselineRunID == "" {
		return out, nil
	}
	rows, err := s.repo.GetNodeResultMetas(ctx, workspaceID)
	if err != nil {
		return out, err
	}
	baseline := rowsForRun(rows, dto.BaselineRunID)
	current := latestSuccessByNode(rows)
	latest := latestResultByNode(rows)
	base := baseline[nodeID]
	cur := current[nodeID]

	bodyIDs := make([]string, 0, 2)
	if base.ID != nil {
		bodyIDs = append(bodyIDs, *base.ID)
	}
	if cur.ID != nil {
		bodyIDs = append(bodyIDs, *cur.ID)
	}
	bodies, err := s.repo.GetNodeResultBodies(ctx, bodyIDs)
	if err != nil {
		return out, err
	}
	byBody := bodiesByID(bodies)
	var baseBody, curBody *model.NodeResultBody
	if base.ID != nil {
		if b, ok := byBody[*base.ID]; ok {
			baseBody = &b
		}
	}
	if cur.ID != nil {
		if b, ok := byBody[*cur.ID]; ok {
			curBody = &b
		}
	}

	var kinds []string
	baseHash := model.StrVal(base.ContentHash)
	curHash := model.StrVal(cur.ContentHash)
	if baseHash != curHash {
		kinds = append(kinds, "content")
		out.Content = &model.DiffPairDTO{
			Old: model.StrVal(nilSafeMarkdown(baseBody)),
			New: model.StrVal(nilSafeMarkdown(curBody)),
		}
	}

	baseLinksHash := model.StrVal(base.LinksHash)
	curLinksHash := model.StrVal(cur.LinksHash)
	if baseLinksHash != curLinksHash {
		kinds = append(kinds, "links")
		baseLinks := linksFromBody(baseBody)
		curLinks := linksFromBody(curBody)
		out.Links = &model.DiffPairDTO{
			Old: prettyLinksJSON(baseLinks),
			New: prettyLinksJSON(curLinks),
		}
	}

	baseFetch := fetchState(base)
	curFetch := fetchState(latest[nodeID])
	if baseFetch != curFetch {
		kinds = append(kinds, "fetch")
		out.Fetch = &model.DiffPairDTO{Old: baseFetch, New: curFetch}
	}
	out.Kinds = sortDiffKinds(kinds)
	return out, nil
}

// nilSafeMarkdown は body が nil のとき nil、あれば Markdown を返す。
func nilSafeMarkdown(body *model.NodeResultBody) *string {
	if body == nil {
		return nil
	}
	return body.Markdown
}

// sortDiffKinds は content / links / fetch の順に kinds を並べる。
func sortDiffKinds(kinds []string) []string {
	order := map[string]int{"content": 0, "links": 1, "fetch": 2}
	cp := append([]string(nil), kinds...)
	sort.Slice(cp, func(i, j int) bool {
		return order[cp[i]] < order[cp[j]]
	})
	return cp
}

// prettyLinksJSON はリンク配列をソートしたインデント JSON にする。
func prettyLinksJSON(links []string) string {
	if len(links) == 0 {
		return "[]"
	}
	cp := append([]string(nil), links...)
	sort.Strings(cp)
	b, _ := json.MarshalIndent(cp, "", "  ")
	return string(b)
}

// fetchState は結果行から取得状態を返す。
// 結果行があれば error / success。結果なし（idle / skipped 含む）は none。
// SaveResultsSnapshot は成功結果のみコピーするため、結果なし skipped を別状態にすると
// 確認済み後も none≠skipped の fetch 差分が残る。
func fetchState(r model.NodeResult) string {
	if model.StrVal(r.ID) != "" {
		if r.Error != nil && *r.Error != "" {
			return "error"
		}
		return "success"
	}
	return "none"
}
