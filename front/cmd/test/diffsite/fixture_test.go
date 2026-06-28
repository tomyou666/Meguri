package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestFixtures は各シナリオの a/b フィクスチャ制約を検証する。
func TestFixtures(t *testing.T) {
	t.Run("content: a と b で本文が異なる", func(t *testing.T) {
		a := readFixture(t, fmt.Sprintf("%s/content/a/index.html", fixturesDir))
		b := readFixture(t, fmt.Sprintf("%s/content/b/index.html", fixturesDir))
		if a == b {
			t.Fatal("content a and b should differ")
		}
		if extractLinks(a) != extractLinks(b) {
			t.Fatal("content scenario links should match between a and b")
		}
	})

	t.Run("links: a と b で出リンクが異なり本文は同一", func(t *testing.T) {
		a := readFixture(t, fmt.Sprintf("%s/links/a/index.html", fixturesDir))
		b := readFixture(t, fmt.Sprintf("%s/links/b/index.html", fixturesDir))
		if stripLinks(a) != stripLinks(b) {
			t.Fatal("links scenario body text should match between a and b")
		}
		if extractLinks(a) == extractLinks(b) {
			t.Fatal("links a and b should differ")
		}
	})

	t.Run("fetch: / の本文は a と b で同一", func(t *testing.T) {
		a := readFixture(t, fmt.Sprintf("%s/fetch/a/index.html", fixturesDir))
		b := readFixture(t, fmt.Sprintf("%s/fetch/b/index.html", fixturesDir))
		if a != b {
			t.Fatal("fetch root pages should be identical")
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

func extractLinks(html string) string {
	var links []string
	for {
		i := strings.Index(html, `href="`)
		if i < 0 {
			break
		}
		html = html[i+6:]
		j := strings.Index(html, `"`)
		if j < 0 {
			break
		}
		links = append(links, html[:j])
		html = html[j+1:]
	}
	return strings.Join(links, "|")
}

func stripLinks(html string) string {
	for strings.Contains(html, "<a ") {
		start := strings.Index(html, "<a ")
		end := strings.Index(html[start:], "</a>")
		if end < 0 {
			break
		}
		html = html[:start] + html[start+end+4:]
	}
	return html
}
