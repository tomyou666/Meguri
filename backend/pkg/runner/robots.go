package runner

import (
	"context"
	"encoding/json"

	"meguri/internal/usecase"
)

// RobotsTxtResult は robots.txt 取得結果（usecase.RobotsTxtResult のエイリアス）。
type RobotsTxtResult = usecase.RobotsTxtResult

// FetchRobotsTxt は host の robots.txt を P3 Fetcher 経由で取得する。
func FetchRobotsTxt(ctx context.Context, host, baseURL string, configLayers ...json.RawMessage) (RobotsTxtResult, error) {
	return usecase.FetchRobotsTxt(ctx, host, baseURL, configLayers...)
}
