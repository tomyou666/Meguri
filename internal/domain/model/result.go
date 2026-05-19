package model

import "net/url"

// Result はパイプライン最終出力。P6 Transformer が組み立てる。
type Result struct {
	URL      *url.URL
	Markdown string
	HTML     string
	RawHTML  string
	JSON     map[string]any
	Links    []*url.URL
	Metadata map[string]string
}
