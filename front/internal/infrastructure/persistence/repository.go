package persistence

import (
	"context"

	"scraperbot-front/internal/model"
)

// Repository は永続化操作の interface。
type Repository interface {
	GetAppConfig(ctx context.Context) (*model.AppConfig, error)
	SaveAppConfig(ctx context.Context, defaultsJSON string) error
	BootstrapAppConfig(ctx context.Context) error

	ListWorkspaces(ctx context.Context) ([]model.WorkspaceListItem, error)
	LoadWorkspaceBundle(ctx context.Context, id string) (*model.WorkspaceBundle, error)
	SaveWorkspaceBundle(ctx context.Context, bundle model.WorkspaceBundle) error
	DeleteWorkspace(ctx context.Context, id string) error

	GetNodeResults(ctx context.Context, workspaceID string) ([]model.NodeResult, error)
	AppendNodeResult(ctx context.Context, row model.NodeResult) error
	DeleteLatestResults(ctx context.Context, workspaceID string, nodeIDs []string) error
	TrimNodeResults(ctx context.Context, workspaceID, nodeID string, keep int) error

	GetCrawlRuns(ctx context.Context, workspaceID string) ([]model.CrawlRun, error)
	BeginCrawlRun(ctx context.Context, run model.CrawlRun) error
	FinishCrawlRun(ctx context.Context, runID string, status string, finishedAt string, summaryJSON, errorMessage *string) error
	TrimCrawlRuns(ctx context.Context, workspaceID string, keep int) error

	PatchGraphNodeStatus(ctx context.Context, workspaceID, nodeID, status string, lastError *string) error
	PatchGraphNodePositions(ctx context.Context, workspaceID string, updates []model.NodePositionPatchDTO) error
	UpsertDiscoveredGraph(ctx context.Context, workspaceID, sourceNodeID, targetNodeID, targetURL string) error
	SetBaselineRunID(ctx context.Context, workspaceID, runID string) error
}
