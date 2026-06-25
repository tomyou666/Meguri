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
	ID                    string          `json:"id"`
	Name                  string          `json:"name"`
	SeedURL               string          `json:"seedUrl"`
	Settings              json.RawMessage `json:"settings"`
	ExcludeURLs           []string        `json:"exclude_urls"`
	Nodes                 []GraphNodeDTO  `json:"nodes"`
	Edges                 []GraphEdgeDTO  `json:"edges"`
	GraphLayoutDirection  string          `json:"graphLayoutDirection"`
	BaselineRunID         string          `json:"baselineRunId,omitempty"`
	CollapsedNodeIDs      []string        `json:"collapsedNodeIds,omitempty"`
	ExpandedDetailNodeIDs []string        `json:"expandedDetailNodeIds,omitempty"`
	CreatedAt             string          `json:"createdAt,omitempty"`
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
	URL            string            `json:"url"`
	Markdown       string            `json:"markdown,omitempty"`
	HTML           string            `json:"html,omitempty"`
	RawHTML        string            `json:"rawHtml,omitempty"`
	JSONBody       string            `json:"jsonBody,omitempty"`
	Links          []string          `json:"links,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	ManuallyEdited bool              `json:"manuallyEdited,omitempty"`
}

// UpdateNodeResultPatchDTO はノード結果の部分更新フィールド。
type UpdateNodeResultPatchDTO struct {
	Markdown *string `json:"markdown,omitempty"`
	HTML     *string `json:"html,omitempty"`
	RawHTML  *string `json:"rawHtml,omitempty"`
	JSONBody *string `json:"jsonBody,omitempty"`
}

// UpdateNodeResultRequest はノード結果手動編集の保存リクエスト。
type UpdateNodeResultRequest struct {
	WorkspaceID string                   `json:"workspaceId"`
	NodeID      string                   `json:"nodeId"`
	Patch       UpdateNodeResultPatchDTO `json:"patch"`
}

// MaximizedNodeResultRequest は最大化ウィンドウ表示用スナップショット。
type MaximizedNodeResultRequest struct {
	Title        string         `json:"title"`
	WorkspaceID  string         `json:"workspaceId"`
	NodeID       string         `json:"nodeId"`
	ActiveFormat string         `json:"activeFormat"`
	MarkdownView string         `json:"markdownView"`
	Formats      []string       `json:"formats"`
	Result       CrawlResultDTO `json:"result"`
}

// ExportSessionNodeDTO はエクスポートツリー用ノード。
type ExportSessionNodeDTO struct {
	// ID はグラフノード ID。
	ID string `json:"id"`
	// URLNormalized は正規化 URL。
	URLNormalized string `json:"urlNormalized"`
	// Label は表示ラベル。
	Label string `json:"label"`
	// Status はノードの crawl 状態。
	Status string `json:"status"`
}

// ExportSessionEdgeDTO はエクスポートツリー構築用エッジ。
type ExportSessionEdgeDTO struct {
	// Source は始点ノード ID。
	Source string `json:"source"`
	// Target は終点ノード ID。
	Target string `json:"target"`
}

// ExportSessionRequest はエクスポートウィンドウ表示用スナップショット。
type ExportSessionRequest struct {
	// Title はウィンドウタイトル。
	Title string `json:"title"`
	// WorkspaceID は対象ワークスペース ID。
	WorkspaceID string `json:"workspaceId"`
	// Mode はエクスポート対象の選び方。
	// "all": status=success の全ノード。
	// "selected": SelectedNodeIDs のみ。
	Mode string `json:"mode"`
	// SeedURL は親決定 BFS の起点 URL。
	SeedURL string `json:"seedUrl"`
	// Nodes はグラフノード一覧。
	Nodes []ExportSessionNodeDTO `json:"nodes"`
	// Edges はグラフエッジ一覧。
	Edges []ExportSessionEdgeDTO `json:"edges"`
	// SelectedNodeIDs は mode=selected 時の対象ノード ID。
	SelectedNodeIDs []string `json:"selectedNodeIds,omitempty"`
}

// ExportZipEntryDTO は ZIP エクスポート用の 1 ファイル分。
type ExportZipEntryDTO struct {
	// Name は ZIP 内のファイル名。
	Name string `json:"name"`
	// Content はファイル本文。
	Content string `json:"content"`
}

// NodeResultUpdatedEvent はノード結果手動編集後の同期イベント。
type NodeResultUpdatedEvent struct {
	WorkspaceID string         `json:"workspaceId"`
	NodeID      string         `json:"nodeId"`
	Result      CrawlResultDTO `json:"result"`
}

// NodeResultContentPatch は node_results 本文列の部分更新。
type NodeResultContentPatch struct {
	Markdown       *string
	HTML           *string
	RawHTML        *string
	JSONBody       *string
	ContentHash    *string
	ManuallyEdited bool
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
	HTML         string `json:"html,omitempty"`
	RawHTML      string `json:"rawHtml,omitempty"`
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

// NodePositionPatchDTO はノード座標の部分更新。
type NodePositionPatchDTO struct {
	NodeID         string      `json:"nodeId"`
	Position       PositionDTO `json:"position"`
	UserPositioned bool        `json:"userPositioned"`
}

// PatchGraphNodePositionsRequest はノード座標のバッチ部分更新。
type PatchGraphNodePositionsRequest struct {
	WorkspaceID string                 `json:"workspaceId"`
	Updates     []NodePositionPatchDTO `json:"updates"`
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
	RunID            string          `json:"runId"`
	WorkspaceID      string          `json:"workspaceId"`
	Mode             int32           `json:"mode"`
	StartNodeID      string          `json:"startNodeId,omitempty"`
	NodeIDs          []string        `json:"nodeIds,omitempty"`
	RescrapeExisting bool            `json:"rescrapeExisting"`
	AppDefaults      json.RawMessage `json:"appDefaults"`
	Workspace        WorkspaceDTO    `json:"workspace"`
}

// CrawlNodeResultDTO は Wails Event 用のノード結果プレビュー。
type CrawlNodeResultDTO struct {
	URL      string            `json:"url"`
	Markdown string            `json:"markdown,omitempty"`
	HTML     string            `json:"html,omitempty"`
	RawHTML  string            `json:"rawHtml,omitempty"`
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
	Mode                  int32  `json:"mode"`
	FinishedAt            string `json:"finishedAt"`
	Enqueued              int    `json:"enqueued"`
	Succeeded             int    `json:"succeeded"`
	Failed                int    `json:"failed"`
	Skipped               int    `json:"skipped"`
	SkippedDuplicateLinks int    `json:"skippedDuplicateLinks"`
	StoppedReason         string `json:"stoppedReason,omitempty"`
}

// RobotsTxtInfoDTO は robots.txt 取得結果。
type RobotsTxtInfoDTO struct {
	Host       string `json:"host"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
	Error      string `json:"error,omitempty"`
}
