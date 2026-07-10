package model

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfig_Validate は Config の入力検証ルールを検証する。
func TestConfig_Validate(t *testing.T) {
	t.Run("正常系: デフォルト設定にtargetsを1件付ければ検証は通る", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}

		err := c.Validate()

		assert.NoError(t, err, "デフォルト+ターゲット指定は検証を通過するはず")
	})

	t.Run("異常系: targetsが空だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = nil

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "targets", "targets についてのエラーメッセージを含むこと")
	})

	t.Run("異常系: targetsがhttp(s)で始まらないとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"ftp://example.com/"}

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "http://")
	})

	t.Run("異常系: request.timeoutが範囲外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Request.Timeout = 500 * time.Millisecond

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request.timeout")
	})

	t.Run("異常系: content.formatsに不正な値があるとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Content.Formats = []OutputFormat{"unknown"}

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content.formats")
	})

	t.Run("異常系: content.formatsの重複はエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Content.Formats = []OutputFormat{FormatMarkdown, FormatMarkdown}

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "重複")
	})

	t.Run("異常系: include_tagsとexclude_tagsに同名タグがあるとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Content.IncludeTags = []string{"article"}
		c.Content.ExcludeTags = []string{"article"}

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "include_tags")
	})

	t.Run("異常系: content.selectorが不正なCSSセレクタだとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Content.Selector = "div[unclosed"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "content.selector")
	})

	t.Run("異常系: pdf.modeが列挙外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.PDF.Mode = "weird"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pdf.mode")
	})

	t.Run("異常系: crawl.include_pathsに不正な正規表現があるとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Crawl.IncludePaths = []string{"["}

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "include_paths")
	})

	t.Run("正常系: request_delay>0 のとき max_concurrency は 1 に強制される", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Crawl.RequestDelay = 500 * time.Millisecond
		c.Crawl.MaxConcurrency = 8

		err := c.Validate()

		assert.NoError(t, err)
		assert.Equal(t, 1, c.Crawl.MaxConcurrency, "request_delay>0 のとき concurrency は強制で1")
	})

	t.Run("異常系: plugins.fetcher が列挙外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.Fetcher = "selenium"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugins.fetcher")
	})

	t.Run("異常系: plugins.fetcher_config.wait_timeout が範囲外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.FetcherConfig.WaitTimeout = 200 * time.Second

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wait_timeout")
	})

	t.Run("異常系: wait_until=selector でセレクタ未指定だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.FetcherConfig.WaitUntil = WaitUntilSelector

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wait_visible_selector")
	})

	t.Run("異常系: plugins.fetcher_config.wait_until が列挙外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.FetcherConfig.WaitUntil = "sleep"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wait_until")
	})

	t.Run("異常系: network_idle_request_max_age が範囲外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.FetcherConfig.NetworkIdleRequestMaxAge = 500 * time.Millisecond

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "network_idle_request_max_age")
	})

	t.Run("異常系: wait_after_load が範囲外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.FetcherConfig.WaitAfterLoad = 45 * time.Second

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "wait_after_load")
	})

	t.Run("異常系: fetch_limits の watermark が不正だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Crawl.FetchLimits.MemoryLowWatermark = 0.9
		c.Crawl.FetchLimits.MemoryHighWatermark = 0.7

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "memory_low_watermark")
	})

	t.Run("正常系: デフォルトの fetch_limits が設定される", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}

		err := c.Validate()

		assert.NoError(t, err)
		assert.Equal(t, DefaultHTTPMaxInflight, c.Crawl.FetchLimits.HTTPMaxInflight)
		assert.Equal(t, DefaultChromiumMaxInflight, c.Crawl.FetchLimits.ChromiumMaxInflight)
		assert.True(t, c.Crawl.FetchLimits.AutoCalibrate)
	})

	t.Run("正常系: デフォルトの fetcher は http", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}

		err := c.Validate()

		assert.NoError(t, err)
		assert.Equal(t, FetcherHTTP, c.Plugins.Fetcher)
		assert.True(t, c.Plugins.Stealth.Chromium.Headless)
		assert.True(t, c.Plugins.Stealth.Chromium.HideAutomation)
		assert.Equal(t, WaitUntilLoad, c.Plugins.FetcherConfig.EffectiveWaitUntil())
		assert.Equal(t, 5*time.Second, c.Plugins.FetcherConfig.WaitTimeout)
		assert.Equal(t, DefaultNetworkIdleDuration, c.Plugins.FetcherConfig.EffectiveNetworkIdleDuration())
		assert.Equal(t, DefaultNetworkIdleRequestMaxAge, c.Plugins.FetcherConfig.EffectiveNetworkIdleRequestMaxAge())
	})

	t.Run("異常系: stealth.http.user_agent に改行があるとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.Stealth.HTTP.UserAgent = "bad\nagent"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "plugins.stealth.http.user_agent")
	})

	t.Run("異常系: stealth.chromium.window_width が範囲外だとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Plugins.Stealth.Chromium.WindowWidth = 100

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "window_width")
	})

	t.Run("異常系: output.file_pattern に未知のプレースホルダがあるとエラー", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"https://example.com/"}
		c.Output.FilePattern = "{unknown}-{host}.{ext}"

		err := c.Validate()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "{unknown}")
	})

	t.Run("異常系: 複数の違反は集約されて返る", func(t *testing.T) {
		c := Default()
		c.Targets = []string{"ftp://x"}
		c.Request.Timeout = 0
		c.PDF.Mode = "weird"

		err := c.Validate()

		assert.Error(t, err)
		msg := err.Error()
		assert.True(t,
			strings.Contains(msg, "targets") &&
				strings.Contains(msg, "request.timeout") &&
				strings.Contains(msg, "pdf.mode"),
			"複数のエラーが集約されているはず: %s", msg)
	})
}

// TestOutputFormat_Valid は出力フォーマット列挙の妥当性判定を検証する。
func TestOutputFormat_Valid(t *testing.T) {
	t.Run("正常系: 列挙値はすべてValid", func(t *testing.T) {
		for _, f := range []OutputFormat{FormatMarkdown, FormatHTML, FormatRawHTML, FormatJSON, FormatLinks, FormatMetadata} {
			assert.True(t, f.Valid(), "Valid: %s", f)
		}
	})
	t.Run("異常系: 列挙外はInvalid", func(t *testing.T) {
		assert.False(t, OutputFormat("xml").Valid())
	})
}
