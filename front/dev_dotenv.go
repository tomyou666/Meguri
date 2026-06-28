//go:build dev

package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	loadDevDotEnv()
}

// loadDevDotEnv は親ディレクトリを辿って .env を探し、未設定の環境変数だけを読み込む。
//
// wails3 dev 起動時は VS Code tasks の envFile が meguri プロセスまで届かないことがあるため、
// dev ビルドでは Go 側でも .env を読む。
func loadDevDotEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for {
		path := filepath.Join(dir, ".env")
		if _, err := os.Stat(path); err == nil {
			_ = applyDevEnvFile(path)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return
		}
		dir = parent
	}
}

func applyDevEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || os.Getenv(key) != "" {
			continue
		}
		value = strings.Trim(value, `"'`)
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}
