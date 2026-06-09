package domain

import "testing"

// TestNormalizeCrawlURL はクロール用 URL の正規化を検証する。
func TestNormalizeCrawlURL(t *testing.T) {
	t.Run("正常系: スキーム・ホストを小文字化しデフォルトポートと末尾スラッシュを整理する", func(t *testing.T) {
		got, err := NormalizeCrawlURL("HTTPS://Example.COM:443/path/")
		if err != nil {
			t.Fatal(err)
		}
		want := "https://example.com/path/"
		if got != want {
			t.Fatalf("got %q want %q", got, want)
		}
	})
}
