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

// buildHTMLFetchTasks は HTML 取得用の chromedp タスク列を組み立てる。
func (c *client) buildHTMLFetchTasks(u string, html *string) []chromedp.Action {
	var idleWatch *networkIdleWatch
	tasks := []chromedp.Action{}

	if c.fetcherCfg.EffectiveWaitUntil() == model.WaitUntilNetworkIdle {
		idleWatch = newNetworkIdleWatch(c.fetcherCfg.EffectiveNetworkIdleDuration())
		tasks = append(tasks, idleWatch.begin())
	}

	tasks = append(tasks, chromiumSetExtraHeadersAction(chromiumExtraHTTPHeaders(c.stealthCfg)))
	tasks = append(tasks, chromedp.Navigate(u))
	c.appendPostNavigateWait(&tasks, idleWatch)
	tasks = append(tasks, chromedp.OuterHTML("html", html, chromedp.ByQuery))
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

// runWithWait は wait_until 待機を含む chromedp アクションを実行する。
func (c *client) runWithWait(ctx context.Context, ua string, build func(context.Context) []chromedp.Action) error {
	waitUntil := c.fetcherCfg.EffectiveWaitUntil()
	if waitUntil == model.WaitUntilNone {
		return c.runWithTab(ctx, ua, func(tabCtx context.Context) error {
			return chromedp.Run(tabCtx, build(tabCtx)...)
		})
	}

	waitCtx, cancel := context.WithTimeout(ctx, c.waitTimeout())
	defer cancel()

	return c.runWithTab(waitCtx, ua, func(tabCtx context.Context) error {
		return chromedp.Run(tabCtx, build(tabCtx)...)
	})
}
