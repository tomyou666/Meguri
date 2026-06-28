package pdf

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/domain/model"
	"meguri/internal/domain/plugin"
)

type configHost struct {
	values map[string]string
}

func (h configHost) RequestConfig() model.RequestConfig { return model.RequestConfig{} }

func (h configHost) FetcherConfig() model.FetcherConfig { return model.FetcherConfig{} }

func (h configHost) Config(key string) (string, bool) {
	v, ok := h.values[key]
	return v, ok
}

func fixturePDF(t *testing.T) []byte {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	path := filepath.Join(filepath.Dir(file), "..", "..", "testdata", "pdf", "minimal-text.pdf")
	b, err := os.ReadFile(path)
	require.NoError(t, err)
	return b
}

// TestParser_Parse は ledongthuc/pdf による PDF テキスト抽出を検証する。
func TestParser_Parse(t *testing.T) {
	t.Run("正常系: minimal-text.pdf から既知文字列を抽出する", func(t *testing.T) {
		p := &parser{host: configHost{values: map[string]string{
			"pdf.mode":      "fast",
			"pdf.max_pages": "0",
		}}}
		u, err := url.Parse("https://example.com/minimal-text.pdf")
		require.NoError(t, err)

		content, err := p.Parse(context.Background(), &model.Response{
			URL:         u,
			ContentType: "application/pdf",
			Body:        fixturePDF(t),
		})
		require.NoError(t, err)
		assert.Contains(t, content.Text, "MEGURI-PDF-FIXTURE-ASCII")
		assert.Equal(t, "ledongthuc", content.Metadata["parse_strategy"])
		assert.Equal(t, "fast", content.Metadata["parse_mode"])
	})

	t.Run("異常系: PDF ヘッダが無い body はエラー", func(t *testing.T) {
		p := &parser{host: configHost{}}
		u, err := url.Parse("https://example.com/doc.pdf")
		require.NoError(t, err)

		_, err = p.Parse(context.Background(), &model.Response{
			URL:         u,
			ContentType: "text/html",
			Body:        []byte("<html><body></body></html>"),
		})
		require.Error(t, err)
	})

	t.Run("正常系: max_pages=1 で 1 ページ目のみ処理する", func(t *testing.T) {
		p := &parser{host: configHost{values: map[string]string{
			"pdf.mode":      "fast",
			"pdf.max_pages": "1",
		}}}
		u, err := url.Parse("https://example.com/minimal-text.pdf")
		require.NoError(t, err)

		content, err := p.Parse(context.Background(), &model.Response{
			URL:         u,
			ContentType: "application/pdf",
			Body:        fixturePDF(t),
		})
		require.NoError(t, err)
		assert.Contains(t, content.Text, "MEGURI-PDF-FIXTURE-ASCII")
	})
}

// TestParser_CanParse は PDF 判定を検証する。
func TestParser_CanParse(t *testing.T) {
	p := &parser{}
	u, err := url.Parse("https://example.com/a.pdf")
	require.NoError(t, err)
	assert.True(t, p.CanParse(&model.Response{URL: u, ContentType: "application/pdf"}))
	assert.True(t, p.CanParse(&model.Response{URL: u, ContentType: "text/html"}))
}

var _ plugin.Host = configHost{}
