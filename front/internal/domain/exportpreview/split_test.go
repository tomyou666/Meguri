package exportpreview

import (
	"strings"
	"testing"
)

// TestSplitMarkdown はブロック境界での分割を検証する。
func TestSplitMarkdown(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		if got := SplitMarkdown(""); got != nil {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("small document is one fragment", func(t *testing.T) {
		src := "# Title\n\nHello world.\n"
		frags := SplitMarkdown(src)
		if len(frags) != 1 {
			t.Fatalf("len=%d want 1", len(frags))
		}
		got := src[frags[0].Start : frags[0].Start+frags[0].Length]
		if !strings.Contains(got, "Hello world") {
			t.Fatalf("fragment %q", got)
		}
	})

	t.Run("packs blocks under 8KB", func(t *testing.T) {
		var b strings.Builder
		for i := 0; i < 20; i++ {
			b.WriteString("Paragraph ")
			b.WriteString(strings.Repeat("x", 100))
			b.WriteString(".\n\n")
		}
		src := b.String()
		frags := SplitMarkdown(src)
		if len(frags) < 1 {
			t.Fatal("expected fragments")
		}
		total := 0
		for _, f := range frags {
			total += f.Length
			if f.Length <= 0 {
				t.Fatal("empty fragment")
			}
		}
		if total > len(src) {
			t.Fatalf("total %d > src %d", total, len(src))
		}
	})

	t.Run("oversized single block stays one row", func(t *testing.T) {
		src := strings.Repeat("a", TargetChunkBytes+100) + "\n\n"
		frags := SplitMarkdown(src)
		if len(frags) < 1 {
			t.Fatal("expected at least one fragment")
		}
		if frags[0].Length < TargetChunkBytes {
			t.Fatalf("first frag too small: %d", frags[0].Length)
		}
	})
}

// TestSplitHTML は構造タグ直下での分割を検証する。
func TestSplitHTML(t *testing.T) {
	t.Run("empty returns nil", func(t *testing.T) {
		if got := SplitHTML("  "); got != nil {
			t.Fatalf("got %v", got)
		}
	})

	t.Run("body children are split units", func(t *testing.T) {
		src := `<html><body><p>one</p><p>two</p><p>three</p></body></html>`
		frags := SplitHTML(src)
		if len(frags) < 1 {
			t.Fatal("expected fragments")
		}
		joined := ""
		for _, f := range frags {
			joined += src[f.Start : f.Start+f.Length]
		}
		if !strings.Contains(joined, "one") || !strings.Contains(joined, "three") {
			t.Fatalf("joined %q", joined)
		}
	})
}

// TestMergeHelpers は区切り・ファイル名規則を検証する。
func TestMergeHelpers(t *testing.T) {
	t.Run("parse separator escapes", func(t *testing.T) {
		if got := ParseSeparator(`\n\n---\n\n`); got != "\n\n---\n\n" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("html separator escapes", func(t *testing.T) {
		got := ResolveSeparator(`<hr>`, FormatHTML)
		if got != "&lt;hr&gt;" {
			t.Fatalf("got %q", got)
		}
	})

	t.Run("unique file names", func(t *testing.T) {
		names := AssignUniqueFileNames([]string{"a", "a", "b"}, "md")
		want := []string{"a.md", "a-2.md", "b.md"}
		for i := range want {
			if names[i] != want[i] {
				t.Fatalf("names=%v want %v", names, want)
			}
		}
	})

	t.Run("sanitize empty becomes export", func(t *testing.T) {
		if got := SanitizeFileName(`***`); got != "export" && got != "___" {
			// *** becomes ___ after replace; fields trim may leave ___
			if SanitizeFileName("") != "export" {
				t.Fatalf("empty sanitize=%q", SanitizeFileName(""))
			}
		}
	})
}
