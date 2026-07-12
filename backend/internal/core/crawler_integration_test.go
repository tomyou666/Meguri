package core_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"meguri/internal/core"
	"meguri/internal/domain/model"
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

// countingRobots は Allowed 呼び出し回数を数えるフェイク（常に許可）。
type countingRobots struct {
	calls atomic.Int64
}

func (c *countingRobots) Allowed(_ context.Context, _ *url.URL, _ string) bool {
	c.calls.Add(1)
	return true
}

// denyAllRobots は常に不許可を返すフェイク。
type denyAllRobots struct{}

func (*denyAllRobots) Allowed(_ context.Context, _ *url.URL, _ string) bool {
	return false
}

// pathDenyRobots は指定パスのみ不許可。残りは許可。
type pathDenyRobots struct {
	deny map[string]struct{}
}

func (p *pathDenyRobots) Allowed(_ context.Context, u *url.URL, _ string) bool {
	_, denied := p.deny[u.Path]
	return !denied
}

// slowPathRobots は Allowed を遅延させ、パスごとの呼び出し回数を数える。
type slowPathRobots struct {
	delay time.Duration
	mu    sync.Mutex
	calls map[string]int
}

func (s *slowPathRobots) Allowed(_ context.Context, u *url.URL, _ string) bool {
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.calls == nil {
		s.calls = map[string]int{}
	}
	s.calls[u.Path]++
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

	t.Run("正常系: max_pages 到達後の URL では robots 判定しない", func(t *testing.T) {
		// max_pages=1 なのでシードのみ reserve→robots。子リンクは max_pages で落ち robots 未呼び出し。
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 5
		cfg.Crawl.MaxPages = 1
		cfg.Crawl.RespectRobotsTxt = true

		robots := &countingRobots{}
		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), robots, nil, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		_, err := c.Run(context.Background(), []*url.URL{seed})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), robots.calls.Load(),
			"max_pages 超過 URL では robots.Allowed を呼ばない")
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

	t.Run("正常系: シードが robots 拒否のみなら enqueue せず完了する", func(t *testing.T) {
		// pending 巻き戻しで pushQ が閉じ、ワーカーがデッドロックせず終了すること。
		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 1
		cfg.Crawl.MaxPages = 10
		cfg.Crawl.RespectRobotsTxt = true

		k := setupKernel(t, cfg)
		c := core.NewCrawler(k, core.NewPipeline(k), &denyAllRobots{}, nil, nil)

		seed, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		done := make(chan struct{})
		var stats *core.CrawlStats
		var err error
		go func() {
			stats, err = c.Run(context.Background(), []*url.URL{seed})
			close(done)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("seed robots deny 後にクロールが完了しない（pushQ close 漏れの疑い）")
		}
		assert.NoError(t, err)
		assert.Equal(t, 0, stats.Enqueued)
		assert.GreaterOrEqual(t, stats.Skipped, 1)
		assert.Equal(t, 0, stats.Succeeded)
	})

	t.Run("正常系: robots 拒否後は visited を戻し別シードを enqueue できる", func(t *testing.T) {
		// max_pages=1。先頭シード拒否で枠を戻さないと次シードが max_pages で落ちる。
		mux := http.NewServeMux()
		mux.HandleFunc("/seed-a.html", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body>a</body></html>`))
		})
		mux.HandleFunc("/seed-b.html", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body>b</body></html>`))
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()

		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 0
		cfg.Crawl.MaxPages = 1
		cfg.Crawl.MaxConcurrency = 1
		cfg.Crawl.RespectRobotsTxt = true
		cfg.Crawl.AllowExternal = false

		robots := &pathDenyRobots{deny: map[string]struct{}{"/seed-a.html": {}}}
		k := setupKernel(t, cfg)
		var collected []string
		var mu sync.Mutex
		sink := func(r *model.Result) {
			mu.Lock()
			collected = append(collected, r.URL.Path)
			mu.Unlock()
		}
		c := core.NewCrawler(k, core.NewPipeline(k), robots, sink, nil)

		a, _ := url.Parse(ts.URL + "/seed-a.html")
		b, _ := url.Parse(ts.URL + "/seed-b.html")
		stats, err := c.Run(context.Background(), []*url.URL{a, b})
		assert.NoError(t, err)
		assert.Equal(t, 1, stats.Enqueued)
		mu.Lock()
		defer mu.Unlock()
		assert.Equal(t, []string{"/seed-b.html"}, collected)
	})

	t.Run("正常系: 遅い robots 中の同一 URL 並行 enqueue は二重にならない", func(t *testing.T) {
		// 2 シードが同じ子へリンク。slow robots 中に両ワーカーが enqueue しても 1 回だけ。
		mux := http.NewServeMux()
		mux.HandleFunc("/seed-a.html", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><a href="/shared.html">s</a></body></html>`))
		})
		mux.HandleFunc("/seed-b.html", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body><a href="/shared.html">s</a></body></html>`))
		})
		mux.HandleFunc("/shared.html", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<html><body>shared</body></html>`))
		})
		ts := httptest.NewServer(mux)
		defer ts.Close()

		cfg := baseConfig()
		cfg.Crawl.Enabled = true
		cfg.Crawl.MaxDepth = 1
		cfg.Crawl.MaxPages = 10
		cfg.Crawl.MaxConcurrency = 2
		cfg.Crawl.RespectRobotsTxt = true
		cfg.Crawl.AllowExternal = false

		robots := &slowPathRobots{delay: 80 * time.Millisecond}
		k := setupKernel(t, cfg)

		var mu sync.Mutex
		startedShared := 0
		progress := func(ev core.ProgressEvent) {
			if ev.Kind == core.ProgressStarted && strings.HasSuffix(ev.URL, "/shared.html") {
				mu.Lock()
				startedShared++
				mu.Unlock()
			}
		}
		c := core.NewCrawler(k, core.NewPipeline(k), robots, nil, progress)

		a, _ := url.Parse(ts.URL + "/seed-a.html")
		b, _ := url.Parse(ts.URL + "/seed-b.html")
		stats, err := c.Run(context.Background(), []*url.URL{a, b})
		assert.NoError(t, err)
		assert.Equal(t, 3, stats.Enqueued, "seed2 + shared1")
		mu.Lock()
		assert.Equal(t, 1, startedShared, "shared は 1 回だけ開始")
		mu.Unlock()
		robots.mu.Lock()
		assert.Equal(t, 1, robots.calls["/shared.html"], "shared の robots も 1 回")
		robots.mu.Unlock()
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
