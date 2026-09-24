package persistence

import (
	"context"

	"meguri-app/internal/model"
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

	// GetNodeResultMetas は WS の全 node_results メタを返す。
	GetNodeResultMetas(ctx context.Context, workspaceID string) ([]model.NodeResult, error)
	// GetNodeResultMetasByNodeIDs は指定ノードのメタを fetched_at 降順で返す。
	GetNodeResultMetasByNodeIDs(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.NodeResult, error)
	// GetNodeResultBodies は result id 指定で本文を返す。
	GetNodeResultBodies(ctx context.Context, ids []string) ([]model.NodeResultBody, error)
	// AppendNodeResult はメタを追加し、成功行なら本文も書く。
	AppendNodeResult(ctx context.Context, meta model.NodeResult, body *model.NodeResultBody) error
	// UpdateLatestNodeResult は最新成功行のメタ／本文を部分更新する。
	UpdateLatestNodeResult(ctx context.Context, workspaceID, nodeID string, patch model.NodeResultContentPatch) error
	// DeleteNodeResult は result id 指定でメタ行を削除する。本文は FK CASCADE。
	DeleteNodeResult(ctx context.Context, id string) error
	// DeleteLatestResults は各ノードの最新 1 行を削除する。
	DeleteLatestResults(ctx context.Context, workspaceID string, nodeIDs []string) error
	// TrimNodeResults はノードごとの履歴を keep 件に切り詰める。
	TrimNodeResults(ctx context.Context, workspaceID, nodeID string, keep int) error

	GetCrawlRuns(ctx context.Context, workspaceID string) ([]model.CrawlRun, error)
	BeginCrawlRun(ctx context.Context, run model.CrawlRun) error
	FinishCrawlRun(ctx context.Context, runID string, status string, finishedAt string, summaryJSON, errorMessage *string) error
	TrimCrawlRuns(ctx context.Context, workspaceID string, keep int) error

	PatchGraphNodeStatus(ctx context.Context, workspaceID, nodeID, status string, lastError *string) error
	GetGraphNodeStatuses(ctx context.Context, workspaceID string, nodeIDs []string) ([]model.GraphNodeStatusDTO, error)
	PatchGraphNodePositions(ctx context.Context, workspaceID string, updates []model.NodePositionPatchDTO) error
	UpsertDiscoveredGraph(ctx context.Context, workspaceID, sourceNodeID, targetNodeID, targetURL string) error
	SetBaselineRunID(ctx context.Context, workspaceID, runID string) error
}
