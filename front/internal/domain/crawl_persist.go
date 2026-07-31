package domain

import (
	"context"
	"time"

	"meguri-app/internal/infrastructure/persistence"
	"meguri-app/internal/model"
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

// GetGraphNodeStatuses は指定ノードの status と lastError を返す。
//
// nodeIDs が空のときは空スライスを返す。
func (s *CrawlPersistService) GetGraphNodeStatuses(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.GraphNodeStatusDTO, error) {
	return s.repo.GetGraphNodeStatuses(ctx, workspaceID, nodeIDs)
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

// BuildSkipScrapeLinkMap は skip scrape 対象 URL の最新成功結果から outbound リンクマップを返す。
//
// skipURLs が空なら nil。結果に links が無い URL はマップに含めない。
func (s *CrawlPersistService) BuildSkipScrapeLinkMap(
	ctx context.Context,
	workspaceID string,
	skipURLs []string,
) (map[string][]string, error) {
	if len(skipURLs) == 0 {
		return nil, nil
	}
	want := make(map[string]struct{}, len(skipURLs))
	for _, u := range skipURLs {
		want[u] = struct{}{}
	}
	rows, err := s.repo.GetNodeResults(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]string)
	for _, row := range latestSuccessByNode(rows) {
		if _, ok := want[row.URL]; !ok {
			continue
		}
		links := linksFromRow(row)
		if len(links) == 0 {
			continue
		}
		out[row.URL] = links
	}
	return out, nil
}

// NowISO は現在時刻 ISO 文字列。
func NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
