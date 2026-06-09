package configloader_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"scraperbot/internal/domain/model"
	"scraperbot/internal/infrastructure/configloader"
)

// TestParseYAML は YAML 設定のパースとデフォルト値マージを検証する。
func TestParseYAML(t *testing.T) {
	t.Run("正常系: 部分指定YAMLでもデフォルト値がマージされる", func(t *testing.T) {
		yaml := []byte(`
targets:
  - https://example.com/
request:
  timeout: 5s
`)
		cfg, err := configloader.ParseYAML(yaml)

		assert.NoError(t, err)
		assert.Equal(t, []string{"https://example.com/"}, cfg.Targets)
		assert.Equal(t, 5*time.Second, cfg.Request.Timeout)
		assert.Equal(t, 2, cfg.Request.RetryCount, "未指定の値はDefault値が引き継がれる")
		assert.Equal(t, "markdown", cfg.Plugins.Transformer)
	})

	t.Run("正常系: ロード結果はValidateを通過する", func(t *testing.T) {
		yaml := []byte(`
targets:
  - https://example.com/
`)
		cfg, err := configloader.ParseYAML(yaml)
		assert.NoError(t, err)

		err = cfg.Validate()

		assert.NoError(t, err, "Validateを通過するはず")
	})

	t.Run("異常系: 不正YAMLはエラー", func(t *testing.T) {
		_, err := configloader.ParseYAML([]byte("targets: [unclosed\n"))
		assert.Error(t, err)
	})

	t.Run("正常系: ContentFormats を列挙でロードできる", func(t *testing.T) {
		yaml := []byte(`
targets: ["https://example.com/"]
content:
  formats: [markdown, links]
`)
		cfg, err := configloader.ParseYAML(yaml)

		assert.NoError(t, err)
		assert.Equal(t, []model.OutputFormat{model.FormatMarkdown, model.FormatLinks}, cfg.Content.Formats)
	})
}
