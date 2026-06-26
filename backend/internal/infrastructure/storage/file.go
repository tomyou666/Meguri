// Package storage はファイルベースの結果出力を提供する。
package storage

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"meguri/internal/domain/model"
)

// FileWriter は OutputConfig に従い *model.Result を保存する。
type FileWriter struct {
	// dir は出力ディレクトリ。
	dir string
	// filePattern はファイル名テンプレート。
	filePattern string
	// formats は書き出す OutputFormat の一覧。
	formats []model.OutputFormat

	// mu は seq 更新の排他制御。
	mu sync.Mutex
	// seq は次に割り当てる連番。
	seq int
}

// NewFileWriter は OutputConfig とフォーマット一覧から FileWriter を構築する。
func NewFileWriter(out model.OutputConfig, formats []model.OutputFormat) *FileWriter {
	return &FileWriter{
		dir:         out.Dir,
		filePattern: out.FilePattern,
		formats:     formats,
	}
}

// Write は要求フォーマットごとに 1 ファイルを書き出す。
// 例: markdown と links を指定したら .md と links.txt の2ファイル。
func (w *FileWriter) Write(r *model.Result) error {
	w.mu.Lock()
	seq := w.seq
	w.seq++
	w.mu.Unlock()

	if err := os.MkdirAll(w.dir, 0o755); err != nil {
		return fmt.Errorf("mkdir output dir: %w", err)
	}

	for _, f := range w.formats {
		name, content, err := w.render(r, f, seq)
		if err != nil {
			return err
		}
		full := filepath.Join(w.dir, name)
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", full, err)
		}
	}
	return nil
}

var pathSanitizer = regexp.MustCompile(`[^a-zA-Z0-9_\-]+`)

// render は 1 フォーマット分のファイル名と本文を生成する。
func (w *FileWriter) render(r *model.Result, f model.OutputFormat, seq int) (string, string, error) {
	ext := ""
	body := ""
	switch f {
	case model.FormatMarkdown:
		ext = "md"
		body = r.Markdown
	case model.FormatHTML:
		ext = "html"
		body = r.HTML
	case model.FormatRawHTML:
		ext = "raw.html"
		body = r.RawHTML
	case model.FormatJSON:
		ext = "json"
		j := map[string]any{
			"url":      urlString(r.URL),
			"metadata": r.Metadata,
			"text":     r.Markdown,
		}
		b, err := json.MarshalIndent(j, "", "  ")
		if err != nil {
			return "", "", err
		}
		body = string(b)
	case model.FormatLinks:
		ext = "links.txt"
		var sb strings.Builder
		for _, l := range r.Links {
			sb.WriteString(l.String())
			sb.WriteByte('\n')
		}
		body = sb.String()
	case model.FormatMetadata:
		ext = "metadata.txt"
		var sb strings.Builder
		for k, v := range r.Metadata {
			sb.WriteString(k)
			sb.WriteString(": ")
			sb.WriteString(v)
			sb.WriteByte('\n')
		}
		body = sb.String()
	default:
		return "", "", fmt.Errorf("unknown format: %s", f)
	}

	name := buildFileName(w.filePattern, seq, r.URL, ext)
	return name, body, nil
}

// buildFileName はテンプレートと URL から出力ファイル名を組み立てる。
func buildFileName(pattern string, seq int, u *url.URL, ext string) string {
	host := ""
	path := ""
	if u != nil {
		host = u.Host
		path = u.Path
	}
	pathSafe := pathSanitizer.ReplaceAllString(strings.Trim(path, "/"), "_")
	if pathSafe == "" {
		pathSafe = "index"
	}
	r := strings.NewReplacer(
		"{seq}", fmt.Sprintf("%05d", seq),
		"{host}", strings.ReplaceAll(host, ":", "_"),
		"{path}", pathSafe,
		"{ext}", ext,
	)
	return r.Replace(pattern)
}

// urlString は nil 安全に URL 文字列を返す。
func urlString(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
