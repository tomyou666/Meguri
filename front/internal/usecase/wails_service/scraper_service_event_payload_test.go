package wails_service

import (
	"encoding/json"
	"testing"

	"meguri-app/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNodeSucceededEventPayload は成功 Event に本文フィールドが乗らないことを検証する。
func TestNodeSucceededEventPayload(t *testing.T) {
	t.Run("正常系: nodeSucceeded payload の JSON に html/markdown/rawHtml/result が無い", func(t *testing.T) {
		payload := model.CrawlEventPayload{
			WorkspaceID: "ws-1",
			RunID:       "run-1",
			NodeID:      "n1",
			URL:         "https://example.com/",
		}
		b, err := json.Marshal(payload)
		require.NoError(t, err)

		var raw map[string]any
		require.NoError(t, json.Unmarshal(b, &raw))

		assert.Equal(t, "n1", raw["nodeId"])
		assert.Equal(t, "run-1", raw["runId"])
		assert.NotContains(t, raw, "result")
		assert.NotContains(t, raw, "html")
		assert.NotContains(t, raw, "markdown")
		assert.NotContains(t, raw, "rawHtml")
	})
}
