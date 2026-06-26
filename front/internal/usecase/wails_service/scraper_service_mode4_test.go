package wails_service

import (
	"testing"

	"meguri-app/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMode4VisitOrder は mode 4 の訪問リスト構築を検証する。
func TestMode4VisitOrder(t *testing.T) {
	t.Run("正常系: 指定 ID のみ入力順で返し未知 ID は除外する", func(t *testing.T) {
		nodeByID := map[string]model.GraphNodeDTO{
			"n1": {ID: "n1", URLNormalized: "https://example.com/a"},
			"n2": {ID: "n2", URLNormalized: "https://example.com/b"},
			"n3": {ID: "n3", URLNormalized: "https://example.com/c"},
		}

		visit := filterExistingNodeIDs([]string{"n3", "missing", "n1", "n2"}, nodeByID)

		assert.Equal(t, []string{"n3", "n1", "n2"}, visit)
	})

	t.Run("正常系: nodeIds 空なら runMode4 はエラー", func(t *testing.T) {
		s := &ScraperService{}
		st := newCrawlState(model.StartCrawlRequest{
			Workspace: model.WorkspaceDTO{},
		})

		err := s.runMode4(t.Context(), model.StartCrawlRequest{}, st, nil, new(int), new(int), new(int), new(int))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "nodeIds")
	})
}
