package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFixtures は pdfsite フィクスチャの制約を検証する。
func TestFixtures(t *testing.T) {
	t.Run("index.html に PDF リンクがある", func(t *testing.T) {
		html := readFixture(t, filepath.Join(fixturesDir, "index.html"))
		if !strings.Contains(html, `href="/`+pdfName+`"`) {
			t.Fatalf("index.html should link to /%s", pdfName)
		}
	})

	t.Run("PDF ファイルが存在し PDF ヘッダを持つ", func(t *testing.T) {
		data := readBytes(t, filepath.Join(pdfDir, pdfName))
		if !strings.HasPrefix(string(data), "%PDF-") {
			t.Fatal("sample PDF should start with %PDF- header")
		}
	})
}

func readFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func readBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}
