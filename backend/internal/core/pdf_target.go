package core

import (
	"net/url"
	"strings"

	"meguri/internal/domain/model"
)

// IsPDFTarget は PDF 取得ルーティング対象 URL かを判定する。
func IsPDFTarget(u *url.URL, cfg *model.Config) bool {
	if u == nil || cfg == nil || !cfg.PDF.Enabled {
		return false
	}
	return strings.HasSuffix(strings.ToLower(u.Path), ".pdf")
}
