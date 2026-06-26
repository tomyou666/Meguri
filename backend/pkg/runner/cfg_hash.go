package runner

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"meguri/internal/domain/model"
)

// scrapeConfigFingerprint は Scrape 実行に影響する設定の正規化サブセット。
type scrapeConfigFingerprint struct {
	Request model.RequestConfig   `json:"request"`
	Content model.ContentConfig   `json:"content"`
	PDF     model.PDFConfig       `json:"pdf"`
	Plugins model.PluginSelection `json:"plugins"`
}

// cfgHash は scrape 用設定の SHA-256 十六進ハッシュを返す。
//
// Targets・ExcludeURLs・Crawl 巡回パラメータは hash 対象外。
func cfgHash(cfg *model.Config) (string, error) {
	if cfg == nil {
		return "", fmt.Errorf("config is nil")
	}
	fp := scrapeConfigFingerprint{
		Request: cfg.Request,
		Content: cfg.Content,
		PDF:     cfg.PDF,
		Plugins: cfg.Plugins,
	}
	data, err := json.Marshal(fp)
	if err != nil {
		return "", fmt.Errorf("marshal scrape config: %w", err)
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}
