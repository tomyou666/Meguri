package wails_service

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// FetchRobotsTxt の HTTP 応答パターンを検証する。
func TestFetchRobotsTxt(t *testing.T) {
	svc := &ScraperService{}

	t.Run("正常系: 200 応答は found と body を返す", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("User-agent: *\nDisallow: /private"))
		}))
		defer srv.Close()

		u, err := url.Parse(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		info, err := svc.FetchRobotsTxt(u.Host, srv.URL+"/seed")
		if err != nil {
			t.Fatal(err)
		}
		if info.Status != "found" {
			t.Fatalf("status=%q", info.Status)
		}
		if info.Body == "" {
			t.Fatal("expected body")
		}
	})

	t.Run("正常系: 404 応答は not_found を返す", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer srv.Close()

		u, err := url.Parse(srv.URL)
		if err != nil {
			t.Fatal(err)
		}
		info, err := svc.FetchRobotsTxt(u.Host, srv.URL+"/seed")
		if err != nil {
			t.Fatal(err)
		}
		if info.Status != "not_found" {
			t.Fatalf("status=%q", info.Status)
		}
	})
}
