package sqlitedsn

import (
	"path/filepath"
	"testing"

	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
)

// TestDSN は DSN 経由で接続した DB に期待 PRAGMA が適用されることを検証する。
func TestDSN(t *testing.T) {
	t.Run("正常系: WAL・synchronous(NORMAL)・busy_timeout が適用される", func(t *testing.T) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test.db")

		db, err := gorm.Open(sqlite.Open(DSN(dbPath)), &gorm.Config{})
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("db: %v", err)
		}
		t.Cleanup(func() { _ = sqlDB.Close() })

		var journalMode string
		if err := sqlDB.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
			t.Fatalf("journal_mode: %v", err)
		}
		if journalMode != "wal" {
			t.Fatalf("journal_mode = %q, want wal", journalMode)
		}

		var synchronous int
		if err := sqlDB.QueryRow("PRAGMA synchronous").Scan(&synchronous); err != nil {
			t.Fatalf("synchronous: %v", err)
		}
		if synchronous != 1 {
			t.Fatalf("synchronous = %d, want 1 (NORMAL)", synchronous)
		}

		var busyTimeout int
		if err := sqlDB.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
			t.Fatalf("busy_timeout: %v", err)
		}
		if busyTimeout != 5000 {
			t.Fatalf("busy_timeout = %d, want 5000", busyTimeout)
		}
	})
}
