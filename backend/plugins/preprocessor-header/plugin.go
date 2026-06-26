// Package header は設定の request.headers をリクエストへ転写する P2 PreProcessor を提供する。
package header

import (
	"context"

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
		Description: "request.headers の値をリクエストに転写する",
	}
}

// Init は plugin.Plugin.Init の実装。
func (p *pp) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

// Close は plugin.Plugin.Close の実装。
func (p *pp) Close(_ context.Context) error { return nil }

// PreProcess は設定の User-Agent 等を req.Headers に転写する。
func (p *pp) PreProcess(_ context.Context, req *model.Request) error {
	if v, ok := p.host.Config("request.headers.User-Agent"); ok && v != "" {
		if req.Headers == nil {
			req.Headers = map[string]string{}
		}
		req.Headers["User-Agent"] = v
	}
	return nil
}
