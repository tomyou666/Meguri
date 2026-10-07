package usecase_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// TestCrawlPageConfigOverridesContent は PageConfig がある URL だけ content を切り替え、
// nil の URL はラン共通パイプラインのままであることを検証する。
func TestCrawlPageConfigOverridesContent(t *testing.T) {
	const body = `<!doctype html><html><body><main>
<article class="target"><p>NODE_BODY</p></article>
<p>OUTSIDE_BODY</p>
</main></body></html>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	cfg := model.Default()
	cfg.Crawl.Enabled = true
	cfg.Crawl.MaxDepth = 0
	cfg.Crawl.MaxPages = 10
	cfg.Crawl.MaxConcurrency = 2
	cfg.Crawl.RespectRobotsTxt = false

	opts := &usecase.RunOptions{
		PageConfig: func(rawURL string) (*model.Config, error) {
			if !strings.Contains(rawURL, "/node") {
				return nil, nil
			}
			page := cfg
			page.Content.Selector = "article.target"
			return &page, nil
		},
	}

	var mu sync.Mutex
	htmlByPath := map[string]string{}
	sink := func(res *model.Result) {
		if res == nil || res.URL == nil {
			return
		}
		mu.Lock()
		htmlByPath[res.URL.Path] = res.HTML
		mu.Unlock()
	}

	stats, err := usecase.NewCrawl(sink).RunWithConfig(
		context.Background(),
		&cfg,
		[]string{srv.URL + "/plain", srv.URL + "/node"},
		nil,
		opts,
	)
	require.NoError(t, err)
	require.NotNil(t, stats)
	assert.Equal(t, 0, stats.Failed)
	assert.Equal(t, 2, stats.Succeeded)

	mu.Lock()
	defer mu.Unlock()
	assert.Contains(t, htmlByPath["/plain"], "OUTSIDE_BODY")
	assert.Contains(t, htmlByPath["/node"], "NODE_BODY")
	assert.NotContains(t, htmlByPath["/node"], "OUTSIDE_BODY")
}
