package plugin

import (
	"context"

	"scraperbot/internal/domain/model"
)

// Filter (P7) は中間表現を絞り込む。HTML→Markdown 変換は行わず、構造の選別のみを担う。
type Filter interface {
	Plugin
	Filter(ctx context.Context, c *model.Content) (*model.Content, error)
}
