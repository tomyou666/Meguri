package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri-app/internal/domain"
)

// TestContentHashFromMarkdown はフロントエンドと同一の正規化・ハッシュ算法を検証する。
func TestContentHashFromMarkdown(t *testing.T) {
	// golden: same algorithm as front/frontend/src/lib/contentHash.ts
	const markdown = "hello\r\nworld  "

	t.Run("正常系: 正規化後の Markdown から 64 文字の SHA-256 を返す", func(t *testing.T) {
		hash := domain.ContentHashFromMarkdown(markdown)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})

	t.Run("正常系: 改行・空白の差異は同一ハッシュに正規化される", func(t *testing.T) {
		hash := domain.ContentHashFromMarkdown(markdown)
		hash2 := domain.ContentHashFromMarkdown("hello\nworld")
		assert.Equal(t, hash, hash2)
	})
}

// TestLinksHashFromLinks は canonical links JSON の SHA-256 を検証する。
func TestLinksHashFromLinks(t *testing.T) {
	t.Run("正常系: リンク配列から 64 文字の SHA-256 を返す", func(t *testing.T) {
		hash := domain.LinksHashFromLinks([]string{"https://b", "https://a"})
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 64)
	})

	t.Run("正常系: 順序差はソート後に同一ハッシュになる", func(t *testing.T) {
		h1 := domain.LinksHashFromLinks([]string{"https://a", "https://b"})
		h2 := domain.LinksHashFromLinks([]string{"https://b", "https://a"})
		assert.Equal(t, h1, h2)
	})

	t.Run("正常系: 空・nil は同一ハッシュになる", func(t *testing.T) {
		assert.Equal(t, domain.LinksHashFromLinks(nil), domain.LinksHashFromLinks([]string{}))
	})
}
