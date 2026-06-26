package httpfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"meguri/internal/domain/model"
)

// get はリトライ付き GET で model.Response を返す。
func (c *client) get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error) {
	hc := &http.Client{Timeout: c.reqCfg.Timeout}
	var lastErr error
	attempts := c.reqCfg.RetryCount + 1
	for i := 0; i < attempts; i++ {
		reqCtx, cancel := context.WithTimeout(ctx, c.reqCfg.Timeout)
		res, err := doOnce(reqCtx, hc, u, headers)
		cancel()

		if err == nil && res.StatusCode < 500 {
			return res, nil
		}
		if err != nil {
			lastErr = err
		} else {
			lastErr = fmt.Errorf("http %d", res.StatusCode)
		}
		if err != nil && !isRetryableHTTPError(err) {
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
		lastErr = errors.New("unknown http error")
	}
	return nil, fmt.Errorf("HTTP取得失敗 (url=%s): %w", u.String(), lastErr)
}

func doOnce(ctx context.Context, hc *http.Client, u *url.URL, headers map[string]string) (*model.Response, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	res, err := hc.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &model.Response{
		URL:         u,
		StatusCode:  res.StatusCode,
		Headers:     flattenHeaders(res.Header),
		ContentType: res.Header.Get("Content-Type"),
		Body:        body,
		FetchedAt:   time.Now(),
	}, nil
}

func isRetryableHTTPError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return false
	}
	return true
}

func flattenHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) > 0 {
			out[k] = v[0]
		}
	}
	return out
}
