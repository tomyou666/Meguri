package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"scraperbot-front/internal/infrastructure/persistence"
	"scraperbot-front/internal/model"
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

// GetNodeResult は最新成功結果を返す。
func (s *ResultsService) GetNodeResult(ctx context.Context, workspaceID, nodeID string) (*model.CrawlResultDTO, error) {
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	row, ok := latestSuccessByNode(rows)[nodeID]
	if !ok {
		return nil, nil
	}
	dto := nodeResultToPreview(row)
	return &dto, nil
}

// GetNodeResults は複数ノードの最新成功結果を返す。
func (s *ResultsService) GetNodeResults(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.CrawlResultDTO, error) {
	out := []model.CrawlResultDTO{}
	for _, nodeID := range nodeIDs {
		r, err := s.GetNodeResult(ctx, workspaceID, nodeID)
		if err != nil {
			return nil, err
		}
		if r != nil {
			out = append(out, *r)
		}
	}
	return out, nil
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
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return err
	}
	latest := latestSuccessByNode(rows)
	for _, nodeID := range nodeIDs {
		source, ok := latest[nodeID]
		if !ok {
			continue
		}
		copy := source
		copy.ID = model.StrPtr(genID())
		copy.RunID = runID
		copy.FetchedAt = time.Now().UTC().Format(time.RFC3339)
		if err := s.repo.AppendNodeResult(ctx, copy); err != nil {
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
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return "", err
	}
	for _, source := range latestSuccessByNode(rows) {
		copy := source
		copy.ID = model.StrPtr(genID())
		copy.RunID = runID
		copy.FetchedAt = time.Now().UTC().Format(time.RFC3339)
		if err := s.repo.AppendNodeResult(ctx, copy); err != nil {
			return "", err
		}
	}
	if err := s.repo.SetBaselineRunID(ctx, workspaceID, runID); err != nil {
		return "", err
	}
	return runID, nil
}

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

// AppendNodeResultRow は crawl 永続化用に結果行を追加する。
func (s *ResultsService) AppendNodeResultRow(ctx context.Context, req model.AppendNodeResultRequest) error {
	row := model.NodeResult{
		ID:          model.StrPtr(genID()),
		RunID:       req.RunID,
		WorkspaceID: req.WorkspaceID,
		NodeID:      req.NodeID,
		URL:         req.URL,
		FetchedAt:   req.FetchedAt,
	}
	if req.Markdown != "" {
		row.Markdown = &req.Markdown
	}
	if req.LinksJSON != "" {
		row.LinksJSON = &req.LinksJSON
	}
	if req.MetadataJSON != "" {
		row.MetadataJSON = &req.MetadataJSON
	}
	if req.Error != "" {
		row.Error = &req.Error
	}
	if req.ContentHash != "" {
		row.ContentHash = &req.ContentHash
	}
	return s.repo.AppendNodeResult(ctx, row)
}

// UpsertBaselineFromRow は baseline 行 upsert 用（未使用だが将来用）。
func (s *ResultsService) UpsertBaselineFromRow(_ context.Context, _ model.NodeResult) error {
	return nil
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
