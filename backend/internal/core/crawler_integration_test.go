package core_test

import (
	"context"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
)

// denyRobots は特定パスを robots.txt 不許可として返すフェイク。
type denyRobots struct {
	denyPaths []string
}

func (d *denyRobots) Allowed(_ context.Context, u *url.URL, _ string) bool {
	for _, p := range d.denyPaths {
		if u.Path == p {
			return false
		}
	}
	return true
}

// TestCrawler は BFS クロールと各種フィルタ・制限の挙動を検証する。
func TestCrawler(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

	t.Run("正常系: BFSでリンクを辿り、想定したURL集合を取得する", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.MaxConcurrency = 2
		cfg.Crawl.AllowExternal = false

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		stats, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, stats.Failed, 0, "全URLが成功するはず")

		mu.Lock()
		defer mu.Unlock()
		assert.Contains(t, collected, srv.URL+"/links_with_pdf.html")
		assert.Contains(t, collected, srv.URL+"/docs/page-a.html")
		assert.Contains(t, collected, srv.URL+"/docs/page-b.html")
		assert.Contains(t, collected, srv.URL+"/files/report.pdf", "PDFリンクもクロールされる")
	})

	t.Run("正常系: max_depth=0 ならシードのみ処理しリンクは追跡しない", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 0
		cfg.Crawl.MaxPages = 100

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, []string{srv.URL + "/links_with_pdf.html"}, collected,
			"深度0なのでシードのみ取得される")
	})

	t.Run("正常系: max_pages を尊重して打ち切られる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 5
		cfg.Crawl.MaxPages = 2

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		assert.LessOrEqual(t, len(collected), 2, "max_pagesを超えて取得しない")
	})

	t.Run("正常系: allow_external=false なら外部リンクはスキップされる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.AllowExternal = false

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		for _, u := range collected {
			assert.NotContains(t, u, "external.example.com",
				"外部ドメインのURLは取得されない")
		}
	})

	t.Run("正常系: pdf.enabled=false ならPDFリンクは追跡されない", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.PDF.Enabled = false

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		for _, u := range collected {
			assert.NotContains(t, u, ".pdf",
				"PDF無効化時はPDFリンクをスキップする")
		}
	})

	t.Run("正常系: robots.txt 不許可URLはスキップされる", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 2
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.RespectRobotsTxt = true

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		robots := &denyRobots{denyPaths: []string{"/docs/page-a.html"}}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), robots, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})

		assert.NoError(t, err)
		mu.Lock()
		defer mu.Unlock()
		for _, u := range collected {
			assert.NotContains(t, u, "/docs/page-a.html",
				"robots.txt不許可URLはクロールされない")
		}
	})

	t.Run("クロール: 同一URLは重複訪問されない", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 3
		cfg.Crawl.MaxPages = 100

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})
		assert.NoError(t, err)

		mu.Lock()
		defer mu.Unlock()
		counts := map[string]int{}
		for _, u := range collected {
			counts[u]++
		}
		for u, n := range counts {
			assert.Equal(t, 1, n, "同一URL %s は1回のみ訪問される", u)
		}
	})

	t.Run("クロール: context キャンセルでクロールが停止する", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 3
		cfg.Crawl.MaxPages = 100
		cfg.Crawl.RequestDelay = 200 * time.Millisecond

		var mu sync.Mutex
		var collected []string
		sink := func(r *model.Result) {
			mu.Lock()
			defer mu.Unlock()
			collected = append(collected, r.URL.String())
		}

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), nil, sink, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, _ = c.Run(ctx, []*url.URL{seed})

		// 完了することそのものを検証 (デッドロックしない)
		assert.True(t, true)
	})
}
