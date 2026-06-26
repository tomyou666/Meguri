package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

const (
	defaultLogFilename = "meguri.log"
	defaultMaxSize     = 10 * 1024 * 1024 // 10MB
	defaultMaxFiles    = 5
)

// AppConfig は Wails 等デスクトップアプリ向けのログ初期化設定。
type AppConfig struct {
	// Console はコンソール出力先。nil のときコンソールへは書かない。
	Console io.Writer
	// Level は slog の最小出力レベル。
	Level slog.Level
	// FileDir はログファイルのディレクトリ。
	FileDir string
	// FileName はログファイル名。空のとき meguri.log。
	FileName string
	// MaxSize はローテート前の最大バイト数。0 のとき 10MB。
	MaxSize int64
	// MaxFiles は保持するローテート済みファイル数。0 のとき 5。
	MaxFiles int
	// Flush はファイル出力の flush 挙動。
	Flush FlushConfig
}

// InitConsole は指定 Writer / Level で slog 標準ロガーを構築し SetDefault する。
func InitConsole(w io.Writer, level slog.Level) {
	Init(w, level)
}

// InitApp はコンソールとローテーション付きファイルへ slog を出力する。
func InitApp(cfg AppConfig) error {
	filename := cfg.FileName
	if filename == "" {
		filename = defaultLogFilename
	}
	maxSize := cfg.MaxSize
	if maxSize == 0 {
		maxSize = defaultMaxSize
	}
	maxFiles := cfg.MaxFiles
	if maxFiles == 0 {
		maxFiles = defaultMaxFiles
	}

	rw, err := newRotatingWriter(cfg.FileDir, filename, maxSize, maxFiles, cfg.Flush)
	if err != nil {
		return fmt.Errorf("init log file: %w", err)
	}
	if globalRotating != nil {
		_ = globalRotating.close()
	}
	globalRotating = rw

	writers := []io.Writer{rw}
	if cfg.Console != nil {
		writers = append([]io.Writer{cfg.Console}, writers...)
	}

	Init(io.MultiWriter(writers...), cfg.Level)
	return nil
}

// InitDefault は標準エラーへ Info レベルでログを書き出す既定設定を適用する。
func InitDefault() {
	InitConsole(os.Stderr, slog.LevelInfo)
}
