// Package usecase はプレゼンテーション層からカーネルへのシナリオを束ねる。
package usecase

import (
	"context"
	"fmt"
	"net/url"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// Scrape は単一URLの取得→出力までを実行するユースケース。
type Scrape struct {
	// Pipeline は 1 URL あたりの処理パイプライン。
	Pipeline *core.Pipeline
}

// NewScrape は単一 URL スクレイプ用ユースケースを構築する。
func NewScrape(pipeline *core.Pipeline) *Scrape {
	return &Scrape{Pipeline: pipeline}
}

// Run は与えられた target URL に対してパイプラインを1回走らせる。
func (s *Scrape) Run(ctx context.Context, target string) (*model.Result, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("invalid target url: %w", err)
	}
	req := model.NewRequest(u, 0)

	out, err := s.Pipeline.Run(ctx, req)
	if err != nil {
		return nil, err
	}
	return out.Result, nil
}
