package usecase

import chromiumfetch "meguri/plugins/fetcher-chromium"

// CloseChromiumBrowsers は chromium 共有プールの全ブラウザセッションを終了する。
func CloseChromiumBrowsers() {
	chromiumfetch.CloseAllBrowserSessions()
}
