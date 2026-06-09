package runner_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/pkg/runner"
)

// TestMergeUIConfigLayers は UI 設定レイヤーのマージと JSON 文字列の展開を検証する。
func TestMergeUIConfigLayers(t *testing.T) {
	t.Run("正常系: JSON 文字列でラップされた設定を展開してマージする", func(t *testing.T) {
		inner := `{"crawl":{"max_depth":3},"request":{"timeout":"30s"}}`
		wrapped, err := json.Marshal(inner)
		require.NoError(t, err)

		merged, err := runner.MergeUIConfigLayers(wrapped, json.RawMessage(`{"content":{"formats":["markdown"]}}`))
		require.NoError(t, err)

		var out map[string]json.RawMessage
		require.NoError(t, json.Unmarshal(merged, &out))
		assert.Contains(t, string(out["crawl"]), "max_depth")
		assert.Contains(t, string(out["content"]), "markdown")
	})

	t.Run("正常系: ネストしたセクションを深くマージする", func(t *testing.T) {
		app := json.RawMessage(`{"crawl":{"max_pages":50},"request":{"retry_count":1}}`)
		ws := json.RawMessage(`{"crawl":{"max_depth":5}}`)

		merged, err := runner.MergeUIConfigLayers(app, ws)
		require.NoError(t, err)

		cfg, err := runner.ParseUIConfig(merged)
		require.NoError(t, err)
		assert.Equal(t, 5, cfg.Crawl.MaxDepth)
		assert.Equal(t, 50, cfg.Crawl.MaxPages)
		assert.Equal(t, 1, cfg.Request.RetryCount)
	})
}
