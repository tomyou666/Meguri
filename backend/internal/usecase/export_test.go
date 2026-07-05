package usecase

import "meguri/internal/domain/model"

// CfgHashForTest はテスト用に cfgHash を公開する。
func CfgHashForTest(cfg *model.Config) (string, error) {
	return cfgHash(cfg)
}

// EntryCountForTest はキャッシュエントリ数を返す（テスト用）。
func (c *ScrapeCache) EntryCountForTest() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// NewScrapeCacheWithMaxForTest は maxEntries を指定して ScrapeCache を構築する（テスト用）。
func NewScrapeCacheWithMaxForTest(max int) *ScrapeCache {
	c := NewScrapeCache()
	c.maxEntries = max
	return c
}
