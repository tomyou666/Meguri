package core_test

import (
	"context"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/pkg/logger"

	// プラグイン副作用 import: 実装プラグインをレジストリへ登録する
	_ "meguri/plugins/fetcher-chromium"
	_ "meguri/plugins/fetcher-http"
	_ "meguri/plugins/filter-maincontent"
	_ "meguri/plugins/filter-selector"
	_ "meguri/plugins/linkextractor-default"
	_ "meguri/plugins/parser-html"
	_ "meguri/plugins/parser-pdf"
	_ "meguri/plugins/preprocessor-header"
	_ "meguri/plugins/transformer-html"
	_ "meguri/plugins/transformer-markdown"
	_ "meguri/plugins/transformer-raw-html"
)

// setupKernel はテスト用に初期化済みカーネルを組み立てる共通関数。
func setupKernel(t *testing.T, cfg *model.Config) *core.Kernel {
	t.Helper()
	logger.Init(io.Discard, slog.LevelInfo)
	host := core.NewHost(cfg)
	k := core.NewKernel(cfg, host, core.Default())
	if err := k.Init(context.Background()); err != nil {
		t.Fatalf("kernel init: %v", err)
	}
	t.Cleanup(func() {
		_ = k.Close(context.Background())
	})
	return k
}

func baseConfig() *model.Config {
	c := model.Default()
	c.Targets = []string{"http://placeholder/"}
	return &c
}

// TestPipeline_SingleURL は単一 URL に対するパイプライン処理を検証する。
func TestPipeline_SingleURL(t *testing.T) {
	srv := newTestWebServer(t)
	defer srv.Close()

	t.Run("正常系: HTMLページからMarkdownが生成されメタデータも抽出される", func(t *testing.T) {
		cfg := baseConfig()
		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/")
		req := model.NewRequest(u, 0)

		out, err := p.Run(context.Background(), req)

		assert.NoError(t, err)
		assert.NotNil(t, out)
		assert.NotNil(t, out.Result)
		assert.Equal(t, "シンプルページ", out.Result.Metadata["title"])
		assert.Contains(t, out.Result.Markdown, "本文タイトル", "main内のh1がMarkdownに残るはず")
		assert.NotContains(t, out.Result.Markdown, "サイトヘッダ（除去対象）",
			"maincontentフィルタによりヘッダは除去されている")
		assert.NotContains(t, out.Result.Markdown, "alert", "scriptタグは除去されている")
	})

	t.Run("正常系: P8 LinkExtractor は同一サイトの相対リンクと外部リンクを抽出する", func(t *testing.T) {
		cfg := baseConfig()
		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/links_with_pdf.html")
		req := model.NewRequest(u, 0)

		out, err := p.Run(context.Background(), req)

		assert.NoError(t, err)

		urls := toURLStrings(out.Links)
		assert.Contains(t, urls, srv.URL+"/docs/page-a.html")
		assert.Contains(t, urls, srv.URL+"/docs/page-b.html")
		assert.Contains(t, urls, srv.URL+"/files/report.pdf")
		assert.Contains(t, urls, "https://external.example.com/x")

		for _, u := range urls {
			assert.False(t, strings.HasPrefix(u, "javascript"), "javascript: は弾かれる")
			assert.False(t, strings.HasSuffix(u, "#section"), "#fragment は除去される")
		}
	})

	t.Run("正常系: PDF リンク (.pdf) を直接たどると PDF パーサーへ振り分けられる", func(t *testing.T) {
		cfg := baseConfig()
		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/files/report.pdf")
		req := model.NewRequest(u, 0)

		out, err := p.Run(context.Background(), req)

		assert.NoError(t, err, "PDFが有効なら処理が通る")
		assert.NotNil(t, out)
		assert.Contains(t, out.Result.Markdown, "FAKE-PDF-CONTENT",
			"PDFパーサーがバイナリからテキストを取り出している")
	})

	t.Run("異常系: pdf.enabled=false の場合 PDF リンクを開くと当該URLはエラー", func(t *testing.T) {
		cfg := baseConfig()
		cfg.PDF.Enabled = false
		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/files/report.pdf")
		req := model.NewRequest(u, 0)

		out, err := p.Run(context.Background(), req)

		assert.Error(t, err, "PDF無効時はエラー")
		assert.Nil(t, out)
		assert.Contains(t, err.Error(), "PDF")
	})

	t.Run("正常系: PreProcessor 'header' が User-Agent をリクエストに付与する", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Request.Headers = map[string]string{"User-Agent": "meguri-test/1.0"}
		cfg.Plugins.PreProcessors = []string{"header"}

		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/")
		req := model.NewRequest(u, 0)

		_, err := p.Run(context.Background(), req)

		assert.NoError(t, err)
		assert.Equal(t, "meguri-test/1.0", req.Headers["User-Agent"],
			"P2 PreProcessor 経由でヘッダが転写されること")
	})

	t.Run("正常系: filter-selector を有効化すると selector 範囲のみが残る", func(t *testing.T) {
		cfg := baseConfig()
		cfg.Content.Selector = "article.target"
		cfg.Plugins.Filters = []string{"selector"}

		k := setupKernel(t, cfg)
		p := core.NewPipeline(k)

		u, _ := url.Parse(srv.URL + "/selector_target.html")
		req := model.NewRequest(u, 0)

		out, err := p.Run(context.Background(), req)

		assert.NoError(t, err)
		assert.Contains(t, out.Result.Markdown, "残るべき", "selectorで残した範囲の本文が出力されている")
		// MarkdownはアンダースコアをエスケープするのでHTMLを検証
		assert.NotContains(t, out.Result.HTML, "NOT_TARGET_CONTENT", "範囲外は除外される")
		assert.NotContains(t, out.Result.HTML, "HEADER_TEXT")
		assert.NotContains(t, out.Result.HTML, "FOOTER_TEXT")
	})
}

func toURLStrings(us []*url.URL) []string {
	out := make([]string, 0, len(us))
	for _, u := range us {
		out = append(out, u.String())
	}
	return out
}
