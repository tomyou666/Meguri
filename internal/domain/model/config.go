package model

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
)

type Config struct {
	Request RequestConfig   `yaml:"request"`
	Content ContentConfig   `yaml:"content"`
	PDF     PDFConfig       `yaml:"pdf"`
	Crawl   CrawlConfig     `yaml:"crawl"`
	Plugins PluginSelection `yaml:"plugins"`
	Targets []string        `yaml:"targets"`
	Output  OutputConfig    `yaml:"output"`
}

type RequestConfig struct {
	Headers       map[string]string `yaml:"headers"`
	Timeout       time.Duration     `yaml:"timeout"`
	RetryCount    int               `yaml:"retry_count"`
	RetryInterval time.Duration     `yaml:"retry_interval"`
}

type ContentConfig struct {
	Formats         []OutputFormat `yaml:"formats"`
	OnlyMainContent bool           `yaml:"only_main_content"`
	IncludeTags     []string       `yaml:"include_tags"`
	ExcludeTags     []string       `yaml:"exclude_tags"`
	Selector        string         `yaml:"selector"`
	ExtractLinks    bool           `yaml:"extract_links"`
	ExtractMetadata bool           `yaml:"extract_metadata"`
}

type PDFConfig struct {
	Enabled  bool         `yaml:"enabled"`
	Mode     PDFParseMode `yaml:"mode"`
	MaxPages int          `yaml:"max_pages"`
	Output   PDFOutput    `yaml:"output"`
}

type CrawlConfig struct {
	Enabled          bool          `yaml:"enabled"`
	MaxDepth         int           `yaml:"max_depth"`
	MaxPages         int           `yaml:"max_pages"`
	IncludePaths     []string      `yaml:"include_paths"`
	ExcludePaths     []string      `yaml:"exclude_paths"`
	AllowExternal    bool          `yaml:"allow_external_links"`
	AllowSubdomains  bool          `yaml:"allow_subdomains"`
	RequestDelay     time.Duration `yaml:"request_delay"`
	MaxConcurrency   int           `yaml:"max_concurrency"`
	RespectRobotsTxt bool          `yaml:"respect_robots_txt"`
}

type PluginSelection struct {
	PreProcessors []string `yaml:"preprocessors"`
	Parsers       []string `yaml:"parsers"`
	Transformer   string   `yaml:"transformer"`
	Filters       []string `yaml:"filters"`
	LinkExtractor string   `yaml:"link_extractor"`
}

type OutputConfig struct {
	Dir         string `yaml:"dir"`
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
		},
		Plugins: PluginSelection{
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
	errs = append(errs, c.validateOutput()...)

	if c.Crawl.RequestDelay > 0 && c.Crawl.MaxConcurrency != 1 {
		c.Crawl.MaxConcurrency = 1
	}

	return errors.Join(errs...)
}

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
	return errs
}

var placeholderRe = regexp.MustCompile(`\{([a-zA-Z0-9_]+)\}`)

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
