package chromiumfetch

import (
	"context"
	"fmt"
	"strings"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	"meguri/internal/domain/model"
)

// chromiumStealthLaunchFlags はステルス設定から chromedp 起動フラグのキーと値を返す。
func chromiumStealthLaunchFlags(s model.ChromiumStealthConfig) map[string]any {
	flags := map[string]any{}
	if s.HideAutomation {
		// chromedp.DefaultExecAllocatorOptions の enable-automation を上書きして外す。
		// excludeSwitches は ChromeDriver 専用で CLI には効かない。
		flags["enable-automation"] = false
		flags["disable-blink-features"] = "AutomationControlled"
	}
	if s.DisableGPU {
		flags["disable-gpu"] = true
	}
	if dir := strings.TrimSpace(s.UserDataDir); dir != "" {
		flags["user-data-dir"] = dir
	}
	if lang := s.EffectiveLang(); lang != "" {
		flags["lang"] = lang
	}
	w := s.EffectiveWindowWidth()
	h := s.EffectiveWindowHeight()
	if w > 0 && h > 0 {
		flags["window-size"] = fmt.Sprintf("%d,%d", w, h)
	}
	return flags
}

// appendChromiumStealthFlags はステルス設定に応じた chromedp 起動フラグを opts に追加する。
func appendChromiumStealthFlags(opts []chromedp.ExecAllocatorOption, s model.ChromiumStealthConfig) []chromedp.ExecAllocatorOption {
	for name, value := range chromiumStealthLaunchFlags(s) {
		opts = append(opts, chromedp.Flag(name, value))
	}
	return opts
}

// chromiumExtraHTTPHeaders は CDP で付与する追加 HTTP ヘッダを返す。
func chromiumExtraHTTPHeaders(s model.ChromiumStealthConfig) map[string]string {
	out := map[string]string{}
	if v := strings.TrimSpace(s.AcceptLanguage); v != "" {
		out["Accept-Language"] = v
	}
	return out
}

// chromiumSetExtraHeadersAction は Navigate 前に CDP extra HTTP ヘッダを設定するアクションを返す。
func chromiumSetExtraHeadersAction(headers map[string]string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if len(headers) == 0 {
			return nil
		}
		h := make(network.Headers, len(headers))
		for k, v := range headers {
			h[k] = v
		}
		return network.SetExtraHTTPHeaders(h).Do(ctx)
	})
}
