// Package main は PDF クロール手動確認用の静的 HTTP サーバー。
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const (
	fixturesDir = "../../../testdata/pdfsite"
	pdfDir      = "../../../testdata/pdf"
	pdfName     = "sample-pdf.pdf"
)

func main() {
	addr := flag.String("addr", ":18766", "listen address")
	flag.Parse()

	indexPath := filepath.Join(fixturesDir, "index.html")
	pdfPath := filepath.Join(pdfDir, pdfName)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveFile(w, indexPath, "text/html; charset=utf-8")
	})
	mux.HandleFunc("/"+pdfName, func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, pdfPath, "application/pdf")
	})

	log.Printf("pdfsite listening on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func serveFile(w http.ResponseWriter, path, contentType string) {
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(data)
}
