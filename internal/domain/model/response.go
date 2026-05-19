package model

import (
	"net/url"
	"time"
)

// Response は HTTP 取得結果を表す（P5 Parser の入力）。
type Response struct {
	URL         *url.URL
	StatusCode  int
	Headers     map[string]string
	ContentType string
	Body        []byte
	FetchedAt   time.Time
}
