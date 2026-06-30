package main

import (
	"fmt"
	"io"
	"time"
)

// trialResult は1回の取得試行のメタ情報を保持する。
type trialResult struct {
	// method は取得方式（http / chromium）。
	method string
	// variant は試行バリアント ID。
	variant string
	// statusCode は HTTP ステータス。取得失敗時は 0。
	statusCode int
	// bodyBytes はレスポンス本文のバイト数。
	bodyBytes int
	// duration は試行にかかった時間。
	duration time.Duration
	// err は試行エラー。成功時は nil。
	err error
}

// writeTrialResults は試行結果を定義順に w へ書き出す。
func writeTrialResults(w io.Writer, results []trialResult) error {
	for _, r := range results {
		if err := writeTrialResult(w, r); err != nil {
			return err
		}
	}
	return nil
}

// writeTrialResult は試行結果を1行で w に書き出す。
func writeTrialResult(w io.Writer, r trialResult) error {
	line := fmt.Sprintf(
		"method=%s variant=%s status=%d bytes=%d duration=%s",
		r.method,
		r.variant,
		r.statusCode,
		r.bodyBytes,
		formatDuration(r.duration),
	)
	if r.err != nil {
		line += fmt.Sprintf(" error=%q", r.err.Error())
	}
	line += "\n"
	_, err := io.WriteString(w, line)
	return err
}

// formatDuration は秒単位の小数3桁で duration を整形する。
func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.3fs", d.Seconds())
}
