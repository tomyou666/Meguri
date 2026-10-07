package exportpreview

import (
	"bytes"
	"io"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
	"golang.org/x/net/html"
)

// TargetChunkBytes はプレビュー行の目標バイト長。
const TargetChunkBytes = 8192

// Fragment は元本文内の 1 行分のバイト範囲。
type Fragment struct {
	// Start は元本文内の開始位置（バイト）。
	Start int
	// Length は断片のバイト長。
	Length int
}

type byteSpan struct {
	start int
	stop  int
}

// SplitMarkdown は markdown をブロック境界で約 TargetChunkBytes ごとに分割する。
//
// 表・リスト・コードブロックなどの途中では割らない。
// 1 ブロックが TargetChunkBytes を超える場合はそのブロックだけで 1 行にする。
func SplitMarkdown(src string) []Fragment {
	if src == "" {
		return nil
	}
	source := []byte(src)
	md := goldmark.New()
	doc := md.Parser().Parse(text.NewReader(source))

	var spans []byteSpan
	for c := doc.FirstChild(); c != nil; c = c.NextSibling() {
		start, stop, ok := nodeByteRange(c)
		if !ok || stop <= start {
			continue
		}
		spans = append(spans, byteSpan{start: start, stop: stop})
	}
	if len(spans) == 0 {
		return []Fragment{{Start: 0, Length: len(source)}}
	}
	return packSpans(spans, TargetChunkBytes)
}

// SplitHTML は HTML を構造タグの子の間で約 TargetChunkBytes ごとに分割する。
//
// html / body / div / section / article / main を潜り、その直下の子の間でまとめる。
// 1 子が TargetChunkBytes を超える場合はその子だけで 1 行にする。
func SplitHTML(src string) []Fragment {
	if strings.TrimSpace(src) == "" {
		return nil
	}
	source := []byte(src)
	spans := htmlChildSpans(source)
	if len(spans) == 0 {
		return []Fragment{{Start: 0, Length: len(source)}}
	}
	return packSpans(spans, TargetChunkBytes)
}

// nodeByteRange はブロックノードが占めるソース範囲を返す。
func nodeByteRange(n ast.Node) (start, stop int, ok bool) {
	start = -1
	stop = -1
	_ = ast.Walk(n, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if node.Type() != ast.TypeBlock {
			return ast.WalkContinue, nil
		}
		lines := node.Lines()
		if lines == nil || lines.Len() == 0 {
			return ast.WalkContinue, nil
		}
		s := lines.At(0).Start
		e := lines.At(lines.Len() - 1).Stop
		if start < 0 || s < start {
			start = s
		}
		if e > stop {
			stop = e
		}
		return ast.WalkContinue, nil
	})
	return start, stop, start >= 0 && stop > start
}

// packSpans は連続する span を maxBytes 以下になるようまとめる。
func packSpans(spans []byteSpan, maxBytes int) []Fragment {
	if len(spans) == 0 {
		return nil
	}
	out := make([]Fragment, 0, len(spans))
	packStart := spans[0].start
	packStop := spans[0].stop
	for i := 1; i < len(spans); i++ {
		next := spans[i]
		if next.stop-packStart <= maxBytes {
			packStop = next.stop
			continue
		}
		out = append(out, Fragment{Start: packStart, Length: packStop - packStart})
		packStart = next.start
		packStop = next.stop
	}
	out = append(out, Fragment{Start: packStart, Length: packStop - packStart})
	return out
}

var structuralTags = map[string]bool{
	"html":    true,
	"body":    true,
	"div":     true,
	"section": true,
	"article": true,
	"main":    true,
}

// htmlChildSpans は分割コンテナ直下の要素ノードのバイト範囲を返す。
func htmlChildSpans(source []byte) []byteSpan {
	root, err := html.Parse(bytes.NewReader(source))
	if err != nil || root == nil {
		return tokenizerTopLevelSpans(source)
	}
	container := findSplitContainer(root)
	if container == nil {
		return tokenizerTopLevelSpans(source)
	}
	path := elementPath(container)
	if len(path) == 0 {
		return tokenizerTopLevelSpans(source)
	}
	spans := spansUnderPath(source, path)
	if len(spans) == 0 {
		return tokenizerTopLevelSpans(source)
	}
	return spans
}

// findSplitContainer は html/body/div/section/article/main を潜った分割点を返す。
func findSplitContainer(n *html.Node) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type == html.ElementNode && structuralTags[n.Data] {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.ElementNode && structuralTags[c.Data] {
				if found := findSplitContainer(c); found != nil {
					return found
				}
			}
		}
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findSplitContainer(c); found != nil {
			return found
		}
	}
	return nil
}

// elementPath は document から n までの要素タグ名パスを返す。
func elementPath(n *html.Node) []string {
	var rev []string
	for cur := n; cur != nil; cur = cur.Parent {
		if cur.Type == html.ElementNode {
			rev = append(rev, cur.Data)
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

// spansUnderPath は path で示す要素の直下子要素のバイト範囲をトークナイザで取る。
func spansUnderPath(source []byte, path []string) []byteSpan {
	z := html.NewTokenizer(bytes.NewReader(source))
	offset := 0
	matched := 0
	inside := false
	var spans []byteSpan
	childStart := -1
	childDepth := 0
	var open []string

	popOpen := func(name string) {
		for i := len(open) - 1; i >= 0; i-- {
			if open[i] == name {
				open = open[:i]
				return
			}
		}
	}

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				if childStart >= 0 && offset > childStart {
					spans = append(spans, byteSpan{start: childStart, stop: offset})
				}
				break
			}
			return nil
		}
		raw := z.Raw()
		tokStart := offset
		offset += len(raw)
		nameBytes, _ := z.TagName()
		name := string(nameBytes)

		switch tt {
		case html.StartTagToken:
			open = append(open, name)
			if !inside {
				if matched < len(path) && name == path[matched] {
					matched++
					if matched == len(path) {
						inside = true
					}
				}
			} else if childStart < 0 {
				childStart = tokStart
				childDepth = 1
			} else {
				childDepth++
			}
			if isVoid(name) {
				popOpen(name)
				if inside && childStart >= 0 {
					childDepth--
					if childDepth == 0 {
						spans = append(spans, byteSpan{start: childStart, stop: offset})
						childStart = -1
					}
				}
				if inside && matched == len(path) && len(open) < len(path) {
					inside = false
					matched = len(open)
				}
			}

		case html.SelfClosingTagToken:
			if inside && childStart < 0 {
				spans = append(spans, byteSpan{start: tokStart, stop: offset})
			}

		case html.EndTagToken:
			if inside && childStart >= 0 {
				childDepth--
				if childDepth == 0 {
					spans = append(spans, byteSpan{start: childStart, stop: offset})
					childStart = -1
				}
			}
			popOpen(name)
			if inside && len(open) < len(path) {
				inside = false
				// パス再同期
				matched = 0
				for i := 0; i < len(open) && matched < len(path); i++ {
					if open[i] == path[matched] {
						matched++
					}
				}
				if matched == len(path) {
					inside = true
				}
			}
		}
	}
	return spans
}

// tokenizerTopLevelSpans は最上位の要素範囲を返す（断片 HTML 用）。
func tokenizerTopLevelSpans(source []byte) []byteSpan {
	z := html.NewTokenizer(bytes.NewReader(source))
	offset := 0
	depth := 0
	var spans []byteSpan
	start := -1

	for {
		tt := z.Next()
		if tt == html.ErrorToken {
			if z.Err() == io.EOF {
				if start >= 0 && offset > start {
					spans = append(spans, byteSpan{start: start, stop: offset})
				}
				break
			}
			return nil
		}
		raw := z.Raw()
		tokStart := offset
		offset += len(raw)
		nameBytes, _ := z.TagName()
		name := string(nameBytes)

		switch tt {
		case html.StartTagToken:
			if depth == 0 {
				start = tokStart
			}
			if isVoid(name) {
				if depth == 0 && start >= 0 {
					spans = append(spans, byteSpan{start: start, stop: offset})
					start = -1
				}
			} else {
				depth++
			}
		case html.SelfClosingTagToken:
			if depth == 0 {
				spans = append(spans, byteSpan{start: tokStart, stop: offset})
			}
		case html.EndTagToken:
			if depth > 0 {
				depth--
			}
			if depth == 0 && start >= 0 {
				spans = append(spans, byteSpan{start: start, stop: offset})
				start = -1
			}
		}
	}
	return spans
}

// isVoid は HTML void 要素かどうかを返す。
func isVoid(name string) bool {
	switch name {
	case "area", "base", "br", "col", "embed", "hr", "img", "input",
		"link", "meta", "param", "source", "track", "wbr":
		return true
	default:
		return false
	}
}
