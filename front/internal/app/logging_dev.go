//go:build dev

package app

import (
	"os"
	"time"

	"log/slog"

	"meguri/pkg/logger"
)

// setLoggerDefaults は dev 向けのログ既定値を cfg に設定する。
func setLoggerDefaults(cfg *logger.AppConfig) {
	cfg.Console = os.Stdout
	cfg.Level = slog.LevelDebug
	cfg.Flush = logger.FlushConfig{
		Policy:   logger.FlushInterval,
		Interval: time.Second,
	}
}
