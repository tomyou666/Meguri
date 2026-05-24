// Package logger はアプリ全体の slog 標準ロガーを初期化する。
package logger

import (
	"io"
	"log/slog"
	"os"
)

// Init は指定 Writer / Level で slog 標準ロガーを構築し SetDefault する。
func Init(w io.Writer, level slog.Level) {
	h := NewHandler(w, level)
	slog.SetDefault(slog.New(h))
}

// InitDefault は標準エラーへ Info レベルでログを書き出す既定設定を適用する。
func InitDefault() {
	Init(os.Stderr, slog.LevelInfo)
}

// NewHandler は TextHandler ベースの slog.Handler を返す。
func NewHandler(w io.Writer, level slog.Level) slog.Handler {
	return slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
}
