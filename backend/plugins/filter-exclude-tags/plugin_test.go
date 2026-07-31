package excludetags

import (
	"context"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// contentHost は ExcludeTags だけを返す Host モック。
type contentHost struct {
	content model.ContentConfig
}

func (h contentHost) Config(string) (string, bool)       { return "", false }
func (h contentHost) RequestConfig() model.RequestConfig { return model.RequestConfig{} }
func (h contentHost) FetcherConfig() model.FetcherConfig { return model.FetcherConfig{} }
func (h contentHost) StealthConfig() model.StealthConfig { return model.StealthConfig{} }
func (h contentHost) FetcherKind() model.FetcherKind     { return model.FetcherHTTP }
func (h contentHost) ContentConfig() model.ContentConfig { return h.content }

var _ plugin.Host = contentHost{}

func htmlContent(t *testing.T, html string) *model.Content {
	t.Helper()
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	require.NoError(t, err)
	return &model.Content{
		Format: "html",
		DOM:    doc,
		Text:   doc.Text(),
	}
}

// TestFilter_ExcludeTags は exclude_tags によるタグ単位の要素除去を検証する。
func TestFilter_ExcludeTags(t *testing.T) {
	const src = `<html><body>
<article>KEEP</article>
<dialog id="modal">REMOVE_DIALOG</dialog>
<script>REMOVE_SCRIPT</script>
</body></html>`

	t.Run("正常系: 指定タグの要素を除去する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeTags: []string{"dialog", "script"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		html, err := out.DOM.(*goquery.Document).Html()
		require.NoError(t, err)
		assert.Contains(t, html, "KEEP")
		assert.NotContains(t, html, "REMOVE_DIALOG")
		assert.NotContains(t, html, "REMOVE_SCRIPT")
		assert.NotContains(t, out.Text, "REMOVE_DIALOG")
	})

	t.Run("正常系: 大文字指定でも除去する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeTags: []string{"DIALOG"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.NotContains(t, out.Text, "REMOVE_DIALOG")
	})

	t.Run("正常系: 空リストなら DOM を変更しない", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeTags: []string{}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "REMOVE_DIALOG")
	})

	t.Run("異常系: タグ名以外のエントリはスキップし他タグは適用する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeTags: []string{"", "  ", "#modal", ".foo", "dialog"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.NotContains(t, out.Text, "REMOVE_DIALOG")
		assert.Contains(t, out.Text, "REMOVE_SCRIPT")
	})

	t.Run("正常系: html 以外のフォーマットは何もしない", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeTags: []string{"dialog"}},
		}))
		c := &model.Content{Format: "pdf", Text: "REMOVE_DIALOG"}

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Equal(t, "REMOVE_DIALOG", out.Text)
	})
}
