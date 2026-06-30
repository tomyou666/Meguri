package main

import (
	"context"
	"io"
	"net/http"
	"time"
)

// httpVariant は HTTP 取得バリアントの定義。
type httpVariant struct {
	// id は stdout に出すバリアント名。
	id string
	// headers はリクエストに付与するヘッダ。nil なら付与しない。
	headers map[string]string
	// utlsTransport は utls 利用時のトランスポート。"http2" / "http1"。空なら標準 net/http。
	utlsTransport string
}

// fetchAllHTTPVariants は HTTP バリアントを並列実行して結果を返す。
func fetchAllHTTPVariants(ctx context.Context, target string) ([]trialResult, error) {
	variants, err := httpVariantsFor(target)
	if err != nil {
		return nil, err
	}
	return fetchHTTPVariants(ctx, target, variants)
}

// fetchHTTPVariants は HTTP バリアントを並列実行し、定義順の結果を返す。
func fetchHTTPVariants(ctx context.Context, target string, variants []httpVariant) ([]trialResult, error) {
	results := make([]trialResult, len(variants))
	runParallel(ctx, len(variants), func(ctx context.Context, i int) {
		if err := ctx.Err(); err != nil {
			results[i] = trialResult{
				method:  "http",
				variant: variants[i].id,
				err:     err,
			}
			return
		}
		results[i] = fetchHTTPOnce(ctx, target, variants[i])
	})
	return results, nil
}

// fetchHTTPOnce は1つの HTTP バリアントで GET を実行する。
func fetchHTTPOnce(ctx context.Context, target string, v httpVariant) trialResult {
	start := time.Now()
	res := trialResult{
		method:  "http",
		variant: v.id,
	}

	reqCtx, cancel := context.WithTimeout(ctx, cfg.HTTPVariantTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, target, nil)
	if err != nil {
		res.duration = time.Since(start)
		res.err = err
		return res
	}
	for k, val := range v.headers {
		req.Header.Set(k, val)
	}

	client := &http.Client{}
	switch v.utlsTransport {
	case "http2":
		client = newUTLSHTTP2Client()
	case "http1":
		client = newUTLSHTTP1Client()
	}
	resp, err := client.Do(req)
	if err != nil {
		res.duration = time.Since(start)
		res.err = err
		return res
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	res.duration = time.Since(start)
	if err != nil {
		res.statusCode = resp.StatusCode
		res.err = err
		return res
	}

	res.statusCode = resp.StatusCode
	res.bodyBytes = len(body)
	return res
}
