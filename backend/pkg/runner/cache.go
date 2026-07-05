package runner

import "meguri/internal/usecase"

// RunnerCache は cfg hash 単位で Kernel を再利用する LRU キャッシュ（usecase.ScrapeCache のエイリアス）。
type RunnerCache = usecase.ScrapeCache

// NewRunnerCache は RunnerCache を構築する。
func NewRunnerCache() *RunnerCache {
	return usecase.NewScrapeCache()
}
