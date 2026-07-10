package usecase

import (
	"encoding/json"
	"fmt"
	"time"

	"meguri/internal/domain/model"
)

// normalizeConfigLayer は Wails 経由で JSON 文字列として渡されたレイヤをオブジェクト JSON に正規化する。
func normalizeConfigLayer(layer json.RawMessage) (json.RawMessage, error) {
	if len(layer) == 0 || string(layer) == "null" {
		return nil, nil
	}
	var asString string
	if err := json.Unmarshal(layer, &asString); err == nil {
		if asString == "" || asString == "null" {
			return nil, nil
		}
		return json.RawMessage(asString), nil
	}
	return layer, nil
}

// MergeUIConfigLayers は UI 由来の PartialConfig JSON を深くマージする。
func MergeUIConfigLayers(layers ...json.RawMessage) (json.RawMessage, error) {
	merged := map[string]json.RawMessage{}
	for _, layer := range layers {
		normalized, err := normalizeConfigLayer(layer)
		if err != nil {
			return nil, err
		}
		if len(normalized) == 0 {
			continue
		}
		var m map[string]json.RawMessage
		if err := json.Unmarshal(normalized, &m); err != nil {
			return nil, fmt.Errorf("merge layer: %w", err)
		}
		for k, v := range m {
			if existing, ok := merged[k]; ok {
				out, err := mergeSection(existing, v)
				if err != nil {
					return nil, err
				}
				merged[k] = out
			} else {
				merged[k] = v
			}
		}
	}
	return json.Marshal(merged)
}

func mergeSection(base, override json.RawMessage) (json.RawMessage, error) {
	var b, o map[string]json.RawMessage
	if err := json.Unmarshal(base, &b); err != nil {
		return override, nil
	}
	if err := json.Unmarshal(override, &o); err != nil {
		return override, nil
	}
	for k, v := range o {
		b[k] = v
	}
	return json.Marshal(b)
}

// ParseUIConfig はマージ済み UI JSON を backend model.Config に変換する。
func ParseUIConfig(raw json.RawMessage) (*model.Config, error) {
	cfg := model.Default()
	if len(raw) == 0 {
		return &cfg, nil
	}
	var ui uiConfigJSON
	if err := json.Unmarshal(raw, &ui); err != nil {
		return nil, fmt.Errorf("parse ui config: %w", err)
	}
	applyUIConfig(&cfg, ui)
	return &cfg, nil
}

type uiConfigJSON struct {
	Request *uiRequestJSON `json:"request"`
	Content *uiContentJSON `json:"content"`
	PDF     *uiPDFJSON     `json:"pdf"`
	Crawl   *uiCrawlJSON   `json:"crawl"`
	Plugins *uiPluginsJSON `json:"plugins"`
	Output  *uiOutputJSON  `json:"output"`
}

type uiRequestJSON struct {
	Headers       map[string]string `json:"headers"`
	Timeout       string            `json:"timeout"`
	RetryCount    int               `json:"retry_count"`
	RetryInterval string            `json:"retry_interval"`
}

type uiContentJSON struct {
	Formats         []string `json:"formats"`
	OnlyMainContent *bool    `json:"only_main_content"`
	IncludeTags     []string `json:"include_tags"`
	ExcludeTags     []string `json:"exclude_tags"`
	Selector        string   `json:"selector"`
	ExtractLinks    *bool    `json:"extract_links"`
	ExtractMetadata *bool    `json:"extract_metadata"`
}

type uiPDFJSON struct {
	Enabled  *bool  `json:"enabled"`
	Mode     string `json:"mode"`
	MaxPages int    `json:"max_pages"`
	Output   string `json:"output"`
}

type uiFetchLimitsJSON struct {
	HTTPMaxInflight     int     `json:"http_max_inflight"`
	ChromiumMaxInflight int     `json:"chromium_max_inflight"`
	AutoCalibrate       *bool   `json:"auto_calibrate"`
	DynamicChromium     *bool   `json:"dynamic_chromium"`
	MemoryHighWatermark float64 `json:"memory_high_watermark"`
	MemoryLowWatermark  float64 `json:"memory_low_watermark"`
}

type uiCrawlJSON struct {
	Enabled          *bool              `json:"enabled"`
	MaxDepth         int                `json:"max_depth"`
	MaxPages         int                `json:"max_pages"`
	IncludePaths     []string           `json:"include_paths"`
	ExcludePaths     []string           `json:"exclude_paths"`
	ExcludeURLs      []string           `json:"exclude_urls"`
	AllowExternal    *bool              `json:"allow_external_links"`
	AllowSubdomains  *bool              `json:"allow_subdomains"`
	RequestDelay     string             `json:"request_delay"`
	MaxConcurrency   int                `json:"max_concurrency"`
	RespectRobotsTxt *bool              `json:"respect_robots_txt"`
	FetchLimits      *uiFetchLimitsJSON `json:"fetch_limits"`
}

type uiPluginsJSON struct {
	Fetcher       string                 `json:"fetcher"`
	FetcherConfig map[string]interface{} `json:"fetcher_config"`
	Stealth       *uiStealthJSON         `json:"stealth"`
	PreProcessors []string               `json:"preprocessors"`
	Parsers       []string               `json:"parsers"`
	Transformer   string                 `json:"transformer"`
	Filters       []string               `json:"filters"`
	LinkExtractor string                 `json:"link_extractor"`
}

type uiHTTPStealthJSON struct {
	UserAgent      string `json:"user_agent"`
	AcceptLanguage string `json:"accept_language"`
	Cookie         string `json:"cookie"`
}

type uiChromiumStealthJSON struct {
	UserAgent      string `json:"user_agent"`
	Headless       *bool  `json:"headless"`
	HideAutomation *bool  `json:"hide_automation"`
	DisableGPU     *bool  `json:"disable_gpu"`
	UserDataDir    string `json:"user_data_dir"`
	Lang           string `json:"lang"`
	WindowWidth    int    `json:"window_width"`
	WindowHeight   int    `json:"window_height"`
	AcceptLanguage string `json:"accept_language"`
}

type uiStealthJSON struct {
	HTTP     *uiHTTPStealthJSON     `json:"http"`
	Chromium *uiChromiumStealthJSON `json:"chromium"`
}

type uiOutputJSON struct {
	Dir         string `json:"dir"`
	FilePattern string `json:"file_pattern"`
}

func applyUIConfig(cfg *model.Config, ui uiConfigJSON) {
	if ui.Request != nil {
		if ui.Request.Headers != nil {
			cfg.Request.Headers = ui.Request.Headers
		}
		if d, err := parseDuration(ui.Request.Timeout); err == nil && d > 0 {
			cfg.Request.Timeout = d
		}
		if ui.Request.RetryCount > 0 {
			cfg.Request.RetryCount = ui.Request.RetryCount
		}
		if d, err := parseDuration(ui.Request.RetryInterval); err == nil && d > 0 {
			cfg.Request.RetryInterval = d
		}
	}
	if ui.Content != nil {
		if len(ui.Content.Formats) > 0 {
			cfg.Content.Formats = make([]model.OutputFormat, len(ui.Content.Formats))
			for i, f := range ui.Content.Formats {
				cfg.Content.Formats[i] = model.OutputFormat(f)
			}
		}
		if ui.Content.OnlyMainContent != nil {
			cfg.Content.OnlyMainContent = *ui.Content.OnlyMainContent
		}
		if ui.Content.IncludeTags != nil {
			cfg.Content.IncludeTags = ui.Content.IncludeTags
		}
		if ui.Content.ExcludeTags != nil {
			cfg.Content.ExcludeTags = ui.Content.ExcludeTags
		}
		if ui.Content.Selector != "" {
			cfg.Content.Selector = ui.Content.Selector
		}
		if ui.Content.ExtractLinks != nil {
			cfg.Content.ExtractLinks = *ui.Content.ExtractLinks
		}
		if ui.Content.ExtractMetadata != nil {
			cfg.Content.ExtractMetadata = *ui.Content.ExtractMetadata
		}
	}
	if ui.PDF != nil {
		if ui.PDF.Enabled != nil {
			cfg.PDF.Enabled = *ui.PDF.Enabled
		}
		if ui.PDF.Mode != "" {
			cfg.PDF.Mode = model.PDFParseMode(ui.PDF.Mode)
		}
		if ui.PDF.MaxPages >= 0 {
			cfg.PDF.MaxPages = ui.PDF.MaxPages
		}
		if ui.PDF.Output != "" {
			cfg.PDF.Output = model.PDFOutput(ui.PDF.Output)
		}
	}
	if ui.Crawl != nil {
		if ui.Crawl.Enabled != nil {
			cfg.Crawl.Enabled = *ui.Crawl.Enabled
		}
		if ui.Crawl.MaxDepth >= 0 {
			cfg.Crawl.MaxDepth = ui.Crawl.MaxDepth
		}
		if ui.Crawl.MaxPages > 0 {
			cfg.Crawl.MaxPages = ui.Crawl.MaxPages
		}
		if ui.Crawl.IncludePaths != nil {
			cfg.Crawl.IncludePaths = ui.Crawl.IncludePaths
		}
		if ui.Crawl.ExcludePaths != nil {
			cfg.Crawl.ExcludePaths = ui.Crawl.ExcludePaths
		}
		if ui.Crawl.ExcludeURLs != nil {
			cfg.Crawl.ExcludeURLs = ui.Crawl.ExcludeURLs
		}
		if ui.Crawl.AllowExternal != nil {
			cfg.Crawl.AllowExternal = *ui.Crawl.AllowExternal
		}
		if ui.Crawl.AllowSubdomains != nil {
			cfg.Crawl.AllowSubdomains = *ui.Crawl.AllowSubdomains
		}
		if d, err := parseDuration(ui.Crawl.RequestDelay); err == nil {
			cfg.Crawl.RequestDelay = d
		}
		if ui.Crawl.MaxConcurrency > 0 {
			cfg.Crawl.MaxConcurrency = ui.Crawl.MaxConcurrency
		}
		if ui.Crawl.RespectRobotsTxt != nil {
			cfg.Crawl.RespectRobotsTxt = *ui.Crawl.RespectRobotsTxt
		}
		if ui.Crawl.FetchLimits != nil {
			fl := ui.Crawl.FetchLimits
			if fl.HTTPMaxInflight > 0 {
				cfg.Crawl.FetchLimits.HTTPMaxInflight = fl.HTTPMaxInflight
			}
			if fl.ChromiumMaxInflight > 0 {
				cfg.Crawl.FetchLimits.ChromiumMaxInflight = fl.ChromiumMaxInflight
			}
			if fl.AutoCalibrate != nil {
				cfg.Crawl.FetchLimits.AutoCalibrate = *fl.AutoCalibrate
			}
			if fl.DynamicChromium != nil {
				cfg.Crawl.FetchLimits.DynamicChromium = *fl.DynamicChromium
			}
			if fl.MemoryHighWatermark > 0 {
				cfg.Crawl.FetchLimits.MemoryHighWatermark = fl.MemoryHighWatermark
			}
			if fl.MemoryLowWatermark > 0 {
				cfg.Crawl.FetchLimits.MemoryLowWatermark = fl.MemoryLowWatermark
			}
		}
	}
	if ui.Plugins != nil {
		if ui.Plugins.Fetcher != "" {
			cfg.Plugins.Fetcher = model.FetcherKind(ui.Plugins.Fetcher)
		}
		if len(ui.Plugins.PreProcessors) > 0 {
			cfg.Plugins.PreProcessors = ui.Plugins.PreProcessors
		}
		if len(ui.Plugins.Parsers) > 0 {
			cfg.Plugins.Parsers = ui.Plugins.Parsers
		}
		if ui.Plugins.Transformer != "" {
			cfg.Plugins.Transformer = ui.Plugins.Transformer
		}
		if len(ui.Plugins.Filters) > 0 {
			cfg.Plugins.Filters = ui.Plugins.Filters
		}
		if ui.Plugins.LinkExtractor != "" {
			cfg.Plugins.LinkExtractor = ui.Plugins.LinkExtractor
		}
		if ui.Plugins.FetcherConfig != nil {
			applyFetcherConfig(&cfg.Plugins.FetcherConfig, ui.Plugins.FetcherConfig)
		}
		if ui.Plugins.Stealth != nil {
			applyStealthConfig(&cfg.Plugins.Stealth, ui.Plugins.Stealth)
		}
	}
	if ui.Output != nil {
		if ui.Output.Dir != "" {
			cfg.Output.Dir = ui.Output.Dir
		}
		if ui.Output.FilePattern != "" {
			cfg.Output.FilePattern = ui.Output.FilePattern
		}
	}
}

func parseDuration(s string) (time.Duration, error) {
	if s == "" {
		return 0, nil
	}
	return time.ParseDuration(s)
}

func applyFetcherConfig(fc *model.FetcherConfig, raw map[string]interface{}) {
	if fc == nil || raw == nil {
		return
	}
	if v, ok := raw["browser_path"].(string); ok {
		fc.BrowserPath = v
	}
	if v, ok := raw["wait_until"].(string); ok && v != "" {
		fc.WaitUntil = model.WaitUntil(v)
	}
	if v, ok := raw["wait_visible_selector"].(string); ok {
		fc.WaitVisibleSelector = v
	}
	if v, ok := raw["wait_timeout"].(string); ok {
		if d, err := parseDuration(v); err == nil {
			fc.WaitTimeout = d
		}
	}
	if v, ok := raw["network_idle_duration"].(string); ok {
		if d, err := parseDuration(v); err == nil {
			fc.NetworkIdleDuration = d
		}
	}
	if v, ok := raw["wait_after_load"].(string); ok {
		if d, err := parseDuration(v); err == nil {
			fc.WaitAfterLoad = d
		}
	}
	if v, ok := raw["network_idle_request_max_age"].(string); ok {
		if d, err := parseDuration(v); err == nil {
			fc.NetworkIdleRequestMaxAge = d
		}
	}
}

func applyStealthConfig(sc *model.StealthConfig, ui *uiStealthJSON) {
	if sc == nil || ui == nil {
		return
	}
	if ui.HTTP != nil {
		if ui.HTTP.UserAgent != "" {
			sc.HTTP.UserAgent = ui.HTTP.UserAgent
		}
		if ui.HTTP.AcceptLanguage != "" {
			sc.HTTP.AcceptLanguage = ui.HTTP.AcceptLanguage
		}
		if ui.HTTP.Cookie != "" {
			sc.HTTP.Cookie = ui.HTTP.Cookie
		}
	}
	if ui.Chromium != nil {
		c := &sc.Chromium
		if ui.Chromium.UserAgent != "" {
			c.UserAgent = ui.Chromium.UserAgent
		}
		if ui.Chromium.Headless != nil {
			c.Headless = *ui.Chromium.Headless
		}
		if ui.Chromium.HideAutomation != nil {
			c.HideAutomation = *ui.Chromium.HideAutomation
		}
		if ui.Chromium.DisableGPU != nil {
			c.DisableGPU = *ui.Chromium.DisableGPU
		}
		if ui.Chromium.UserDataDir != "" {
			c.UserDataDir = ui.Chromium.UserDataDir
		}
		if ui.Chromium.Lang != "" {
			c.Lang = ui.Chromium.Lang
		}
		if ui.Chromium.WindowWidth > 0 {
			c.WindowWidth = ui.Chromium.WindowWidth
		}
		if ui.Chromium.WindowHeight > 0 {
			c.WindowHeight = ui.Chromium.WindowHeight
		}
		if ui.Chromium.AcceptLanguage != "" {
			c.AcceptLanguage = ui.Chromium.AcceptLanguage
		}
	}
}

// DeriveContentFormats は transformer と extract フラグから content.formats を導出する。
func DeriveContentFormats(cfg *model.Config) {
	if cfg == nil {
		return
	}
	t := cfg.Plugins.Transformer
	if t == "" {
		t = "markdown"
	}
	formats := []model.OutputFormat{model.OutputFormat(t)}
	if cfg.Content.ExtractMetadata {
		formats = append(formats, model.FormatMetadata)
	}
	if cfg.Content.ExtractLinks {
		formats = append(formats, model.FormatLinks)
	}
	cfg.Content.Formats = formats
}
