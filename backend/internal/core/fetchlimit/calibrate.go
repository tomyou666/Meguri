package fetchlimit

import (
	"context"
	"log/slog"
	"runtime"
	"strings"
	"time"

	"github.com/chromedp/chromedp"

	"scraperbot/internal/core/fetchlimit/memprobe"
	"scraperbot/internal/domain/model"
)

const calibrationMemoryFraction = 0.6

// CalibrateChromium は 1 回の Chromium 起動計測から同時実行上限を調整する。
//
// browserPath は呼び出し元が解決済みのブラウザ実行ファイルパス。
// 空の場合は静的既定 2 にフォールバックする。
func CalibrateChromium(ctx context.Context, lim *FetchLimiter, browserPath string) error {
	if lim == nil {
		return nil
	}
	fallback := func(reason string, err error) {
		lim.SetChromiumCapacity(model.DefaultChromiumMaxInflight)
		if err != nil {
			slog.Warn("chromium fetch calibration failed; using default limit",
				"reason", reason, "limit", model.DefaultChromiumMaxInflight, "err", err.Error())
			return
		}
		slog.Warn("chromium fetch calibration skipped; using default limit",
			"reason", reason, "limit", model.DefaultChromiumMaxInflight)
	}

	if strings.TrimSpace(browserPath) == "" {
		fallback("empty browser path", nil)
		return nil
	}

	available, err := memprobe.AvailableBytes()
	if err != nil {
		fallback("memory probe", err)
		return nil
	}

	before := currentProcessRSS()
	perInstance, err := measureChromiumRSS(ctx, browserPath)
	if err != nil {
		fallback("chromium probe", err)
		return nil
	}
	after := currentProcessRSS()
	if perInstance == 0 && after > before {
		perInstance = after - before
	}
	if perInstance == 0 {
		perInstance = 256 * 1024 * 1024
	}

	max := int(float64(available) * calibrationMemoryFraction / float64(perInstance))
	if max < 1 {
		max = 1
	}
	if max > model.MaxChromiumMaxInflight {
		max = model.MaxChromiumMaxInflight
	}
	lim.SetChromiumCapacity(max)
	slog.Info("chromium fetch limit calibrated",
		"chromium_max_inflight", max,
		"per_instance_bytes", perInstance,
		"available_bytes", available,
	)
	return nil
}

func measureChromiumRSS(ctx context.Context, browserPath string) (uint64, error) {
	probeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(browserPath),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("headless", true),
	)
	allocCtx, allocCancel := chromedp.NewExecAllocator(probeCtx, opts...)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	before := currentProcessRSS()
	if err := chromedp.Run(browserCtx, chromedp.Navigate("about:blank")); err != nil {
		return 0, err
	}
	after := currentProcessRSS()
	if after > before {
		return after - before, nil
	}
	return 128 * 1024 * 1024, nil
}

func currentProcessRSS() uint64 {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.Sys
}
