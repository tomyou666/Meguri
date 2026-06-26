package rawhtml

import (
	"bytes"
	"context"
	"net/url"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
)

// TestTransformer_Transform は生 HTML Transformer の変換結果を検証する。
func TestTransformer_Transform(t *testing.T) {
	t.Parallel()

	tr := &transformer{}

	t.Run("正常系: Content.RawHTML が Result.RawHTML にコピーされる", func(t *testing.T) {
		t.Parallel()

		raw := `<html><head><title>T</title></head><body><p>raw</p></body></html>`
		u, err := url.Parse("https://example.com/page.html")
		require.NoError(t, err)

		c := &model.Content{
			URL:     u,
			Format:  "html",
			RawHTML: raw,
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Equal(t, raw, r.RawHTML)
	})

	t.Run("正常系: RawHTML はフィルタ前の生本文のまま保持される", func(t *testing.T) {
		t.Parallel()

		raw := `<html><body><header>HEADER</header><main><p>body</p></main></body></html>`
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(raw)))
		require.NoError(t, err)
		doc.Find("header").Remove()

		u, err := url.Parse("https://example.com/")
		require.NoError(t, err)

		c := &model.Content{
			URL:     u,
			Format:  "html",
			RawHTML: raw,
			DOM:     doc,
			Text:    doc.Text(),
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Contains(t, r.RawHTML, "HEADER", "生 HTML にはヘッダが残る")
		assert.NotContains(t, doc.Text(), "HEADER", "DOM はフィルタ済み")
	})

	t.Run("正常系: PDF では RawHTML は空で Transform は成功する", func(t *testing.T) {
		t.Parallel()

		u, err := url.Parse("https://example.com/doc.pdf")
		require.NoError(t, err)

		c := &model.Content{
			URL:    u,
			Format: "pdf",
			Text:   "pdf text",
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Empty(t, r.RawHTML)
	})
}
