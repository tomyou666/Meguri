package shared

import _ "embed"

// DefaultAppConfigJSON はアプリ既定設定の JSON ソース。
//
//go:embed defaults.json
var DefaultAppConfigJSON string
