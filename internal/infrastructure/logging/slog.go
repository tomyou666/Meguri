package logging

import (
	"io"
	"log/slog"
	"os"

	"scraperbot/internal/domain/plugin"
)

// SlogAdapter は log/slog を plugin.Logger 抽象に適合させる。
type SlogAdapter struct {
	l *slog.Logger
}

// NewDefault は標準エラーへ JSON ロガーを書き出す既定実装を返す。
func NewDefault() *SlogAdapter {
	return New(os.Stderr, slog.LevelInfo)
}

// New は任意の Writer / Level でロガーを作る（テストで bytes.Buffer を渡せる）。
func New(w io.Writer, level slog.Level) *SlogAdapter {
	h := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	return &SlogAdapter{l: slog.New(h)}
}

func (s *SlogAdapter) Debug(msg string, kv ...any) { s.l.Debug(msg, kv...) }
func (s *SlogAdapter) Info(msg string, kv ...any)  { s.l.Info(msg, kv...) }
func (s *SlogAdapter) Warn(msg string, kv ...any)  { s.l.Warn(msg, kv...) }
func (s *SlogAdapter) Error(msg string, kv ...any) { s.l.Error(msg, kv...) }

// 静的に plugin.Logger を満たすことを保証。
var _ plugin.Logger = (*SlogAdapter)(nil)
