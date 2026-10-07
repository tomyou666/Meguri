package exportpreview

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// Format はエクスポート形式。
type Format string

const (
	// FormatMarkdown は Markdown 出力。
	FormatMarkdown Format = "markdown"
	// FormatHTML は HTML 出力。
	FormatHTML Format = "html"
)

// HeadingField は見出しに使う項目。
type HeadingField string

const (
	// HeadingURL は正規化 URL を見出しにする。
	HeadingURL HeadingField = "url"
	// HeadingLabel はラベルを見出しにする。
	HeadingLabel HeadingField = "label"
)

// NodeMeta は見出し・ファイル名用のノード情報。
type NodeMeta struct {
	// ID はグラフノード ID。
	ID string
	// URLNormalized は正規化 URL。
	URLNormalized string
	// Label は表示ラベル。
	Label string
}

// MergeSettings は連結・保存の設定。
type MergeSettings struct {
	// Format は出力形式。
	Format Format
	// Separator はノード間区切り（エスケープ前の生文字列）。
	Separator string
	// IncludeHeading は見出しを付けるか。
	IncludeHeading bool
	// HeadingField は見出しに使う項目。
	HeadingField HeadingField
	// SplitSave はノードごとに ZIP 分割するか。
	SplitSave bool
}

var unsafeFileNameChars = regexp.MustCompile(`[\\/:*?"<>|]`)

// ParseSeparator は区切り文字のエスケープシーケンスを実文字に変換する。
func ParseSeparator(raw string) string {
	return strings.NewReplacer(
		`\r\n`, "\r\n",
		`\n`, "\n",
		`\r`, "\r",
		`\t`, "\t",
		`\\`, `\`,
	).Replace(raw)
}

// ResolveSeparator は形式に応じた区切り文字を返す。
//
// HTML のときは & < > " ' をエスケープする。
func ResolveSeparator(raw string, format Format) string {
	parsed := ParseSeparator(raw)
	if format == FormatHTML {
		return escapeHTMLText(parsed)
	}
	return parsed
}

// escapeHTMLText は HTML テキスト用にエスケープする。
func escapeHTMLText(text string) string {
	replacer := strings.NewReplacer(
		`&`, "&amp;",
		`<`, "&lt;",
		`>`, "&gt;",
		`"`, "&quot;",
		`'`, "&#39;",
	)
	return replacer.Replace(text)
}

// HeadingLine はノード見出し行（## text）を返す。
func HeadingLine(meta NodeMeta, field HeadingField) string {
	text := meta.URLNormalized
	if field == HeadingLabel {
		text = meta.Label
	}
	return "## " + text
}

// WrapNodeContent は見出し付きなら本文の前に見出しを付ける。
func WrapNodeContent(body string, meta NodeMeta, settings MergeSettings) string {
	if body == "" {
		return ""
	}
	if !settings.IncludeHeading {
		return body
	}
	return HeadingLine(meta, settings.HeadingField) + "\n\n" + body
}

// JoinNodeContents は複数ノード本文を区切りで連結する。
//
// contents は空でない本文のみを順序どおりに渡す。先頭の前には区切りを置かない。
func JoinNodeContents(contents []string, settings MergeSettings) string {
	if len(contents) == 0 {
		return ""
	}
	sep := ResolveSeparator(settings.Separator, settings.Format)
	return strings.Join(contents, sep)
}

// SanitizeFileName はファイル名に使えない文字を除去する。
func SanitizeFileName(raw string) string {
	trimmed := unsafeFileNameChars.ReplaceAllString(raw, "_")
	trimmed = strings.Join(strings.Fields(trimmed), " ")
	trimmed = strings.TrimSpace(trimmed)
	if len(trimmed) > 120 {
		trimmed = trimmed[:120]
	}
	if trimmed == "" {
		return "export"
	}
	return trimmed
}

// BaseNameForMeta は ZIP エントリの基の名前を返す。
func BaseNameForMeta(meta NodeMeta, field HeadingField) string {
	if field == HeadingLabel {
		return meta.Label
	}
	u, err := url.Parse(meta.URLNormalized)
	if err != nil {
		return meta.URLNormalized
	}
	path := strings.TrimSuffix(u.Path, "/")
	parts := strings.Split(path, "/")
	var segment string
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" {
			segment = parts[i]
			break
		}
	}
	if segment != "" {
		return segment
	}
	if u.Hostname() != "" {
		return u.Hostname()
	}
	return meta.URLNormalized
}

// AssignUniqueFileNames は基の名前から一意なファイル名を付ける。
//
// 1 件目は stem.ext、2 件目以降は stem-2.ext、stem-3.ext（stem-1 は使わない）。
func AssignUniqueFileNames(bases []string, ext string) []string {
	seen := map[string]int{}
	out := make([]string, len(bases))
	for i, base := range bases {
		stem := SanitizeFileName(base)
		count := seen[stem]
		seen[stem] = count + 1
		if count == 0 {
			out[i] = fmt.Sprintf("%s.%s", stem, ext)
		} else {
			out[i] = fmt.Sprintf("%s-%d.%s", stem, count+1, ext)
		}
	}
	return out
}

// RowID はノード ID と開始位置から行 ID を作る。
func RowID(nodeID string, start int) string {
	return fmt.Sprintf("%s:%d", nodeID, start)
}

// ExtForFormat は形式に対応する拡張子を返す。
func ExtForFormat(format Format) string {
	if format == FormatHTML {
		return "html"
	}
	return "md"
}
