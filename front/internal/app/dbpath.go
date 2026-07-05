package app

import (
	"fmt"
	"meguri-app/internal/sqlitedsn"
	"path/filepath"
)

// ResolveDBPath は SQLite ファイルの絶対パスを返す。
func ResolveDBPath() (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbFileName), nil
}

// SQLiteDSN は GORM 用 SQLite DSN を返す。
func SQLiteDSN(dbPath string) string {
	return sqlitedsn.DSN(dbPath)
}

// sqliteURL は golang-migrate 用の sqlite:// URL を返す。
func sqliteURL(dbPath string) (string, error) {
	abs, err := filepath.Abs(dbPath)
	if err != nil {
		return "", fmt.Errorf("abs db path: %w", err)
	}
	return "sqlite://" + filepath.ToSlash(abs), nil
}
