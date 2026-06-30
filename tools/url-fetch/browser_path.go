package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// envBrowserPath はブラウザ実行ファイルを明示指定する環境変数名。
const envBrowserPath = "MEGURI_CHROMIUM_PATH"

var chromiumNames = []string{
	"chromium",
	"chromium-browser",
	"google-chrome",
	"google-chrome-stable",
	"chrome",
}

var edgeNames = []string{
	"microsoft-edge",
	"microsoft-edge-stable",
	"msedge",
}

var chromiumFixedPaths = []string{
	"/usr/bin/chromium",
	"/usr/bin/chromium-browser",
	"/snap/bin/chromium",
}

var edgeFixedPaths = []string{
	"/usr/bin/microsoft-edge",
	"/usr/bin/microsoft-edge-stable",
	"/opt/microsoft/msedge/msedge",
}

// resolveBrowserPath は使用するブラウザ実行ファイルのパスを解決する。
//
// 優先順位:
// 1) 環境変数 MEGURI_CHROMIUM_PATH
// 2) Chromium 系候補（PATH / 固定パス）
// 3) Edge 系候補（PATH / 固定パス）
func resolveBrowserPath() (string, error) {
	if p := strings.TrimSpace(os.Getenv(envBrowserPath)); p != "" {
		return validateExecutable(p)
	}
	if p, err := findFirstExecutable(chromiumNames, chromiumFixedPaths); err == nil {
		return p, nil
	}
	if p, err := findFirstExecutable(edgeNames, edgeFixedPaths); err == nil {
		return p, nil
	}
	return "", fmt.Errorf("browser not found (install Chromium or Edge, or set %s)", envBrowserPath)
}

// findFirstExecutable は names と fixed から最初に見つかった実行ファイルを返す。
func findFirstExecutable(names, fixed []string) (string, error) {
	for _, name := range names {
		if p, err := exec.LookPath(name); err == nil {
			if path, err := validateExecutable(p); err == nil {
				return path, nil
			}
		}
	}
	for _, p := range fixed {
		if path, err := validateExecutable(p); err == nil {
			return path, nil
		}
	}
	if runtime.GOOS == "windows" {
		for _, p := range windowsProgramFilesPaths(names) {
			if path, err := validateExecutable(p); err == nil {
				return path, nil
			}
		}
	}
	return "", fmt.Errorf("not found")
}

// isEdgeNameList は names が Edge 系かどうかを返す。
func isEdgeNameList(names []string) bool {
	for _, name := range names {
		switch name {
		case "msedge", "microsoft-edge", "microsoft-edge-stable":
			return true
		}
	}
	return false
}

// windowsProgramFilesPaths は Windows の Program Files 配下候補パスを返す。
func windowsProgramFilesPaths(names []string) []string {
	var out []string
	roots := []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")}
	for _, root := range roots {
		if root == "" {
			continue
		}
		if isEdgeNameList(names) {
			out = append(out, filepath.Join(root, "Microsoft", "Edge", "Application", "msedge.exe"))
		} else {
			out = append(out, filepath.Join(root, "Google", "Chrome", "Application", "chrome.exe"))
		}
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		for _, name := range names {
			out = append(out, filepath.Join(root, name, name+".exe"))
		}
	}
	return out
}

// validateExecutable は path が実行可能ファイルであることを検証する。
func validateExecutable(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("empty path")
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("browser path %q: %w", path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("browser path %q is a directory", path)
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return abs, nil
	}
	return abs, nil
}
