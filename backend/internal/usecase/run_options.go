package usecase

import (
	"meguri/internal/core"
	"meguri/internal/core/fetchlimit"
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
}
