package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupScheduledTaskDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE scheduled_tasks (
			name TEXT PRIMARY KEY,
			last_run DATETIME,
			interval_seconds INTEGER NOT NULL
		);
	`)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestScheduledTaskRepository_UpsertAndGet(t *testing.T) {
	db := setupScheduledTaskDB(t)
	defer db.Close()
	repo := NewScheduledTaskRepository(db)
	ctx := context.Background()

	now := time.Now().Truncate(time.Second)

	// Upsert a new task
	err := repo.Upsert(ctx, &ScheduledTask{
		Name:            "library-scan",
		LastRun:         &now,
		IntervalSeconds: 86400,
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// Get it back
	task, err := repo.GetByName(ctx, "library-scan")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if task == nil {
		t.Fatal("expected task, got nil")
	}
	if task.Name != "library-scan" {
		t.Errorf("expected name 'library-scan', got %q", task.Name)
	}
	if task.IntervalSeconds != 86400 {
		t.Errorf("expected interval 86400, got %d", task.IntervalSeconds)
	}
	if task.LastRun == nil {
		t.Fatal("expected last_run to be set")
	}

	// Upsert again (update)
	newTime := now.Add(1 * time.Hour)
	err = repo.Upsert(ctx, &ScheduledTask{
		Name:            "library-scan",
		LastRun:         &newTime,
		IntervalSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("Upsert (update) failed: %v", err)
	}

	task, err = repo.GetByName(ctx, "library-scan")
	if err != nil {
		t.Fatalf("GetByName after update failed: %v", err)
	}
	if task.IntervalSeconds != 3600 {
		t.Errorf("expected updated interval 3600, got %d", task.IntervalSeconds)
	}
}

func TestScheduledTaskRepository_GetByName_NotFound(t *testing.T) {
	db := setupScheduledTaskDB(t)
	defer db.Close()
	repo := NewScheduledTaskRepository(db)
	ctx := context.Background()

	task, err := repo.GetByName(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetByName should not error for missing: %v", err)
	}
	if task != nil {
		t.Errorf("expected nil, got %+v", task)
	}
}

func TestScheduledTaskRepository_UpdateLastRun(t *testing.T) {
	db := setupScheduledTaskDB(t)
	defer db.Close()
	repo := NewScheduledTaskRepository(db)
	ctx := context.Background()

	// Insert first
	err := repo.Upsert(ctx, &ScheduledTask{
		Name:            "session-cleanup",
		IntervalSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// UpdateLastRun
	now := time.Now().Truncate(time.Second)
	err = repo.UpdateLastRun(ctx, "session-cleanup", now)
	if err != nil {
		t.Fatalf("UpdateLastRun failed: %v", err)
	}

	task, err := repo.GetByName(ctx, "session-cleanup")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if task.LastRun == nil {
		t.Fatal("expected last_run to be set after update")
	}
}

func TestScheduledTaskRepository_UpdateInterval(t *testing.T) {
	db := setupScheduledTaskDB(t)
	defer db.Close()
	repo := NewScheduledTaskRepository(db)
	ctx := context.Background()

	// Insert first
	err := repo.Upsert(ctx, &ScheduledTask{
		Name:            "library-scan",
		IntervalSeconds: 86400,
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	// UpdateInterval
	err = repo.UpdateInterval(ctx, "library-scan", 43200)
	if err != nil {
		t.Fatalf("UpdateInterval failed: %v", err)
	}

	task, err := repo.GetByName(ctx, "library-scan")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if task.IntervalSeconds != 43200 {
		t.Errorf("expected 43200, got %d", task.IntervalSeconds)
	}
}

func TestScheduledTaskRepository_UpsertWithNilLastRun(t *testing.T) {
	db := setupScheduledTaskDB(t)
	defer db.Close()
	repo := NewScheduledTaskRepository(db)
	ctx := context.Background()

	// Upsert without LastRun
	err := repo.Upsert(ctx, &ScheduledTask{
		Name:            "transcode-cleanup",
		IntervalSeconds: 3600,
	})
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	task, err := repo.GetByName(ctx, "transcode-cleanup")
	if err != nil {
		t.Fatalf("GetByName failed: %v", err)
	}
	if task.LastRun != nil {
		t.Errorf("expected nil LastRun, got %v", task.LastRun)
	}
}
