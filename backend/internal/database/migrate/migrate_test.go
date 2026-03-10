package migrate

import (
	"database/sql"
	"testing"

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

func TestRealMigrations_FreshDB(t *testing.T) {
	db := openTestDB(t)
	runner := New(db, All())

	if err := runner.Up(); err != nil {
		t.Fatalf("real migrations Up() error: %v", err)
	}

	// Verify core tables exist
	for _, table := range []string{"libraries", "media", "progress"} {
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

	if err := runner.Rollback(); err != nil {
		t.Fatalf("Rollback() error: %v", err)
	}

	// After rollback of 001, tables should be gone
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='libraries'").Scan(&name)
	if err == nil {
		t.Error("libraries table still exists after rollback")
	}
}
