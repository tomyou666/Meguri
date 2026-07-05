package runner

import (
	"context"
	"encoding/json"

	"meguri/internal/usecase"
)

//go:generate go tool gowrap gen -p meguri/pkg/runner -i RobotsFetcher -t templates/slog_debug.gotmpl -o robots_fetcher_with_debug_log.go

// RobotsTxtResult は robots.txt 取得結果（usecase.RobotsTxtResult のエイリアス）。
type RobotsTxtResult = usecase.RobotsTxtResult

// RobotsFetcher は robots.txt 取得を抽象化する。
type RobotsFetcher interface {
	// FetchRobotsTxt は host の robots.txt を P3 Fetcher 経由で取得する。
	FetchRobotsTxt(ctx context.Context, host, baseURL string, configLayers ...json.RawMessage) (RobotsTxtResult, error)
}

type robotsFetcherImpl struct{}

func (robotsFetcherImpl) FetchRobotsTxt(ctx context.Context, host, baseURL string, configLayers ...json.RawMessage) (RobotsTxtResult, error) {
	return usecase.FetchRobotsTxt(ctx, host, baseURL, configLayers...)
}

// FetchRobotsTxt は host の robots.txt を P3 Fetcher 経由で取得する。
func FetchRobotsTxt(ctx context.Context, host, baseURL string, configLayers ...json.RawMessage) (RobotsTxtResult, error) {
	return defaultRobotsFetcher.FetchRobotsTxt(ctx, host, baseURL, configLayers...)
}
