package runner

import (
	"context"
	"encoding/json"

	"meguri/internal/core"
	"meguri/internal/core/fetchlimit"
	"meguri/internal/domain/model"
	"meguri/internal/usecase"
)

// debugLogMap は gowrap debug ログ用に map 値を読める形へ整形する。
func debugLogMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = formatDebugValue(v)
	}
	return out
}

func formatDebugValue(v any) any {
	if v == nil {
		return nil
	}

	switch x := v.(type) {
	case context.Context:
		return formatDebugContext(x)
	case *model.Config:
		return formatDebugConfig(x)
	case ProgressSink:
		return formatDebugProgressSink(x)
	case *usecase.RunOptions:
		return formatDebugRunOptions(x)
	case *model.Result:
		return formatDebugResult(x)
	case *core.CrawlStats:
		return formatDebugCrawlStats(x)
	case usecase.RobotsTxtResult:
		return formatDebugRobotsTxtResult(x)
	case *fetchlimit.FetchLimiter:
		return formatDebugFetchLimiter(x)
	case json.RawMessage:
		return formatDebugRawMessage(x)
	case []json.RawMessage:
		return formatDebugRawMessageSlice(x)
	case error:
		if x == nil {
			return nil
		}
		return x.Error()
	default:
		return v
	}
}

type debugContextSummary struct {
	// Done はコンテキストがキャンセル済みか。
	Done bool `json:"done"`
	// Err はキャンセル理由（あれば）。
	Err string `json:"err,omitempty"`
	// HasDeadline は期限が設定されているか。
	HasDeadline bool `json:"has_deadline"`
}

func formatDebugContext(ctx context.Context) debugContextSummary {
	if ctx == nil {
		return debugContextSummary{}
	}
	s := debugContextSummary{}
	if err := ctx.Err(); err != nil {
		s.Done = true
		s.Err = err.Error()
	}
	if _, ok := ctx.Deadline(); ok {
		s.HasDeadline = true
	}
	return s
}

type debugConfigSummary struct {
	// Targets は処理対象 URL 一覧。
	Targets []string `json:"targets"`
	// Fetcher は使用フェッチャ種別。
	Fetcher string `json:"fetcher"`
	// CrawlEnabled はクロール有効か。
	CrawlEnabled bool `json:"crawl_enabled"`
	// CrawlMaxDepth は最大深度。
	CrawlMaxDepth int `json:"crawl_max_depth"`
	// CrawlMaxPages は最大ページ数。
	CrawlMaxPages int `json:"crawl_max_pages"`
	// CrawlMaxConcurrency は同時実行ワーカー数。
	CrawlMaxConcurrency int `json:"crawl_max_concurrency"`
	// RespectRobotsTxt は robots.txt 遵守か。
	RespectRobotsTxt bool `json:"respect_robots_txt"`
	// ContentFormats は出力フォーマット一覧。
	ContentFormats []string `json:"content_formats"`
	// OutputDir は出力ディレクトリ。
	OutputDir string `json:"output_dir"`
}

func formatDebugConfig(cfg *model.Config) any {
	if cfg == nil {
		return nil
	}
	formats := make([]string, 0, len(cfg.Content.Formats))
	for _, f := range cfg.Content.Formats {
		formats = append(formats, string(f))
	}
	fetcher := string(cfg.Plugins.Fetcher)
	if fetcher == "" {
		fetcher = string(model.FetcherHTTP)
	}
	return debugConfigSummary{
		Targets:             append([]string(nil), cfg.Targets...),
		Fetcher:             fetcher,
		CrawlEnabled:        cfg.Crawl.Enabled,
		CrawlMaxDepth:       cfg.Crawl.MaxDepth,
		CrawlMaxPages:       cfg.Crawl.MaxPages,
		CrawlMaxConcurrency: cfg.Crawl.MaxConcurrency,
		RespectRobotsTxt:    cfg.Crawl.RespectRobotsTxt,
		ContentFormats:      formats,
		OutputDir:           cfg.Output.Dir,
	}
}

func formatDebugProgressSink(sink ProgressSink) any {
	if sink == nil {
		return nil
	}
	return "<callback>"
}

type debugRunOptionsSummary struct {
	// HasPause は Pause が設定されているか。
	HasPause bool `json:"has_pause"`
	// HasCache は Cache が設定されているか。
	HasCache bool `json:"has_cache"`
	// HasFetchLimiter は FetchLimiter が設定されているか。
	HasFetchLimiter bool `json:"has_fetch_limiter"`
}

func formatDebugRunOptions(opts *usecase.RunOptions) any {
	if opts == nil {
		return nil
	}
	return debugRunOptionsSummary{
		HasPause:        opts.Pause != nil,
		HasCache:        opts.Cache != nil,
		HasFetchLimiter: opts.FetchLimiter != nil,
	}
}

type debugResultSummary struct {
	// URL は結果の対象 URL。
	URL string `json:"url,omitempty"`
	// MarkdownLen は Markdown 本文の文字数。
	MarkdownLen int `json:"markdown_len"`
	// HTMLLen は HTML 本文の文字数。
	HTMLLen int `json:"html_len"`
	// RawHTMLLen は生 HTML の文字数。
	RawHTMLLen int `json:"raw_html_len"`
	// LinksCount は抽出リンク数。
	LinksCount int `json:"links_count"`
	// MetadataKeys はメタデータのキー一覧。
	MetadataKeys []string `json:"metadata_keys"`
	// JSONKeys は JSON 出力のキー一覧。
	JSONKeys []string `json:"json_keys"`
}

func formatDebugResult(r *model.Result) any {
	if r == nil {
		return nil
	}
	s := debugResultSummary{
		MarkdownLen: len(r.Markdown),
		HTMLLen:     len(r.HTML),
		RawHTMLLen:  len(r.RawHTML),
		LinksCount:  len(r.Links),
	}
	if r.URL != nil {
		s.URL = r.URL.String()
	}
	if len(r.Metadata) > 0 {
		s.MetadataKeys = make([]string, 0, len(r.Metadata))
		for k := range r.Metadata {
			s.MetadataKeys = append(s.MetadataKeys, k)
		}
	}
	if len(r.JSON) > 0 {
		s.JSONKeys = make([]string, 0, len(r.JSON))
		for k := range r.JSON {
			s.JSONKeys = append(s.JSONKeys, k)
		}
	}
	return s
}

func formatDebugCrawlStats(stats *core.CrawlStats) any {
	if stats == nil {
		return nil
	}
	return *stats
}

type debugRobotsTxtSummary struct {
	// Host は対象ホスト名。
	Host string `json:"host"`
	// Status は取得状態。
	Status string `json:"status"`
	// StatusCode は HTTP ステータスコード。
	StatusCode int `json:"status_code"`
	// BodyLen は本文の文字数。
	BodyLen int `json:"body_len"`
	// Error はエラー詳細。
	Error string `json:"error,omitempty"`
}

func formatDebugRobotsTxtResult(r usecase.RobotsTxtResult) debugRobotsTxtSummary {
	return debugRobotsTxtSummary{
		Host:       r.Host,
		Status:     r.Status,
		StatusCode: r.StatusCode,
		BodyLen:    len(r.Body),
		Error:      r.Error,
	}
}

func formatDebugFetchLimiter(l *fetchlimit.FetchLimiter) any {
	if l == nil {
		return nil
	}
	return map[string]bool{"present": true}
}

type debugRawMessageSummary struct {
	// Len は JSON バイト長。
	Len int `json:"len"`
	// Preview は先頭プレビュー（最大 120 文字）。
	Preview string `json:"preview,omitempty"`
}

func formatDebugRawMessage(raw json.RawMessage) debugRawMessageSummary {
	s := debugRawMessageSummary{Len: len(raw)}
	if len(raw) == 0 {
		return s
	}
	const previewMax = 120
	preview := string(raw)
	if len(preview) > previewMax {
		preview = preview[:previewMax] + "..."
	}
	s.Preview = preview
	return s
}

func formatDebugRawMessageSlice(layers []json.RawMessage) any {
	if layers == nil {
		return nil
	}
	out := make([]debugRawMessageSummary, len(layers))
	for i, layer := range layers {
		out[i] = formatDebugRawMessage(layer)
	}
	return out
}
