// Package header は設定の request.headers をリクエストへ転写する P2 PreProcessor を提供する。
package header

import (
	"context"

	"scraperbot/internal/core"
	"scraperbot/internal/domain/model"
	"scraperbot/internal/domain/plugin"
)

func init() {
	core.RegisterPreProcessor("header", func() plugin.PreProcessor { return &pp{} })
}

type pp struct {
	host plugin.Host
}

func (p *pp) Metadata() plugin.Metadata {
	return plugin.Metadata{
		Name:        "header",
		Version:     "0.1.0",
		Kind:        plugin.KindPreProcessor,
		Description: "request.headers の値をリクエストに転写する",
	}
}

func (p *pp) Init(_ context.Context, host plugin.Host) error {
	p.host = host
	return nil
}

func (p *pp) Close(_ context.Context) error { return nil }

func (p *pp) PreProcess(_ context.Context, req *model.Request) error {
	if v, ok := p.host.Config("request.headers.User-Agent"); ok && v != "" {
		if req.Headers == nil {
			req.Headers = map[string]string{}
		}
		req.Headers["User-Agent"] = v
	}
	return nil
}
