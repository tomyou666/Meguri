package model

import "net/url"

// Request はパイプラインで扱うスクレイピングリクエストを表す。
// P2 PreProcessor はこの構造体に対して副作用を加えてよい。
type Request struct {
	URL     *url.URL
	Method  string
	Headers map[string]string
	Depth   int
	Meta    map[string]any
}

// NewRequest は GET メソッドのリクエストを構築するヘルパ。
func NewRequest(u *url.URL, depth int) *Request {
	return &Request{
		URL:     u,
		Method:  "GET",
		Headers: map[string]string{},
		Depth:   depth,
		Meta:    map[string]any{},
	}
}
