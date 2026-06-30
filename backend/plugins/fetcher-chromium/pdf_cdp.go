package chromiumfetch

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"

	"meguri/internal/domain/model"
)

// fetchPDFViaCDP は CDP Fetch ドメインで Response 段階をインターセプトし PDF バイナリを取得する。
func (c *client) fetchPDFViaCDP(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	ua := resolveUserAgent(c.fetcherCfg, headers)
	opts := c.chromedpAllocatorOptions(ua)

	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	targetURL := u.String()
	var (
		body        []byte
		contentType string
		statusCode  int64
	)
	ready := make(chan error, 1)

	chromedp.ListenTarget(browserCtx, func(ev any) {
		e, ok := ev.(*fetch.EventRequestPaused)
		if !ok {
			return
		}
		if !shouldInterceptPDFRequest(e.Request.URL, targetURL) {
			go func(requestID fetch.RequestID) {
				_ = chromedp.Run(browserCtx, fetch.ContinueRequest(requestID))
			}(e.RequestID)
			return
		}
		if e.ResponseStatusCode == 0 {
			go func(requestID fetch.RequestID) {
				_ = chromedp.Run(browserCtx, fetch.ContinueRequest(requestID))
			}(e.RequestID)
			return
		}

		go func(ev *fetch.EventRequestPaused) {
			err := chromedp.Run(browserCtx, chromedp.ActionFunc(func(ctx context.Context) error {
				b, err := fetch.GetResponseBody(ev.RequestID).Do(ctx)
				if err != nil {
					_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
					return err
				}
				mime := "application/pdf"
				for _, h := range ev.ResponseHeaders {
					if strings.EqualFold(h.Name, "content-type") {
						mime = h.Value
						break
					}
				}
				body = b
				contentType = mime
				statusCode = ev.ResponseStatusCode
				return fetch.ContinueRequest(ev.RequestID).Do(ctx)
			}))
			ready <- err
		}(e)
	})

	err := chromedp.Run(browserCtx,
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{
			URLPattern:   "*",
			RequestStage: fetch.RequestStageResponse,
		}}),
		chromedp.Navigate(targetURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			select {
			case err := <-ready:
				return err
			case <-ctx.Done():
				return ctx.Err()
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("pdf fetch: no body captured")
	}

	ct := contentType
	if ct == "" {
		ct = "application/pdf"
	}
	sc := int(statusCode)
	if sc == 0 {
		sc = 200
	}

	return &model.Response{
		URL:         u,
		StatusCode:  sc,
		Headers:     map[string]string{"Content-Type": ct},
		ContentType: ct,
		Body:        body,
		FetchedAt:   time.Now(),
	}, nil
}

// shouldInterceptPDFRequest は Fetch インターセプト対象の PDF リクエストかを返す。
func shouldInterceptPDFRequest(requestURL, targetURL string) bool {
	if strings.HasSuffix(strings.ToLower(requestURL), ".pdf") {
		return true
	}
	return requestURL == targetURL
}
