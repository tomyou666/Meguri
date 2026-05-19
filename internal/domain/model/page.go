package model

import "net/url"

// Page は将来の出力ストレージ層が使うエンティティ。MVPでは Result を主に使う。
type Page struct {
	URL      *url.URL
	Title    string
	Metadata map[string]string
}
