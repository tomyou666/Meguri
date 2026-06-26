//go:build !dev

package app

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// dataDir は本番用アプリデータディレクトリを返す。
func dataDir() (string, error) {
	dir := filepath.Join(xdg.DataHome, "meguri")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
