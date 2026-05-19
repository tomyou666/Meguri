package core_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// newTestWebServer は testdata/html 配下の HTML を返すテストWebサーバーを起動するヘルパ。
// 仕様:
//   - パス末尾が .pdf の場合は application/pdf として PDF 用の固定バイト列を返す
//   - 末尾が .html の場合は testdata/html/<basename> を返す
//   - "/" は simple_page.html を返す
//   - 上記以外は 404
func newTestWebServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		switch {
		case path == "/":
			serveTestFile(t, w, "simple_page.html", "text/html; charset=utf-8")
		case strings.HasSuffix(path, ".pdf"):
			w.Header().Set("Content-Type", "application/pdf")
			// 「ASCII テキストを含む擬似 PDF」。実 PDF パーサーは未実装だが、
			// CanParse / 振り分け / Stub テキスト抽出の動作確認には十分。
			_, _ = w.Write([]byte("%PDF-1.4\nFAKE-PDF-CONTENT line1\nFAKE-PDF-CONTENT line2\n%%EOF\n"))
		case strings.HasSuffix(path, ".html"):
			base := filepath.Base(path)
			// "page-a.html" -> "page_a.html" の柔軟マッピング
			base = strings.ReplaceAll(base, "-", "_")
			serveTestFile(t, w, base, "text/html; charset=utf-8")
		default:
			http.NotFound(w, r)
		}
	})

	return httptest.NewServer(mux)
}

func serveTestFile(t *testing.T, w http.ResponseWriter, name, contentType string) {
	t.Helper()
	path := filepath.Join(testdataDir(t), "html", name)
	b, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "testdata not found: "+name+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(b)
}

func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// internal/core/testserver_test.go の場所からプロジェクトルートへ。
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata")
}
