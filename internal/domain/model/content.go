package model

import "net/url"

// Content は P5 Parser が出力する中間表現。P6/P7/P8 に渡される。
type Content struct {
	URL         *url.URL
	Format      string
	Text        string
	DOM         any
	Metadata    map[string]string
	Attachments []Attachment
}

type Attachment struct {
	URL  *url.URL
	Kind string
	Data []byte
}
