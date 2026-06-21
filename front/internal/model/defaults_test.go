package model

import (
	"encoding/json"
	"testing"
)

// TestDefaultAppConfigJSONParses は embed された defaults JSON がパース可能か検証する。
func TestDefaultAppConfigJSONParses(t *testing.T) {
	t.Run("正常系: defaults JSON を map にパースできる", func(t *testing.T) {
		var cfg map[string]any
		if err := json.Unmarshal([]byte(DefaultAppConfigJSON), &cfg); err != nil {
			t.Fatal(err)
		}
		if cfg["request"] == nil {
			t.Fatal("request section missing")
		}
	})
}
