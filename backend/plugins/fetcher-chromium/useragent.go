package chromiumfetch

import (
	"strings"

	"meguri/internal/domain/model"
)

// DefaultUserAgent は chromedp 既定の HeadlessChrome UA を避けるための一般的な Chromium UA。
const DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

// resolveUserAgent は chromium フェッチで使う User-Agent を決定する。
//
// 優先順位:
// 1) plugins.stealth.chromium.user_agent
// 2) DefaultUserAgent
func resolveUserAgent(s model.ChromiumStealthConfig) string {
	if ua := strings.TrimSpace(s.UserAgent); ua != "" {
		return ua
	}
	return DefaultUserAgent
}
