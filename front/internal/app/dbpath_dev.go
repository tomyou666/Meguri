//go:build dev

package app

import (
	"os"
	"path/filepath"
)

// dataDir は開発用 DB ディレクトリ（リポジトリ data/）を返す。
func dataDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(wd, "data")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
