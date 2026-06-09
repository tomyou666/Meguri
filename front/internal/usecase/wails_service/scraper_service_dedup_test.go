package wails_service

import (
	"testing"

	"scraperbot-front/internal/model"

	"github.com/stretchr/testify/assert"
)

// TestCrawlState はクロール中のリンク重複判定とスキップ対象 URL の収集を検証する。
func TestCrawlState(t *testing.T) {
	t.Run("正常系: 既存ノード・同一実行内の重複を skip reason で区別する", func(t *testing.T) {
		st := newCrawlState(model.StartCrawlRequest{
			Workspace: model.WorkspaceDTO{
				Nodes: []model.GraphNodeDTO{
					{ID: "n1", URLNormalized: "https://example.com/existing"},
				},
			},
		})

		assert.Equal(t, "duplicate_existing", st.linkSkipReason("https://example.com/existing"))

		st.markMaterialized("https://example.com/new")
		assert.Equal(t, "duplicate_in_run", st.linkSkipReason("https://example.com/new"))
		assert.Equal(t, "", st.linkSkipReason("https://example.com/unknown"))
	})

	t.Run("正常系: rescrapeExisting=false なら success 済み URL を skip_scrape に含める", func(t *testing.T) {
		st := newCrawlState(model.StartCrawlRequest{
			RescrapeExisting: false,
			Workspace: model.WorkspaceDTO{
				Nodes: []model.GraphNodeDTO{
					{ID: "n1", URLNormalized: "https://example.com/a", Status: "success"},
					{ID: "n2", URLNormalized: "https://example.com/b", Status: "idle"},
				},
			},
		})
		urls := st.skipScrapeURLs()
		assert.ElementsMatch(t, []string{"https://example.com/a"}, urls)

		st.rescrapeExisting = true
		assert.Nil(t, st.skipScrapeURLs())
	})
}
