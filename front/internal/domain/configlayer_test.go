package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalConfigMap_unwrapsJSONString(t *testing.T) {
	inner := `{"crawl":{"max_depth":3},"request":{"timeout":"30s"}}`
	wrapped, err := json.Marshal(inner)
	require.NoError(t, err)

	m, err := unmarshalConfigMap(string(wrapped))
	require.NoError(t, err)
	assert.Contains(t, string(m["crawl"]), "max_depth")
}

func TestSettingsJSONFromRaw_object(t *testing.T) {
	out, err := settingsJSONFromRaw(json.RawMessage(`{"request":{"timeout":"30s"}}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"request":{"timeout":"30s"}}`, out)
}

func TestSettingsJSONFromRaw_wrappedString(t *testing.T) {
	inner := `{"crawl":{"max_depth":2}}`
	wrapped, err := json.Marshal(inner)
	require.NoError(t, err)

	out, err := settingsJSONFromRaw(wrapped)
	require.NoError(t, err)
	assert.JSONEq(t, inner, out)
}
