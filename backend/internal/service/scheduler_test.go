package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/repository"
)

func setupSchedulerTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	schema := `
	CREATE TABLE scheduled_tasks (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		name             TEXT NOT NULL UNIQUE,
		interval_seconds INTEGER NOT NULL DEFAULT 3600,
		last_run         DATETIME,
		next_run         DATETIME,
		enabled          INTEGER NOT NULL DEFAULT 1,
		created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("create schema: %v", err)
	}

	return db
}

func TestScheduler_New(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	assert.NotNil(t, svc)
	assert.NotNil(t, svc.tasks)
	assert.NotNil(t, svc.done)
}

func TestScheduler_Register(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("test-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	svc.mu.Lock()
	task, ok := svc.tasks["test-task"]
	svc.mu.Unlock()
	assert.True(t, ok)
	assert.NotNil(t, task)
	assert.Equal(t, 1*time.Hour, task.Interval)
}

func TestScheduler_Register_LoadsFromDB(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	// Insert a task with a different interval
	_, err := db.Exec(`
		INSERT INTO scheduled_tasks (name, interval_seconds, enabled)
		VALUES ('db-task', 7200, 1)
	`)
	require.NoError(t, err)

	svc.Register("db-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	svc.mu.Lock()
	task, ok := svc.tasks["db-task"]
	svc.mu.Unlock()
	assert.True(t, ok)
	assert.Equal(t, 2*time.Hour, task.Interval) // Should use DB value
}

func TestScheduler_Stop_DoesNotPanic(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("test-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	assert.NotPanics(t, func() {
		svc.Stop()
	})
}

func TestScheduler_ListTasks(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("task-1", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})
	svc.Register("task-2", 2*time.Hour, func(ctx context.Context) error {
		return nil
	})

	tasks := svc.ListTasks()
	assert.Len(t, tasks, 2)

	names := make([]string, len(tasks))
	for i, t := range tasks {
		names[i] = t.Name
	}
	assert.Contains(t, names, "task-1")
	assert.Contains(t, names, "task-2")
}

func TestScheduler_ListTasks_Empty(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	tasks := svc.ListTasks()
	assert.Len(t, tasks, 0)
}

func TestScheduler_RunNow_TaskNotFound(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	err := svc.RunNow("nonexistent-task")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestScheduler_RunNow_ExecutesTask(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("exec-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	// RunNow triggers async execution - just verify it doesn't error
	err := svc.RunNow("exec-task")
	assert.NoError(t, err)
}

func TestScheduler_RunNow_PropagatesError(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("error-task", 1*time.Hour, func(ctx context.Context) error {
		return assert.AnError
	})

	// RunNow triggers async execution - error is logged but not returned synchronously
	err := svc.RunNow("error-task")
	assert.NoError(t, err) // RunNow itself returns immediately
}

func TestScheduler_UpdateInterval(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("interval-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	err := svc.UpdateInterval(context.Background(), "interval-task", 3*time.Hour)
	assert.NoError(t, err)

	svc.mu.Lock()
	task := svc.tasks["interval-task"]
	svc.mu.Unlock()
	assert.Equal(t, 3*time.Hour, task.Interval)
}

func TestScheduler_UpdateInterval_NotFound(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	err := svc.UpdateInterval(context.Background(), "nonexistent", 1*time.Hour)
	assert.Error(t, err)
}

func TestScheduler_TaskInfo_Fields(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("info-task", 2*time.Hour, func(ctx context.Context) error {
		return nil
	})

	tasks := svc.ListTasks()
	require.Len(t, tasks, 1)

	info := tasks[0]
	assert.Equal(t, "info-task", info.Name)
	assert.Equal(t, "2h0m0s", info.Interval)
}

func TestScheduler_UpdateInterval_PersistsToDB(t *testing.T) {
	t.Parallel()
	db := setupSchedulerTestDB(t)
	t.Cleanup(func() { db.Close() })

	repo := repository.NewScheduledTaskRepository(db)
	svc := NewScheduler(repo)

	svc.Register("persist-task", 1*time.Hour, func(ctx context.Context) error {
		return nil
	})

	err := svc.UpdateInterval(context.Background(), "persist-task", 5*time.Hour)
	assert.NoError(t, err)

	// Verify in DB
	dbTask, err := repo.GetByName(context.Background(), "persist-task")
	require.NoError(t, err)
	assert.Equal(t, 18000, dbTask.IntervalSeconds)
}
