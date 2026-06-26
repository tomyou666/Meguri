package app

import (
	"meguri/pkg/logger"
)

// InitLogger は Wails アプリ向けに slog を初期化する。
// backend（meguri/*）の slog 呼び出しも同一出力先へまとまる。
func InitLogger() error {
	dir, err := logDir()
	if err != nil {
		return err
	}
	cfg := defaultLoggerConfig(dir)
	return logger.InitApp(cfg)
}

func defaultLoggerConfig(dir string) logger.AppConfig {
	cfg := logger.AppConfig{FileDir: dir}
	setLoggerDefaults(&cfg)
	return cfg
}
