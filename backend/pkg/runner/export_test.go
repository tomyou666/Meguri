package runner

// CfgHashForTest はテスト用に cfgHash を公開する。
func CfgHashForTest(cfg *Config) (string, error) {
	return cfgHash(cfg)
}

// EntryCountForTest はキャッシュエントリ数を返す（テスト用）。
func (c *RunnerCache) EntryCountForTest() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

// NewRunnerCacheWithMaxForTest は maxEntries を指定して RunnerCache を構築する（テスト用）。
func NewRunnerCacheWithMaxForTest(max int) *RunnerCache {
	c := NewRunnerCache()
	c.maxEntries = max
	return c
}
