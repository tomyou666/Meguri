// Package defaultlinks は <a href> から相対URLを解決して抽出する P8 LinkExtractor を提供する。
package defaultlinks

import (
	"context"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterLinkExtractor("default", func() plugin.LinkExtractor { return &extractor{} })
}

type extractor struct {
	host plugin.Host
}

func (e *extractor) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "default",
		Version:     "0.1.0",
		Kind:        plugin.KindLinkExtractor,
		Description: "<a href> から URL を抽出し base に対して解決する",
	}
}

func (e *extractor) Init(_ context.Context, host plugin.Host) error {
	e.host = host
	return nil
}

func (e *extractor) Close(_ context.Context) error { return nil }

func (e *extractor) Extract(_ context.Context, c *model.Content, base *url.URL) ([]*url.URL, error) {
	if c.Format != "html" {
		return nil, nil
	}
	doc, ok := c.DOM.(*goquery.Document)
	if !ok {
		return nil, nil
	}

	var out []*url.URL
	seen := map[string]struct{}{}

	doc.Find("a[href]").Each(func(_ int, s *goquery.Selection) {
		raw, _ := s.Attr("href")
		raw = strings.TrimSpace(raw)
		if raw == "" || strings.HasPrefix(raw, "#") || strings.HasPrefix(raw, "javascript:") {
			return
		}
		ref, err := url.Parse(raw)
		if err != nil {
			return
		}
		resolved := base.ResolveReference(ref)
		if resolved.Scheme != "http" && resolved.Scheme != "https" {
			return
		}
		resolved.Fragment = ""
		key := resolved.String()
		if _, dup := seen[key]; dup {
			return
		}
		seen[key] = struct{}{}
		out = append(out, resolved)
	})

	return out, nil
}
