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

// TestClient_buildHTMLFetchTasks は wait_until ごとのタスク数を検証する。
func TestClient_buildHTMLFetchTasks(t *testing.T) {
	t.Run("正常系: none は extra headers・Navigate・OuterHTML", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilNone}}
		tasks := c.buildHTMLFetchTasks("https://example.com", new(string))
		assert.Len(t, tasks, 3)
	})

	t.Run("正常系: load は待機アクションを挟む", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilLoad}}
		tasks := c.buildHTMLFetchTasks("https://example.com", new(string))
		assert.Len(t, tasks, 4)
	})

	t.Run("正常系: network_idle は監視開始を Navigate 前に挟む", func(t *testing.T) {
		c := &client{fetcherCfg: model.FetcherConfig{WaitUntil: model.WaitUntilNetworkIdle}}
		tasks := c.buildHTMLFetchTasks("https://example.com", new(string))
		assert.Len(t, tasks, 5)
	})
}
