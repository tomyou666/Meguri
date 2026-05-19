package plugin

import (
	"context"

	"scraperbot/internal/domain/model"
)

// Parser (P5) はレスポンスをコンテンツ中間表現へ変換する。
// CanParse は Content-Type や URL 拡張子で処理可否を判定する。
type Parser interface {
	Plugin
	CanParse(res *model.Response) bool
	Parse(ctx context.Context, res *model.Response) (*model.Content, error)
}
