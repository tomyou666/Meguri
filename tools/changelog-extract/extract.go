package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

var versionRE = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

// repoRoot はリポジトリルート（go.work があるディレクトリ）を返す。
func repoRoot() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("caller path unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../..")), nil
}

// parseVersionArg は X.Y.Z 形式の版文字列を検証する。
func parseVersionArg(input string) (string, error) {
	if input == "" || strings.HasPrefix(input, "v") {
		return "", fmt.Errorf(`invalid version %q: use X.Y.Z without "v" prefix`, input)
	}
	if !versionRE.MatchString(input) {
		return "", fmt.Errorf(
			`invalid version %q: expected X.Y.Z (e.g. 1.0.0); pre-release suffixes are not supported`,
			input,
		)
	}
	return input, nil
}

// sectionHeadingRE は Keep a Changelog 形式の版見出し行にマッチする。
func sectionHeadingRE(version string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(version)
	return regexp.MustCompile(`^## \[` + escaped + `\](?:\s*-\s*\d{4}-\d{2}-\d{2})?\s*$`)
}

var nextSectionRE = regexp.MustCompile(`^## \[`)

// extractSection は CHANGELOG 本文から指定版の節本文を返す。
func extractSection(content, version string) (string, error) {
	heading := sectionHeadingRE(version)
	lines := strings.Split(content, "\n")
	start := -1
	for i, line := range lines {
		if heading.MatchString(line) {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return "", fmt.Errorf("changelog section for version %q not found", version)
	}
	end := len(lines)
	for i := start; i < len(lines); i++ {
		if nextSectionRE.MatchString(lines[i]) {
			end = i
			break
		}
	}
	body := strings.TrimRight(strings.Join(lines[start:end], "\n"), " \t")
	if body == "" {
		return "", fmt.Errorf("changelog section for version %q is empty", version)
	}
	return body + "\n", nil
}

// defaultChangelogPath はリポジトリルートの CHANGELOG.md パスを返す。
func defaultChangelogPath() (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "CHANGELOG.md"), nil
}
