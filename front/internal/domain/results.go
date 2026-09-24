package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
)

// ResultsService は node_results / baseline 操作。
type ResultsService struct {
	repo persistence.Repository
	ws   *WorkspaceService
}

// NewResultsService は ResultsService を構築する。
func NewResultsService(repo persistence.Repository, ws *WorkspaceService) *ResultsService {
	return &ResultsService{repo: repo, ws: ws}
}

// GetNodeResult は最新成功結果（本文付き）を返す。
func (s *ResultsService) GetNodeResult(ctx context.Context, workspaceID, nodeID string) (*model.CrawlResultDTO, error) {
	rows, err := s.repo.GetNodeResultMetasByNodeIDs(ctx, workspaceID, []string{nodeID})
	if err != nil {
		return nil, err
	}
	row, ok := latestSuccessByNode(rows)[nodeID]
	if !ok {
		return nil, nil
	}
	var body *model.NodeResultBody
	if row.ID != nil {
		bodies, err := s.repo.GetNodeResultBodies(ctx, []string{*row.ID})
		if err != nil {
			return nil, err
		}
		if len(bodies) > 0 {
			body = &bodies[0]
		}
	}
	dto := nodeResultToPreview(row, body)
	return &dto, nil
}

// GetNodeResults は複数ノードの最新成功結果を返す。
//
// 返却順は nodeIDs の順。結果が無い ID はスキップする。
func (s *ResultsService) GetNodeResults(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.CrawlResultDTO, error) {
	out := []model.CrawlResultDTO{}
	if len(nodeIDs) == 0 {
		return out, nil
	}
	rows, err := s.repo.GetNodeResultMetasByNodeIDs(ctx, workspaceID, nodeIDs)
	if err != nil {
		return nil, err
	}
	byNode := latestSuccessByNode(rows)
	ids := make([]string, 0, len(byNode))
	for _, nodeID := range nodeIDs {
		row, ok := byNode[nodeID]
		if !ok || row.ID == nil {
			continue
		}
		ids = append(ids, *row.ID)
	}
	bodies, err := s.repo.GetNodeResultBodies(ctx, ids)
	if err != nil {
		return nil, err
	}
	byBody := bodiesByID(bodies)
	for _, nodeID := range nodeIDs {
		row, ok := byNode[nodeID]
		if !ok {
			continue
		}
		var body *model.NodeResultBody
		if row.ID != nil {
			if b, ok := byBody[*row.ID]; ok {
				body = &b
			}
		}
		out = append(out, nodeResultToPreview(row, body))
	}
	return out, nil
}

// UpdateNodeResult はノードの最新成功結果を手動編集で更新する。
func (s *ResultsService) UpdateNodeResult(
	ctx context.Context,
	req model.UpdateNodeResultRequest,
) (*model.CrawlResultDTO, error) {
	patch := model.NodeResultContentPatch{ManuallyEdited: true}
	hasField := false
	if req.Patch.Markdown != nil {
		patch.Markdown = req.Patch.Markdown
		hash := ContentHashFromMarkdown(*req.Patch.Markdown)
		patch.ContentHash = &hash
		hasField = true
	}
	if req.Patch.HTML != nil {
		patch.HTML = req.Patch.HTML
		hasField = true
	}
	if req.Patch.RawHTML != nil {
		patch.RawHTML = req.Patch.RawHTML
		hasField = true
	}
	if req.Patch.JSONBody != nil {
		patch.JSONBody = req.Patch.JSONBody
		hasField = true
	}
	if !hasField {
		return nil, fmt.Errorf("no editable fields in patch")
	}
	if err := s.repo.UpdateLatestNodeResult(ctx, req.WorkspaceID, req.NodeID, patch); err != nil {
		return nil, err
	}
	return s.GetNodeResult(ctx, req.WorkspaceID, req.NodeID)
}

// MergeResults は markdown を連結する。
func (s *ResultsService) MergeResults(ctx context.Context, workspaceID string, nodeIDs []string, formats []string) (model.MergeResultsResponseDTO, error) {
	if len(formats) == 0 {
		formats = []string{"markdown"}
	}
	dto, err := s.ws.Load(ctx, workspaceID)
	if err != nil || dto == nil {
		return model.MergeResultsResponseDTO{}, fmt.Errorf("workspace not found")
	}
	ids := nodeIDs
	if ids == nil {
		for _, n := range dto.Nodes {
			if n.Status == "success" {
				ids = append(ids, n.ID)
			}
		}
	}
	previews, err := s.GetNodeResults(ctx, workspaceID, ids)
	if err != nil {
		return model.MergeResultsResponseDTO{}, err
	}
	var parts []string
	for _, p := range previews {
		if contains(formats, "markdown") && p.Markdown != "" {
			parts = append(parts, fmt.Sprintf("## %s\n\n%s", p.URL, p.Markdown))
		}
	}
	return model.MergeResultsResponseDTO{
		Merged:    strings.Join(parts, "\n\n---\n\n"),
		Format:    "markdown",
		NodeCount: len(previews),
	}, nil
}

// contains は ss に v が含まれるかを返す。
func contains(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// SaveResults は baseline run に最新成功行をコピーする。
func (s *ResultsService) SaveResults(ctx context.Context, workspaceID string, nodeIDs []string) error {
	bundle, err := s.repo.LoadWorkspaceBundle(ctx, workspaceID)
	if err != nil || bundle == nil {
		return fmt.Errorf("workspace not found")
	}
	runID, err := s.ensureBaselineRun(ctx, bundle)
	if err != nil {
		return err
	}
	rows, err := s.repo.GetNodeResultMetas(ctx, workspaceID)
	if err != nil {
		return err
	}
	latest := latestSuccessByNode(rows)
	for _, nodeID := range nodeIDs {
		source, ok := latest[nodeID]
		if !ok {
			continue
		}
		if err := s.copyResultToRun(ctx, source, runID); err != nil {
			return err
		}
	}
	return nil
}

// DeleteResults は最新 1 行を削除する。
func (s *ResultsService) DeleteResults(ctx context.Context, workspaceID string, nodeIDs []string) error {
	return s.repo.DeleteLatestResults(ctx, workspaceID, nodeIDs)
}

// SaveResultsSnapshot は baseline_run_id を更新し結果を snapshot する。
//
// runID が空なら合成 run を作る。渡された runID が最新成功行の run と同じなら
// UNIQUE (run_id, node_id) を避けるためコピーせず、その run を baseline にする。
func (s *ResultsService) SaveResultsSnapshot(ctx context.Context, workspaceID, runID string) (string, error) {
	if runID == "" {
		runID = genID()
	}
	runs, err := s.repo.GetCrawlRuns(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	found := false
	for _, r := range runs {
		if model.StrVal(r.ID) == runID {
			found = true
			break
		}
	}
	if !found {
		now := time.Now().UTC().Format(time.RFC3339)
		if err := s.repo.BeginCrawlRun(ctx, model.CrawlRun{
			ID: model.StrPtr(runID), WorkspaceID: workspaceID, Mode: 1,
			Status: model.StrPtr("completed"), StartedAt: now, FinishedAt: &now,
		}); err != nil {
			return "", err
		}
	}
	rows, err := s.repo.GetNodeResultMetas(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	for _, source := range latestSuccessByNode(rows) {
		if err := s.copyResultToRun(ctx, source, runID); err != nil {
			return "", err
		}
	}
	if err := s.repo.SetBaselineRunID(ctx, workspaceID, runID); err != nil {
		return "", err
	}
	return runID, nil
}

// copyResultToRun は source を指定 run へコピーする。
//
// source が既に runID の行なら何もしない（初回 baseline はクロール run をそのまま使う）。
// 同じ (run_id, node_id) の別行があるときは削除してから挿入する。
func (s *ResultsService) copyResultToRun(ctx context.Context, source model.NodeResult, runID string) error {
	if source.RunID == runID {
		return nil
	}
	existing, err := s.repo.GetNodeResultMetasByNodeIDs(ctx, source.WorkspaceID, []string{source.NodeID})
	if err != nil {
		return err
	}
	for _, row := range existing {
		if row.RunID == runID && row.ID != nil {
			if err := s.repo.DeleteNodeResult(ctx, *row.ID); err != nil {
				return err
			}
		}
	}
	var body *model.NodeResultBody
	if source.ID != nil {
		bodies, err := s.repo.GetNodeResultBodies(ctx, []string{*source.ID})
		if err != nil {
			return err
		}
		if len(bodies) > 0 {
			b := bodies[0]
			body = &b
		}
	}
	copyMeta := source
	newID := genID()
	copyMeta.ID = model.StrPtr(newID)
	copyMeta.RunID = runID
	copyMeta.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	var copyBody *model.NodeResultBody
	if body != nil {
		b := *body
		b.ID = model.StrPtr(newID)
		copyBody = &b
	}
	return s.repo.AppendNodeResult(ctx, copyMeta, copyBody)
}

// ensureBaselineRun は baseline run が無ければ合成 run を作り、その ID を返す。
func (s *ResultsService) ensureBaselineRun(ctx context.Context, bundle *model.WorkspaceBundle) (string, error) {
	if bundle.Workspace.BaselineRunID != nil && *bundle.Workspace.BaselineRunID != "" {
		return *bundle.Workspace.BaselineRunID, nil
	}
	runID := genID()
	now := time.Now().UTC().Format(time.RFC3339)
	if err := s.repo.BeginCrawlRun(ctx, model.CrawlRun{
		ID: model.StrPtr(runID), WorkspaceID: model.StrVal(bundle.Workspace.ID), Mode: 1,
		Status: model.StrPtr("completed"), StartedAt: now, FinishedAt: &now,
	}); err != nil {
		return "", err
	}
	if err := s.repo.SetBaselineRunID(ctx, model.StrVal(bundle.Workspace.ID), runID); err != nil {
		return "", err
	}
	bundle.Workspace.BaselineRunID = &runID
	return runID, nil
}

// MarshalSummary は summary を JSON 文字列にする。
func MarshalSummary(v any) (*string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}
