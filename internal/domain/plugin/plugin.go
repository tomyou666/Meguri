package plugin

import (
	"context"
	"net/url"
)

type Kind string

const (
	KindPreProcessor  Kind = "preprocessor"
	KindParser        Kind = "parser"
	KindTransformer   Kind = "transformer"
	KindFilter        Kind = "filter"
	KindLinkExtractor Kind = "link_extractor"
)

type Metadata struct {
	Name        string
	Version     string
	Kind        Kind
	Description string
}

type Plugin interface {
	Metadata() Metadata
	Init(ctx context.Context, host Host) error
	Close(ctx context.Context) error
}

// Host はプラグインに渡される最小の依存集合。
// グローバル変数を介さずにここから取得する。
type Host interface {
	Logger() Logger
	Config(key string) (string, bool)
	HTTP() HTTPClient
}

type Logger interface {
	Debug(msg string, kv ...any)
	Info(msg string, kv ...any)
	Warn(msg string, kv ...any)
	Error(msg string, kv ...any)
}

type HTTPClient interface {
	Do(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error)
}

type HTTPRequest struct {
	Method  string
	URL     *url.URL
	Headers map[string]string
	Body    []byte
}

type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}
