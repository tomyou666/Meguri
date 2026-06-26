package htmlfmt

import (
	"bytes"
	"context"
	"net/url"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
)

// TestTransformer_Transform は HTML 整形 Transformer の変換結果を検証する。
func TestTransformer_Transform(t *testing.T) {
	t.Parallel()

	tr := &transformer{}

	t.Run("正常系: HTML DOM を整形し Result.HTML に改行付きで出力する", func(t *testing.T) {
		t.Parallel()

		minified := `<html><body><div><p>hello</p></div></body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(minified))
		require.NoError(t, err)

		u, err := url.Parse("https://example.com/")
		require.NoError(t, err)

		c := &model.Content{
			URL:    u,
			Format: "html",
			DOM:    doc,
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Contains(t, r.HTML, "\n", "gohtml 整形後は改行が入る")
		assert.Contains(t, r.HTML, "hello")
		assert.NotEqual(t, minified, strings.TrimSpace(r.HTML))
	})

	t.Run("正常系: PDF は section ラップの HTML を出力する", func(t *testing.T) {
		t.Parallel()

		u, err := url.Parse("https://example.com/doc.pdf")
		require.NoError(t, err)

		c := &model.Content{
			URL:    u,
			Format: "pdf",
			Text:   "page one text",
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Contains(t, r.HTML, "<section>")
		assert.Contains(t, r.HTML, "page one text")
		assert.Contains(t, r.HTML, "\n")
	})

	t.Run("正常系: Text フォールバックでも整形 HTML を出力する", func(t *testing.T) {
		t.Parallel()

		c := &model.Content{
			Format: "html",
			Text:   "<div><span>fallback</span></div>",
		}

		r, err := tr.Transform(context.Background(), c)
		require.NoError(t, err)
		assert.Contains(t, r.HTML, "fallback")
		assert.Contains(t, r.HTML, "\n")
	})
}

// TestSerializeHTML は DOM から HTML 文字列を取り出す処理を検証する。
func TestSerializeHTML(t *testing.T) {
	t.Parallel()

	t.Run("正常系: goquery Document から HTML をシリアライズする", func(t *testing.T) {
		t.Parallel()

		body := `<html><body><p>test</p></body></html>`
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader([]byte(body)))
		require.NoError(t, err)

		got, err := serializeHTML(&model.Content{DOM: doc})
		require.NoError(t, err)
		assert.Contains(t, got, "test")
	})
}
