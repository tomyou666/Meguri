package domain

import (
	"context"
	"time"

	"scraperbot-front/internal/infrastructure/persistence"
	"scraperbot-front/internal/model"
)

// CrawlPersistService は crawl run 永続化。
type CrawlPersistService struct {
	repo persistence.Repository
}

// NewCrawlPersistService は CrawlPersistService を構築する。
func NewCrawlPersistService(repo persistence.Repository) *CrawlPersistService {
	return &CrawlPersistService{repo: repo}
}

// BeginCrawlRun は crawl run を開始する。
func (s *CrawlPersistService) BeginCrawlRun(ctx context.Context, req model.BeginCrawlRunRequest) error {
	return s.repo.BeginCrawlRun(ctx, model.CrawlRun{
		ID:          model.StrPtr(req.RunID),
		WorkspaceID: req.WorkspaceID,
		Mode:        req.Mode,
		Status:      model.StrPtr("running"),
		StartedAt:   req.StartedAt,
	})
}

// FinishCrawlRun は crawl run を終了する。
func (s *CrawlPersistService) FinishCrawlRun(ctx context.Context, req model.FinishCrawlRunRequest) error {
	var summary, errMsg *string
	if len(req.SummaryJSON) > 0 {
		s := string(req.SummaryJSON)
		summary = &s
	}
	if req.ErrorMessage != "" {
		errMsg = &req.ErrorMessage
	}
	return s.repo.FinishCrawlRun(ctx, req.RunID, req.Status, req.FinishedAt, summary, errMsg)
}

// PatchGraphNodeStatus はノード status を更新する。
func (s *CrawlPersistService) PatchGraphNodeStatus(ctx context.Context, req model.PatchGraphNodeStatusRequest) error {
	return s.repo.PatchGraphNodeStatus(ctx, req.WorkspaceID, req.NodeID, req.Status, strPtr(req.LastError))
}

// UpsertDiscoveredGraph は crawl 中に発見したノードとエッジを永続化する。
func (s *CrawlPersistService) UpsertDiscoveredGraph(ctx context.Context, req model.UpsertDiscoveredGraphRequest) error {
	return s.repo.UpsertDiscoveredGraph(ctx, req.WorkspaceID, req.SourceID, req.TargetID, req.TargetURL)
}

// AppendNodeResult は crawl 中のノード結果行を追加する。
func (s *CrawlPersistService) AppendNodeResult(ctx context.Context, req model.AppendNodeResultRequest) error {
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
	if req.HTML != "" {
		row.HTML = &req.HTML
	}
	if req.RawHTML != "" {
		row.RawHTML = &req.RawHTML
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

// NowISO は現在時刻 ISO 文字列。
func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
