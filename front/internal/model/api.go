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

// OpenScrbResponse は .scrb インポート結果。
type OpenScrbResponse struct {
	WorkspaceID string `json:"workspaceId"`
}
