package wails_service

import (
	"testing"

	"scraperbot-front/internal/model"

	"github.com/stretchr/testify/assert"
)

func TestCrawlStateLinkSkipReason(t *testing.T) {
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
}

func TestCrawlStateSkipScrapeURLs(t *testing.T) {
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
}
