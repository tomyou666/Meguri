package runner

import (
	"meguri/internal/core"
	"meguri/internal/domain/model"
)

// ProgressKind はクロール進捗イベントの種別（core.ProgressKind のエイリアス）。
type ProgressKind = core.ProgressKind

// Progress イベント種別定数。
const (
	ProgressStarted        = core.ProgressStarted
	ProgressSucceeded      = core.ProgressSucceeded
	ProgressFailed         = core.ProgressFailed
	ProgressSkipped        = core.ProgressSkipped
	ProgressLinkDiscovered = core.ProgressLinkDiscovered
	ProgressCompleted      = core.ProgressCompleted
	ProgressError          = core.ProgressError
)

// ProgressEvent は URL 単位またはジョブ単位の進捗通知。
type ProgressEvent = core.ProgressEvent

// ProgressSink は進捗イベントを受け取るコールバック。
type ProgressSink = core.ProgressSink

// CrawlStats はクロールの最終サマリ。
type CrawlStats = core.CrawlStats

// Config は backend 実行設定（model.Config のエイリアス）。
type Config = model.Config

// Result は 1 URL の取得結果（model.Result のエイリアス）。
type Result = model.Result
