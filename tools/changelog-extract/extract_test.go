package main

import (
	"strings"
	"testing"
)

// TestExtractSection は CHANGELOG から版節を抽出する条件を検証する。
func TestExtractSection(t *testing.T) {
	sample := `# 変更履歴

## [Unreleased]

### 追加
- 未リリース

## [1.0.0] - 2026-01-15

### 追加
- 初回リリース

### 修正
- バグ修正

## [0.9.0]

### 変更
- ベータ版
`

	t.Run("日付付き見出しから節を抽出する", func(t *testing.T) {
		got, err := extractSection(sample, "1.0.0")
		if err != nil {
			t.Fatalf("extractSection: %v", err)
		}
		if !strings.Contains(got, "### 追加") || !strings.Contains(got, "初回リリース") {
			t.Fatalf("unexpected section: %q", got)
		}
		if strings.Contains(got, "0.9.0") {
			t.Fatalf("included next section: %q", got)
		}
	})

	t.Run("日付なし見出しから節を抽出する", func(t *testing.T) {
		got, err := extractSection(sample, "0.9.0")
		if err != nil {
			t.Fatalf("extractSection: %v", err)
		}
		if !strings.Contains(got, "ベータ版") {
			t.Fatalf("unexpected section: %q", got)
		}
	})

	t.Run("存在しない版はエラー", func(t *testing.T) {
		if _, err := extractSection(sample, "9.9.9"); err == nil {
			t.Fatal("expected error for missing version")
		}
	})

	t.Run("Unreleased は版引数として拒否する", func(t *testing.T) {
		if _, err := parseVersionArg("Unreleased"); err == nil {
			t.Fatal("expected error for Unreleased pseudo-version")
		}
	})

	t.Run("空節はエラー", func(t *testing.T) {
		content := "## [2.0.0] - 2026-06-01\n\n## [1.0.0]\n\n### 追加\n- item\n"
		if _, err := extractSection(content, "2.0.0"); err == nil {
			t.Fatal("expected error for empty section")
		}
	})
}

// TestParseVersionArg は版引数の検証を行う。
func TestParseVersionArg(t *testing.T) {
	t.Run("X.Y.Z を受け付ける", func(t *testing.T) {
		got, err := parseVersionArg("1.2.3")
		if err != nil || got != "1.2.3" {
			t.Fatalf("parseVersionArg: got %q err %v", got, err)
		}
	})

	t.Run("v プレフィックスと pre-release は拒否", func(t *testing.T) {
		for _, input := range []string{"", "v1.0.0", "1.0.0-beta", "abc"} {
			if _, err := parseVersionArg(input); err == nil {
				t.Fatalf("parseVersionArg(%q) expected error", input)
			}
		}
	})
}
