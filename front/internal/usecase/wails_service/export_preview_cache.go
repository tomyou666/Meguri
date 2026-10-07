package wails_service

import (
	"sync"

	"meguri-app/internal/domain/exportpreview"
)

const topicExportSaveProgress = "export:save-progress"

// exportPreviewRow はセッション中に保持する行オフセット。
type exportPreviewRow struct {
	// id は行 ID。
	id string
	// nodeID はグラフノード ID。
	nodeID string
	// url は基準 URL。
	url string
	// byteLength は断片バイト長。
	byteLength int
	// isFirst はそのノードの先頭行か。
	isFirst bool
	// start は元本文内の開始位置。
	start int
	// length は断片長。
	length int
}

// exportPreviewSession はプレビュー分割のセッションキャッシュ。
type exportPreviewSession struct {
	workspaceID string
	format      exportpreview.Format
	generation  int64
	nodeIDs     []string
	rows        []exportPreviewRow
}

// exportPreviewCache はプレビュー行メタのプロセス内キャッシュ。
type exportPreviewCache struct {
	mu      sync.Mutex
	session *exportPreviewSession
}

// clear はキャッシュを捨てる。
func (c *exportPreviewCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = nil
}

// store は新しいセッションを保存する。
func (c *exportPreviewCache) store(s *exportPreviewSession) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.session = s
}

// get は世代が一致するセッションを返す。
func (c *exportPreviewCache) get(workspaceID string, format exportpreview.Format, generation int64) *exportPreviewSession {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.session == nil {
		return nil
	}
	if c.session.workspaceID != workspaceID || c.session.format != format || c.session.generation != generation {
		return nil
	}
	return c.session
}
