package model

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
