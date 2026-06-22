package domain

import (
	"testing"

	"scraperbot-front/internal/model"
)

// TestNodeResultToPreviewHTML は html / raw_html が DTO にマップされることを検証する。
func TestNodeResultToPreviewHTML(t *testing.T) {
	html := "<p>filtered</p>"
	raw := "<html>raw</html>"
	jsonBody := `{"k":"v"}`
	row := model.NodeResult{
		URL:            "https://example.com",
		HTML:           &html,
		RawHTML:        &raw,
		JSONBody:       &jsonBody,
		ManuallyEdited: 1,
	}
	dto := nodeResultToPreview(row)
	if dto.HTML != html {
		t.Fatalf("HTML: got %q want %q", dto.HTML, html)
	}
	if dto.RawHTML != raw {
		t.Fatalf("RawHTML: got %q want %q", dto.RawHTML, raw)
	}
	if dto.JSONBody != jsonBody {
		t.Fatalf("JSONBody: got %q want %q", dto.JSONBody, jsonBody)
	}
	if !dto.ManuallyEdited {
		t.Fatal("expected manuallyEdited true")
	}
}
