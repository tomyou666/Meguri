package model

// WorkspaceBundle は WS 関連行の集合。
type WorkspaceBundle struct {
	Workspace Workspace
	Nodes     []GraphNode
	Edges     []GraphEdge
	UIState   *GraphUIState
	// Results は .crawlproj 交換用の最新成功結果（永続 Save では使わない）。
	Results []NodeResult
}

// WorkspaceListItem は WS 一覧用。
type WorkspaceListItem struct {
	ID        string
	Name      string
	UpdatedAt string
}

const MaxCrawlRunHistory = 20
const MaxNodeResultsPerNode = 20

// StrPtr は非空文字列のポインタを返す。
func StrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// StrVal は string ポインタを値にする。
func StrVal(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// Int32Ptr は int32 ポインタを返す。
func Int32Ptr(v int32) *int32 {
	return &v
}

// Int32Val は int32 ポインタを値にする。
func Int32Val(p *int32) int32 {
	if p == nil {
		return 0
	}
	return *p
}
