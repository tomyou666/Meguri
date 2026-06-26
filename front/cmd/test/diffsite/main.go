// Package main は差分 UI 手動確認用の静的 HTTP サーバー。
package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed fixtures/**
var fixturesFS embed.FS

// variant は配信するフィクスチャを表す。
//
// 取りうる値:
// content-a / content-b: 本文差分のみ（links は同一）
// links-a / links-b: 出リンク差分のみ（本文は同一）
// fetch-a / fetch-b: /error.html の HTTP ステータス差分（本文・links は同一）
type variant string

const (
	variantContentA variant = "content-a"
	variantContentB variant = "content-b"
	variantLinksA   variant = "links-a"
	variantLinksB   variant = "links-b"
	variantFetchA   variant = "fetch-a"
	variantFetchB   variant = "fetch-b"
)

func parseVariant(raw string) (variant, error) {
	v := variant(raw)
	switch v {
	case variantContentA, variantContentB, variantLinksA, variantLinksB, variantFetchA, variantFetchB:
		return v, nil
	default:
		return "", fmt.Errorf("unknown variant %q", raw)
	}
}

func (v variant) scenario() string {
	parts := strings.SplitN(string(v), "-", 2)
	return parts[0]
}

func (v variant) phase() string {
	parts := strings.SplitN(string(v), "-", 2)
	if len(parts) < 2 {
		return "a"
	}
	return parts[1]
}

func main() {
	variantFlag := flag.String("variant", string(variantContentA), "fixture variant: content-a|content-b|links-a|links-b|fetch-a|fetch-b")
	addr := flag.String("addr", ":18765", "listen address")
	flag.Parse()

	v, err := parseVariant(*variantFlag)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		serveFixture(w, v.scenario(), v.phase(), "index.html")
	})
	mux.HandleFunc("/error.html", func(w http.ResponseWriter, r *http.Request) {
		if v == variantFetchB {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		phase := "a"
		if v.scenario() == "fetch" {
			phase = v.phase()
		}
		path := fmt.Sprintf("fixtures/fetch/%s/error.html", phase)
		data, err := fs.ReadFile(fixturesFS, path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	log.Printf("diffsite listening on %s variant=%s", *addr, v)
	log.Fatal(http.ListenAndServe(*addr, mux))
}

func serveFixture(w http.ResponseWriter, scenario, phase, name string) {
	path := fmt.Sprintf("fixtures/%s/%s/%s", scenario, phase, name)
	data, err := fs.ReadFile(fixturesFS, path)
	if err != nil {
		http.Error(w, "fixture not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}
