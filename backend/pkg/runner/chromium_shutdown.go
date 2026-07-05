package runner

import "meguri/internal/usecase"

//go:generate go tool gowrap gen -p meguri/pkg/runner -i ChromiumShutdown -t templates/slog_debug.gotmpl -o chromium_shutdown_with_debug_log.go

// ChromiumShutdown は chromium ブラウザプール終了を抽象化する。
type ChromiumShutdown interface {
	// CloseChromiumBrowsers は chromium 共有プールの全ブラウザセッションを終了する。
	CloseChromiumBrowsers()
}

type chromiumShutdownImpl struct{}

func (chromiumShutdownImpl) CloseChromiumBrowsers() {
	usecase.CloseChromiumBrowsers()
}

// CloseChromiumBrowsers は chromium 共有プールの全ブラウザセッションを終了する。
func CloseChromiumBrowsers() {
	defaultChromiumShutdown.CloseChromiumBrowsers()
}
