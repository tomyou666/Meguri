package core

import (
	"fmt"
	"strings"

	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

type hostImpl struct {
	logger plugin.Logger
	cfg    *model.Config
	http   plugin.HTTPClient
}

// NewHost はプラグインに渡す Host 実装を構築する。
func NewHost(logger plugin.Logger, cfg *model.Config, http plugin.HTTPClient) plugin.Host {
	return &hostImpl{logger: logger, cfg: cfg, http: http}
}

func (h *hostImpl) Logger() plugin.Logger   { return h.logger }
func (h *hostImpl) HTTP() plugin.HTTPClient { return h.http }

// Config はフラットキーで設定値を文字列として取得する軽量 API。
// 例: "request.headers.User-Agent" / "content.selector" / "pdf.mode"
func (h *hostImpl) Config(key string) (string, bool) {
	if h.cfg == nil {
		return "", false
	}
	return lookupFlat(h.cfg, key)
}

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
