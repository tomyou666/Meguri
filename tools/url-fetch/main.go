// Command url-fetch は URL を HTTP / Chromium の複数バリアントで取得し、メタ情報を標準出力へ書き出す。
package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"sync"
	"time"
)

// cfg は url-fetch の実行設定を一括で保持する。
var cfg = struct {
	// DefaultTargetURL は引数未指定時に取得する URL。
	DefaultTargetURL string
	// HTTPVariantTimeout は HTTP バリアント1回あたりのタイムアウト。
	HTTPVariantTimeout time.Duration
	// ChromiumOverallTimeout は Chromium 群全体のタイムアウト。
	ChromiumOverallTimeout time.Duration
	// ChromeUserAgent は backend fetcher-chromium の DefaultUserAgent と同じ UA。
	ChromeUserAgent string
	// MaxWaitAfterNavigate は headless_wait バリアントの Navigate 後待機上限。
	MaxWaitAfterNavigate time.Duration
}{
	DefaultTargetURL:       "https://example.com",
	HTTPVariantTimeout:     5 * time.Second,
	ChromiumOverallTimeout: 60 * time.Second,
	ChromeUserAgent:        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	MaxWaitAfterNavigate:   2 * time.Second,
}

func main() {
	target, err := resolveTargetURL(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		fmt.Fprintln(os.Stderr, "Usage: go run ./tools/url-fetch [URL]")
		os.Exit(1)
	}

	if err := runAllVariants(target, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runAllVariants は HTTP と Chromium を並列実行し、HTTP 結果→Chromium 結果の順で書き出す。
func runAllVariants(target string, w io.Writer) error {
	httpCtx, httpCancel := context.WithCancel(context.Background())
	defer httpCancel()

	chromiumCtx, chromiumCancel := context.WithTimeout(context.Background(), cfg.ChromiumOverallTimeout)
	defer chromiumCancel()

	var (
		httpResults     []trialResult
		chromiumResults []trialResult
		httpErr         error
		chromiumErr     error
		wg              sync.WaitGroup
	)

	wg.Add(2)
	go func() {
		defer wg.Done()
		httpResults, httpErr = fetchAllHTTPVariants(httpCtx, target)
	}()
	go func() {
		defer wg.Done()
		chromiumResults, chromiumErr = fetchAllChromiumVariants(chromiumCtx, target)
	}()
	wg.Wait()

	if httpErr != nil {
		return fmt.Errorf("http: %w", httpErr)
	}
	if chromiumErr != nil {
		return fmt.Errorf("chromium: %w", chromiumErr)
	}
	if err := writeTrialResults(w, httpResults); err != nil {
		return err
	}
	return writeTrialResults(w, chromiumResults)
}

// httpVariantsFor は target 向け HTTP バリアント一覧を返す。
func httpVariantsFor(target string) ([]httpVariant, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	referer := parsed.Scheme + "://" + parsed.Host + "/"
	return []httpVariant{
		{id: "default"},
		{id: "chrome_ua", headers: map[string]string{"User-Agent": cfg.ChromeUserAgent}},
		{
			id: "chrome_ua_lang",
			headers: map[string]string{
				"User-Agent":      cfg.ChromeUserAgent,
				"Accept-Language": "ja",
				"Accept":          "text/html",
			},
		},
		{
			id: "chrome_ua_referer",
			headers: map[string]string{
				"User-Agent":      cfg.ChromeUserAgent,
				"Accept-Language": "ja",
				"Accept":          "text/html",
				"Referer":         referer,
			},
		},
		{
			id: "utls_chrome_ua",
			headers: map[string]string{
				"User-Agent": cfg.ChromeUserAgent,
			},
			utlsTransport: "http2",
		},
		{
			id: "utls_chrome_ua_http1",
			headers: map[string]string{
				"User-Agent": cfg.ChromeUserAgent,
			},
			utlsTransport: "http1",
		},
	}, nil
}

// chromiumVariants は Chromium バリアント一覧を返す。
func chromiumVariants() []chromiumVariant {
	return []chromiumVariant{
		{id: "headless_default", headless: true},
		{id: "headless_chrome_ua", headless: true, userAgentOverride: cfg.ChromeUserAgent},
		{id: "headless_wait", headless: true, userAgentOverride: cfg.ChromeUserAgent, waitAfterNavigate: true},
		{id: "chrome_ua", headless: false},
	}
}

// resolveTargetURL は引数から取得 URL を決定する。
//
// args が空なら cfg.DefaultTargetURL を使う。
// args が1件なら http / https の絶対 URL として検証する。
func resolveTargetURL(args []string) (string, error) {
	if len(args) == 0 {
		return cfg.DefaultTargetURL, nil
	}
	if len(args) > 1 {
		return "", fmt.Errorf("too many arguments")
	}
	raw := args[0]
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid URL: scheme must be http or https")
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("invalid URL: host is required")
	}
	return raw, nil
}
