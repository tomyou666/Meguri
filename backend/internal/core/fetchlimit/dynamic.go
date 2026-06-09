package fetchlimit

import (
	"log/slog"
	"time"

	"scraperbot/internal/core/fetchlimit/memprobe"
	"scraperbot/internal/domain/model"
)

const dynamicPollInterval = 5 * time.Second

// StartDynamicChromium はメモリ水位に応じて Chromium 上限を調整する監視を開始する。
func StartDynamicChromium(lim *FetchLimiter, cfg model.FetchLimitsConfig) {
	if lim == nil || !cfg.DynamicChromium {
		return
	}
	lim.mu.Lock()
	if lim.stopDynamic != nil {
		lim.mu.Unlock()
		return
	}
	lim.stopDynamic = make(chan struct{})
	lim.dynamicDone = make(chan struct{})
	stop := lim.stopDynamic
	done := lim.dynamicDone
	high := cfg.EffectiveMemoryHighWatermark()
	low := cfg.EffectiveMemoryLowWatermark()
	lim.mu.Unlock()

	go func() {
		defer close(done)
		ticker := time.NewTicker(dynamicPollInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				adjustChromiumByMemory(lim, high, low)
			}
		}
	}()
}

func adjustChromiumByMemory(lim *FetchLimiter, high, low float64) {
	ratio, err := memprobe.UsedRatio()
	if err != nil {
		slog.Debug("dynamic chromium adjust: memory probe failed", "err", err.Error())
		return
	}

	lim.mu.Lock()
	current := lim.chromiumCap
	staticMax := lim.staticChromium
	if lim.chromiumMax > staticMax {
		staticMax = lim.chromiumMax
	}
	lim.mu.Unlock()

	switch {
	case ratio > high && current > 1:
		lim.SetChromiumCapacity(current - 1)
		slog.Info("dynamic chromium limit decreased",
			"used_ratio", ratio, "chromium_max_inflight", current-1)
	case ratio < low && current < staticMax:
		lim.SetChromiumCapacity(current + 1)
		slog.Info("dynamic chromium limit increased",
			"used_ratio", ratio, "chromium_max_inflight", current+1)
	}
}
