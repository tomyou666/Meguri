package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"meguri/internal/core"
)

const robotsBodyLimit = 512 * 1024

// RobotsTxtResult は robots.txt 取得結果。
type RobotsTxtResult struct {
	// Host は対象ホスト名。
	Host string
	// Status は取得状態（found / not_found / error）。
	Status string
	// StatusCode は HTTP ステータスコード。
	StatusCode int
	// Body は robots.txt 本文（found 時）。
	Body string
	// Error はエラー詳細（error 時）。
	Error string
}

// FetchRobotsTxt は host の robots.txt を P3 Fetcher 経由で取得する。
//
// baseURL は scheme 推定用（ノード URL）。空の場合は https を使用する。
//
// configLayers は MergeUIConfigLayers に渡す PartialConfig JSON（appDefaults + workspace settings 等）。
func FetchRobotsTxt(ctx context.Context, host, baseURL string, configLayers ...json.RawMessage) (RobotsTxtResult, error) {
	host = strings.TrimSpace(strings.ToLower(host))
	if host == "" {
		return RobotsTxtResult{}, fmt.Errorf("host is required")
	}

	scheme := "https"
	if baseURL != "" {
		if u, err := url.Parse(baseURL); err == nil && u.Scheme != "" {
			scheme = u.Scheme
		}
	}

	merged, err := MergeUIConfigLayers(configLayers...)
	if err != nil {
		return RobotsTxtResult{Host: host, Status: "error", Error: err.Error()}, nil
	}
	cfg, err := ParseUIConfig(merged)
	if err != nil {
		return RobotsTxtResult{Host: host, Status: "error", Error: err.Error()}, nil
	}

	robotsURL, err := url.Parse(fmt.Sprintf("%s://%s/robots.txt", scheme, host))
	if err != nil {
		return RobotsTxtResult{Host: host, Status: "error", Error: err.Error()}, nil
	}

	kHost := core.NewHost(cfg)
	k := core.NewKernel(cfg, kHost, core.Default())
	if err := k.Init(ctx); err != nil {
		return RobotsTxtResult{Host: host, Status: "error", Error: err.Error()}, nil
	}
	defer func() { _ = k.Close(ctx) }()

	res, err := k.Fetcher().Get(ctx, robotsURL, cfg.Request.Headers)
	if err != nil {
		return RobotsTxtResult{Host: host, Status: "error", Error: err.Error()}, nil
	}

	return mapRobotsResponse(host, res.StatusCode, res.Body), nil
}

func mapRobotsResponse(host string, statusCode int, body []byte) RobotsTxtResult {
	if statusCode == http.StatusNotFound {
		return RobotsTxtResult{
			Host:       host,
			Status:     "not_found",
			StatusCode: statusCode,
		}
	}
	if statusCode < 200 || statusCode >= 400 {
		return RobotsTxtResult{
			Host:       host,
			Status:     "error",
			StatusCode: statusCode,
			Error:      fmt.Sprintf("HTTP %d", statusCode),
		}
	}

	if len(body) > robotsBodyLimit {
		body = body[:robotsBodyLimit]
	}
	return RobotsTxtResult{
		Host:       host,
		Status:     "found",
		StatusCode: statusCode,
		Body:       string(body),
	}
}
