package plugin

import (
	"context"
	"net/url"

	"meguri/internal/domain/model"
)

// Fetcher は P3 URL 取得プラグインの契約。
type Fetcher interface {
	Plugin
	Get(ctx context.Context, u *url.URL, headers map[string]string) (*model.Response, error)
}
