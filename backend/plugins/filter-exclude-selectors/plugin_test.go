package excludeselectors

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

// contentHost は ExcludeSelectors だけを返す Host モック。
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

// TestFilter_ExcludeSelectors は exclude_selectors による要素除去を検証する。
func TestFilter_ExcludeSelectors(t *testing.T) {
	const src = `<html><body>
<article>KEEP</article>
<div class="ad">REMOVE_AD</div>
<aside id="promo">REMOVE_PROMO</aside>
</body></html>`

	t.Run("正常系: 複数セレクタにマッチする要素を除去する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeSelectors: []string{".ad", "#promo"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		html, err := out.DOM.(*goquery.Document).Html()
		require.NoError(t, err)
		assert.Contains(t, html, "KEEP")
		assert.NotContains(t, html, "REMOVE_AD")
		assert.NotContains(t, html, "REMOVE_PROMO")
		assert.Contains(t, out.Text, "KEEP")
		assert.NotContains(t, out.Text, "REMOVE_AD")
	})

	t.Run("正常系: 空リストなら DOM を変更しない", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeSelectors: []string{}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.Contains(t, out.Text, "REMOVE_AD")
		assert.Contains(t, out.Text, "REMOVE_PROMO")
	})

	t.Run("正常系: 空文字エントリはスキップし他セレクタは適用する", func(t *testing.T) {
		f := &filter{}
		require.NoError(t, f.Init(context.Background(), contentHost{
			content: model.ContentConfig{ExcludeSelectors: []string{"", "  ", ".ad"}},
		}))
		c := htmlContent(t, src)

		out, err := f.Filter(context.Background(), c)

		require.NoError(t, err)
		assert.NotContains(t, out.Text, "REMOVE_AD")
		assert.Contains(t, out.Text, "REMOVE_PROMO")
	})
}
