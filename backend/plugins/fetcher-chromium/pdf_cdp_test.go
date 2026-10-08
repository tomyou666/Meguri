package chromiumfetch

import (
	"context"
	"errors"
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

// TestApplyPDFCaptureResponse は PDF インターセプト応答列の累積結果を検証する。
func TestApplyPDFCaptureResponse(t *testing.T) {
	t.Parallel()

	pdfBody := []byte("%PDF-1.4 fixture")
	htmlBody := []byte("<html>ok</html>")
	err405 := []byte("method not allowed")
	err403 := []byte("forbidden")

	type step struct {
		status int64
		mime   string
		body   []byte
	}
	tests := []struct {
		name           string
		steps          []step
		wantFinalized  bool
		wantHas4xx     bool
		wantStatus     int64
		wantReadyOnTO  bool
		wantBodyPrefix string
	}{
		{
			name: "405のあとPDFの200は成功",
			steps: []step{
				{405, "application/pdf; charset=UTF-8", err405},
				{200, "application/pdf", pdfBody},
			},
			wantFinalized:  true,
			wantHas4xx:     true,
			wantStatus:     200,
			wantReadyOnTO:  true,
			wantBodyPrefix: "%PDF-",
		},
		{
			name: "PDFの200のあと405は成功のまま",
			steps: []step{
				{200, "application/pdf", pdfBody},
				{405, "application/pdf; charset=UTF-8", err405},
			},
			wantFinalized:  true,
			wantHas4xx:     false,
			wantStatus:     200,
			wantReadyOnTO:  true,
			wantBodyPrefix: "%PDF-",
		},
		{
			name: "405のあと403は最初の405",
			steps: []step{
				{405, "application/pdf; charset=UTF-8", err405},
				{403, "text/html", err403},
			},
			wantFinalized:  false,
			wantHas4xx:     true,
			wantStatus:     405,
			wantReadyOnTO:  true,
			wantBodyPrefix: "method not allowed",
		},
		{
			name: "200のHTMLと202は未確定",
			steps: []step{
				{200, "text/html", htmlBody},
				{202, "text/html", htmlBody},
			},
			wantFinalized: false,
			wantHas4xx:    false,
			wantStatus:    0,
			wantReadyOnTO: false,
		},
		{
			name: "400番台だけは最初の4xxをタイムアウトで確定できる",
			steps: []step{
				{405, "application/pdf; charset=UTF-8", err405},
			},
			wantFinalized:  false,
			wantHas4xx:     true,
			wantStatus:     405,
			wantReadyOnTO:  true,
			wantBodyPrefix: "method not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var snap pdfCaptureSnapshot
			for _, s := range tt.steps {
				snap = applyPDFCaptureResponse(snap, s.status, s.mime, s.body)
			}
			assert.Equal(t, tt.wantFinalized, snap.Finalized)
			assert.Equal(t, tt.wantHas4xx, snap.Has4xx)
			assert.Equal(t, tt.wantStatus, snap.StatusCode)
			assert.Equal(t, tt.wantReadyOnTO, pdfCaptureReadyOnTimeout(snap))
			if tt.wantBodyPrefix != "" {
				assert.True(t, strings.HasPrefix(string(snap.Body), tt.wantBodyPrefix))
			} else {
				assert.Empty(t, snap.Body)
			}
		})
	}
}

// TestPDFCaptureOnWaitStop は待ち終了時の期限切れと中断の扱いを検証する。
func TestPDFCaptureOnWaitStop(t *testing.T) {
	t.Parallel()

	with4xx := pdfCaptureSnapshot{
		Body:        []byte("method not allowed"),
		ContentType: "application/pdf; charset=UTF-8",
		StatusCode:  405,
		Has4xx:      true,
	}
	withPDF := pdfCaptureSnapshot{
		Finalized:   true,
		Body:        []byte("%PDF-1.4"),
		ContentType: "application/pdf",
		StatusCode:  200,
	}
	empty := pdfCaptureSnapshot{}

	t.Run("期限切れかつ4xx記憶ありは出口へ渡す", func(t *testing.T) {
		t.Parallel()
		use, err := pdfCaptureOnWaitStop(with4xx, context.DeadlineExceeded)
		assert.True(t, use)
		assert.NoError(t, err)
	})
	t.Run("期限切れかつ成功PDFは出口へ渡す", func(t *testing.T) {
		t.Parallel()
		use, err := pdfCaptureOnWaitStop(withPDF, context.DeadlineExceeded)
		assert.True(t, use)
		assert.NoError(t, err)
	})
	t.Run("期限切れかつ記憶なしはタイムアウト", func(t *testing.T) {
		t.Parallel()
		use, err := pdfCaptureOnWaitStop(empty, context.DeadlineExceeded)
		assert.False(t, use)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("中断は4xxを捨ててCanceled", func(t *testing.T) {
		t.Parallel()
		use, err := pdfCaptureOnWaitStop(with4xx, context.Canceled)
		assert.False(t, use)
		assert.ErrorIs(t, err, context.Canceled)
	})
	t.Run("reqErrがnilならCanceled", func(t *testing.T) {
		t.Parallel()
		use, err := pdfCaptureOnWaitStop(with4xx, nil)
		assert.False(t, use)
		assert.True(t, errors.Is(err, context.Canceled))
	})
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
