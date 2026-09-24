package domain

import (
	"testing"

	"meguri-app/internal/model"
)

// TestNodeResultToPreviewHTML は html / raw_html が DTO にマップされることを検証する。
func TestNodeResultToPreviewHTML(t *testing.T) {
	html := "<p>filtered</p>"
	raw := "<html>raw</html>"
	jsonBody := `{"k":"v"}`
	meta := model.NodeResult{
		URL:            "https://example.com",
		ManuallyEdited: 1,
	}
	body := &model.NodeResultBody{
		HTML:     &html,
		RawHTML:  &raw,
		JSONBody: &jsonBody,
	}
	dto := nodeResultToPreview(meta, body)
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
