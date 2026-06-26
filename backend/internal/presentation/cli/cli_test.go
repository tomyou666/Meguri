package cli_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/presentation/cli"

	// プラグイン副作用 import: CLI テストでもプラグインを利用するため
	_ "meguri/plugins/fetcher-chromium"
	_ "meguri/plugins/fetcher-http"
	_ "meguri/plugins/filter-maincontent"
	_ "meguri/plugins/filter-selector"
	_ "meguri/plugins/linkextractor-default"
	_ "meguri/plugins/parser-html"
	_ "meguri/plugins/parser-pdf"
	_ "meguri/plugins/preprocessor-header"
	_ "meguri/plugins/transformer-html"
	_ "meguri/plugins/transformer-markdown"
	_ "meguri/plugins/transformer-raw-html"
)

func newCLITestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<!doctype html><html><head><title>CLIテスト</title></head><body><main><h1>こんにちは</h1><p>CLIから取得した本文です。</p></main></body></html>`))
			return
		}
		http.NotFound(w, r)
	})
	return httptest.NewServer(mux)
}

// TestCLI_RunApp は CLI のエンドツーエンド実行と終了コードを検証する。
func TestCLI_RunApp(t *testing.T) {
	srv := newCLITestServer(t)
	defer srv.Close()

	t.Run("正常系: --url で単一URLを取得し --stdout でMarkdownを出力する", func(t *testing.T) {
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		app := &cli.App{
			Args:   []string{"--url", srv.URL + "/", "--stdout"},
			Stdout: out,
			Stderr: errOut,
		}

		code := app.RunApp()

		assert.Equal(t, 0, code, "正常終了するはず: stderr=%s", errOut.String())
		assert.Contains(t, out.String(), "こんにちは", "Markdownが標準出力に書かれる")
	})

	t.Run("正常系: --output-dir に Markdown ファイルが書き出される", func(t *testing.T) {
		tmp := t.TempDir()

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		app := &cli.App{
			Args: []string{
				"--url", srv.URL + "/",
				"--output-dir", tmp,
				"--output-pattern", "{seq}-{host}-{path}.{ext}",
			},
			Stdout: out,
			Stderr: errOut,
		}

		code := app.RunApp()

		assert.Equal(t, 0, code, "stderr=%s", errOut.String())

		entries, err := os.ReadDir(tmp)
		assert.NoError(t, err)
		assert.NotEmpty(t, entries, "出力ディレクトリに少なくとも1つのファイル")
		var found bool
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".md") {
				body, err := os.ReadFile(filepath.Join(tmp, e.Name()))
				assert.NoError(t, err)
				assert.Contains(t, string(body), "こんにちは")
				found = true
			}
		}
		assert.True(t, found, ".md ファイルが見つかる")
	})

	t.Run("異常系: 設定検証に失敗するURLは終了コード2", func(t *testing.T) {
		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		app := &cli.App{
			Args:   []string{"--url", "ftp://invalid"},
			Stdout: out,
			Stderr: errOut,
		}

		code := app.RunApp()

		assert.Equal(t, 2, code)
		assert.Contains(t, errOut.String(), "設定検証エラー")
	})

	t.Run("正常系: YAML設定ファイルを読み込みCLIフラグで上書きできる", func(t *testing.T) {
		tmp := t.TempDir()
		cfgPath := filepath.Join(tmp, "config.yaml")
		err := os.WriteFile(cfgPath, []byte(`
targets: ["https://will-be-overridden.example.com/"]
content:
  formats: [markdown]
crawl:
  enabled: false
`), 0o644)
		assert.NoError(t, err)

		out := &bytes.Buffer{}
		errOut := &bytes.Buffer{}

		app := &cli.App{
			Args: []string{
				"--config", cfgPath,
				"--url", srv.URL + "/", // YAMLのtargetsをCLIで上書き
				"--stdout",
			},
			Stdout: out,
			Stderr: errOut,
		}

		code := app.RunApp()

		assert.Equal(t, 0, code, "stderr=%s", errOut.String())
		assert.Contains(t, out.String(), "こんにちは", "CLIで上書きしたURLが使われる")
	})
}
