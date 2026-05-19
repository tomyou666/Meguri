package plugin

import (
	"context"

	"scraperbot/internal/domain/model"
)

// PreProcessor (P2) はリクエスト送信前のリクエスト改変を担う。
type PreProcessor interface {
	Plugin
	PreProcess(ctx context.Context, req *model.Request) error
}
