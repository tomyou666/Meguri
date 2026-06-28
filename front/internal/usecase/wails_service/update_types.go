package wails_service

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

const (
	topicUpdateAvailable = "meguri:update:available"

	githubMetadataReleaseHTMLURL = "github.release.htmlURL"

	updateStatusUpToDate    = "up_to_date"
	updateStatusAvailable   = "available"
	updateStatusReady       = "ready"
	updateStatusUnavailable = "unavailable"

	promptActionConfirmed   = "confirmed"
	promptActionOpenRelease = "open_release"
	promptActionDismissed   = "dismissed"

	// PromptActionConfirmed は更新適用の確認。
	PromptActionConfirmed = promptActionConfirmed
	// PromptActionOpenRelease はリリースページを開く選択。
	PromptActionOpenRelease = promptActionOpenRelease
	// PromptActionDismissed は更新確認を後回しにする選択。
	PromptActionDismissed = promptActionDismissed
)

const defaultGitHubRepository = "tomyou666/Meguri"

// UpdateAvailableEvent は起動時更新通知のイベント payload。
type UpdateAvailableEvent struct {
	// Version は利用可能なリリース版。
	Version string `json:"version"`
	// ReleaseURL は GitHub リリースページ URL。
	ReleaseURL string `json:"releaseURL"`
}

// UpdateStatus はフロント向けの更新状態。
type UpdateStatus struct {
	// Status は up_to_date / available / ready / unavailable。
	Status string `json:"status"`
	// Version は対象リリース版（available / ready のとき）。
	Version string `json:"version,omitempty"`
	// ReleaseURL は GitHub リリースページ URL。
	ReleaseURL string `json:"releaseURL,omitempty"`
}

// UpdatePromptResult はネイティブ更新確認ダイアログの結果。
type UpdatePromptResult struct {
	// Action は confirmed / open_release / dismissed。
	Action string `json:"action"`
	// Version は対象リリース版。
	Version string `json:"version,omitempty"`
	// ReleaseURL は GitHub リリースページ URL。
	ReleaseURL string `json:"releaseURL,omitempty"`
}

// CheckForUpdatesResult は手動更新確認の結果。
type CheckForUpdatesResult struct {
	// Status は up_to_date / available（ダイアログ表示後）など。
	Status string `json:"status"`
	// Action は PromptUpdate の結果（confirmed / open_release / dismissed）。
	Action string `json:"action,omitempty"`
	// Version は対象リリース版。
	Version string `json:"version,omitempty"`
	// ReleaseURL は GitHub リリースページ URL。
	ReleaseURL string `json:"releaseURL,omitempty"`
}

// releaseURLFrom は Release から GitHub リリースページ URL を得る。
func releaseURLFrom(rel *updater.Release) string {
	if rel == nil {
		return ""
	}
	if rel.Metadata != nil {
		if raw, ok := rel.Metadata[githubMetadataReleaseHTMLURL]; ok {
			if s, ok := raw.(string); ok && s != "" {
				return s
			}
		}
	}
	if rel.Version == "" {
		return ""
	}
	return fmt.Sprintf("https://github.com/%s/releases/tag/v%s", defaultGitHubRepository, rel.Version)
}
