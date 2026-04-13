package repository

import (
	"context"
	"database/sql"
	"time"
)

type ScheduledTask struct {
	Name            string     `json:"name"`
	LastRun         *time.Time `json:"last_run"`
	IntervalSeconds int        `json:"interval_seconds"`
}

type ScheduledTaskRepository struct {
	db *sql.DB
}

func NewScheduledTaskRepository(db *sql.DB) *ScheduledTaskRepository {
	return &ScheduledTaskRepository{db: db}
}

func (r *ScheduledTaskRepository) GetByName(ctx context.Context, name string) (*ScheduledTask, error) {
	var task ScheduledTask
	var lastRun sql.NullTime

	err := r.db.QueryRowContext(ctx, "SELECT name, last_run, interval_seconds FROM scheduled_tasks WHERE name = ?", name).
		Scan(&task.Name, &lastRun, &task.IntervalSeconds)

	if err == sql.ErrNoRows {
		return nil, nil // Not found is fine, we use default
	}
	if err != nil {
		return nil, err
	}

	if lastRun.Valid {
		task.LastRun = &lastRun.Time
	}

	return &task, nil
}

func (r *ScheduledTaskRepository) Upsert(ctx context.Context, task *ScheduledTask) error {
	var lastRun any
	if task.LastRun != nil {
		lastRun = *task.LastRun
	} else {
		lastRun = nil
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO scheduled_tasks (name, last_run, interval_seconds)
		VALUES (?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			last_run = excluded.last_run,
			interval_seconds = excluded.interval_seconds
	`, task.Name, lastRun, task.IntervalSeconds)

	return err
}

func (r *ScheduledTaskRepository) UpdateLastRun(ctx context.Context, name string, lastRun time.Time) error {
	_, err := r.db.ExecContext(ctx, "UPDATE scheduled_tasks SET last_run = ? WHERE name = ?", lastRun, name)
	return err
}

func (r *ScheduledTaskRepository) UpdateInterval(ctx context.Context, name string, intervalSeconds int) error {
	_, err := r.db.ExecContext(ctx, "UPDATE scheduled_tasks SET interval_seconds = ? WHERE name = ?", intervalSeconds, name)
	// If the row doesn't exist yet, we should insert it right away to persist the user's choice.
	if err == nil {
		// Just ensure it exists
		var exists int
		_ = r.db.QueryRowContext(ctx, "SELECT 1 FROM scheduled_tasks WHERE name = ?", name).Scan(&exists)
		if exists == 0 {
			_, err = r.db.ExecContext(ctx, "INSERT INTO scheduled_tasks (name, interval_seconds) VALUES (?, ?)", name, intervalSeconds)
		}
	}
	return err
}
