package model

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
)

const (
	// DefaultHTTPMaxInflight は HTTP 取得の既定同時実行上限。
	DefaultHTTPMaxInflight = 16
	// DefaultChromiumMaxInflight は Chromium 取得の既定同時実行上限。
	DefaultChromiumMaxInflight = 2
	// MaxChromiumMaxInflight は Chromium 同時実行上限の絶対最大値。
	MaxChromiumMaxInflight = 8
	// DefaultNetworkIdleDuration は network_idle 待機の既定静止時間。
	DefaultNetworkIdleDuration = 500 * time.Millisecond
	// DefaultHTTPUserAgent は http フェッチの既定 User-Agent。
	DefaultHTTPUserAgent = "meguri/0.1"
	// MinStealthWindowSize はステルス用ウィンドウサイズの最小値。
	MinStealthWindowSize = 320
	// MaxStealthWindowSize はステルス用ウィンドウサイズの最大値。
	MaxStealthWindowSize = 7680
)

// Config は meguri 全体の実行設定を表すルート構造体。
type Config struct {
	// Request は HTTP 取得に関する設定。
	Request RequestConfig `yaml:"request"`
	// Content は本文抽出・出力フォーマットに関する設定。
	Content ContentConfig `yaml:"content"`
	// PDF は PDF 取得・解析に関する設定。
	PDF PDFConfig `yaml:"pdf"`
	// Crawl はサイト横断クロールに関する設定。
	Crawl CrawlConfig `yaml:"crawl"`
	// Plugins は使用するプラグイン名の選択。
	Plugins PluginSelection `yaml:"plugins"`
	// Targets は処理対象 URL の一覧。
	Targets []string `yaml:"targets"`
	// Output は結果ファイルの出力先設定。
	Output OutputConfig `yaml:"output"`
}

// RequestConfig は HTTP リクエストのタイムアウト・リトライ・ヘッダを保持する。
type RequestConfig struct {
	// Headers は追加するリクエストヘッダ（キーはそのまま送信される）。
	Headers map[string]string `yaml:"headers"`
	// Timeout は 1 リクエストあたりの最大待ち時間。
	Timeout time.Duration `yaml:"timeout"`
	// RetryCount は失敗時の再試行回数（0 は再試行なし）。
	RetryCount int `yaml:"retry_count"`
	// RetryInterval は再試行の間隔。
	RetryInterval time.Duration `yaml:"retry_interval"`
}

// ContentConfig は HTML 本文の抽出方針と出力フォーマットを保持する。
type ContentConfig struct {
	// Formats は書き出す出力フォーマットの一覧。
	Formats []OutputFormat `yaml:"formats"`
	// OnlyMainContent はメインコンテンツ領域のみを抽出するか。
	OnlyMainContent bool `yaml:"only_main_content"`
	// IncludeTags は抽出対象に含める HTML タグ名。
	IncludeTags []string `yaml:"include_tags"`
	// ExcludeTags は抽出から除外する HTML タグ名。
	ExcludeTags []string `yaml:"exclude_tags"`
	// Selector は本文を絞り込む CSS セレクタ（空なら全体）。
	Selector string `yaml:"selector"`
	// ExtractLinks は結果にリンク一覧を含めるか。
	ExtractLinks bool `yaml:"extract_links"`
	// ExtractMetadata はメタデータ抽出を行うか。
	ExtractMetadata bool `yaml:"extract_metadata"`
}

// PDFConfig は PDF 処理の有効化と解析モードを保持する。
type PDFConfig struct {
	// Enabled は PDF の取得・解析を許可するか。
	Enabled bool `yaml:"enabled"`
	// Mode は PDF 解析モード（PDFParseMode を参照）。
	Mode PDFParseMode `yaml:"mode"`
	// MaxPages は解析する最大ページ数（0 は無制限）。
	MaxPages int `yaml:"max_pages"`
	// Output は PDF からの出力形式（PDFOutput を参照）。
	Output PDFOutput `yaml:"output"`
}

// CrawlConfig は BFS クロールの深度・件数・フィルタを保持する。
type CrawlConfig struct {
	// Enabled は複数 URL のクロールを有効にするか。
	Enabled bool `yaml:"enabled"`
	// MaxDepth はシードからの最大リンク深度。
	MaxDepth int `yaml:"max_depth"`
	// MaxPages は訪問する最大ページ数。
	MaxPages int `yaml:"max_pages"`
	// IncludePaths は許可する URL パスの正規表現（空なら制限なし）。
	IncludePaths []string `yaml:"include_paths"`
	// ExcludePaths は除外する URL パスの正規表現。
	ExcludePaths []string `yaml:"exclude_paths"`
	// ExcludeURLs は完全一致でスキップする正規化 URL 一覧（exclude_paths とは別）。
	ExcludeURLs []string `yaml:"exclude_urls"`
	// SkipScrapeURLs は fetch をスキップする正規化 URL 一覧（UI オーケストレーション用。exclude_urls とは別）。
	SkipScrapeURLs []string `yaml:"skip_scrape_urls"`
	// AllowExternal は登録ドメイン外へのリンク追跡を許可するか。
	AllowExternal bool `yaml:"allow_external_links"`
	// AllowSubdomains はサブドメインへの追跡を許可するか。
	AllowSubdomains bool `yaml:"allow_subdomains"`
	// RequestDelay は連続リクエスト間の待機時間（>0 のとき並行度は 1 に制限される）。
	RequestDelay time.Duration `yaml:"request_delay"`
	// MaxConcurrency は同時に走るワーカー数。
	MaxConcurrency int `yaml:"max_concurrency"`
	// RespectRobotsTxt は robots.txt に従うか。
	RespectRobotsTxt bool `yaml:"respect_robots_txt"`
	// FetchLimits はフェッチャ種別ごとの同時取得上限と動的調整設定。
	FetchLimits FetchLimitsConfig `yaml:"fetch_limits"`
}

// FetchLimitsConfig は HTTP / Chromium の同時取得上限とメモリ連動調整を保持する。
//
// Chromium 実効上限はジョブ開始時に ChromiumMaxInflight を初期値とし、
// AutoCalibrate で 1 回上書きした後、DynamicChromium が実行中に ±1 する。
type FetchLimitsConfig struct {
	// HTTPMaxInflight は HTTP フェッチの同時実行上限（0 は既定 16）。
	// AutoCalibrate / DynamicChromium の対象外。
	HTTPMaxInflight int `yaml:"http_max_inflight"`
	// ChromiumMaxInflight は Chromium の静的上限（0 は既定 2）。
	// 両方オフなら常にこの値。キャリブレーション失敗時のフォールバック兼、動的調整の戻り上限の基準。
	ChromiumMaxInflight int `yaml:"chromium_max_inflight"`
	// AutoCalibrate はジョブ開始時に Chromium を 1 回起動して上限を 1〜8 で再計算するか。
	AutoCalibrate bool `yaml:"auto_calibrate"`
	// DynamicChromium は実行中にメモリ使用率で Chromium 上限を ±1 調整するか。
	DynamicChromium bool `yaml:"dynamic_chromium"`
	// MemoryHighWatermark は使用率がこれを超えたら上限を下げる（0 は既定 0.80）。
	MemoryHighWatermark float64 `yaml:"memory_high_watermark"`
	// MemoryLowWatermark は使用率がこれ未満なら上限を上げる（0 は既定 0.60）。
	MemoryLowWatermark float64 `yaml:"memory_low_watermark"`
}

// EffectiveHTTPMaxInflight は正規化済みの HTTP 同時実行上限を返す。
func (f FetchLimitsConfig) EffectiveHTTPMaxInflight() int {
	if f.HTTPMaxInflight <= 0 {
		return DefaultHTTPMaxInflight
	}
	return f.HTTPMaxInflight
}

// EffectiveChromiumMaxInflight は正規化済みの Chromium 同時実行上限を返す。
func (f FetchLimitsConfig) EffectiveChromiumMaxInflight() int {
	if f.ChromiumMaxInflight <= 0 {
		return DefaultChromiumMaxInflight
	}
	return f.ChromiumMaxInflight
}

// EffectiveMemoryHighWatermark は正規化済みの高水位しきい値を返す。
func (f FetchLimitsConfig) EffectiveMemoryHighWatermark() float64 {
	if f.MemoryHighWatermark <= 0 {
		return 0.80
	}
	return f.MemoryHighWatermark
}

// EffectiveMemoryLowWatermark は正規化済みの低水位しきい値を返す。
func (f FetchLimitsConfig) EffectiveMemoryLowWatermark() float64 {
	if f.MemoryLowWatermark <= 0 {
		return 0.60
	}
	return f.MemoryLowWatermark
}

// HTTPStealthConfig は http フェッチャ向けのステルス（リクエスト偽装）設定を保持する。
type HTTPStealthConfig struct {
	// UserAgent は HTTP 取得時の User-Agent（空なら DefaultHTTPUserAgent）。
	UserAgent string `yaml:"user_agent"`
	// AcceptLanguage は Accept-Language ヘッダ値。
	AcceptLanguage string `yaml:"accept_language"`
	// Cookie は Cookie ヘッダ値。
	Cookie string `yaml:"cookie"`
}

// ChromiumStealthConfig は chromium フェッチャ向けのステルス設定を保持する。
type ChromiumStealthConfig struct {
	// UserAgent は chromium 取得時の User-Agent（空なら既定 Chromium UA）。
	UserAgent string `yaml:"user_agent"`
	// Headless はヘッドレス実行を有効にするか。
	Headless bool `yaml:"headless"`
	// HideAutomation は --enable-automation を外し disable-blink-features=AutomationControlled を付与する。
	// 自動テスト情報バー非表示と navigator.webdriver 検知回避に使う。
	HideAutomation bool `yaml:"hide_automation"`
	// DisableGPU は GPU 無効化フラグを付与するか。
	DisableGPU bool `yaml:"disable_gpu"`
	// UserDataDir は永続プロファイル用ディレクトリ（空なら ephemeral）。
	UserDataDir string `yaml:"user_data_dir"`
	// Lang はブラウザの言語設定（--lang）。
	Lang string `yaml:"lang"`
	// WindowWidth はウィンドウ幅（0 なら既定 1920）。
	WindowWidth int `yaml:"window_width"`
	// WindowHeight はウィンドウ高さ（0 なら既定 1080）。
	WindowHeight int `yaml:"window_height"`
	// AcceptLanguage は CDP extra HTTP ヘッダの Accept-Language。
	AcceptLanguage string `yaml:"accept_language"`
}

// StealthConfig はフェッチャ種別ごとのステルス設定を保持する。
type StealthConfig struct {
	// HTTP は http フェッチャ向け設定。
	HTTP HTTPStealthConfig `yaml:"http"`
	// Chromium は chromium フェッチャ向け設定。
	Chromium ChromiumStealthConfig `yaml:"chromium"`
}

// EffectiveUserAgent は http フェッチ用の実効 User-Agent を返す。
func (h HTTPStealthConfig) EffectiveUserAgent() string {
	if ua := strings.TrimSpace(h.UserAgent); ua != "" {
		return ua
	}
	return DefaultHTTPUserAgent
}

// EffectiveLang は chromium の実効 --lang を返す。
func (c ChromiumStealthConfig) EffectiveLang() string {
	if lang := strings.TrimSpace(c.Lang); lang != "" {
		return lang
	}
	return "ja-JP"
}

// EffectiveWindowWidth は chromium の実効ウィンドウ幅を返す。
func (c ChromiumStealthConfig) EffectiveWindowWidth() int {
	if c.WindowWidth > 0 {
		return c.WindowWidth
	}
	return 1920
}

// EffectiveWindowHeight は chromium の実効ウィンドウ高さを返す。
func (c ChromiumStealthConfig) EffectiveWindowHeight() int {
	if c.WindowHeight > 0 {
		return c.WindowHeight
	}
	return 1080
}

// FetcherConfig は chromium フェッチャ専用の実行時設定を保持する。
type FetcherConfig struct {
	// BrowserPath は使用するブラウザ実行ファイルのパス（空なら自動検出）。
	BrowserPath string `yaml:"browser_path"`
	// WaitUntil は Navigate 後のページ読み込み待機モード（空なら load）。
	WaitUntil WaitUntil `yaml:"wait_until"`
	// WaitTimeout は wait_until 待機フェーズ全体の上限（0 なら request.timeout を使用）。
	WaitTimeout time.Duration `yaml:"wait_timeout"`
	// WaitVisibleSelector は wait_until=selector のときに可視になるまで待つ CSS セレクタ。
	WaitVisibleSelector string `yaml:"wait_visible_selector"`
	// NetworkIdleDuration は wait_until=network_idle のとき、接続ゼロが続く必要がある時間。
	NetworkIdleDuration time.Duration `yaml:"network_idle_duration"`
}

// EffectiveWaitUntil は未設定時 load を返す実効 wait_until を返す。
func (fc FetcherConfig) EffectiveWaitUntil() WaitUntil {
	if fc.WaitUntil == "" {
		return WaitUntilLoad
	}
	return fc.WaitUntil
}

// EffectiveNetworkIdleDuration は未設定時の既定 network_idle_duration を返す。
func (fc FetcherConfig) EffectiveNetworkIdleDuration() time.Duration {
	if fc.NetworkIdleDuration <= 0 {
		return DefaultNetworkIdleDuration
	}
	return fc.NetworkIdleDuration
}

// PluginSelection はパイプライン各段で使うプラグイン名を保持する。
type PluginSelection struct {
	// Fetcher は URL フェッチ実装の種別（http / chromium）。
	Fetcher FetcherKind `yaml:"fetcher"`
	// FetcherConfig は Fetcher が chromium のときに使う実行時設定。
	FetcherConfig FetcherConfig `yaml:"fetcher_config"`
	// Stealth はフェッチャ種別ごとのステルス設定。
	Stealth StealthConfig `yaml:"stealth"`
	// PreProcessors は P2 で実行する PreProcessor 名の順序付き一覧。
	PreProcessors []string `yaml:"preprocessors"`
	// Parsers は P5 で登録される Parser 名の一覧。
	Parsers []string `yaml:"parsers"`
	// Transformer は P6 で使う Transformer 名（1 件）。
	Transformer string `yaml:"transformer"`
	// Filters は P7 で実行する Filter 名の順序付き一覧。
	Filters []string `yaml:"filters"`
	// LinkExtractor は P8 で使う LinkExtractor 名（1 件）。
	LinkExtractor string `yaml:"link_extractor"`
}

// OutputConfig は結果ファイルの保存先と命名規則を保持する。
type OutputConfig struct {
	// Dir は出力ディレクトリのパス。
	Dir string `yaml:"dir"`
	// FilePattern はファイル名テンプレート（{seq},{host},{path},{ext} が使える）。
	FilePattern string `yaml:"file_pattern"`
}

// Default は設計書で確定したデフォルト値を適用した Config を返す。
func Default() Config {
	return Config{
		Request: RequestConfig{
			Headers:       map[string]string{},
			Timeout:       60 * time.Second,
			RetryCount:    2,
			RetryInterval: 1 * time.Second,
		},
		Content: ContentConfig{
			Formats:         []OutputFormat{FormatMarkdown},
			OnlyMainContent: true,
			IncludeTags:     []string{},
			ExcludeTags:     []string{"script", "style", "noscript"},
			Selector:        "",
			ExtractLinks:    true,
			ExtractMetadata: true,
		},
		PDF: PDFConfig{
			Enabled:  true,
			Mode:     PDFModeAuto,
			MaxPages: 0,
			Output:   PDFOutputText,
		},
		Crawl: CrawlConfig{
			Enabled:          false,
			MaxDepth:         2,
			MaxPages:         100,
			IncludePaths:     nil,
			ExcludePaths:     nil,
			AllowExternal:    false,
			AllowSubdomains:  false,
			RequestDelay:     0,
			MaxConcurrency:   4,
			RespectRobotsTxt: true,
			FetchLimits: FetchLimitsConfig{
				HTTPMaxInflight:     DefaultHTTPMaxInflight,
				ChromiumMaxInflight: DefaultChromiumMaxInflight,
				AutoCalibrate:       true,
				DynamicChromium:     true,
				MemoryHighWatermark: 0.80,
				MemoryLowWatermark:  0.60,
			},
		},
		Plugins: PluginSelection{
			Fetcher: FetcherHTTP,
			FetcherConfig: FetcherConfig{
				WaitUntil:           WaitUntilLoad,
				WaitTimeout:         5 * time.Second,
				NetworkIdleDuration: DefaultNetworkIdleDuration,
			},
			Stealth: StealthConfig{
				Chromium: ChromiumStealthConfig{
					Headless:       true,
					HideAutomation: true,
					DisableGPU:     true,
					Lang:           "ja-JP",
					WindowWidth:    1920,
					WindowHeight:   1080,
				},
			},
			PreProcessors: nil,
			Parsers:       []string{"html", "pdf"},
			Transformer:   "markdown",
			Filters:       []string{"maincontent"},
			LinkExtractor: "default",
		},
		Output: OutputConfig{
			Dir:         "./out",
			FilePattern: "{seq}-{host}-{path}.{ext}",
		},
	}
}

// Validate は設計書の検証ルールを集中して評価する。
// 違反は errors.Join で集約して返す。
func (c *Config) Validate() error {
	var errs []error

	errs = append(errs, c.validateTargets()...)
	errs = append(errs, c.validateRequest()...)
	errs = append(errs, c.validateContent()...)
	errs = append(errs, c.validatePDF()...)
	errs = append(errs, c.validateCrawl()...)
	errs = append(errs, c.validatePlugins()...)
	errs = append(errs, c.validateOutput()...)

	if c.Crawl.RequestDelay > 0 && c.Crawl.MaxConcurrency != 1 {
		c.Crawl.MaxConcurrency = 1
	}

	c.warnFetchConcurrencyMismatch()

	return errors.Join(errs...)
}

// warnFetchConcurrencyMismatch は Chromium 時にワーカー数が取得上限を超える場合に警告する。
func (c *Config) warnFetchConcurrencyMismatch() {
	fetcher := c.Plugins.Fetcher
	if fetcher == "" {
		fetcher = FetcherHTTP
	}
	if fetcher != FetcherChromium {
		return
	}
	chromiumLimit := c.Crawl.FetchLimits.EffectiveChromiumMaxInflight()
	if c.Crawl.MaxConcurrency > chromiumLimit {
		slog.Warn("crawl.max_concurrency exceeds chromium fetch limit; workers may wait at fetch",
			"max_concurrency", c.Crawl.MaxConcurrency,
			"chromium_max_inflight", chromiumLimit,
		)
	}
}

// validateTargets は targets の件数と URL 形式を検証する。
func (c *Config) validateTargets() []error {
	if len(c.Targets) == 0 {
		return []error{errors.New("targets: 少なくとも1件のURLが必要です")}
	}
	var errs []error
	for i, t := range c.Targets {
		if !strings.HasPrefix(t, "http://") && !strings.HasPrefix(t, "https://") {
			errs = append(errs, fmt.Errorf("targets[%d]: http:// または https:// で始まる必要があります: %q", i, t))
			continue
		}
		if _, err := url.Parse(t); err != nil {
			errs = append(errs, fmt.Errorf("targets[%d]: URLとしてパースできません: %w", i, err))
		}
	}
	return errs
}

// validateRequest は request セクションの数値・ヘッダを検証する。
func (c *Config) validateRequest() []error {
	var errs []error
	if c.Request.Timeout < time.Second || c.Request.Timeout > 300*time.Second {
		errs = append(errs, fmt.Errorf("request.timeout: 1s 以上 300s 以下 (現在: %s)", c.Request.Timeout))
	}
	if c.Request.RetryCount < 0 || c.Request.RetryCount > 10 {
		errs = append(errs, fmt.Errorf("request.retry_count: 0 以上 10 以下 (現在: %d)", c.Request.RetryCount))
	}
	if c.Request.RetryInterval < 100*time.Millisecond || c.Request.RetryInterval > 60*time.Second {
		errs = append(errs, fmt.Errorf("request.retry_interval: 100ms 以上 60s 以下 (現在: %s)", c.Request.RetryInterval))
	}
	for k, v := range c.Request.Headers {
		if strings.TrimSpace(k) == "" || strings.TrimSpace(v) == "" {
			errs = append(errs, fmt.Errorf("request.headers: 空のキーまたは値は許可されません (key=%q)", k))
		}
		if strings.ContainsAny(k, "\r\n") || strings.ContainsAny(v, "\r\n") {
			errs = append(errs, fmt.Errorf("request.headers: 改行を含むヘッダは許可されません (key=%q)", k))
		}
	}
	return errs
}

// validateContent は content セクションのフォーマット・タグ・セレクタを検証する。
func (c *Config) validateContent() []error {
	var errs []error
	seen := map[OutputFormat]bool{}
	for _, f := range c.Content.Formats {
		if !f.Valid() {
			errs = append(errs, fmt.Errorf("content.formats: 不正なフォーマット %q", f))
			continue
		}
		if seen[f] {
			errs = append(errs, fmt.Errorf("content.formats: 重複したフォーマット %q", f))
		}
		seen[f] = true
	}
	incSet := map[string]bool{}
	for _, t := range c.Content.IncludeTags {
		incSet[t] = true
	}
	for _, t := range c.Content.ExcludeTags {
		if incSet[t] {
			errs = append(errs, fmt.Errorf("content.exclude_tags: include_tags と同名タグは指定できません: %q", t))
		}
	}
	if s := strings.TrimSpace(c.Content.Selector); s != "" {
		if _, err := cascadia.Compile(s); err != nil {
			errs = append(errs, fmt.Errorf("content.selector: CSSセレクタとしてパースできません: %w", err))
		}
	}
	return errs
}

// validatePDF は pdf セクションの mode・output・max_pages を検証する。
func (c *Config) validatePDF() []error {
	var errs []error
	if !c.PDF.Mode.Valid() {
		errs = append(errs, fmt.Errorf("pdf.mode: 不正な値 %q", c.PDF.Mode))
	}
	if !c.PDF.Output.Valid() {
		errs = append(errs, fmt.Errorf("pdf.output: 不正な値 %q", c.PDF.Output))
	}
	if c.PDF.MaxPages < 0 || c.PDF.MaxPages > 10000 {
		errs = append(errs, fmt.Errorf("pdf.max_pages: 0 以上 10000 以下 (現在: %d)", c.PDF.MaxPages))
	}
	return errs
}

// validateCrawl は crawl セクションの深度・件数・正規表現を検証する。
func (c *Config) validateCrawl() []error {
	var errs []error
	if c.Crawl.MaxDepth < 0 || c.Crawl.MaxDepth > 10 {
		errs = append(errs, fmt.Errorf("crawl.max_depth: 0 以上 10 以下 (現在: %d)", c.Crawl.MaxDepth))
	}
	if c.Crawl.MaxPages < 1 || c.Crawl.MaxPages > 100000 {
		errs = append(errs, fmt.Errorf("crawl.max_pages: 1 以上 100000 以下 (現在: %d)", c.Crawl.MaxPages))
	}
	if c.Crawl.MaxConcurrency < 1 || c.Crawl.MaxConcurrency > 64 {
		errs = append(errs, fmt.Errorf("crawl.max_concurrency: 1 以上 64 以下 (現在: %d)", c.Crawl.MaxConcurrency))
	}
	if c.Crawl.RequestDelay < 0 || c.Crawl.RequestDelay > 60*time.Second {
		errs = append(errs, fmt.Errorf("crawl.request_delay: 0s 以上 60s 以下 (現在: %s)", c.Crawl.RequestDelay))
	}
	for i, p := range c.Crawl.IncludePaths {
		if _, err := regexp.Compile(p); err != nil {
			errs = append(errs, fmt.Errorf("crawl.include_paths[%d]: 不正な正規表現 %q: %w", i, p, err))
		}
	}
	for i, p := range c.Crawl.ExcludePaths {
		if _, err := regexp.Compile(p); err != nil {
			errs = append(errs, fmt.Errorf("crawl.exclude_paths[%d]: 不正な正規表現 %q: %w", i, p, err))
		}
	}
	for i, raw := range c.Crawl.ExcludeURLs {
		if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
			errs = append(errs, fmt.Errorf("crawl.exclude_urls[%d]: http:// または https:// で始まる必要があります: %q", i, raw))
			continue
		}
		if _, err := url.Parse(raw); err != nil {
			errs = append(errs, fmt.Errorf("crawl.exclude_urls[%d]: URLとしてパースできません: %w", i, err))
		}
	}
	errs = append(errs, c.validateFetchLimits()...)
	return errs
}

// validateFetchLimits は crawl.fetch_limits の範囲を検証する。
func (c *Config) validateFetchLimits() []error {
	fl := c.Crawl.FetchLimits
	var errs []error
	if fl.HTTPMaxInflight < 0 || fl.HTTPMaxInflight > 64 {
		errs = append(errs, fmt.Errorf("crawl.fetch_limits.http_max_inflight: 0（既定16）または 1 以上 64 以下 (現在: %d)", fl.HTTPMaxInflight))
	}
	if fl.ChromiumMaxInflight < 0 || fl.ChromiumMaxInflight > MaxChromiumMaxInflight {
		errs = append(errs, fmt.Errorf("crawl.fetch_limits.chromium_max_inflight: 0（既定2）または 1 以上 %d 以下 (現在: %d)", MaxChromiumMaxInflight, fl.ChromiumMaxInflight))
	}
	low := fl.EffectiveMemoryLowWatermark()
	high := fl.EffectiveMemoryHighWatermark()
	if fl.MemoryLowWatermark != 0 && (fl.MemoryLowWatermark < 0.5 || fl.MemoryLowWatermark >= 0.95) {
		errs = append(errs, fmt.Errorf("crawl.fetch_limits.memory_low_watermark: 0（既定0.60）または 0.5 以上 0.95 未満 (現在: %g)", fl.MemoryLowWatermark))
	}
	if fl.MemoryHighWatermark != 0 && (fl.MemoryHighWatermark <= 0.5 || fl.MemoryHighWatermark > 0.95) {
		errs = append(errs, fmt.Errorf("crawl.fetch_limits.memory_high_watermark: 0（既定0.80）または 0.5 超 0.95 以下 (現在: %g)", fl.MemoryHighWatermark))
	}
	if low >= high {
		errs = append(errs, fmt.Errorf("crawl.fetch_limits: memory_low_watermark (%g) は memory_high_watermark (%g) より小さくする必要があります", low, high))
	}
	return errs
}

var placeholderRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

// validatePlugins は plugins セクションのフェッチャ種別と fetcher_config を検証する。
func (c *Config) validatePlugins() []error {
	var errs []error
	fetcher := c.Plugins.Fetcher
	if fetcher == "" {
		fetcher = FetcherHTTP
	}
	if !fetcher.Valid() {
		errs = append(errs, fmt.Errorf("plugins.fetcher: 不正な値 %q (http / chromium)", fetcher))
	}
	waitUntil := c.Plugins.FetcherConfig.EffectiveWaitUntil()
	if c.Plugins.FetcherConfig.WaitUntil != "" && !c.Plugins.FetcherConfig.WaitUntil.Valid() {
		errs = append(errs, fmt.Errorf("plugins.fetcher_config.wait_until: 不正な値 %q (none / load / network_idle / selector)", c.Plugins.FetcherConfig.WaitUntil))
	}
	if waitUntil == WaitUntilSelector && strings.TrimSpace(c.Plugins.FetcherConfig.WaitVisibleSelector) == "" {
		errs = append(errs, errors.New("plugins.fetcher_config.wait_visible_selector: wait_until=selector のとき必須"))
	}
	if c.Plugins.FetcherConfig.WaitTimeout < 0 || c.Plugins.FetcherConfig.WaitTimeout > 120*time.Second {
		errs = append(errs, fmt.Errorf("plugins.fetcher_config.wait_timeout: 0s 以上 120s 以下 (現在: %s)", c.Plugins.FetcherConfig.WaitTimeout))
	}
	idle := c.Plugins.FetcherConfig.NetworkIdleDuration
	if idle < 0 || idle > 30*time.Second {
		errs = append(errs, fmt.Errorf("plugins.fetcher_config.network_idle_duration: 0s 以上 30s 以下 (現在: %s)", idle))
	} else if idle > 0 && idle < 100*time.Millisecond {
		errs = append(errs, fmt.Errorf("plugins.fetcher_config.network_idle_duration: 100ms 以上 30s 以下 (現在: %s)", idle))
	}
	errs = append(errs, c.validateStealth()...)
	return errs
}

// validateStealth は plugins.stealth の文字列・ウィンドウサイズを検証する。
func (c *Config) validateStealth() []error {
	var errs []error
	checkNoNewline := func(path, val string) {
		if strings.TrimSpace(val) != "" && strings.ContainsAny(val, "\r\n") {
			errs = append(errs, fmt.Errorf("%s: 改行を含む値は許可されません", path))
		}
	}
	s := c.Plugins.Stealth
	checkNoNewline("plugins.stealth.http.user_agent", s.HTTP.UserAgent)
	checkNoNewline("plugins.stealth.http.accept_language", s.HTTP.AcceptLanguage)
	checkNoNewline("plugins.stealth.http.cookie", s.HTTP.Cookie)
	checkNoNewline("plugins.stealth.chromium.user_agent", s.Chromium.UserAgent)
	checkNoNewline("plugins.stealth.chromium.accept_language", s.Chromium.AcceptLanguage)
	for _, dim := range []struct {
		path string
		val  int
	}{
		{"plugins.stealth.chromium.window_width", s.Chromium.WindowWidth},
		{"plugins.stealth.chromium.window_height", s.Chromium.WindowHeight},
	} {
		if dim.val != 0 && (dim.val < MinStealthWindowSize || dim.val > MaxStealthWindowSize) {
			errs = append(errs, fmt.Errorf("%s: 0（既定）または %d 以上 %d 以下 (現在: %d)",
				dim.path, MinStealthWindowSize, MaxStealthWindowSize, dim.val))
		}
	}
	return errs
}

// validateOutput は output.file_pattern のプレースホルダを検証する。
func (c *Config) validateOutput() []error {
	allowed := map[string]bool{"seq": true, "host": true, "path": true, "ext": true}
	var errs []error
	for _, m := range placeholderRe.FindAllStringSubmatch(c.Output.FilePattern, -1) {
		if !allowed[m[1]] {
			errs = append(errs, fmt.Errorf("output.file_pattern: 未知のプレースホルダ {%s}", m[1]))
		}
	}
	return errs
}
