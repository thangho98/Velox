package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// ScanJobRepo handles scan_jobs database operations
type ScanJobRepo struct {
	db DBTX
}

func NewScanJobRepo(db DBTX) *ScanJobRepo {
	return &ScanJobRepo{db: db}
}

// Create inserts a new scan job
func (r *ScanJobRepo) Create(ctx context.Context, job *model.ScanJob) error {
	query := `INSERT INTO scan_jobs (library_id, status) VALUES (?, ?) RETURNING id, created_at`
	row := r.db.QueryRowContext(ctx, query, job.LibraryID, job.Status)
	return row.Scan(&job.ID, &job.CreatedAt)
}

// GetByID retrieves a scan job by ID
func (r *ScanJobRepo) GetByID(ctx context.Context, id int64) (*model.ScanJob, error) {
	var job model.ScanJob
	var startedAt, finishedAt sql.NullString

	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, status, total_files,
		scanned_files, new_files, errors, error_log, started_at, finished_at, created_at
		FROM scan_jobs WHERE id = ?`, id).
		Scan(&job.ID, &job.LibraryID, &job.Status, &job.TotalFiles,
			&job.ScannedFiles, &job.NewFiles, &job.Errors, &job.ErrorLog,
			&startedAt, &finishedAt, &job.CreatedAt)
	if err != nil {
		return nil, err
	}
	if startedAt.Valid {
		job.StartedAt = &startedAt.String
	}
	if finishedAt.Valid {
		job.FinishedAt = &finishedAt.String
	}
	return &job, nil
}

// UpdateStatus updates job status
func (r *ScanJobRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE scan_jobs SET status = ? WHERE id = ?", status, id)
	return err
}

// Start marks job as started
func (r *ScanJobRepo) Start(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET status = 'scanning', started_at = CURRENT_TIMESTAMP WHERE id = ?",
		id)
	return err
}

// Complete marks job as completed
func (r *ScanJobRepo) Complete(ctx context.Context, id int64, totalFiles, newFiles, errors int, errorLog string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE scan_jobs SET
		status = 'completed', total_files = ?, scanned_files = ?, new_files = ?,
		errors = ?, error_log = ?, finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		totalFiles, totalFiles, newFiles, errors, errorLog, id)
	return err
}

// Fail marks job as failed
func (r *ScanJobRepo) Fail(ctx context.Context, id int64, errorLog string) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET status = 'failed', error_log = ?, finished_at = CURRENT_TIMESTAMP WHERE id = ?",
		errorLog, id)
	return err
}

// UpdateProgress updates scanned files count
func (r *ScanJobRepo) UpdateProgress(ctx context.Context, id int64, scannedFiles int) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET scanned_files = ? WHERE id = ?",
		scannedFiles, id)
	return err
}

// IncrementNewFiles increments new files count
func (r *ScanJobRepo) IncrementNewFiles(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE scan_jobs SET new_files = new_files + 1 WHERE id = ?", id)
	return err
}

// IncrementErrors increments error count and appends to log
func (r *ScanJobRepo) IncrementErrors(ctx context.Context, id int64, errorMsg string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE scan_jobs SET
		errors = errors + 1,
		error_log = CASE WHEN error_log = '' THEN ? ELSE error_log || '\n' || ? END
		WHERE id = ?`, errorMsg, errorMsg, id)
	return err
}

// ListByLibrary retrieves scan jobs for a library
func (r *ScanJobRepo) ListByLibrary(ctx context.Context, libraryID int64, limit int) ([]model.ScanJob, error) {
	if limit <= 0 {
		limit = 10
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, library_id, status, total_files,
		scanned_files, new_files, errors, error_log, started_at, finished_at, created_at
		FROM scan_jobs WHERE library_id = ? ORDER BY created_at DESC LIMIT ?`,
		libraryID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing scan jobs: %w", err)
	}
	defer rows.Close()

	var jobs []model.ScanJob
	for rows.Next() {
		var job model.ScanJob
		var startedAt, finishedAt sql.NullString

		if err := rows.Scan(&job.ID, &job.LibraryID, &job.Status, &job.TotalFiles,
			&job.ScannedFiles, &job.NewFiles, &job.Errors, &job.ErrorLog,
			&startedAt, &finishedAt, &job.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.String
		}
		if finishedAt.Valid {
			job.FinishedAt = &finishedAt.String
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// ListRecent retrieves recent scan jobs
func (r *ScanJobRepo) ListRecent(ctx context.Context, limit int) ([]model.ScanJob, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, library_id, status, total_files,
		scanned_files, new_files, errors, error_log, started_at, finished_at, created_at
		FROM scan_jobs ORDER BY created_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("listing scan jobs: %w", err)
	}
	defer rows.Close()

	var jobs []model.ScanJob
	for rows.Next() {
		var job model.ScanJob
		var startedAt, finishedAt sql.NullString

		if err := rows.Scan(&job.ID, &job.LibraryID, &job.Status, &job.TotalFiles,
			&job.ScannedFiles, &job.NewFiles, &job.Errors, &job.ErrorLog,
			&startedAt, &finishedAt, &job.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}
		if startedAt.Valid {
			job.StartedAt = &startedAt.String
		}
		if finishedAt.Valid {
			job.FinishedAt = &finishedAt.String
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// Delete removes a scan job
func (r *ScanJobRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM scan_jobs WHERE id = ?", id)
	return err
}

// DeleteOld removes completed/failed jobs older than given days
func (r *ScanJobRepo) DeleteOld(ctx context.Context, days int) error {
	if days <= 0 {
		return nil
	}
	query := fmt.Sprintf("DELETE FROM scan_jobs WHERE status IN ('completed', 'failed') AND created_at < datetime('now', '-%d days')", days)
	_, err := r.db.ExecContext(ctx, query)
	return err
}
