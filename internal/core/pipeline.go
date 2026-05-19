package core

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

// Fetcher は HTTP 取得のみを抽象化したインタフェース。
// テスト時に差し替えられるよう、Pipeline は具体実装ではなくこれに依存する。
type Fetcher interface {
	Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error)
}

// Pipeline は 1 URL 分のスクレイピング処理を実行する。
type Pipeline struct {
	kernel  *Kernel
	fetcher Fetcher
}

// NewPipeline はカーネルとフェッチャを受け取りパイプラインを構築する。
func NewPipeline(k *Kernel, f Fetcher) *Pipeline {
	return &Pipeline{kernel: k, fetcher: f}
}

// PipelineOutput はパイプライン出力。リンクは P8 抽出結果。
type PipelineOutput struct {
	Result *model.Result
	Links  []*url.URL
}

// Run は単一 URL のパイプラインを実行する。
// 設計書 05 の流れ: P2 → HTTP → コンテンツ種別検出 → P5 → P7 → P6 → P8。
func (p *Pipeline) Run(ctx context.Context, req *model.Request) (*PipelineOutput, error) {
	// P2: PreProcessor チェーン
	for _, pp := range p.kernel.PreProcessors() {
		if err := pp.PreProcess(ctx, req); err != nil {
			return nil, fmt.Errorf("preprocess %s: %w", pp.Metadata().Name, err)
		}
	}

	// HTTP 取得（リトライ・タイムアウトはフェッチャ実装）。
	res, err := p.fetcher.Get(ctx, req.URL, req.Headers)
	if err != nil {
		return nil, err
	}

	// コンテンツ種別検出 → P5 Parser 選択。
	parser, err := p.selectParser(res)
	if err != nil {
		return nil, err
	}

	content, err := parser.Parse(ctx, res)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", parser.Metadata().Name, err)
	}

	// P7: Filter チェーン
	for _, f := range p.kernel.Filters() {
		content, err = f.Filter(ctx, content)
		if err != nil {
			return nil, fmt.Errorf("filter %s: %w", f.Metadata().Name, err)
		}
	}

	// P6: Transformer
	result, err := p.kernel.Transformer().Transform(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("transform: %w", err)
	}

	// P8: LinkExtractor（失敗してもページは成功扱い）
	links, lerr := p.kernel.LinkExtractor().Extract(ctx, content, req.URL)
	if lerr != nil {
		p.kernel.Host().Logger().Warn("link extractor failed", "url", req.URL.String(), "err", lerr.Error())
		links = nil
	}
	if result.Links == nil {
		result.Links = links
	}

	return &PipelineOutput{Result: result, Links: links}, nil
}

func (p *Pipeline) selectParser(res *model.Response) (plugin.Parser, error) {
	ct := strings.ToLower(res.ContentType)
	cfg := p.kernel.Config()

	isPDF := strings.Contains(ct, "application/pdf") ||
		strings.HasSuffix(strings.ToLower(res.URL.Path), ".pdf")
	if isPDF && !cfg.PDF.Enabled {
		return nil, errors.New("PDFは設定で無効化されています")
	}

	isHTML := strings.Contains(ct, "text/html") ||
		strings.Contains(ct, "application/xhtml+xml") ||
		strings.HasSuffix(strings.ToLower(res.URL.Path), ".html") ||
		strings.HasSuffix(strings.ToLower(res.URL.Path), ".htm")

	// 1) Content-Type / 拡張子で先に決まる場合は名前で優先選択する。
	preferred := ""
	switch {
	case isPDF:
		preferred = "pdf"
	case isHTML:
		preferred = "html"
	}

	if preferred != "" {
		for _, parser := range p.kernel.Parsers() {
			if parser.Metadata().Name == preferred && parser.CanParse(res) {
				return parser, nil
			}
		}
	}

	// 2) 登録されたパーサーを順に CanParse で問い合わせる。
	for _, parser := range p.kernel.Parsers() {
		if parser.CanParse(res) {
			return parser, nil
		}
	}

	return nil, fmt.Errorf("適用可能なパーサーがありません (content-type=%q)", res.ContentType)
}
