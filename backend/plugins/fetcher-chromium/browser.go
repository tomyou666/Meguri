package chromiumfetch

import (
	"scraperbot/internal/infrastructure/chromium"
)

// EnvBrowserPath はブラウザ実行ファイルを明示指定する環境変数名。
const EnvBrowserPath = chromium.EnvBrowserPath

// resolveBrowserPath は使用するブラウザ実行ファイルのパスを解決する。
func resolveBrowserPath(explicit string) (string, error) {
	return chromium.ResolveBrowserPath(explicit)
}
