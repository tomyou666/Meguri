// Package header は plugins.stealth.http の値を http リクエストへ転写する P2 PreProcessor を提供する。
package header

import (
	"context"
	"strings"

	"meguri/internal/core"
	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

func init() {
	core.RegisterPreProcessor("header", func() plugin.PreProcessor { return &pp{} })
}

// pp はヘッダ転写用 P2 PreProcessor の実装。
type pp struct {
	// host は Init で受け取る Host。
	host plugin.Host
}

// Metadata は plugin.PreProcessor.Metadata の実装。
func (p *pp) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "header",
		Version:     "0.1.0",
		Kind:        plugin.KindPreProcessor,
		Description: "plugins.stealth.http の値を HTTP リクエストに転写する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (p *pp) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (p *pp) Close(_ context.Context) error { return nil }

// PreProcess は stealth.http の User-Agent 等を req.Headers に転写する。
func (p *pp) PreProcess(_ context.Context, req *model.Request) error {
	if p.host == nil || p.host.FetcherKind() != model.FetcherHTTP {
		return nil
	}
	s := p.host.StealthConfig().HTTP
	setHeader := func(name, value string) {
		v := strings.TrimSpace(value)
		if v == "" {
			return
		}
		if req.Headers == nil {
			req.Headers = map[string]string{}
		}
		req.Headers[name] = v
	}
	setHeader("User-Agent", s.EffectiveUserAgent())
	setHeader("Accept-Language", s.AcceptLanguage)
	setHeader("Cookie", s.Cookie)
	return nil
}
