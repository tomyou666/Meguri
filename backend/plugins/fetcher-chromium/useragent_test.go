package chromiumfetch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/domain/model"
)

// TestResolveUserAgent は User-Agent 解決の優先順位を検証する。
func TestResolveUserAgent(t *testing.T) {
	t.Parallel()

	t.Run("fetcher_config が最優先", func(t *testing.T) {
		ua := resolveUserAgent(model.FetcherConfig{UserAgent: "Custom/1.0"}, map[string]string{
			"User-Agent": "Header/2.0",
		})
		assert.Equal(t, "Custom/1.0", ua)
	})

	t.Run("request headers が次点", func(t *testing.T) {
		ua := resolveUserAgent(model.FetcherConfig{}, map[string]string{
			"User-Agent": "Header/2.0",
		})
		assert.Equal(t, "Header/2.0", ua)
	})

	t.Run("未指定時はデフォルトUA", func(t *testing.T) {
		ua := resolveUserAgent(model.FetcherConfig{}, nil)
		assert.Equal(t, DefaultUserAgent, ua)
		assert.NotContains(t, ua, "HeadlessChrome")
	})
}
