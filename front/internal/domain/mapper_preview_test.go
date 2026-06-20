package domain

import (
	"testing"

	"scraperbot-front/internal/model"
)

// TestNodeResultToPreviewHTML は html / raw_html が DTO にマップされることを検証する。
func TestNodeResultToPreviewHTML(t *testing.T) {
	html := "<p>filtered</p>"
	raw := "<html>raw</html>"
	row := model.NodeResult{
		URL:     "https://example.com",
		HTML:    &html,
		RawHTML: &raw,
	}
	dto := nodeResultToPreview(row)
	if dto.HTML != html {
		t.Fatalf("HTML: got %q want %q", dto.HTML, html)
	}
	if dto.RawHTML != raw {
		t.Fatalf("RawHTML: got %q want %q", dto.RawHTML, raw)
	}
}
