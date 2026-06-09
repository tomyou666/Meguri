package chromiumfetch

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveBrowserPath は Chromium 実行ファイルの解決優先順位を検証する。
func TestResolveBrowserPath(t *testing.T) {
	t.Run("正常系: 明示パスが存在すればそのパスを返す", func(t *testing.T) {
		t.Parallel()
		tmp := t.TempDir()
		bin := filepath.Join(tmp, "fake-chromium")
		require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

		path, err := resolveBrowserPath(bin)
		require.NoError(t, err)
		assert.Equal(t, bin, path)
	})

	t.Run("正常系: 環境変数 SCRAPERBOT_CHROMIUM_PATH を参照する", func(t *testing.T) {
		tmp := t.TempDir()
		bin := filepath.Join(tmp, "env-chromium")
		require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

		t.Setenv(EnvBrowserPath, bin)
		path, err := resolveBrowserPath("")
		require.NoError(t, err)
		assert.Equal(t, bin, path)
	})

	t.Run("異常系: 存在しないパスはエラー", func(t *testing.T) {
		t.Setenv(EnvBrowserPath, "")
		_, err := resolveBrowserPath("/nonexistent/browser/binary")
		require.Error(t, err)
	})

	t.Run("正常系: 明示パスは環境変数より優先される", func(t *testing.T) {
		tmp := t.TempDir()
		explicit := filepath.Join(tmp, "explicit")
		envBin := filepath.Join(tmp, "from-env")
		require.NoError(t, os.WriteFile(explicit, []byte{0}, 0o755))
		require.NoError(t, os.WriteFile(envBin, []byte{0}, 0o755))
		t.Setenv(EnvBrowserPath, envBin)

		path, err := resolveBrowserPath(explicit)
		require.NoError(t, err)
		assert.Equal(t, explicit, path)
	})

	t.Run("正常系: Linux ではシステム chromium をフォールバックする", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("linux のみ")
		}
		if _, err := os.Stat("/usr/bin/chromium"); err != nil {
			t.Skip("chromium not installed")
		}
		path, err := resolveBrowserPath("")
		require.NoError(t, err)
		assert.NotEmpty(t, path)
	})

	t.Run("正常系: Windows では Google Chrome を優先する", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("windows のみ")
		}
		chrome := filepath.Join(os.Getenv("ProgramFiles"), "Google", "Chrome", "Application", "chrome.exe")
		if _, err := os.Stat(chrome); err != nil {
			t.Skip("google chrome not installed")
		}
		path, err := resolveBrowserPath("")
		require.NoError(t, err)
		assert.Equal(t, chrome, path)
	})
}
