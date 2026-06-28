package pdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	gopdf "github.com/ledongthuc/pdf"
)

const pdfMagic = "%PDF-"

// extractPlainText は ledongthuc/pdf で PDF からプレーンテキストを抽出する（fast モード）。
func extractPlainText(body []byte, maxPages int) (string, error) {
	if !bytes.HasPrefix(body, []byte(pdfMagic)) {
		return "", fmt.Errorf("invalid PDF: missing %q header", pdfMagic)
	}

	reader, err := gopdf.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		return "", fmt.Errorf("pdf read: %w", err)
	}

	pageCount := reader.NumPage()
	if pageCount < 1 {
		return "", fmt.Errorf("pdf: no pages")
	}

	limit := pageCount
	if maxPages > 0 && maxPages < limit {
		limit = maxPages
	}

	if limit == pageCount {
		r, err := reader.GetPlainText()
		if err != nil {
			return "", fmt.Errorf("pdf extract: %w", err)
		}
		b, err := io.ReadAll(r)
		if err != nil {
			return "", fmt.Errorf("pdf read text: %w", err)
		}
		return strings.TrimSpace(string(b)), nil
	}

	fonts := make(map[string]*gopdf.Font)
	var parts []string
	for i := 1; i <= limit; i++ {
		p := reader.Page(i)
		for _, name := range p.Fonts() {
			if _, ok := fonts[name]; !ok {
				f := p.Font(name)
				fonts[name] = &f
			}
		}
		text, err := p.GetPlainText(fonts)
		if err != nil {
			return "", fmt.Errorf("pdf extract page %d: %w", i, err)
		}
		if trimmed := strings.TrimSpace(text); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n")), nil
}
