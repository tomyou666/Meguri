package main

import (
	"strings"
	"testing"
)

// TestParseVersion は X.Y.Z 専用の版パースと拒否条件を検証する。
func TestParseVersion(t *testing.T) {
	t.Run("accepts X.Y.Z", func(t *testing.T) {
		got, err := parseVersion("1.0.0")
		if err != nil {
			t.Fatalf("parseVersion: %v", err)
		}
		if got.AppVersion != "1.0.0" || got.FileVersion != "1.0.0.0" {
			t.Fatalf("got %+v", got)
		}
	})

	t.Run("rejects v prefix and pre-release suffix", func(t *testing.T) {
		for _, input := range []string{"", "v1.0.0", "1.0.0-beta", "abc"} {
			if _, err := parseVersion(input); err == nil {
				t.Fatalf("parseVersion(%q) expected error", input)
			}
		}
	})
}

// TestUpdateConfigYml は info.version のみを更新し先頭の schema version は触らない。
func TestUpdateConfigYml(t *testing.T) {
	input := "version: '3'\n\ninfo:\n  version: \"0.0.1\"\n"
	got, err := updateConfigYml(input, "9.9.9")
	if err != nil {
		t.Fatalf("updateConfigYml: %v", err)
	}
	if !strings.Contains(got, `version: '3'`) {
		t.Fatalf("schema version changed: %q", got)
	}
	if !strings.Contains(got, `  version: "9.9.9"`) {
		t.Fatalf("info.version not updated: %q", got)
	}
}

// TestUpdateMessagesTS は messages.ts の version 行を単一引用符付きで置換する。
func TestUpdateMessagesTS(t *testing.T) {
	input := "export const messages = {\n\tversion: '0.0.1',\n}\n"

	t.Run("updates version", func(t *testing.T) {
		got, err := updateMessagesTS(input, "9.9.9")
		if err != nil {
			t.Fatalf("updateMessagesTS: %v", err)
		}
		if !strings.Contains(got, "version: '9.9.9'") {
			t.Fatalf("version not updated: %q", got)
		}
	})

	t.Run("same version is no-op", func(t *testing.T) {
		got, err := updateMessagesTS(input, "0.0.1")
		if err != nil {
			t.Fatalf("updateMessagesTS: %v", err)
		}
		if got != input {
			t.Fatalf("expected unchanged content, got %q", got)
		}
	})
}

// TestUpdateInfoJSON は ProductVersion と file_version を同期する。
func TestUpdateInfoJSON(t *testing.T) {
	input := "{\n\t\"fixed\": {\n\t\t\"file_version\": \"0.0.1.0\"\n\t},\n\t\"info\": {\n\t\t\"0000\": {\n\t\t\t\"ProductVersion\": \"0.0.1\"\n\t\t}\n\t}\n}\n"
	got, err := updateInfoJSON(input, "9.9.9", "9.9.9.0")
	if err != nil {
		t.Fatalf("updateInfoJSON: %v", err)
	}
	if !strings.Contains(got, `"ProductVersion": "9.9.9"`) {
		t.Fatalf("ProductVersion not updated: %q", got)
	}
	if !strings.Contains(got, `"file_version": "9.9.9.0"`) {
		t.Fatalf("file_version not updated: %q", got)
	}
}
