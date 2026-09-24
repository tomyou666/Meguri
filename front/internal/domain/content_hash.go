package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
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

// CanonicalLinksJSON はリンク配列をソートした JSON 配列文字列にする。
//
// 空・nil は "[]"。diff の links 判定と links_hash の入力で共有する。
func CanonicalLinksJSON(links []string) string {
	if len(links) == 0 {
		return "[]"
	}
	cp := append([]string(nil), links...)
	sort.Strings(cp)
	b, _ := json.Marshal(cp)
	return string(b)
}

// LinksHashFromLinks は canonical links JSON の SHA-256 十六進を返す。
func LinksHashFromLinks(links []string) string {
	canonical := CanonicalLinksJSON(links)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}

// LinksHashFromLinksJSON は links_json 文字列から links_hash を返す。
//
// 空・不正 JSON は空配列相当として扱う。
func LinksHashFromLinksJSON(linksJSON string) string {
	if linksJSON == "" {
		return LinksHashFromLinks(nil)
	}
	var links []string
	if err := json.Unmarshal([]byte(linksJSON), &links); err != nil {
		return LinksHashFromLinks(nil)
	}
	return LinksHashFromLinks(links)
}

// NewRunID は crawl run 用の一意 ID を生成する。
func NewRunID() string {
	return genID()
}
