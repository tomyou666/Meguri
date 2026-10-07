package usecase

import (
	"meguri/internal/core"
	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
)

// PauseController はクロール一時停止制御（core.PauseController のエイリアス）。
type PauseController = core.PauseController

// NewPauseController は PauseController を構築する。
func NewPauseController() *PauseController {
	return core.NewPauseController()
}

// RunOptions は Crawl / Scrape の実行オプション。
type RunOptions struct {
	// Pause は一時停止制御。nil の場合は pause なし。
	Pause *PauseController
	// Cache は Scrape 用 Kernel キャッシュ。nil の場合は毎回 Init。
	Cache *ScrapeCache
	// FetchLimiter は取得並列上限。nil の場合は PrepareFetchLimiter が生成する。
	FetchLimiter *fetchlimit.FetchLimiter
	// PageConfig は BFS 中の 1 URL 向けスクレイプ設定。
	// nil を返すとラン共通 cfg のパイプラインを使う。
	// クロール上限（depth / pages など）は戻り値ではなく RunWithConfig に渡した cfg を使う。
	PageConfig func(rawURL string) (*model.Config, error)
}
