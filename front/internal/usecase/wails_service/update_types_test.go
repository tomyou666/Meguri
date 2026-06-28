package wails_service

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/updater"
)

func TestReleaseURLFrom_metadata(t *testing.T) {
	rel := &updater.Release{
		Version: "1.2.3",
		Metadata: map[string]any{
			githubMetadataReleaseHTMLURL: "https://github.com/tomyou666/Meguri/releases/tag/v1.2.3",
		},
	}
	if got := releaseURLFrom(rel); got != "https://github.com/tomyou666/Meguri/releases/tag/v1.2.3" {
		t.Fatalf("releaseURLFrom() = %q", got)
	}
}

func TestReleaseURLFrom_fallback(t *testing.T) {
	rel := &updater.Release{Version: "2.0.0"}
	want := "https://github.com/tomyou666/Meguri/releases/tag/v2.0.0"
	if got := releaseURLFrom(rel); got != want {
		t.Fatalf("releaseURLFrom() = %q, want %q", got, want)
	}
}
