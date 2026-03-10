package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"time"
)

// Migration defines a single schema migration.
type Migration struct {
	Version int
	Name    string
	Up      func(tx *sql.Tx) error
	Down    func(tx *sql.Tx) error
}

// Runner applies migrations to a database.
type Runner struct {
	db         *sql.DB
	migrations []Migration
}

// New creates a migration runner with the given migrations.
func New(db *sql.DB, migrations []Migration) *Runner {
	sorted := make([]Migration, len(migrations))
	copy(sorted, migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})
	return &Runner{db: db, migrations: sorted}
}

// Init creates the schema_migrations tracking table.
func (r *Runner) Init() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    INTEGER PRIMARY KEY,
			name       TEXT NOT NULL,
			applied_at DATETIME NOT NULL
		)
	`)
	return err
}

// Applied returns a set of already-applied migration versions.
func (r *Runner) Applied() (map[int]bool, error) {
	rows, err := r.db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// Up applies all pending migrations in order.
func (r *Runner) Up() error {
	if err := r.Init(); err != nil {
		return fmt.Errorf("init migrations table: %w", err)
	}

	applied, err := r.Applied()
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	for _, m := range r.migrations {
		if applied[m.Version] {
			continue
		}

		log.Printf("migrate: applying %03d_%s", m.Version, m.Name)

		tx, err := r.db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx for %03d_%s: %w", m.Version, m.Name, err)
		}

		if err := m.Up(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migrate %03d_%s up: %w", m.Version, m.Name, err)
		}

		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, name, applied_at) VALUES (?, ?, ?)",
			m.Version, m.Name, time.Now().UTC(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %03d_%s: %w", m.Version, m.Name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %03d_%s: %w", m.Version, m.Name, err)
		}

		log.Printf("migrate: applied %03d_%s", m.Version, m.Name)
	}

	return nil
}

// Rollback reverts the last applied migration.
func (r *Runner) Rollback() error {
	if err := r.Init(); err != nil {
		return err
	}

	var version int
	var name string
	err := r.db.QueryRow(
		"SELECT version, name FROM schema_migrations ORDER BY version DESC LIMIT 1",
	).Scan(&version, &name)
	if err == sql.ErrNoRows {
		log.Println("migrate: nothing to rollback")
		return nil
	}
	if err != nil {
		return err
	}

	// Find the migration
	var target *Migration
	for i := range r.migrations {
		if r.migrations[i].Version == version {
			target = &r.migrations[i]
			break
		}
	}
	if target == nil {
		return fmt.Errorf("migrate: migration %03d_%s not found in registry", version, name)
	}
	if target.Down == nil {
		return fmt.Errorf("migrate: %03d_%s has no Down function", version, name)
	}

	log.Printf("migrate: rolling back %03d_%s", version, name)

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if err := target.Down(tx); err != nil {
		tx.Rollback()
		return fmt.Errorf("migrate %03d_%s down: %w", version, name, err)
	}

	if _, err := tx.Exec("DELETE FROM schema_migrations WHERE version = ?", version); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("migrate: rolled back %03d_%s", version, name)
	return nil
}

// Status returns information about all migrations and their applied state.
type MigrationStatus struct {
	Version   int
	Name      string
	Applied   bool
	AppliedAt *time.Time
}

func (r *Runner) Status() ([]MigrationStatus, error) {
	if err := r.Init(); err != nil {
		return nil, err
	}

	rows, err := r.db.Query("SELECT version, name, applied_at FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	appliedMap := make(map[int]time.Time)
	for rows.Next() {
		var v int
		var n string
		var at time.Time
		if err := rows.Scan(&v, &n, &at); err != nil {
			return nil, err
		}
		appliedMap[v] = at
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var statuses []MigrationStatus
	for _, m := range r.migrations {
		s := MigrationStatus{Version: m.Version, Name: m.Name}
		if at, ok := appliedMap[m.Version]; ok {
			s.Applied = true
			s.AppliedAt = &at
		}
		statuses = append(statuses, s)
	}
	return statuses, nil
}
