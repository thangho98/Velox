package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

type LibraryRepo struct {
	db DBTX
}

func NewLibraryRepo(db DBTX) *LibraryRepo {
	return &LibraryRepo{db: db}
}

// populatePaths fetches library_paths for a slice of libraries and sets Paths on each.
func (r *LibraryRepo) populatePaths(ctx context.Context, libs []model.Library) error {
	if len(libs) == 0 {
		return nil
	}

	// Build a map from id → index for fast lookup
	idx := make(map[int64]int, len(libs))
	for i, l := range libs {
		idx[l.ID] = i
		libs[i].Paths = []string{} // ensure non-nil
	}

	rows, err := r.db.QueryContext(ctx,
		"SELECT library_id, path FROM library_paths ORDER BY id")
	if err != nil {
		return fmt.Errorf("querying library paths: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var libID int64
		var path string
		if err := rows.Scan(&libID, &path); err != nil {
			return fmt.Errorf("scanning library path: %w", err)
		}
		if i, ok := idx[libID]; ok {
			libs[i].Paths = append(libs[i].Paths, path)
		}
	}
	return rows.Err()
}

func (r *LibraryRepo) List(ctx context.Context) ([]model.Library, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, type, created_at FROM libraries ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var libs []model.Library
	for rows.Next() {
		var l model.Library
		if err := rows.Scan(&l.ID, &l.Name, &l.Type, &l.CreatedAt); err != nil {
			return nil, err
		}
		libs = append(libs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.populatePaths(ctx, libs); err != nil {
		return nil, err
	}
	return libs, nil
}

func (r *LibraryRepo) GetByID(ctx context.Context, id int64) (*model.Library, error) {
	var l model.Library
	err := r.db.QueryRowContext(ctx,
		"SELECT id, name, type, created_at FROM libraries WHERE id = ?", id).
		Scan(&l.ID, &l.Name, &l.Type, &l.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get library by id %d: %w", id, err)
	}

	libs := []model.Library{l}
	if err := r.populatePaths(ctx, libs); err != nil {
		return nil, fmt.Errorf("populate library paths: %w", err)
	}
	// populatePaths modifies libs in place; return the updated copy
	result := libs[0]
	return &result, nil
}

// Create inserts a new library with one or more root paths.
// The first path is also stored in libraries.path for backward compatibility.
func (r *LibraryRepo) Create(ctx context.Context, name, libType string, paths []string) (*model.Library, error) {
	if len(paths) == 0 {
		return nil, fmt.Errorf("at least one path is required")
	}

	res, err := r.db.ExecContext(ctx,
		"INSERT INTO libraries (name, path, type) VALUES (?, ?, ?)", name, paths[0], libType)
	if err != nil {
		return nil, err
	}
	libID, _ := res.LastInsertId()

	for _, p := range paths {
		if _, err := r.db.ExecContext(ctx,
			"INSERT INTO library_paths (library_id, path) VALUES (?, ?)", libID, p); err != nil {
			return nil, fmt.Errorf("inserting library path %q: %w", p, err)
		}
	}

	return r.GetByID(ctx, libID)
}

func (r *LibraryRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM libraries WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete library %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetStats returns per-library statistics
func (r *LibraryRepo) GetStats(ctx context.Context) ([]model.LibraryStats, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			l.id, l.name, l.type,
			(SELECT COUNT(*) FROM media m WHERE m.library_id = l.id) +
			(SELECT COUNT(*) FROM series sr WHERE sr.library_id = l.id) as item_count,
			(SELECT COUNT(*) FROM media_files mf
			 JOIN media m2 ON mf.media_id = m2.id WHERE m2.library_id = l.id) as file_count,
			(SELECT COALESCE(SUM(mf2.file_size), 0) FROM media_files mf2
			 JOIN media m3 ON mf2.media_id = m3.id WHERE m3.library_id = l.id) as total_size,
			COALESCE((SELECT MAX(sj.finished_at) FROM scan_jobs sj
			          WHERE sj.library_id = l.id AND sj.status = 'completed'), '') as last_scanned
		FROM libraries l
		ORDER BY l.name`)
	if err != nil {
		return nil, fmt.Errorf("querying library stats: %w", err)
	}
	defer rows.Close()

	var stats []model.LibraryStats
	for rows.Next() {
		var ls model.LibraryStats
		if err := rows.Scan(&ls.ID, &ls.Name, &ls.Type, &ls.ItemCount, &ls.FileCount, &ls.TotalSize, &ls.LastScanned); err != nil {
			return nil, fmt.Errorf("scanning library stats: %w", err)
		}
		stats = append(stats, ls)
	}
	return stats, rows.Err()
}
