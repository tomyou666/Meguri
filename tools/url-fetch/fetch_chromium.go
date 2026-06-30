package main

import (
	"context"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/chromedp"
)

// chromiumVariant は Chromium 取得バリアントの定義。
type chromiumVariant struct {
	// id は stdout に出すバリアント名。
	id string
	// headless はヘッドレス実行するかどうか。
	headless bool
	// userAgentOverride は空でなければ emulation で UA を上書きする。
	userAgentOverride string
	// waitAfterNavigate は Navigate 後に待機するかどうか。
	waitAfterNavigate bool
}

// fetchAllChromiumVariants は Chromium バリアントを並列実行して結果を返す。
func fetchAllChromiumVariants(ctx context.Context, target string) ([]trialResult, error) {
	variants := chromiumVariants()

	browserPath, err := resolveBrowserPath()
	if err != nil {
		results := make([]trialResult, len(variants))
		for i, v := range variants {
			results[i] = trialResult{
				method:  "chromium",
				variant: v.id,
				err:     err,
			}
		}
		return results, nil
	}
	return fetchChromiumVariants(ctx, browserPath, target, variants)
}

// fetchChromiumVariants は Chromium バリアントを並列実行し、定義順の結果を返す。
//
// 各バリアントは専用ブラウザを起動する（並列 navigate のため）。
func fetchChromiumVariants(ctx context.Context, browserPath, target string, variants []chromiumVariant) ([]trialResult, error) {
	results := make([]trialResult, len(variants))
	runParallel(ctx, len(variants), func(ctx context.Context, i int) {
		if err := ctx.Err(); err != nil {
			results[i] = trialResult{
				method:  "chromium",
				variant: variants[i].id,
				err:     err,
			}
			return
		}
		results[i] = runChromiumVariant(ctx, browserPath, target, variants[i])
	})
	return results, nil
}

// runChromiumVariant は1バリアント分のブラウザを起動して取得する。
func runChromiumVariant(ctx context.Context, browserPath, target string, v chromiumVariant) trialResult {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.UserAgent(cfg.ChromeUserAgent),
		chromedp.Flag("headless", v.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	return fetchChromiumOnce(browserCtx, ctx, target, v)
}

// fetchChromiumOnce は起動済みブラウザで1バリアント分の取得を行う。
func fetchChromiumOnce(browserCtx, parentCtx context.Context, target string, v chromiumVariant) trialResult {
	start := time.Now()
	res := trialResult{
		method:  "chromium",
		variant: v.id,
	}

	var html string
	tasks := []chromedp.Action{
		chromedp.Navigate(target),
	}
	if v.userAgentOverride != "" {
		tasks = []chromedp.Action{
			setUserAgentAction(v.userAgentOverride),
			chromedp.Navigate(target),
		}
	}
	if v.waitAfterNavigate {
		wait := waitDuration(parentCtx)
		if wait > 0 {
			tasks = append(tasks, chromedp.Sleep(wait))
		}
	}
	tasks = append(tasks, chromedp.OuterHTML("html", &html, chromedp.ByQuery))

	if err := chromedp.Run(browserCtx, tasks...); err != nil {
		res.duration = time.Since(start)
		res.err = err
		return res
	}

	res.statusCode = 200
	res.bodyBytes = len(html)
	res.duration = time.Since(start)
	return res
}

// waitDuration は Navigate 後待機に使う時間を返す。
//
// 上限は cfg.MaxWaitAfterNavigate。
// 親 context の残り時間がそれより短い場合は残り時間を使う。
func waitDuration(ctx context.Context) time.Duration {
	wait := cfg.MaxWaitAfterNavigate
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return 0
		}
		if remaining < wait {
			wait = remaining
		}
	}
	return wait
}

// setUserAgentAction はナビゲーション前に User-Agent を上書きする Action を返す。
func setUserAgentAction(ua string) chromedp.Action {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		return emulation.SetUserAgentOverride(ua).Do(ctx)
	})
}
