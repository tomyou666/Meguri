package chromiumfetch

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"meguri/internal/domain/model"
)

// TestClient_waitTimeout は wait_timeout 未設定時のフォールバックを検証する。
func TestClient_waitTimeout(t *testing.T) {
	t.Run("正常系: wait_timeout 設定時はそれを返す", func(t *testing.T) {
		c := &client{
			reqCfg: model.RequestConfig{Timeout: 60 * time.Second},
			fetcherCfg: model.FetcherConfig{
				WaitTimeout: 12 * time.Second,
			},
		}
		assert.Equal(t, 12*time.Second, c.waitTimeout())
	})

	t.Run("正常系: wait_timeout 未設定時は request.timeout を返す", func(t *testing.T) {
		c := &client{
			reqCfg:     model.RequestConfig{Timeout: 45 * time.Second},
			fetcherCfg: model.FetcherConfig{},
		}
		assert.Equal(t, 45*time.Second, c.waitTimeout())
	})
}

// TestClient_buildNavigateAndWaitTasks は wait_until ごとの Navigate+待機タスク数を検証する。
func TestClient_buildNavigateAndWaitTasks(t *testing.T) {
	t.Run("正常系: none は extra headers・Navigate のみ", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilNone}}
		tasks := c.buildNavigateAndWaitTasks("https://example.com")
		assert.Len(t, tasks, 2)
	})

	t.Run("正常系: load は WaitReady を挟む", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilLoad}}
		tasks := c.buildNavigateAndWaitTasks("https://example.com")
		assert.Len(t, tasks, 3)
	})

	t.Run("正常系: network_idle は監視開始を Navigate 前に挟む", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilNetworkIdle}}
		tasks := c.buildNavigateAndWaitTasks("https://example.com")
		assert.Len(t, tasks, 4)
	})
}

// TestShouldSleepAfterLoad は wait_after_load の追加 sleep 発火条件を検証する。
func TestShouldSleepAfterLoad(t *testing.T) {
	t.Run("正常系: load かつ wait_after_load>0 なら true", func(t *testing.T) {
		assert.True(t, shouldSleepAfterLoad(model.WaitUntilLoad, 5*time.Second))
	})

	t.Run("正常系: load でも wait_after_load=0 なら false", func(t *testing.T) {
		assert.False(t, shouldSleepAfterLoad(model.WaitUntilLoad, 0))
	})

	t.Run("正常系: none は wait_after_load>0 でも false", func(t *testing.T) {
		assert.False(t, shouldSleepAfterLoad(model.WaitUntilNone, 5*time.Second))
	})

	t.Run("正常系: network_idle は wait_after_load>0 でも false", func(t *testing.T) {
		assert.False(t, shouldSleepAfterLoad(model.WaitUntilNetworkIdle, 5*time.Second))
	})

	t.Run("正常系: selector は wait_after_load>0 でも false", func(t *testing.T) {
		assert.False(t, shouldSleepAfterLoad(model.WaitUntilSelector, 5*time.Second))
	})
}
