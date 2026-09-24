package app

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"meguri-app/internal/domain"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/libtnb/sqlite"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations は未適用マイグレーションを Up し、必要なら links_hash を埋める。
func RunMigrations(dbPath string) error {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return fmt.Errorf("mkdir data: %w", err)
	}

	url, err := sqliteURL(dbPath)
	if err != nil {
		return err
	}

	source, err := iofs.New(migrationFS, "migrations")
	if err != nil {
		return fmt.Errorf("migration source: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, url)
	if err != nil {
		return fmt.Errorf("migrate new: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate up: %w", err)
	}
	if err := backfillLinksHash(dbPath); err != nil {
		return fmt.Errorf("backfill links_hash: %w", err)
	}
	return nil
}

// backfillLinksHash は本文がある行の links_json から links_hash を埋める。
func backfillLinksHash(dbPath string) error {
	db, err := gorm.Open(sqlite.Open(SQLiteDSN(dbPath)), &gorm.Config{})
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	rows, err := sqlDB.Query(`
		SELECT nr.id, COALESCE(b.links_json, '')
		FROM node_results nr
		INNER JOIN node_result_bodies b ON b.id = nr.id
		WHERE nr.links_hash IS NULL
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pair struct {
		id   string
		hash string
	}
	var updates []pair
	for rows.Next() {
		var id, linksJSON string
		if err := rows.Scan(&id, &linksJSON); err != nil {
			return err
		}
		updates = append(updates, pair{id: id, hash: domain.LinksHashFromLinksJSON(linksJSON)})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if len(updates) == 0 {
		return nil
	}

	tx, err := sqlDB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`UPDATE node_results SET links_hash = ? WHERE id = ?`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()
	for _, u := range updates {
		if _, err := stmt.Exec(u.hash, u.id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
