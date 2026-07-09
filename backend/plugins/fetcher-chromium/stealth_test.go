package chromiumfetch

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"meguri/internal/domain/model"
)

// TestChromiumStealthLaunchFlags は hide_automation 等の起動フラグ組み立てを検証する。
func TestChromiumStealthLaunchFlags(t *testing.T) {
	t.Parallel()

	t.Run("hide_automation 有効時は enable-automation 無効化と AutomationControlled 抑止", func(t *testing.T) {
		flags := chromiumStealthLaunchFlags(model.ChromiumStealthConfig{
			HideAutomation: true,
		})
		assert.Equal(t, false, flags["enable-automation"])
		assert.Equal(t, "AutomationControlled", flags["disable-blink-features"])
	})

	t.Run("hide_automation 無効時は automation 関連フラグを付与しない", func(t *testing.T) {
		flags := chromiumStealthLaunchFlags(model.ChromiumStealthConfig{
			HideAutomation: false,
		})
		_, hasAutomation := flags["enable-automation"]
		_, hasBlink := flags["disable-blink-features"]
		assert.False(t, hasAutomation)
		assert.False(t, hasBlink)
	})

	t.Run("user_data_dir と window-size を付与", func(t *testing.T) {
		flags := chromiumStealthLaunchFlags(model.ChromiumStealthConfig{
			UserDataDir:  "/tmp/profile",
			WindowWidth:  1280,
			WindowHeight: 720,
			Lang:         "ja-JP",
		})
		assert.Equal(t, "/tmp/profile", flags["user-data-dir"])
		assert.Equal(t, "ja-JP", flags["lang"])
		assert.Equal(t, "1280,720", flags["window-size"])
	})
}

// TestChromiumExtraHTTPHeaders は Accept-Language ヘッダ組み立てを検証する。
func TestChromiumExtraHTTPHeaders(t *testing.T) {
	t.Parallel()

	t.Run("accept_language 未設定なら空", func(t *testing.T) {
		assert.Empty(t, chromiumExtraHTTPHeaders(model.ChromiumStealthConfig{}))
	})

	t.Run("accept_language を返す", func(t *testing.T) {
		h := chromiumExtraHTTPHeaders(model.ChromiumStealthConfig{
			AcceptLanguage: "ja,en-US;q=0.9",
		})
		assert.Equal(t, "ja,en-US;q=0.9", h["Accept-Language"])
	})
}
