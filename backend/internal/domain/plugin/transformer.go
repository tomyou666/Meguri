package plugin

import (
	"context"

	"meguri/internal/domain/model"
)

// Transformer (P6) は中間表現を最終 Result に変換する。
type Transformer interface {
	Plugin
	Transform(ctx context.Context, c *model.Content) (*model.Result, error)
}
