package core

import (
	"fmt"
	"strings"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

// hostImpl は plugin.Host の具象実装。
type hostImpl struct {
	// cfg はフラットキー参照用の設定スナップショット。
	cfg *model.Config
}

// NewHost はプラグインに渡す Host 実装を構築する。
func NewHost(cfg *model.Config) plugin.Host {
	return &hostImpl{cfg: cfg}
}

// RequestConfig は plugin.Host.RequestConfig の実装。
func (h *hostImpl) RequestConfig() model.RequestConfig {
	if h.cfg == nil {
		return model.RequestConfig{}
	}
	return h.cfg.Request
}

// FetcherConfig は plugin.Host.FetcherConfig の実装。
func (h *hostImpl) FetcherConfig() model.FetcherConfig {
	if h.cfg == nil {
		return model.FetcherConfig{}
	}
	return h.cfg.Plugins.FetcherConfig
}

// Config はフラットキーで設定値を文字列として取得する軽量 API。
// 例: "request.headers.User-Agent" / "content.selector" / "pdf.mode"
func (h *hostImpl) Config(key string) (string, bool) {
	if h.cfg == nil {
		return "", false
	}
	return lookupFlat(h.cfg, key)
}

// lookupFlat はドット区切りキーから設定値を文字列で引く。
func lookupFlat(c *model.Config, key string) (string, bool) {
	switch {
	case strings.HasPrefix(key, "request.headers."):
		name := strings.TrimPrefix(key, "request.headers.")
		v, ok := c.Request.Headers[name]
		return v, ok
	case key == "content.selector":
		return c.Content.Selector, true
	case key == "pdf.mode":
		return string(c.PDF.Mode), true
	case key == "pdf.output":
		return string(c.PDF.Output), true
	case key == "pdf.max_pages":
		return fmt.Sprintf("%d", c.PDF.MaxPages), true
	}
	return "", false
}
