package chromiumfetch

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/chromedp"

	"meguri/internal/domain/model"
)

const pdfMagic = "%PDF-"

// pdfCaptureSnapshot は PDF インターセプトの累積状態。
type pdfCaptureSnapshot struct {
	// Finalized は %PDF- の 2xx を確定したか。
	Finalized bool
	// Body は成功 PDF、または最初の 4xx 本文。
	Body []byte
	// ContentType は Body に対応する Content-Type。
	ContentType string
	// StatusCode は Body に対応する HTTP ステータス。
	StatusCode int64
	// Has4xx は 4xx を 1 件でも記憶したか。
	Has4xx bool
}

// applyPDFCaptureResponse は 1 件の応答を取り込み、新しい状態を返す。
// Finalized が true なら成功 PDF が確定済み（以降の応答は無視する）。
func applyPDFCaptureResponse(prev pdfCaptureSnapshot, status int64, mime string, body []byte) pdfCaptureSnapshot {
	if prev.Finalized {
		return prev
	}
	hasMagic := bytes.HasPrefix(body, []byte(pdfMagic))
	if status >= 200 && status < 300 && hasMagic {
		return pdfCaptureSnapshot{
			Finalized:   true,
			Body:        append([]byte(nil), body...),
			ContentType: mime,
			StatusCode:  status,
			Has4xx:      prev.Has4xx,
		}
	}
	if status >= 400 {
		if prev.Has4xx {
			return prev
		}
		return pdfCaptureSnapshot{
			Body:        append([]byte(nil), body...),
			ContentType: mime,
			StatusCode:  status,
			Has4xx:      true,
		}
	}
	return prev
}

// pdfCaptureReadyOnTimeout は制限時間切れ時に出口判定へ渡せる状態かを返す。
// true なら成功 PDF、または記憶した最初の 4xx を結果として使う。
func pdfCaptureReadyOnTimeout(s pdfCaptureSnapshot) bool {
	return s.Finalized || s.Has4xx
}

// pdfCaptureOnWaitStop は待ち終了時の扱いを決める。
// useSnap が true なら記憶した成功 PDF / 最初の 4xx を出口判定へ渡す。
// 取得期限切れかつ記憶ありのときだけ useSnap。ユーザー中断は reqErr のまま返す。
func pdfCaptureOnWaitStop(snap pdfCaptureSnapshot, reqErr error) (useSnap bool, err error) {
	if errors.Is(reqErr, context.DeadlineExceeded) && pdfCaptureReadyOnTimeout(snap) {
		return true, nil
	}
	if reqErr != nil {
		return false, reqErr
	}
	return false, context.Canceled
}

// fetchPDFViaCDP は CDP Fetch ドメインで Response 段階をインターセプトし PDF バイナリを取得する。
func (c *client) fetchPDFViaCDP(ctx context.Context, u *url.URL, headers map[string]string, ua string) (*model.Response, error) {
	targetURL := u.String()
	reqCtx := ctx
	var (
		mu   sync.Mutex
		snap pdfCaptureSnapshot
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
				var finalized bool
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
					mu.Lock()
					snap = applyPDFCaptureResponse(snap, ev.ResponseStatusCode, mime, b)
					finalized = snap.Finalized
					mu.Unlock()
					return fetch.ContinueRequest(ev.RequestID).Do(ctx)
				}))
				if err != nil {
					select {
					case ready <- err:
					default:
					}
					return
				}
				if finalized {
					select {
					case ready <- nil:
					default:
					}
				}
			}(e)
		})

		// 制限時間は reqCtx（get が付ける Timeout）だけを使う。
		// tabCtx は期限切れ時に Canceled になるため、判定には reqCtx.Err() を見る。
		return chromedp.Run(tabCtx,
			fetch.Enable().WithPatterns([]*fetch.RequestPattern{{
				URLPattern:   "*",
				RequestStage: fetch.RequestStageResponse,
			}}),
			chromedp.Navigate(targetURL),
			chromedp.ActionFunc(func(_ context.Context) error {
				select {
				case err := <-ready:
					return err
				case <-tabCtx.Done():
					mu.Lock()
					s := snap
					mu.Unlock()
					useSnap, waitErr := pdfCaptureOnWaitStop(s, reqCtx.Err())
					if useSnap {
						return nil
					}
					return waitErr
				}
			}),
		)
	})
	if err != nil {
		return nil, err
	}

	mu.Lock()
	body := snap.Body
	contentType := snap.ContentType
	statusCode := snap.StatusCode
	mu.Unlock()

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
