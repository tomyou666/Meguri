// Package pdf は PDF レスポンスをテキスト抽出する P5 Parser を提供する。
package pdf

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterParser("pdf", func() plugin.Parser { return &parser{} })
}

// parser は PDF 用 P5 Parser の実装。
type parser struct {
	// host は Init で受け取る Host。
	host plugin.Host
}

// Metadata は plugin.Parser.Metadata の実装。
func (p *parser) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "pdf",
		Version:     "0.2.0",
		Kind:        plugin.KindParser,
		Description: "ledongthuc/pdf による PDF テキスト抽出（fast モード）",
	}
}

// Init は plugin.Plugin.Init の実装。
func (p *parser) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (p *parser) Close(_ context.Context) error { return nil }

// CanParse は application/pdf または .pdf パスを判定する。
func (p *parser) CanParse(res *model.Response) bool {
	ct := strings.ToLower(res.ContentType)
	if strings.Contains(ct, "application/pdf") {
		return true
	}
	return strings.HasSuffix(strings.ToLower(res.URL.Path), ".pdf")
}

// Parse は PDF レスポンスを ledongthuc/pdf でテキスト化した Content として返す。
func (p *parser) Parse(_ context.Context, res *model.Response) (*model.Content, error) {
	mode := p.resolveMode()
	maxPages := p.resolveMaxPages()

	text, err := extractPlainText(res.Body, maxPages)
	if err != nil {
		return nil, fmt.Errorf("pdf parse: %w", err)
	}

	meta := map[string]string{
		"content_type":   res.ContentType,
		"bytes_total":    sprintInt(len(res.Body)),
		"parse_strategy": "ledongthuc",
		"parse_mode":     mode,
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

// resolveMode は pdf.mode を解決し、未対応モードは fast にフォールバックする。
func (p *parser) resolveMode() string {
	mode := string(model.PDFModeFast)
	if p.host != nil {
		if v, ok := p.host.Config("pdf.mode"); ok && v != "" {
			mode = v
		}
	}
	switch model.PDFParseMode(mode) {
	case model.PDFModeFast:
		return string(model.PDFModeFast)
	case model.PDFModeAuto, model.PDFModeOCR:
		slog.Warn("pdf mode not implemented; falling back to fast", "mode", mode)
		return string(model.PDFModeFast)
	default:
		return string(model.PDFModeFast)
	}
}

// resolveMaxPages は pdf.max_pages 設定を返す（0 は無制限）。
func (p *parser) resolveMaxPages() int {
	if p.host == nil {
		return 0
	}
	v, ok := p.host.Config("pdf.max_pages")
	if !ok || v == "" {
		return 0
	}
	n, err := parsePositiveInt(v)
	if err != nil {
		return 0
	}
	return n
}

func parsePositiveInt(s string) (int, error) {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("invalid integer %q", s)
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
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
