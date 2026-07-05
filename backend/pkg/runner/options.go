package runner

import "meguri/internal/usecase"

// RunOptions は CrawlWithProgress / ScrapeWithConfig の実行オプション（usecase.RunOptions のエイリアス）。
type RunOptions = usecase.RunOptions

// PauseController はクロール一時停止制御（usecase.PauseController のエイリアス）。
type PauseController = usecase.PauseController

// NewPauseController は PauseController を構築する。
func NewPauseController() *PauseController {
	return usecase.NewPauseController()
}
