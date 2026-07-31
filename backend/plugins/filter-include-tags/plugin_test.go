package includetags

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

// contentHost は IncludeTags だけを返す Host モック。
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

// TestFilter_IncludeTags は include_tags によるタグ単位の whitelist 絞り込みを検証する。
func TestFilter_IncludeTags(t *testing.T) {
	const src = `<html><body>
<article>KEEP_ARTICLE</article>
<section>DROP_SECTION</section>
<dialog>DROP_DIALOG</dialog>
</body></html>`

	t.Run("正常系: 指定タグの要素だけを残す", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{"article"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "KEEP_ARTICLE")
		assert.NotContains(t, out.Text, "DROP_SECTION")
		assert.NotContains(t, out.Text, "DROP_DIALOG")
	})

	t.Run("正常系: 大文字指定でも残す", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{"ARTICLE"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "KEEP_ARTICLE")
		assert.NotContains(t, out.Text, "DROP_SECTION")
	})

	t.Run("正常系: 空リストなら DOM を変更しない", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "DROP_SECTION")
	})

	t.Run("正常系: 一致ゼロなら空 Content にする", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{"main"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Empty(t, strings.TrimSpace(out.Text))
	})

	t.Run("異常系: タグ名以外のエントリはスキップし他タグは適用する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{"", "  ", "#x", ".foo", "article"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "KEEP_ARTICLE")
		assert.NotContains(t, out.Text, "DROP_SECTION")
	})

	t.Run("正常系: html 以外のフォーマットは何もしない", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{IncludeTags: []string{"article"}},
		}))
		c := &model.Content{Format: "pdf", Text: "KEEP"}

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Equal(t, "KEEP", out.Text)
	})
}
