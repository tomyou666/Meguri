package cli

import (
	"encoding/json"
	"io"

	"meguri/internal/core"
)

// progressJSONLine は NDJSON 1 行分の進捗イベント。
type progressJSONLine struct {
	Kind       string `json:"kind"`
	URL        string `json:"url,omitempty"`
	ParentURL  string `json:"parentUrl,omitempty"`
	Depth      int    `json:"depth,omitempty"`
	Error      string `json:"error,omitempty"`
	SkipReason string `json:"skipReason,omitempty"`
}

// newProgressJSONSink は stderr 等へ Progress を NDJSON 出力する Sink を返す。
func newProgressJSONSink(w io.Writer) core.ProgressSink {
	enc := json.NewEncoder(w)
	return func(ev core.ProgressEvent) {
		_ = enc.Encode(progressJSONLine{
			Kind:       string(ev.Kind),
			URL:        ev.URL,
			ParentURL:  ev.ParentURL,
			Depth:      ev.Depth,
			Error:      ev.Error,
			SkipReason: ev.SkipReason,
		})
	}
}
