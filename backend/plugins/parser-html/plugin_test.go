package html

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"scraperbot/internal/domain/model"
)

// TestParser_Parse は HTML パース結果を検証する。
func TestParser_Parse(t *testing.T) {
	t.Parallel()

	p := &parser{}

	t.Run("正常系: Response.Body が RawHTML に保存される", func(t *testing.T) {
		t.Parallel()

		body := `<html><head><title>Test</title></head><body><p>hello</p></body></html>`
		u, err := url.Parse("https://example.com/page.html")
		require.NoError(t, err)

		res := &model.Response{
			URL:         u,
			ContentType: "text/html",
			Body:        []byte(body),
		}

		c, err := p.Parse(context.Background(), res)
		require.NoError(t, err)
		assert.Equal(t, body, c.RawHTML)
		assert.Equal(t, "html", c.Format)
		assert.Equal(t, "Test", c.Metadata["title"])
		assert.Equal(t, "hello", c.Text)
	})
}
