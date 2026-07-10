package chromiumfetch

import (
	"context"
	"strings"
	"time"

	"github.com/chromedp/chromedp"

	"meguri/internal/domain/model"
)

// waitTimeout は wait_until 待機フェーズに使う上限時間を返す。
func (c *client) waitTimeout() time.Duration {
	if c.fetcherCfg.WaitTimeout > 0 {
		return c.fetcherCfg.WaitTimeout
	}
	if c.reqCfg.Timeout > 0 {
		return c.reqCfg.Timeout
	}
	return 60 * time.Second
}

// buildNavigateAndWaitTasks は Navigate と wait_until 待機までの chromedp タスク列を組み立てる。
func (c *client) buildNavigateAndWaitTasks(u string) []chromedp.Action {
	var idleWatch *networkIdleWatch
	tasks := []chromedp.Action{}

	if c.fetcherCfg.EffectiveWaitUntil() == model.WaitUntilNetworkIdle {
		idleWatch = newNetworkIdleWatch(
			c.fetcherCfg.EffectiveNetworkIdleDuration(),
			c.fetcherCfg.EffectiveNetworkIdleRequestMaxAge(),
		)
		tasks = append(tasks, idleWatch.begin())
	}

	tasks = append(tasks, chromiumSetExtraHeadersAction(chromiumExtraHTTPHeaders(c.stealthCfg)))
	tasks = append(tasks, chromedp.Navigate(u))
	c.appendPostNavigateWait(&tasks, idleWatch)
	return tasks
}

// appendPostNavigateWait は Navigate 後の wait_until 待機アクションを tasks に追加する。
func (c *client) appendPostNavigateWait(tasks *[]chromedp.Action, idleWatch *networkIdleWatch) {
	switch c.fetcherCfg.EffectiveWaitUntil() {
	case model.WaitUntilNone:
		return
	case model.WaitUntilLoad:
		*tasks = append(*tasks, chromedp.WaitReady("body", chromedp.ByQuery))
	case model.WaitUntilSelector:
		sel := strings.TrimSpace(c.fetcherCfg.WaitVisibleSelector)
		if sel != "" {
			*tasks = append(*tasks, chromedp.WaitVisible(sel, chromedp.ByQuery))
		}
	case model.WaitUntilNetworkIdle:
		if idleWatch != nil {
			*tasks = append(*tasks, idleWatch.wait())
		}
	}
}

// shouldSleepAfterLoad は wait_until=load 成功後に wait_after_load の追加 sleep を行うべきかを返す。
// selector・network_idle・none では常に false（load 限定の追加待機のため）。
func shouldSleepAfterLoad(waitUntil model.WaitUntil, waitAfterLoad time.Duration) bool {
	return waitUntil == model.WaitUntilLoad && waitAfterLoad > 0
}

// fetchHTML は wait_until 設定に従いページ HTML を取得する。
func (c *client) fetchHTML(ctx context.Context, ua string, u string, html *string) error {
	return c.runWithTab(ctx, ua, func(tabCtx context.Context) error {
		waitUntil := c.fetcherCfg.EffectiveWaitUntil()

		if waitUntil == model.WaitUntilNone {
			return chromedp.Run(tabCtx, append(
				c.buildNavigateAndWaitTasks(u),
				chromedp.OuterHTML("html", html, chromedp.ByQuery),
			)...)
		}

		waitCtx, cancel := context.WithTimeout(tabCtx, c.waitTimeout())
		defer cancel()

		if err := chromedp.Run(waitCtx, c.buildNavigateAndWaitTasks(u)...); err != nil {
			return err
		}

		if shouldSleepAfterLoad(waitUntil, c.fetcherCfg.WaitAfterLoad) {
			if err := chromedp.Run(tabCtx, chromedp.Sleep(c.fetcherCfg.WaitAfterLoad)); err != nil {
				return err
			}
		}

		return chromedp.Run(tabCtx, chromedp.OuterHTML("html", html, chromedp.ByQuery))
	})
}
