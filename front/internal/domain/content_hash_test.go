package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"scraperbot-front/internal/domain"
)

func TestContentHashFromMarkdownMatchesCanonical(t *testing.T) {
	// golden: same algorithm as front/frontend/src/lib/contentHash.ts
	const markdown = "hello\r\nworld  "
	hash := domain.ContentHashFromMarkdown(markdown)
	assert.NotEmpty(t, hash)
	assert.Len(t, hash, 64)

	hash2 := domain.ContentHashFromMarkdown("hello\nworld")
	assert.Equal(t, hash, hash2)
}
