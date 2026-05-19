// Package pdf は PDF レスポンスを最小限テキスト抽出する P5 Parser を提供する。
//
// 注: 完全な PDF テキスト抽出（fast/auto/ocr モード切替）は将来の置換可能なライブラリに委ねる前提で、
// 本 MVP では PDF と判定された場合に Content を組み立てるルーティングが成立することを優先する。
package pdf

import (
	"context"
	"strings"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterParser("pdf", func() plugin.Parser { return &parser{} })
}

type parser struct {
	host plugin.Host
}

func (p *parser) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "pdf",
		Version:     "0.1.0",
		Kind:        plugin.KindParser,
		Description: "PDF レスポンスを Content にラップする（MVP: テキスト擬似抽出）",
	}
}

func (p *parser) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

func (p *parser) Close(_ context.Context) error { return nil }

func (p *parser) CanParse(res *model.Response) bool {
	ct := strings.ToLower(res.ContentType)
	if strings.Contains(ct, "application/pdf") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(res.URL.Path), ".pdf")
}

// Parse は PDF レスポンスを暫定的にテキストに見立てた Content にして返す。
// 実 PDF パース（埋め込みテキスト/OCR）は今後のプラグイン拡張で差し替える。
func (p *parser) Parse(_ context.Context, res *model.Response) (*model.Content, error) {
	text := extractText(res.Body)

	meta := map[string]string{
		"content_type":   res.ContentType,
		"bytes_total":    sprintInt(len(res.Body)),
		"parse_strategy": "stub",
	}

	return &model.Content{
		URL:      res.URL,
		Format:   "pdf",
		Text:     text,
		Metadata: meta,
		Attachments: []model.Attachment{
			{URL: res.URL, Kind: "pdf", Data: res.Body},
		},
	}, nil
}

// extractText は PDF バイナリから ASCII テキスト断片を保守的に抜き出す簡易実装。
// 真の PDF コンテンツストリーム解析は行わないが、テストデータでは目視可能な文字列を返せる。
func extractText(b []byte) string {
	var sb strings.Builder
	prevPrintable := false
	for _, ch := range b {
		if ch >= 0x20 && ch <= 0x7e {
			sb.WriteByte(ch)
			prevPrintable = true
		} else if ch == '\n' || ch == '\r' || ch == '\t' {
			if prevPrintable {
				sb.WriteByte(' ')
			}
			prevPrintable = false
		} else {
			if prevPrintable {
				sb.WriteByte(' ')
			}
			prevPrintable = false
		}
	}
	return strings.Join(strings.Fields(sb.String()), " ")
}

func sprintInt(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
