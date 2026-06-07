package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// CanonicalizeMarkdown は Markdown 本文をハッシュ入力用に正規化する。
func CanonicalizeMarkdown(text string) string {
	return strings.TrimSpace(strings.ReplaceAll(text, "\r\n", "\n"))
}

// ContentHashFromMarkdown は canonical markdown の SHA-256 十六進を返す。
func ContentHashFromMarkdown(markdown string) string {
	canonical := CanonicalizeMarkdown(markdown)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// NewRunID は crawl run 用の一意 ID を生成する。
func NewRunID() string {
	return genID()
}
