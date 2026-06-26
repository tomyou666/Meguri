//go:build dev

package app

import (
	"os"
	"path/filepath"
)

// logDir は開発用ログディレクトリ（リポジトリ logs/）を返す。
func logDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(wd, "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
