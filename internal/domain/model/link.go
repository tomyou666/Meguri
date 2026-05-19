package model

import "net/url"

// Link はクロール時の追跡対象URLを表す。
type Link struct {
	URL   *url.URL
	Depth int
}
