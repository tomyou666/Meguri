//go:build !dev

package app

import (
	"log/slog"

	"meguri/pkg/logger"
)

// setLoggerDefaults は prod 向けのログ既定値を cfg に設定する。
func setLoggerDefaults(cfg *logger.AppConfig) {
	cfg.Console = nil
	cfg.Level = slog.LevelInfo
	cfg.Flush = logger.FlushConfig{
		Policy: logger.FlushImmediate,
	}
}
