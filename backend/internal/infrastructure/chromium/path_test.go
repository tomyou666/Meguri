package chromium_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"meguri/internal/infrastructure/chromium"
)

// TestResolveBrowserPath は Chromium 実行ファイルの解決優先順位を検証する。
func TestResolveBrowserPath(t *testing.T) {
	t.Run("正常系: 明示パスが存在すればそのパスを返す", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		bin := filepath.Join(tmp, "fake-chromium")
		require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

		path, err := chromium.ResolveBrowserPath(bin)
		require.NoError(t, err)
		assert.Equal(t, bin, path)
	})

	t.Run("正常系: 環境変数 MEGURI_CHROMIUM_PATH を参照する", func(t *testing.T) {
		tmp := t.TempDir()
		bin := filepath.Join(tmp, "env-chromium")
		require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

		t.Setenv(chromium.EnvBrowserPath, bin)
		path, err := chromium.ResolveBrowserPath("")
		require.NoError(t, err)
		assert.Equal(t, bin, path)
	})

	t.Run("異常系: 存在しないパスはエラー", func(t *testing.T) {
		t.Setenv(chromium.EnvBrowserPath, "")
		_, err := chromium.ResolveBrowserPath("/nonexistent/browser/binary")
		require.Error(t, err)
	})

	t.Run("正常系: Windows では Google Chrome を優先する", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows のみ")
		}
		chrome := filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe")
		if _, err := os.Stat(chrome); err != nil {
			t.Skip("google chrome not installed")
		}
		path, err := chromium.ResolveBrowserPath("")
		require.NoError(t, err)
		assert.Equal(t, chrome, path)
	})
}
