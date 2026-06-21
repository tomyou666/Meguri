package runner_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"scraperbot/pkg/runner"
)

// testRobotsConfigLayer は httptest 向けの最小 HTTP fetcher 設定を返す。
func testRobotsConfigLayer(t *testing.T) json.RawMessage {
	t.Helper()
	return json.RawMessage(`{"plugins":{"fetcher":"http"},"request":{"timeout":"10s","retry_count":0}}`)
}

// TestFetchRobotsTxt は P3 Fetcher 経由の robots.txt 取得パターンを検証する。
func TestFetchRobotsTxt(t *testing.T) {
	cfgLayer := testRobotsConfigLayer(t)

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
		res, err := runner.FetchRobotsTxt(context.Background(), u.Host, srv.URL+"/seed", cfgLayer)
		if err != nil {
			t.Fatal(err)
		}
		if res.Status != "found" {
			t.Fatalf("status=%q", res.Status)
		}
		if res.Body == "" {
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
		res, err := runner.FetchRobotsTxt(context.Background(), u.Host, srv.URL+"/seed", cfgLayer)
		if err != nil {
			t.Fatal(err)
		}
		if res.Status != "not_found" {
			t.Fatalf("status=%q", res.Status)
		}
	})

	t.Run("異常系: host 未指定は Go error", func(t *testing.T) {
		_, err := runner.FetchRobotsTxt(context.Background(), "", "https://example.com/seed", cfgLayer)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}
