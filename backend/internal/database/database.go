package database

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/thawng/velox/internal/database/migrate"

	_ "github.com/mattn/go-sqlite3"
)

func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1) // SQLite doesn't handle concurrent writes well

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate applies all pending migrations.
func Migrate(db *sql.DB) error {
	runner := migrate.New(db, migrate.All())
	return runner.Up()
}

// MigrateRollback reverts the last applied migration.
func MigrateRollback(db *sql.DB) error {
	runner := migrate.New(db, migrate.All())
	return runner.Rollback()
}

// MigrateStatus returns the status of all migrations.
func MigrateStatus(db *sql.DB) ([]migrate.MigrationStatus, error) {
	runner := migrate.New(db, migrate.All())
	return runner.Status()
}
