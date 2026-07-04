package chromiumfetch

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"

	"meguri/internal/domain/model"
)

const pdfMagic = "%PDF-"

// fetchPDFViaCDP は CDP Fetch ドメインで Response 段階をインターセプトし PDF バイナリを取得する。
func (c *client) fetchPDFViaCDP(ctx context.Context, u *url.URL, headers map[string]string, ua string) (*model.Response, error) {
	targetURL := u.String()
	var (
		body        []byte
		contentType string
		statusCode  int64
	)
	ready := make(chan error, 1)

	err := c.runWithTab(ctx, ua, func(tabCtx context.Context) error {
		chromedp.ListenTarget(tabCtx, func(ev any) {
			e, ok := ev.(*fetch.EventRequestPaused)
			if !ok {
				return
			}
			if !shouldInterceptPDFRequest(e.Request.URL, targetURL) {
				go func(requestID fetch.RequestID) {
					_ = chromedp.Run(tabCtx, fetch.ContinueRequest(requestID))
				}(e.RequestID)
				return
			}
			if e.ResponseStatusCode == 0 {
				go func(requestID fetch.RequestID) {
					_ = chromedp.Run(tabCtx, fetch.ContinueRequest(requestID))
				}(e.RequestID)
				return
			}

			go func(ev *fetch.EventRequestPaused) {
				final := false
				err := chromedp.Run(tabCtx, chromedp.ActionFunc(func(ctx context.Context) error {
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
					hasMagic := bytes.HasPrefix(b, []byte(pdfMagic))
					// PDF 本体、または 4xx の確定失敗のみ完了とする。
					// WAF challenge (202 等) はブラウザ側の再取得を待つ。
					if hasMagic || ev.ResponseStatusCode >= 400 {
						body = b
						contentType = mime
						statusCode = ev.ResponseStatusCode
						final = true
					}
					return fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}))
				if err != nil {
					select {
					case ready <- err:
					default:
					}
					return
				}
				if final {
					select {
					case ready <- nil:
					default:
					}
				}
			}(e)
		})

		return chromedp.Run(tabCtx,
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
				case <-time.After(c.reqCfg.Timeout):
					return ctx.Err()
				}
			}),
		)
	})
	if err != nil {
		return nil, err
	}
	if len(body) == 0 {
		return nil, fmt.Errorf("pdf取得失敗: 本文を取得できませんでした")
	}

	ct := contentType
	if ct == "" {
		ct = "application/pdf"
	}
	sc := int(statusCode)
	if sc == 0 {
		sc = 200
	}
	if sc < 200 || sc >= 300 {
		return nil, fmt.Errorf("pdf取得失敗: HTTP %d (content-type=%s)", sc, ct)
	}
	if !bytes.HasPrefix(body, []byte(pdfMagic)) {
		return nil, fmt.Errorf("pdf取得失敗: PDFではない応答 (HTTP %d, content-type=%s)", sc, ct)
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
