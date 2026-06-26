package core

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// Pipeline は 1 URL 分のスクレイピング処理を実行する。
type Pipeline struct {
	// kernel はプラグインと設定へのアクセス。
	kernel *Kernel
}

// NewPipeline はカーネルを受け取りパイプラインを構築する。
func NewPipeline(k *Kernel) *Pipeline {
	return &Pipeline{kernel: k}
}

// PipelineOutput はパイプライン出力。リンクは P8 抽出結果。
type PipelineOutput struct {
	// Result は P6 まで完了した最終結果。
	Result *model.Result
	// Links は P8 で抽出した追跡候補 URL。
	Links []*url.URL
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

	// P3: Fetcher による URL 取得（リトライ・タイムアウトはプラグイン実装）。
	res, err := p.kernel.Fetcher().Get(ctx, req.URL, req.Headers)
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
		slog.Warn("link extractor failed", "url", req.URL.String(), "err", lerr.Error())
		links = nil
	}
	if result.Links == nil {
		result.Links = links
	}

	return &PipelineOutput{Result: result, Links: links}, nil
}

// selectParser は Content-Type と URL から適用する Parser を選ぶ。
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
