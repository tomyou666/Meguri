package header

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

type stealthHost struct {
	fetcher model.FetcherKind
	stealth model.StealthConfig
}

func (h stealthHost) Config(string) (string, bool) { return "", false }
func (h stealthHost) RequestConfig() model.RequestConfig {
	return model.RequestConfig{}
}
func (h stealthHost) FetcherConfig() model.FetcherConfig { return model.FetcherConfig{} }
func (h stealthHost) StealthConfig() model.StealthConfig { return h.stealth }
func (h stealthHost) FetcherKind() model.FetcherKind     { return h.fetcher }

var _ plugin.Host = stealthHost{}

// TestPreProcess は stealth.http ヘッダ転写を検証する。
func TestPreProcess(t *testing.T) {
	t.Run("正常系: http 時に stealth.http をヘッダへ転写", func(t *testing.T) {
		p := &pp{}
		require.NoError(t, p.Init(context.Background(), stealthHost{
			fetcher: model.FetcherHTTP,
			stealth: model.StealthConfig{
				HTTP: model.HTTPStealthConfig{
					UserAgent:      "Test/1.0",
					AcceptLanguage: "ja,en-US;q=0.9",
					Cookie:         "session=abc",
				},
			},
		}))
		req := &model.Request{}
		require.NoError(t, p.PreProcess(context.Background(), req))
		assert.Equal(t, "Test/1.0", req.Headers["User-Agent"])
		assert.Equal(t, "ja,en-US;q=0.9", req.Headers["Accept-Language"])
		assert.Equal(t, "session=abc", req.Headers["Cookie"])
	})

	t.Run("正常系: chromium 時はヘッダを付与しない", func(t *testing.T) {
		p := &pp{}
		require.NoError(t, p.Init(context.Background(), stealthHost{
			fetcher: model.FetcherChromium,
			stealth: model.StealthConfig{
				HTTP: model.HTTPStealthConfig{UserAgent: "Ignored/1.0"},
			},
		}))
		req := &model.Request{}
		require.NoError(t, p.PreProcess(context.Background(), req))
		assert.Empty(t, req.Headers)
	})

	t.Run("正常系: user_agent 未設定時は既定 UA", func(t *testing.T) {
		p := &pp{}
		require.NoError(t, p.Init(context.Background(), stealthHost{
			fetcher: model.FetcherHTTP,
		}))
		req := &model.Request{}
		require.NoError(t, p.PreProcess(context.Background(), req))
		assert.Equal(t, model.DefaultHTTPUserAgent, req.Headers["User-Agent"])
	})
}
