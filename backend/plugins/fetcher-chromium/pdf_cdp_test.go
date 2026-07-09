package chromiumfetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// TestShouldInterceptPDFRequest は Fetch インターセプト対象 URL 判定を検証する。
func TestShouldInterceptPDFRequest(t *testing.T) {
	t.Parallel()
	assert.True(t, shouldInterceptPDFRequest("https://example.com/a.pdf", "https://example.com/a.pdf"))
	assert.True(t, shouldInterceptPDFRequest("https://example.com/A.PDF", "https://example.com/other"))
	assert.False(t, shouldInterceptPDFRequest("https://example.com/page.html", "https://example.com/a.pdf"))
}

// TestClient_Get_PDF は CDP Fetch インターセプトで PDF バイナリを取得できることを検証する。
func TestClient_Get_PDF(t *testing.T) {
	if _, err := resolveBrowserPath(""); err != nil {
		t.Skip("chromium browser not available: " + err.Error())
	}

	pdfBytes, err := os.ReadFile(filepath.Join(testdataDir(t), "pdf", "minimal-text.pdf"))
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write(pdfBytes)
	}))
	t.Cleanup(srv.Close)

	cfg := model.Default()
	host := core.NewHost(&cfg)
	c := &client{}
	require.NoError(t, c.Init(context.Background(), host))
	t.Cleanup(func() { _ = c.Close(context.Background()) })

	u, err := url.Parse(srv.URL + "/files/report.pdf")
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)

	res, err := c.Get(ctx, u, nil)
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.True(t, strings.HasPrefix(string(res.Body), "%PDF"))
	assert.Contains(t, string(res.Body), "MEGURI-PDF-FIXTURE-ASCII")
	assert.Contains(t, strings.ToLower(res.ContentType), "application/pdf")
}

// testdataDir は backend/testdata のパスを返す。
func testdataDir(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "testdata"))
}
