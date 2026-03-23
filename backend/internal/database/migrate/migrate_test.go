package migrate

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func testMigrations() []Migration {
	return []Migration{
		{
			Version: 1,
			Name:    "create_users",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT NOT NULL)`)
				return err
			},
			Down: func(tx *sql.Tx) error {
				_, err := tx.Exec(`DROP TABLE users`)
				return err
			},
		},
		{
			Version: 2,
			Name:    "add_email",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec(`ALTER TABLE users ADD COLUMN email TEXT DEFAULT ''`)
				return err
			},
			Down: func(tx *sql.Tx) error {
				// SQLite doesn't support DROP COLUMN before 3.35.0, recreate table
				_, err := tx.Exec(`
					CREATE TABLE users_new (id INTEGER PRIMARY KEY, name TEXT NOT NULL);
					INSERT INTO users_new SELECT id, name FROM users;
					DROP TABLE users;
					ALTER TABLE users_new RENAME TO users;
				`)
				return err
			},
		},
		{
			Version: 3,
			Name:    "create_posts",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec(`CREATE TABLE posts (id INTEGER PRIMARY KEY, user_id INTEGER, title TEXT)`)
				return err
			},
			Down: func(tx *sql.Tx) error {
				_, err := tx.Exec(`DROP TABLE posts`)
				return err
			},
		},
	}
}

func TestUp_FreshDB(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, testMigrations())

	if err := runner.Up(); err != nil {
		t.Fatalf("Up() error: %v", err)
	}

	// Verify all 3 migrations applied
	applied, err := runner.Applied()
	if err != nil {
		t.Fatalf("Applied() error: %v", err)
	}
	if len(applied) != 3 {
		t.Errorf("expected 3 applied, got %d", len(applied))
	}

	// Verify tables exist
	for _, table := range []string{"users", "posts"} {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found: %v", table, err)
		}
	}

	// Verify email column exists on users
	_, err = db.Exec("INSERT INTO users (name, email) VALUES ('test', 'test@test.com')")
	if err != nil {
		t.Errorf("email column missing: %v", err)
	}
}

func TestUp_Idempotent(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, testMigrations())

	if err := runner.Up(); err != nil {
		t.Fatalf("first Up() error: %v", err)
	}

	// Running Up() again should be a no-op
	if err := runner.Up(); err != nil {
		t.Fatalf("second Up() error: %v", err)
	}

	applied, _ := runner.Applied()
	if len(applied) != 3 {
		t.Errorf("expected 3 applied after double-up, got %d", len(applied))
	}
}

func TestUp_Incremental(t *testing.T) {
	db := openTestDB(t)

	// Apply only first 2 migrations
	runner := New(db, testMigrations()[:2])
	if err := runner.Up(); err != nil {
		t.Fatalf("Up(2) error: %v", err)
	}

	applied, _ := runner.Applied()
	if len(applied) != 2 {
		t.Errorf("expected 2 applied, got %d", len(applied))
	}

	// Now apply all 3
	runner = New(db, testMigrations())
	if err := runner.Up(); err != nil {
		t.Fatalf("Up(3) error: %v", err)
	}

	applied, _ = runner.Applied()
	if len(applied) != 3 {
		t.Errorf("expected 3 applied, got %d", len(applied))
	}
}

func TestRollback(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, testMigrations())

	if err := runner.Up(); err != nil {
		t.Fatalf("Up() error: %v", err)
	}

	// Rollback last migration (posts table)
	if err := runner.Rollback(); err != nil {
		t.Fatalf("Rollback() error: %v", err)
	}

	applied, _ := runner.Applied()
	if len(applied) != 2 {
		t.Errorf("expected 2 applied after rollback, got %d", len(applied))
	}

	// posts table should be gone
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='posts'").Scan(&name)
	if err == nil {
		t.Error("posts table still exists after rollback")
	}

	// users table should still exist
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&name)
	if err != nil {
		t.Error("users table was dropped by rollback")
	}
}

func TestRollback_Empty(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, testMigrations())

	// Rollback on empty DB should not error
	if err := runner.Rollback(); err != nil {
		t.Fatalf("Rollback() on empty: %v", err)
	}
}

func TestStatus(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, testMigrations())

	// Apply first 2
	partial := New(db, testMigrations()[:2])
	if err := partial.Up(); err != nil {
		t.Fatalf("Up(2) error: %v", err)
	}

	statuses, err := runner.Status()
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}

	if len(statuses) != 3 {
		t.Fatalf("expected 3 statuses, got %d", len(statuses))
	}

	if !statuses[0].Applied || !statuses[1].Applied {
		t.Error("first 2 migrations should be applied")
	}
	if statuses[2].Applied {
		t.Error("third migration should be pending")
	}
}

func TestUp_TransactionRollbackOnError(t *testing.T) {
	db := openTestDB(t)

	badMigrations := []Migration{
		testMigrations()[0], // good
		{
			Version: 2,
			Name:    "bad_migration",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec("THIS IS NOT VALID SQL")
				return err
			},
			Down: func(tx *sql.Tx) error { return nil },
		},
	}

	runner := New(db, badMigrations)
	err := runner.Up()
	if err == nil {
		t.Fatal("expected error from bad migration")
	}

	// First migration should be applied, second should not
	applied, _ := runner.Applied()
	if !applied[1] {
		t.Error("migration 1 should be applied")
	}
	if applied[2] {
		t.Error("migration 2 should NOT be applied (it failed)")
	}
}

func TestMigrationOrder(t *testing.T) {
	db := openTestDB(t)

	// Provide migrations out of order
	unordered := []Migration{
		testMigrations()[2], // version 3
		testMigrations()[0], // version 1
		testMigrations()[1], // version 2
	}

	runner := New(db, unordered)
	if err := runner.Up(); err != nil {
		t.Fatalf("Up() error: %v", err)
	}

	// Should still apply in order 1, 2, 3
	statuses, _ := runner.Status()
	for i, s := range statuses {
		if s.Version != i+1 {
			t.Errorf("status[%d] version = %d, want %d", i, s.Version, i+1)
		}
	}
}

// ── Migration 021 idempotency ──────────────────────────────────────────────────
// These three tests lock in the three DB states that up021 must handle without
// error: fresh tables (columns missing), fully-patched tables (columns already
// present), and partially-patched tables (some columns present, some missing).

// minimalMediaSchema is the DDL for the media/series tables as they exist after
// migration 020, i.e. without the columns that 021 adds.
func setupMediaSeriesTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS media (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL DEFAULT ''
		);
		CREATE TABLE IF NOT EXISTS series (
			id    INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL DEFAULT ''
		);
	`)
	if err != nil {
		t.Fatalf("setup tables: %v", err)
	}
}

func hasColumn(t *testing.T, db *sql.DB, table, column string) bool {
	t.Helper()
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notNull, pk int
		var name, colType string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
}

// TestUp021_ColumnsAbsent: fresh tables with none of the 021 columns.
// up021 must add all three columns without error.
func TestUp021_ColumnsAbsent(t *testing.T) {
	db := openTestDB(t)
	setupMediaSeriesTables(t, db)

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := up021(tx); err != nil {
		tx.Rollback()
		t.Fatalf("up021 with no existing columns: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	for _, tc := range []struct{ table, col string }{
		{"media", "tagline"},
		{"media", "metadata_locked"},
		{"series", "metadata_locked"},
	} {
		if !hasColumn(t, db, tc.table, tc.col) {
			t.Errorf("column %s.%s missing after up021", tc.table, tc.col)
		}
	}
}

// TestUp021_ColumnsPresent: tables already have all three columns (manually patched).
// up021 must be a no-op — no error, columns still present.
func TestUp021_ColumnsPresent(t *testing.T) {
	db := openTestDB(t)
	setupMediaSeriesTables(t, db)

	// Simulate manual patch
	_, err := db.Exec(`
		ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT '';
		ALTER TABLE media ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
		ALTER TABLE series ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
	`)
	if err != nil {
		t.Fatalf("manual column add: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := up021(tx); err != nil {
		tx.Rollback()
		t.Fatalf("up021 with all columns already present: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	for _, tc := range []struct{ table, col string }{
		{"media", "tagline"},
		{"media", "metadata_locked"},
		{"series", "metadata_locked"},
	} {
		if !hasColumn(t, db, tc.table, tc.col) {
			t.Errorf("column %s.%s unexpectedly dropped", tc.table, tc.col)
		}
	}
}

// TestUp021_ColumnsPartial: only some columns present (partial manual patch).
// up021 must add only the missing ones without error.
func TestUp021_ColumnsPartial(t *testing.T) {
	db := openTestDB(t)
	setupMediaSeriesTables(t, db)

	// Add tagline manually but leave the two metadata_locked columns absent.
	if _, err := db.Exec(`ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT ''`); err != nil {
		t.Fatalf("partial manual patch: %v", err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := up021(tx); err != nil {
		tx.Rollback()
		t.Fatalf("up021 with partial columns: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	for _, tc := range []struct{ table, col string }{
		{"media", "tagline"},
		{"media", "metadata_locked"},
		{"series", "metadata_locked"},
	} {
		if !hasColumn(t, db, tc.table, tc.col) {
			t.Errorf("column %s.%s missing after up021 (partial start)", tc.table, tc.col)
		}
	}
}

// TestUp_Skipped021WithManualColumns: integration test for the exact production
// edge case — a DB has schema_migrations recording versions 22/23/24 but NOT 21,
// and the 021 columns were added manually. Runner.Up() must succeed end-to-end.
func TestUp_Skipped021WithManualColumns(t *testing.T) {
	db := openTestDB(t)

	// Build the migrations list that reflects the broken window: 1-20 and 22-25
	// present in All(), 21 now restored. We simulate with a minimal set.
	base := []Migration{
		{
			Version: 20,
			Name:    "base",
			Up: func(tx *sql.Tx) error {
				_, err := tx.Exec(`
					CREATE TABLE media  (id INTEGER PRIMARY KEY, title TEXT NOT NULL DEFAULT '');
					CREATE TABLE series (id INTEGER PRIMARY KEY, title TEXT NOT NULL DEFAULT '');
				`)
				return err
			},
			Down: func(tx *sql.Tx) error { return nil },
		},
		{
			Version: 21,
			Name:    "metadata_lock",
			Up:      up021,
			Down:    down021,
		},
		{
			Version: 22,
			Name:    "after_gap",
			Up:      func(tx *sql.Tx) error { return nil },
			Down:    func(tx *sql.Tx) error { return nil },
		},
	}

	// Apply version 20 only via the runner so schema_migrations is created.
	runner := New(db, base[:1])
	if err := runner.Up(); err != nil {
		t.Fatalf("apply v20: %v", err)
	}

	// Simulate the "broken window": manually record 22 as applied (not 21).
	if _, err := db.Exec(
		`INSERT INTO schema_migrations (version, name, applied_at) VALUES (22, 'after_gap', ?)`,
		time.Now(),
	); err != nil {
		t.Fatalf("fake schema_migrations entry: %v", err)
	}

	// Manually add the columns (as a user would to silence SQL errors).
	if _, err := db.Exec(`
		ALTER TABLE media ADD COLUMN tagline TEXT NOT NULL DEFAULT '';
		ALTER TABLE media ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
		ALTER TABLE series ADD COLUMN metadata_locked INTEGER NOT NULL DEFAULT 0;
	`); err != nil {
		t.Fatalf("manual column patch: %v", err)
	}

	// Now run All() equivalent (full base set). Runner must apply 21 and skip 22.
	runner = New(db, base)
	if err := runner.Up(); err != nil {
		t.Fatalf("Up() with skipped-021 + manual columns: %v", err)
	}

	applied, _ := runner.Applied()
	for _, v := range []int{20, 21, 22} {
		if !applied[v] {
			t.Errorf("version %d not recorded in schema_migrations", v)
		}
	}

	// Columns must still be present.
	for _, tc := range []struct{ table, col string }{
		{"media", "tagline"},
		{"media", "metadata_locked"},
		{"series", "metadata_locked"},
	} {
		if !hasColumn(t, db, tc.table, tc.col) {
			t.Errorf("column %s.%s missing after integration run", tc.table, tc.col)
		}
	}
}

func TestRealMigrations_FreshDB(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, All())

	if err := runner.Up(); err != nil {
		t.Fatalf("real migrations Up() error: %v", err)
	}

	// Verify core tables exist after all migrations
	coreTables := []string{
		"libraries", "media", "media_files",
		"series", "seasons", "episodes",
		"genres", "media_genres", "people", "credits",
		"scan_jobs", "subtitles", "audio_tracks",
	}
	for _, table := range coreTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s not found after real migration: %v", table, err)
		}
	}

	// Verify we can insert and query
	_, err := db.Exec("INSERT INTO libraries (name, path) VALUES ('Movies', '/movies')")
	if err != nil {
		t.Fatalf("insert library: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM libraries").Scan(&count)
	if count != 1 {
		t.Errorf("expected 1 library, got %d", count)
	}
}

func TestRealMigrations_Rollback(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, All())

	if err := runner.Up(); err != nil {
		t.Fatalf("Up() error: %v", err)
	}

	// Rollback migrations 009, 008, 007, 006, 005, and 004 to test rollback of 004
	for i := 0; i < 6; i++ {
		if err := runner.Rollback(); err != nil {
			t.Fatalf("Rollback() error at iteration %d: %v", i, err)
		}
	}

	// After rollback of 004 (genres_people), genres/people tables should be gone
	// but libraries/media should still exist
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='credits'").Scan(&name)
	if err == nil {
		t.Error("credits table still exists after rollback of 004")
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='genres'").Scan(&name)
	if err == nil {
		t.Error("genres table still exists after rollback of 004")
	}

	// Core tables from earlier migrations should still exist
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='libraries'").Scan(&name)
	if err != nil {
		t.Error("libraries table was dropped by rollback of 004 (should only drop 004 tables)")
	}

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='media'").Scan(&name)
	if err != nil {
		t.Error("media table was dropped by rollback of 004")
	}
}
