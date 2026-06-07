package core

import "scraperbot/internal/domain/model"

// ProgressKind はクロール進捗イベントの種別。
type ProgressKind string

const (
	// ProgressStarted は URL の処理開始を表す。
	ProgressStarted ProgressKind = "started"
	// ProgressSucceeded は URL の処理成功を表す。
	ProgressSucceeded ProgressKind = "succeeded"
	// ProgressFailed は URL の処理失敗を表す。
	ProgressFailed ProgressKind = "failed"
	// ProgressSkipped は URL がスキップされたことを表す。
	ProgressSkipped ProgressKind = "skipped"
	// ProgressLinkDiscovered はリンク発見を表す。
	ProgressLinkDiscovered ProgressKind = "linkDiscovered"
	// ProgressCompleted はクロールジョブ完了を表す。
	ProgressCompleted ProgressKind = "completed"
	// ProgressError はジョブ全体の失敗を表す。
	ProgressError ProgressKind = "error"
)

// ProgressEvent は URL 単位またはジョブ単位の進捗通知。
type ProgressEvent struct {
	// Kind はイベント種別。
	Kind ProgressKind
	// URL は対象 URL（linkDiscovered では child、started/succeeded/failed/skipped で処理対象）。
	URL string
	// ParentURL は発見元または親 URL（started / linkDiscovered 用。ルートは空）。
	ParentURL string
	// Depth はシードからの深度。
	Depth int
	// Result は成功時の取得結果。
	Result *model.Result
	// Error は失敗時のエラー文言。
	Error string
	// SkipReason はスキップ理由（exclude_urls / robots / max_pages / duplicate / already_success 等）。
	SkipReason string
	// Stats は completed 時のサマリ。
	Stats *CrawlStats
}

// ProgressSink は進捗イベントを受け取るコールバック。
type ProgressSink func(ProgressEvent)

// emitProgress は sink が非 nil のときイベントを送る。
func emitProgress(sink ProgressSink, ev ProgressEvent) {
	if sink != nil {
		sink(ev)
	}
}

// EmitProgress は進捗イベントを送る（pkg/runner 等の外部パッケージ用）。
func EmitProgress(sink ProgressSink, ev ProgressEvent) {
	emitProgress(sink, ev)
}
