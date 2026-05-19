package plugin

import (
	"context"
	"net/url"

	"scraperbot/internal/domain/model"
)

// LinkExtractor (P8) は中間表現からクロール候補リンクを抽出する。
type LinkExtractor interface {
	Plugin
	Extract(ctx context.Context, c *model.Content, baseURL *url.URL) ([]*url.URL, error)
}
