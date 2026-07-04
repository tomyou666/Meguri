package wails_service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScraperService_ServiceShutdown はアプリ終了時の crawl 停止と待機を検証する。
func TestScraperService_ServiceShutdown(t *testing.T) {
	t.Run("正常系: active job なしでもエラーを返さない", func(t *testing.T) {
		s := NewScraperService(nil)
		require.NoError(t, s.ServiceShutdown())
	})

	t.Run("正常系: cancel 後に goroutine 終了を待ってから返る", func(t *testing.T) {
		s := NewScraperService(nil)
		done := make(chan struct{})

		s.mu.Lock()
		s.job = &activeCrawlJob{
			cancel: func() {
				go func() {
					time.Sleep(20 * time.Millisecond)
					s.releaseActiveJobResources()
					close(done)
				}()
			},
			done: done,
		}
		s.mu.Unlock()

		start := time.Now()
		require.NoError(t, s.ServiceShutdown())
		assert.GreaterOrEqual(t, time.Since(start), 15*time.Millisecond)
		assert.Nil(t, s.job)
	})
}
