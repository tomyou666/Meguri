package domain

import "testing"

func TestNormalizeCrawlURL(t *testing.T) {
	got, err := NormalizeCrawlURL("HTTPS://Example.COM:443/path/")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://example.com/path/"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
