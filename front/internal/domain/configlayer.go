package domain

import "encoding/json"

// normalizeConfigLayer は Wails 経由で JSON 文字列として渡されたレイヤをオブジェクト JSON に正規化する。
func normalizeConfigLayer(layer json.RawMessage) (json.RawMessage, error) {
	if len(layer) == 0 || string(layer) == "null" {
		return nil, nil
	}
	var asString string
	if err := json.Unmarshal(layer, &asString); err == nil {
		if asString == "" || asString == "null" {
			return nil, nil
		}
		return json.RawMessage(asString), nil
	}
	return layer, nil
}

// unmarshalConfigMap は設定 JSON を map にデコードする（二重エンコード文字列にも対応）。
func unmarshalConfigMap(raw string) (map[string]json.RawMessage, error) {
	normalized, err := normalizeConfigLayer(json.RawMessage(raw))
	if err != nil {
		return nil, err
	}
	if len(normalized) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(normalized, &m); err != nil {
		return nil, err
	}
	if m == nil {
		return map[string]json.RawMessage{}, nil
	}
	return m, nil
}

// settingsJSONFromRaw は RawMessage を DB 保存用の設定 JSON 文字列に正規化する。
func settingsJSONFromRaw(raw json.RawMessage) (string, error) {
	normalized, err := normalizeConfigLayer(raw)
	if err != nil {
		return "{}", err
	}
	if len(normalized) == 0 {
		return "{}", nil
	}
	return string(normalized), nil
}
