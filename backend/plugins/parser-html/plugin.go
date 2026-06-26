// Package html は HTML レスポンスを Content に変換する P5 Parser を提供する。
package html

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterParser("html", func() plugin.Parser { return &parser{} })
}

// parser は HTML 用 P5 Parser の実装。
type parser struct {
	// host は Init で受け取る Host。
	host plugin.Host
}

// Metadata は plugin.Parser.Metadata の実装。
func (p *parser) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "html",
		Version:     "0.1.0",
		Kind:        plugin.KindParser,
		Description: "HTML レスポンスを goquery でパースする",
	}
}

// Init は plugin.Plugin.Init の実装。
func (p *parser) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (p *parser) Close(_ context.Context) error { return nil }

// CanParse は HTML 系 Content-Type または .html/.htm パスを判定する。
func (p *parser) CanParse(res *model.Response) bool {
	ct := strings.ToLower(res.ContentType)
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml") {
		return true
	}
	path := strings.ToLower(res.URL.Path)
	return strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") || path == "" || path == "/"
}

// Parse は HTML 本文を goquery で解析し model.Content を返す。
func (p *parser) Parse(_ context.Context, res *model.Response) (*model.Content, error) {
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(res.Body))
	if err != nil {
		return nil, fmt.Errorf("goquery parse: %w", err)
	}

	meta := extractMetadata(doc)
	text := strings.TrimSpace(doc.Find("body").Text())

	return &model.Content{
		URL:      res.URL,
		Format:   "html",
		Text:     text,
		RawHTML:  string(res.Body),
		DOM:      doc,
		Metadata: meta,
	}, nil
}

func extractMetadata(doc *goquery.Document) map[string]string {
	m := map[string]string{}
	if t := strings.TrimSpace(doc.Find("title").First().Text()); t != "" {
		m["title"] = t
	}
	doc.Find(`meta[name="description"]`).Each(func(_ int, s *goquery.Selection) {
		if v, ok := s.Attr("content"); ok {
			m["description"] = v
		}
	})
	doc.Find(`meta[property^="og:"]`).Each(func(_ int, s *goquery.Selection) {
		k, _ := s.Attr("property")
		if v, ok := s.Attr("content"); ok && k != "" {
			m[k] = v
		}
	})
	return m
}
