package logger

import "time"

// FlushPolicy はファイルログの flush タイミングを表す。
type FlushPolicy int

const (
	// FlushImmediate は毎 Write 後に Sync する。
	FlushImmediate FlushPolicy = iota
	// FlushInterval は Interval 経過ごとに Sync する。
	FlushInterval
	// FlushEveryN は EveryN 回の Write ごとに Sync する。
	FlushEveryN
)

// FlushConfig は rotatingWriter の flush 挙動を設定する。
type FlushConfig struct {
	// Policy は flush のタイミング方式。
	Policy FlushPolicy
	// Interval は FlushInterval 時の同期間隔。
	Interval time.Duration
	// EveryN は FlushEveryN 時の Write 回数閾値。
	EveryN int
}

// Flush は保留中のファイルログをディスクへ反映する。
func Flush() error {
	if globalRotating == nil {
		return nil
	}
	return globalRotating.flushNow()
}

// Shutdown はファイルログを flush してハンドルを閉じる。
func Shutdown() error {
	if globalRotating == nil {
		return nil
	}
	if err := globalRotating.flushNow(); err != nil {
		return err
	}
	err := globalRotating.close()
	globalRotating = nil
	return err
}
