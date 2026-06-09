package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigLayer は設定 JSON の文字列ラップ解除と正規化を検証する。
func TestConfigLayer(t *testing.T) {
	t.Run("正常系: JSON 文字列でラップされた設定を map に展開する", func(t *testing.T) {
		inner := `{"crawl":{"max_depth":3},"request":{"timeout":"30s"}}`
		wrapped, err := json.Marshal(inner)
		require.NoError(t, err)

		m, err := unmarshalConfigMap(string(wrapped))
		require.NoError(t, err)
		assert.Contains(t, string(m["crawl"]), "max_depth")
	})

	t.Run("正常系: オブジェクト形式の RawMessage はそのまま JSON 文字列化する", func(t *testing.T) {
		out, err := settingsJSONFromRaw(json.RawMessage(`{"request":{"timeout":"30s"}}`))
		require.NoError(t, err)
		assert.JSONEq(t, `{"request":{"timeout":"30s"}}`, out)
	})

	t.Run("正常系: 文字列ラップされた RawMessage を展開する", func(t *testing.T) {
		inner := `{"crawl":{"max_depth":2}}`
		wrapped, err := json.Marshal(inner)
		require.NoError(t, err)

		out, err := settingsJSONFromRaw(wrapped)
		require.NoError(t, err)
		assert.JSONEq(t, inner, out)
	})
}
