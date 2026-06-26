//go:build !dev

package app

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// logDir は本番用ログディレクトリ（アプリデータ logs/）を返す。
func logDir() (string, error) {
	dir := filepath.Join(xdg.DataHome, "meguri", "logs")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
