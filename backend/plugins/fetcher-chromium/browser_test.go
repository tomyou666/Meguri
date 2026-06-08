package chromiumfetch

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveBrowserPath_explicit(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "fake-chromium")
	require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

	path, err := resolveBrowserPath(bin)
	require.NoError(t, err)
	assert.Equal(t, bin, path)
}

func TestResolveBrowserPath_env(t *testing.T) {
	tmp := t.TempDir()
	bin := filepath.Join(tmp, "env-chromium")
	require.NoError(t, os.WriteFile(bin, []byte{0}, 0o755))

	t.Setenv(EnvBrowserPath, bin)
	path, err := resolveBrowserPath("")
	require.NoError(t, err)
	assert.Equal(t, bin, path)
}

func TestResolveBrowserPath_notFound(t *testing.T) {
	t.Setenv(EnvBrowserPath, "")
	_, err := resolveBrowserPath("/nonexistent/browser/binary")
	require.Error(t, err)
}

func TestResolveBrowserPath_prefersExplicitOverEnv(t *testing.T) {
	tmp := t.TempDir()
	explicit := filepath.Join(tmp, "explicit")
	envBin := filepath.Join(tmp, "from-env")
	require.NoError(t, os.WriteFile(explicit, []byte{0}, 0o755))
	require.NoError(t, os.WriteFile(envBin, []byte{0}, 0o755))
	t.Setenv(EnvBrowserPath, envBin)

	path, err := resolveBrowserPath(explicit)
	require.NoError(t, err)
	assert.Equal(t, explicit, path)
}

func TestResolveBrowserPath_systemChromium(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("linux のみ")
	}
	if _, err := os.Stat("/usr/bin/chromium"); err != nil {
		t.Skip("chromium not installed")
	}
	path, err := resolveBrowserPath("")
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestResolveBrowserPath_windowsPrefersChrome(t *testing.T) {
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
}
