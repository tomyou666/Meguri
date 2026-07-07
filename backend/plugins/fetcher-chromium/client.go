package chromiumfetch

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/chromedp"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// get はリトライ付きで URL を取得する。
func (c *client) get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	ua := resolveUserAgent(c.fetcherCfg, headers)
	if err := c.joinPool(ctx, ua); err != nil {
		return nil, err
	}

	var lastErr error
	attempts := c.reqCfg.RetryCount + 1
	for i := 0; i < attempts; i++ {
		reqCtx, cancel := context.WithTimeout(ctx, c.reqCfg.Timeout)
		res, err := c.fetchOnce(reqCtx, u, headers, ua)
		cancel()

		if err == nil {
			return res, nil
		}
		lastErr = err
		if !isRetryableFetchError(err) {
			break
		}
		if i+1 < attempts {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.reqCfg.RetryInterval):
			}
		}
	}
	if lastErr == nil {
		lastErr = errors.New("unknown chromium fetch error")
	}
	return nil, fmt.Errorf("chromium取得失敗 (url=%s): %w", u.String(), lastErr)
}

// fetchOnce は 1 回分の取得を実行する。PDF 対象なら CDP 経由で取得する。
func (c *client) fetchOnce(ctx context.Context, u *url.URL, headers map[string]string, ua string) (*model.Response, error) {
	if core.IsPDFTarget(u, c.pdfCfg) {
		return c.fetchPDFViaCDP(ctx, u, headers, ua)
	}

	var html string
	err := c.runWithWait(ctx, ua, func(tabCtx context.Context) []chromedp.Action {
		return c.buildHTMLFetchTasks(u.String(), &html)
	})
	if err != nil {
		return nil, err
	}

	return &model.Response{
		URL:         u,
		StatusCode:  200,
		Headers:     map[string]string{"Content-Type": "text/html; charset=utf-8"},
		ContentType: "text/html; charset=utf-8",
		Body:        []byte(html),
		FetchedAt:   time.Now(),
	}, nil
}

// chromedpAllocatorOptions はブラウザ起動用の chromedp オプションを返す。
func (c *client) chromedpAllocatorOptions(ua string) []chromedp.ExecAllocatorOption {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(c.browserPath),
		chromedp.UserAgent(ua),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	if c.fetcherCfg.Headless {
		opts = append(opts, chromedp.Flag("headless", true))
	} else {
		opts = append(opts, chromedp.Flag("headless", false))
	}
	return opts
}

// isRetryableFetchError はリトライ可能な取得エラーかどうかを返す。
// タイムアウト・キャンセル・実行ファイル未検出はリトライしない。
func isRetryableFetchError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "executable") && strings.Contains(msg, "not found") {
		return false
	}
	return true
}
