package usecase_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// TestMergeUIConfigLayers は UI 設定レイヤーのマージと JSON 文字列の展開を検証する。
func TestMergeUIConfigLayers(t *testing.T) {
	t.Run("正常系: JSON 文字列でラップされた設定を展開してマージする", func(t *testing.T) {
		inner := `{"crawl":{"max_depth":3},"request":{"timeout":"30s"}}`
		wrapped, err := json.Marshal(inner)
		require.NoError(t, err)

		merged, err := usecase.MergeUIConfigLayers(wrapped, json.RawMessage(`{"content":{"formats":["markdown"]}}`))
		require.NoError(t, err)

		var out map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(merged, &out))
		assert.Contains(t, string(out["crawl"]), "max_depth")
		assert.Contains(t, string(out["content"]), "markdown")
	})

	t.Run("正常系: ネストしたセクションを深くマージする", func(t *testing.T) {
		app := json.RawMessage(`{"crawl":{"max_pages":50},"request":{"retry_count":1}}`)
		ws := json.RawMessage(`{"crawl":{"max_depth":5}}`)

		merged, err := usecase.MergeUIConfigLayers(app, ws)
		require.NoError(t, err)

		cfg, err := usecase.ParseUIConfig(merged)
		require.NoError(t, err)
		assert.Equal(t, 5, cfg.Crawl.MaxDepth)
		assert.Equal(t, 50, cfg.Crawl.MaxPages)
		assert.Equal(t, 1, cfg.Request.RetryCount)
	})

	t.Run("正常系: fetch_limits をマージ結果に反映する", func(t *testing.T) {
		raw := json.RawMessage(`{"crawl":{"fetch_limits":{"http_max_inflight":8,"chromium_max_inflight":3,"auto_calibrate":false}}}`)
		cfg, err := usecase.ParseUIConfig(raw)
		require.NoError(t, err)
		assert.Equal(t, 8, cfg.Crawl.FetchLimits.HTTPMaxInflight)
		assert.Equal(t, 3, cfg.Crawl.FetchLimits.ChromiumMaxInflight)
		assert.False(t, cfg.Crawl.FetchLimits.AutoCalibrate)
	})

	t.Run("正常系: fetcher_config をマージ結果に反映する", func(t *testing.T) {
		raw := json.RawMessage(`{"plugins":{"fetcher":"chromium","fetcher_config":{"browser_path":"/bin/chromium","wait_until":"selector","wait_visible_selector":"h1","wait_timeout":"10s","network_idle_duration":"750ms"},"stealth":{"chromium":{"user_agent":"Test/1","headless":false,"hide_automation":true}}}}`)
		cfg, err := usecase.ParseUIConfig(raw)
		require.NoError(t, err)
		assert.Equal(t, model.FetcherChromium, cfg.Plugins.Fetcher)
		assert.Equal(t, "/bin/chromium", cfg.Plugins.FetcherConfig.BrowserPath)
		assert.Equal(t, "Test/1", cfg.Plugins.Stealth.Chromium.UserAgent)
		assert.False(t, cfg.Plugins.Stealth.Chromium.Headless)
		assert.True(t, cfg.Plugins.Stealth.Chromium.HideAutomation)
		assert.Equal(t, model.WaitUntilSelector, cfg.Plugins.FetcherConfig.WaitUntil)
		assert.Equal(t, "h1", cfg.Plugins.FetcherConfig.WaitVisibleSelector)
		assert.Equal(t, 10*time.Second, cfg.Plugins.FetcherConfig.WaitTimeout)
		assert.Equal(t, 750*time.Millisecond, cfg.Plugins.FetcherConfig.NetworkIdleDuration)
	})

	t.Run("正常系: stealth.http をマージ結果に反映する", func(t *testing.T) {
		raw := json.RawMessage(`{"plugins":{"stealth":{"http":{"user_agent":"HTTP/1","accept_language":"ja","cookie":"a=1"}}}}`)
		cfg, err := usecase.ParseUIConfig(raw)
		require.NoError(t, err)
		assert.Equal(t, "HTTP/1", cfg.Plugins.Stealth.HTTP.UserAgent)
		assert.Equal(t, "ja", cfg.Plugins.Stealth.HTTP.AcceptLanguage)
		assert.Equal(t, "a=1", cfg.Plugins.Stealth.HTTP.Cookie)
	})
}
