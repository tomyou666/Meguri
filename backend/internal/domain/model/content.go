package model

import "net/url"

// Content は P5 Parser が出力する中間表現。P6/P7/P8 に渡される。
type Content struct {
	// URL はコンテンツの元 URL。
	URL *url.URL
	// Format はコンテンツ種別（"html", "pdf" など）。
	Format string
	// Text は抽出済みプレーンテキストまたは中間表現。
	Text string
	// RawHTML は HTTP レスポンス本文の生 HTML（P5 で設定、P7 では変更しない）。
	RawHTML string
	// DOM は HTML 等の構造化表現（パーサ実装依存）。
	DOM any
	// Metadata はタイトル・description 等のメタ情報。
	Metadata map[string]string
	// Attachments は PDF バイナリ等の添付データ。
	Attachments []Attachment
}

// Attachment は Content に紐づくバイナリ添付を表す。
type Attachment struct {
	// URL は添付の取得元 URL。
	URL *url.URL
	// Kind は種別識別子（"pdf" など）。
	Kind string
	// Data は生バイト列。
	Data []byte
}
