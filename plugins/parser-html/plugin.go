// Package html は HTML レスポンスを Content に変換する P5 Parser を提供する。
package html

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterParser("html", func() plugin.Parser { return &parser{} })
}

type parser struct {
	host plugin.Host
}

func (p *parser) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "html",
		Version:     "0.1.0",
		Kind:        plugin.KindParser,
		Description: "HTML レスポンスを goquery でパースする",
	}
}

func (p *parser) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

func (p *parser) Close(_ context.Context) error { return nil }

func (p *parser) CanParse(res *model.Response) bool {
	ct := strings.ToLower(res.ContentType)
	if strings.Contains(ct, "text/html") || strings.Contains(ct, "application/xhtml+xml") {
		return true
	}
	path := strings.ToLower(res.URL.Path)
	return strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") || path == "" || path == "/"
}

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
