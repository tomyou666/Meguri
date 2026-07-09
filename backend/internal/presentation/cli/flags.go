package cli

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"meguri/internal/domain/model"
)

// Flags は CLI 引数のパース結果を保持する。
type Flags struct {
	// ConfigPath は --config で指定した YAML パス。
	ConfigPath string

	// Targets は --url または位置引数で指定した URL 一覧。
	Targets []string
	// Headers は --header で指定したリクエストヘッダ。
	Headers map[string]string

	// Timeout は --timeout（0 なら YAML/デフォルトを維持）。
	Timeout time.Duration
	// RetryCount は --retry（負なら未指定）。
	RetryCount int
	// RetryInterval は --retry-interval。
	RetryInterval time.Duration

	// Formats は --format で指定した出力フォーマット。
	Formats []model.OutputFormat
	// OnlyMainContent は --only-main の 3 値フラグ。
	OnlyMainContent boolFlag
	// IncludeTags は --include-tag の繰り返し指定。
	IncludeTags stringSlice
	// ExcludeTags は --exclude-tag の繰り返し指定。
	ExcludeTags stringSlice
	// Selector は --selector。
	Selector string
	// ExtractLinks は --extract-links。
	ExtractLinks boolFlag
	// ExtractMetadata は --extract-metadata。
	ExtractMetadata boolFlag

	// PDF は --pdf。
	PDF boolFlag
	// PDFMode は --pdf-mode（fast/auto/ocr）。
	PDFMode string
	// PDFMaxPages は --pdf-max-pages。
	PDFMaxPages int
	// PDFOutput は --pdf-output（text/markdown/raw）。
	PDFOutput string

	// Crawl は --crawl。
	Crawl boolFlag
	// MaxDepth は --max-depth。
	MaxDepth int
	// MaxPages は --max-pages。
	MaxPages int
	// IncludePaths は --include-path。
	IncludePaths stringSlice
	// ExcludePaths は --exclude-path。
	ExcludePaths stringSlice
	// ExcludeURLs は --exclude-url。
	ExcludeURLs stringSlice
	// AllowExternal は --allow-external。
	AllowExternal boolFlag
	// AllowSubdomains は --allow-subdomains。
	AllowSubdomains boolFlag
	// RequestDelay は --delay。
	RequestDelay time.Duration
	// MaxConcurrency は --concurrency。
	MaxConcurrency int
	// FetchHTTPInflight は --fetch-http-inflight（HTTP 同時取得。キャリブレーション対象外）。
	FetchHTTPInflight int
	// FetchChromiumInflight は --fetch-chromium-inflight（Chromium 静的上限・フォールバック値）。
	FetchChromiumInflight int
	// FetchAutoCalibrate は --fetch-auto-calibrate（true でジョブ開始時に上限を 1〜8 で再計算）。
	FetchAutoCalibrate boolFlag
	// FetchDynamicChromium は --fetch-dynamic-chromium（true で実行中にメモリ使用率で ±1）。
	FetchDynamicChromium boolFlag
	// RespectRobotsTxt は --respect-robots。
	RespectRobotsTxt boolFlag

	// PreProcessors は --preprocessor。
	PreProcessors stringSlice
	// Parsers は --parser。
	Parsers stringSlice
	// Transformer は --transformer。
	Transformer string
	// Filters は --filter。
	Filters stringSlice
	// LinkExtractor は --link-extractor。
	LinkExtractor string

	// Fetcher は --fetcher（http / chromium）。
	Fetcher string
	// FetcherBrowserPath は --fetcher-browser-path。
	FetcherBrowserPath string

	// OutputDir は --output-dir。
	OutputDir string
	// OutputPattern は --output-pattern。
	OutputPattern string

	// Stdout は --stdout（単一 URL 時に標準出力へ出す）。
	Stdout bool
	// ProgressJSON は --progress-json（クロール進捗を stderr に NDJSON 出力）。
	ProgressJSON bool
}

// boolFlag は flag.Bool では区別できない「指定されたか」を扱う 3 値型。
type boolFlag struct {
	// set は CLI で明示指定されたか。
	set bool
	// v は真偽値。
	v bool
}

// String は flag.Value の文字列表現。
func (b *boolFlag) String() string { return fmt.Sprintf("%t", b.v) }

// Set は flag.Value のパース（true/false/1/0 等）。
func (b *boolFlag) Set(s string) error {
	switch strings.ToLower(s) {
	case "true", "1", "yes", "":
		b.v = true
	case "false", "0", "no":
		b.v = false
	default:
		return fmt.Errorf("invalid bool: %s", s)
	}
	b.set = true
	return nil
}

// IsBoolFlag は flag パッケージ向けのブールフラグ識別。
func (b *boolFlag) IsBoolFlag() bool { return true }

type stringSlice struct {
	// set は 1 回以上 Set されたか。
	set bool
	// values は蓄積した文字列。
	values []string
}

// String は flag.Value の文字列表現。
func (s *stringSlice) String() string { return strings.Join(s.values, ",") }

// Set は繰り返し指定を values に追加する。
func (s *stringSlice) Set(v string) error {
	s.values = append(s.values, v)
	s.set = true
	return nil
}

type headerFlag struct {
	// out は KEY=VAL をパースしたヘッダマップ。
	out map[string]string
}

// String は flag.Value（常に空）。
func (h *headerFlag) String() string { return "" }

// Set は KEY=VAL 形式の 1 ヘッダを out に追加する。
func (h *headerFlag) Set(v string) error {
	idx := strings.Index(v, "=")
	if idx < 0 {
		return fmt.Errorf("--header は KEY=VAL 形式で指定してください: %q", v)
	}
	if h.out == nil {
		h.out = map[string]string{}
	}
	h.out[v[:idx]] = v[idx+1:]
	return nil
}

type formatFlag struct {
	// out は --format の蓄積先スライスへのポインタ。
	out *[]model.OutputFormat
}

// String は flag.Value（常に空）。
func (f *formatFlag) String() string { return "" }

// Set は 1 つの OutputFormat を out に追加する。
func (f *formatFlag) Set(v string) error {
	*f.out = append(*f.out, model.OutputFormat(v))
	return nil
}

// ParseArgs はサブコマンド引数群をパースする。
func ParseArgs(args []string) (*Flags, error) {
	fs := flag.NewFlagSet("meguri-cli", flag.ContinueOnError)
	f := &Flags{Headers: map[string]string{}}

	fs.StringVar(&f.ConfigPath, "config", "", "設定ファイルパス (YAML)")
	fs.Var(&headerFlag{out: f.Headers}, "header", "リクエストヘッダ KEY=VAL (繰り返し可)")

	fs.DurationVar(&f.Timeout, "timeout", 0, "リクエストタイムアウト")
	fs.IntVar(&f.RetryCount, "retry", -1, "リトライ回数 (0以上)")
	fs.DurationVar(&f.RetryInterval, "retry-interval", 0, "リトライ間隔")

	fs.Var(&formatFlag{out: &f.Formats}, "format", "出力フォーマット (繰り返し可)")
	fs.Var(&f.OnlyMainContent, "only-main", "メインコンテンツのみ抽出")
	fs.Var(&f.IncludeTags, "include-tag", "include する HTML タグ (繰り返し可)")
	fs.Var(&f.ExcludeTags, "exclude-tag", "exclude する HTML タグ (繰り返し可)")
	fs.StringVar(&f.Selector, "selector", "", "CSSセレクタで本文を絞り込む")
	fs.Var(&f.ExtractLinks, "extract-links", "リンク抽出を有効化")
	fs.Var(&f.ExtractMetadata, "extract-metadata", "メタデータ抽出を有効化")

	fs.Var(&f.PDF, "pdf", "PDFを有効化")
	fs.StringVar(&f.PDFMode, "pdf-mode", "", "PDF解析モード fast/auto/ocr")
	fs.IntVar(&f.PDFMaxPages, "pdf-max-pages", -1, "PDF最大ページ数 (0=無制限)")
	fs.StringVar(&f.PDFOutput, "pdf-output", "", "PDF出力形式 text/markdown/raw")

	fs.Var(&f.Crawl, "crawl", "クロールを有効化")
	fs.IntVar(&f.MaxDepth, "max-depth", -1, "クロール最大深度")
	fs.IntVar(&f.MaxPages, "max-pages", -1, "クロール最大ページ数")
	fs.Var(&f.IncludePaths, "include-path", "クロール許可パス正規表現 (繰り返し可)")
	fs.Var(&f.ExcludePaths, "exclude-path", "クロール除外パス正規表現 (繰り返し可)")
	fs.Var(&f.ExcludeURLs, "exclude-url", "完全一致でスキップする URL (繰り返し可)")
	fs.Var(&f.AllowExternal, "allow-external", "外部リンクの追跡を許可")
	fs.Var(&f.AllowSubdomains, "allow-subdomains", "サブドメインの追跡を許可")
	fs.DurationVar(&f.RequestDelay, "delay", -1, "リクエスト間遅延")
	fs.IntVar(&f.MaxConcurrency, "concurrency", -1, "並行ワーカー数")
	fs.IntVar(&f.FetchHTTPInflight, "fetch-http-inflight", -1, "HTTP同時取得上限 (1-64)。キャリブレーション対象外")
	fs.IntVar(&f.FetchChromiumInflight, "fetch-chromium-inflight", -1, "Chromium静的上限 (1-8)。auto_calibrate/dynamicの基準値")
	fs.Var(&f.FetchAutoCalibrate, "fetch-auto-calibrate", "true=ジョブ開始時にメモリ計測でChromium上限を1-8に再計算")
	fs.Var(&f.FetchDynamicChromium, "fetch-dynamic-chromium", "true=実行中にメモリ使用率でChromium上限を±1調整")
	fs.Var(&f.RespectRobotsTxt, "respect-robots", "robots.txtを尊重")

	fs.Var(&f.PreProcessors, "preprocessor", "PreProcessor プラグイン名 (繰り返し可)")
	fs.Var(&f.Parsers, "parser", "Parser プラグイン名 (繰り返し可)")
	fs.StringVar(&f.Transformer, "transformer", "", "Transformer プラグイン名")
	fs.Var(&f.Filters, "filter", "Filter プラグイン名 (繰り返し可)")
	fs.StringVar(&f.LinkExtractor, "link-extractor", "", "LinkExtractor プラグイン名")

	fs.StringVar(&f.Fetcher, "fetcher", "", "URLフェッチ実装 (http / chromium)")
	fs.StringVar(&f.FetcherBrowserPath, "fetcher-browser-path", "", "chromium フェッチ用ブラウザ実行ファイルパス")

	fs.StringVar(&f.OutputDir, "output-dir", "", "出力ディレクトリ")
	fs.StringVar(&f.OutputPattern, "output-pattern", "", "出力ファイル名パターン")

	fs.BoolVar(&f.Stdout, "stdout", false, "結果を標準出力に出す（単一URLモード）")
	fs.BoolVar(&f.ProgressJSON, "progress-json", false, "クロール進捗を stderr に NDJSON 出力")

	var url string
	fs.StringVar(&url, "url", "", "対象URL (1件指定)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if url != "" {
		f.Targets = append(f.Targets, url)
	}
	f.Targets = append(f.Targets, fs.Args()...)
	return f, nil
}

// Merge はデフォルト+YAML済みの Config に対して CLI フラグを上書き適用する。
func Merge(cfg *model.Config, f *Flags) {
	if len(f.Targets) > 0 {
		cfg.Targets = f.Targets
	}
	if len(f.Headers) > 0 {
		if cfg.Request.Headers == nil {
			cfg.Request.Headers = map[string]string{}
		}
		for k, v := range f.Headers {
			cfg.Request.Headers[k] = v
		}
	}

	if f.Timeout > 0 {
		cfg.Request.Timeout = f.Timeout
	}
	if f.RetryCount >= 0 {
		cfg.Request.RetryCount = f.RetryCount
	}
	if f.RetryInterval > 0 {
		cfg.Request.RetryInterval = f.RetryInterval
	}

	if len(f.Formats) > 0 {
		cfg.Content.Formats = f.Formats
	}
	if f.OnlyMainContent.set {
		cfg.Content.OnlyMainContent = f.OnlyMainContent.v
	}
	if f.IncludeTags.set {
		cfg.Content.IncludeTags = f.IncludeTags.values
	}
	if f.ExcludeTags.set {
		cfg.Content.ExcludeTags = f.ExcludeTags.values
	}
	if f.Selector != "" {
		cfg.Content.Selector = f.Selector
	}
	if f.ExtractLinks.set {
		cfg.Content.ExtractLinks = f.ExtractLinks.v
	}
	if f.ExtractMetadata.set {
		cfg.Content.ExtractMetadata = f.ExtractMetadata.v
	}

	if f.PDF.set {
		cfg.PDF.Enabled = f.PDF.v
	}
	if f.PDFMode != "" {
		cfg.PDF.Mode = model.PDFParseMode(f.PDFMode)
	}
	if f.PDFMaxPages >= 0 {
		cfg.PDF.MaxPages = f.PDFMaxPages
	}
	if f.PDFOutput != "" {
		cfg.PDF.Output = model.PDFOutput(f.PDFOutput)
	}

	if f.Crawl.set {
		cfg.Crawl.Enabled = f.Crawl.v
	}
	if f.MaxDepth >= 0 {
		cfg.Crawl.MaxDepth = f.MaxDepth
	}
	if f.MaxPages > 0 {
		cfg.Crawl.MaxPages = f.MaxPages
	}
	if f.IncludePaths.set {
		cfg.Crawl.IncludePaths = f.IncludePaths.values
	}
	if f.ExcludePaths.set {
		cfg.Crawl.ExcludePaths = f.ExcludePaths.values
	}
	if f.ExcludeURLs.set {
		cfg.Crawl.ExcludeURLs = append(cfg.Crawl.ExcludeURLs, f.ExcludeURLs.values...)
	}
	if f.AllowExternal.set {
		cfg.Crawl.AllowExternal = f.AllowExternal.v
	}
	if f.AllowSubdomains.set {
		cfg.Crawl.AllowSubdomains = f.AllowSubdomains.v
	}
	if f.RequestDelay >= 0 {
		cfg.Crawl.RequestDelay = f.RequestDelay
	}
	if f.MaxConcurrency > 0 {
		cfg.Crawl.MaxConcurrency = f.MaxConcurrency
	}
	if f.FetchHTTPInflight > 0 {
		cfg.Crawl.FetchLimits.HTTPMaxInflight = f.FetchHTTPInflight
	}
	if f.FetchChromiumInflight > 0 {
		cfg.Crawl.FetchLimits.ChromiumMaxInflight = f.FetchChromiumInflight
	}
	if f.FetchAutoCalibrate.set {
		cfg.Crawl.FetchLimits.AutoCalibrate = f.FetchAutoCalibrate.v
	}
	if f.FetchDynamicChromium.set {
		cfg.Crawl.FetchLimits.DynamicChromium = f.FetchDynamicChromium.v
	}
	if f.RespectRobotsTxt.set {
		cfg.Crawl.RespectRobotsTxt = f.RespectRobotsTxt.v
	}

	if f.PreProcessors.set {
		cfg.Plugins.PreProcessors = f.PreProcessors.values
	}
	if f.Parsers.set {
		cfg.Plugins.Parsers = f.Parsers.values
	}
	if f.Transformer != "" {
		cfg.Plugins.Transformer = f.Transformer
	}
	if f.Filters.set {
		cfg.Plugins.Filters = f.Filters.values
	}
	if f.LinkExtractor != "" {
		cfg.Plugins.LinkExtractor = f.LinkExtractor
	}

	if f.Fetcher != "" {
		cfg.Plugins.Fetcher = model.FetcherKind(f.Fetcher)
	}
	if f.FetcherBrowserPath != "" {
		cfg.Plugins.FetcherConfig.BrowserPath = f.FetcherBrowserPath
	}

	if f.OutputDir != "" {
		cfg.Output.Dir = f.OutputDir
	}
	if f.OutputPattern != "" {
		cfg.Output.FilePattern = f.OutputPattern
	}
}
