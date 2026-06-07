package model

import "encoding/json"

// PositionDTO はグラフノード座標。
type PositionDTO struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// GraphNodeDTO は Wails 公開ノード。
type GraphNodeDTO struct {
	ID             string          `json:"id"`
	URLNormalized  string          `json:"urlNormalized"`
	Label          string          `json:"label"`
	Position       PositionDTO     `json:"position"`
	UserPositioned bool            `json:"userPositioned,omitempty"`
	NodeSettings   json.RawMessage `json:"nodeSettings"`
	CrawlExclude   bool            `json:"crawlExclude"`
	Origin         string          `json:"origin"`
	Status         string          `json:"status"`
	LastError      string          `json:"lastError,omitempty"`
	LastResult     *CrawlResultDTO `json:"lastResult,omitempty"`
}

// GraphEdgeDTO は Wails 公開エッジ。
type GraphEdgeDTO struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

// WorkspaceDTO は Wails 公開ワークスペース。
type WorkspaceDTO struct {
	ID                    string                     `json:"id"`
	Name                  string                     `json:"name"`
	SeedURL               string                     `json:"seedUrl"`
	Settings              json.RawMessage            `json:"settings"`
	ExcludeURLs           []string                   `json:"exclude_urls"`
	Nodes                 []GraphNodeDTO             `json:"nodes"`
	Edges                 []GraphEdgeDTO             `json:"edges"`
	GraphLayoutDirection  string                     `json:"graphLayoutDirection"`
	DomainSettings        map[string]json.RawMessage `json:"domainSettings"`
	BaselineRunID         string                     `json:"baselineRunId,omitempty"`
	CollapsedNodeIDs      []string                   `json:"collapsedNodeIds,omitempty"`
	ExpandedDetailNodeIDs []string                   `json:"expandedDetailNodeIds,omitempty"`
	CreatedAt             string                     `json:"createdAt,omitempty"`
}

// WorkspaceListItemDTO は WS 一覧行。
type WorkspaceListItemDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	UpdatedAt string `json:"updatedAt"`
}

// SaveSettingsResponseDTO は設定保存レスポンス。
type SaveSettingsResponseDTO struct {
	OK    bool   `json:"ok"`
	Scope string `json:"scope"`
}

// CrawlResultDTO はノード結果プレビュー。
type CrawlResultDTO struct {
	URL      string            `json:"url"`
	Markdown string            `json:"markdown,omitempty"`
	Links    []string          `json:"links,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// MergeResultsResponseDTO はマージ結果。
type MergeResultsResponseDTO struct {
	Merged    string `json:"merged"`
	Format    string `json:"format"`
	NodeCount int    `json:"nodeCount"`
}

// NodeDiffDTO は差分ノード。
type NodeDiffDTO struct {
	NodeID string   `json:"nodeId"`
	URL    string   `json:"url"`
	Kinds  []string `json:"kinds"`
}

// WorkspaceDiffDTO は WS 差分。
type WorkspaceDiffDTO struct {
	WorkspaceID   string        `json:"workspaceId"`
	HasDiff       bool          `json:"hasDiff"`
	BaselineRunID string        `json:"baselineRunId,omitempty"`
	Nodes         []NodeDiffDTO `json:"nodes"`
	Summary       struct {
		Content int `json:"content"`
		Links   int `json:"links"`
		Fetch   int `json:"fetch"`
	} `json:"summary"`
}

// BeginCrawlRunRequest は crawl run 開始。
type BeginCrawlRunRequest struct {
	WorkspaceID string `json:"workspaceId"`
	RunID       string `json:"runId"`
	Mode        int32  `json:"mode"`
	StartedAt   string `json:"startedAt"`
}

// FinishCrawlRunRequest は crawl run 終了。
type FinishCrawlRunRequest struct {
	WorkspaceID  string          `json:"workspaceId"`
	RunID        string          `json:"runId"`
	Status       string          `json:"status"`
	FinishedAt   string          `json:"finishedAt"`
	SummaryJSON  json.RawMessage `json:"summaryJson,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

// AppendNodeResultRequest は結果行追加。
type AppendNodeResultRequest struct {
	WorkspaceID  string `json:"workspaceId"`
	RunID        string `json:"runId"`
	NodeID       string `json:"nodeId"`
	URL          string `json:"url"`
	Markdown     string `json:"markdown,omitempty"`
	LinksJSON    string `json:"linksJson,omitempty"`
	MetadataJSON string `json:"metadataJson,omitempty"`
	Error        string `json:"error,omitempty"`
	FetchedAt    string `json:"fetchedAt"`
	ContentHash  string `json:"contentHash,omitempty"`
}

// PatchGraphNodeStatusRequest はノード status 更新。
type PatchGraphNodeStatusRequest struct {
	WorkspaceID string `json:"workspaceId"`
	NodeID      string `json:"nodeId"`
	Status      string `json:"status"`
	LastError   string `json:"lastError,omitempty"`
}

// UpsertDiscoveredGraphRequest は crawl 中に発見したノードとエッジを永続化する。
type UpsertDiscoveredGraphRequest struct {
	WorkspaceID string `json:"workspaceId"`
	SourceID    string `json:"sourceId"`
	TargetID    string `json:"targetId"`
	TargetURL   string `json:"targetUrl"`
}

// OpenScrbResponse は .scrb インポート結果。
type OpenScrbResponse struct {
	WorkspaceID string `json:"workspaceId"`
}

// StartCrawlRequest はクロール開始 RPC 入力。
type StartCrawlRequest struct {
	RunID       string          `json:"runId"`
	WorkspaceID string          `json:"workspaceId"`
	Mode        int32           `json:"mode"`
	StartNodeID string          `json:"startNodeId,omitempty"`
	NodeIDs     []string        `json:"nodeIds,omitempty"`
	AppDefaults json.RawMessage `json:"appDefaults"`
	Workspace   WorkspaceDTO    `json:"workspace"`
}

// CrawlNodeResultDTO は Wails Event 用のノード結果プレビュー。
type CrawlNodeResultDTO struct {
	URL      string            `json:"url"`
	Markdown string            `json:"markdown,omitempty"`
	Links    []string          `json:"links,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// CrawlEventPayload は scraper:crawl:* Event の共通フィールド。
type CrawlEventPayload struct {
	WorkspaceID string              `json:"workspaceId"`
	RunID       string              `json:"runId"`
	NodeID      string              `json:"nodeId,omitempty"`
	URL         string              `json:"url,omitempty"`
	Result      *CrawlNodeResultDTO `json:"result,omitempty"`
	Error       string              `json:"error,omitempty"`
	Reason      string              `json:"reason,omitempty"`
	SourceID    string              `json:"sourceId,omitempty"`
	TargetID    string              `json:"targetId,omitempty"`
	TargetURL   string              `json:"targetUrl,omitempty"`
	Summary     *CrawlSummaryDTO    `json:"summary,omitempty"`
	Message     string              `json:"message,omitempty"`
}

// CrawlSummaryDTO は crawl 完了サマリ。
type CrawlSummaryDTO struct {
	Mode          int32  `json:"mode"`
	FinishedAt    string `json:"finishedAt"`
	Enqueued      int    `json:"enqueued"`
	Succeeded     int    `json:"succeeded"`
	Failed        int    `json:"failed"`
	Skipped       int    `json:"skipped"`
	StoppedReason string `json:"stoppedReason,omitempty"`
}
