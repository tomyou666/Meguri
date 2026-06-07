package runner

import "scraperbot/internal/core"

// PauseController はクロール一時停止制御（core.PauseController のエイリアス）。
type PauseController = core.PauseController

// NewPauseController は PauseController を構築する。
func NewPauseController() *PauseController {
	return core.NewPauseController()
}

// RunOptions は CrawlWithProgress / ScrapeWithConfig の実行オプション。
type RunOptions struct {
	// Pause は一時停止制御。nil の場合は pause なし。
	Pause *PauseController
	// Cache は ScrapeWithConfig 用 Kernel キャッシュ。nil の場合は毎回 Init。
	Cache *RunnerCache
}
