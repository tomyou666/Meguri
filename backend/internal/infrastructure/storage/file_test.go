package storage_test

import (
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/domain/model"
	"meguri/internal/infrastructure/storage"
)

// TestFileWriter_Write は結果ファイルの書き出しパターンと連番を検証する。
func TestFileWriter_Write(t *testing.T) {
	t.Run("正常系: Markdown 1ファイルが {seq}-{host}-{path}.md で書き出される", func(t *testing.T) {
		tmp := t.TempDir()
		u, _ := url.Parse("https://example.com/docs/intro")

		w := storage.NewFileWriter(
			model.OutputConfig{Dir: tmp, FilePattern: "{seq}-{host}-{path}.{ext}"},
			[]model.OutputFormat{model.FormatMarkdown},
		)

		err := w.Write(&model.Result{URL: u, Markdown: "# タイトル"})
		assert.NoError(t, err)

		entries, _ := os.ReadDir(tmp)
		assert.Len(t, entries, 1)
		name := entries[0].Name()
		assert.Contains(t, name, "00000-example.com-docs_intro.md")

		body, _ := os.ReadFile(filepath.Join(tmp, name))
		assert.Equal(t, "# タイトル", string(body))
	})

	t.Run("正常系: 複数フォーマット指定でフォーマットごとに別ファイルが出る", func(t *testing.T) {
		tmp := t.TempDir()
		u, _ := url.Parse("https://example.com/")
		link, _ := url.Parse("https://example.com/next")

		w := storage.NewFileWriter(
			model.OutputConfig{Dir: tmp, FilePattern: "{seq}-{path}.{ext}"},
			[]model.OutputFormat{model.FormatMarkdown, model.FormatLinks},
		)

		err := w.Write(&model.Result{URL: u, Markdown: "body", Links: []*url.URL{link}})

		assert.NoError(t, err)
		entries, _ := os.ReadDir(tmp)
		assert.Len(t, entries, 2)

		var hasMD, hasLinks bool
		for _, e := range entries {
			switch {
			case filepath.Ext(e.Name()) == ".md":
				hasMD = true
			case e.Name() == "00000-index.links.txt":
				hasLinks = true
			}
		}
		assert.True(t, hasMD, ".md が存在する")
		assert.True(t, hasLinks, ".links.txt が存在する: %v", entries)
	})

	t.Run("正常系: 連番(seq)はWriteごとにインクリメントされる", func(t *testing.T) {
		tmp := t.TempDir()
		u1, _ := url.Parse("https://a.example.com/x")
		u2, _ := url.Parse("https://b.example.com/y")

		w := storage.NewFileWriter(
			model.OutputConfig{Dir: tmp, FilePattern: "{seq}.{ext}"},
			[]model.OutputFormat{model.FormatMarkdown},
		)

		assert.NoError(t, w.Write(&model.Result{URL: u1, Markdown: "1"}))
		assert.NoError(t, w.Write(&model.Result{URL: u2, Markdown: "2"}))

		_, err := os.Stat(filepath.Join(tmp, "00000.md"))
		assert.NoError(t, err)
		_, err = os.Stat(filepath.Join(tmp, "00001.md"))
		assert.NoError(t, err)
	})
}
