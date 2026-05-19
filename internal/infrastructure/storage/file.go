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

	"scraperbot/internal/domain/model"
)

// FileWriter は OutputConfig に従い *model.Result を保存する。
type FileWriter struct {
	dir         string
	filePattern string
	formats     []model.OutputFormat

	mu  sync.Mutex
	seq int
}

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

func urlString(u *url.URL) string {
	if u == nil {
		return ""
	}
	return u.String()
}
