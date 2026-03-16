package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// MediaMarkerRepo handles media_markers database operations
type MediaMarkerRepo struct {
	db DBTX
}

func NewMediaMarkerRepo(db DBTX) *MediaMarkerRepo {
	return &MediaMarkerRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction
func (r *MediaMarkerRepo) WithTx(tx *sql.Tx) *MediaMarkerRepo {
	return &MediaMarkerRepo{db: tx}
}

// Create inserts a new media marker
func (r *MediaMarkerRepo) Create(ctx context.Context, m *model.MediaMarker) error {
	query := `INSERT INTO media_markers
		(media_file_id, marker_type, start_sec, end_sec, source, confidence, label)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		m.MediaFileID, m.MarkerType, m.StartSec, m.EndSec,
		m.Source, m.Confidence, m.Label)

	return row.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// GetByID retrieves a marker by ID
func (r *MediaMarkerRepo) GetByID(ctx context.Context, id int64) (*model.MediaMarker, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, media_file_id, marker_type, start_sec, end_sec, source, confidence, label, created_at, updated_at
		FROM media_markers WHERE id = ?`, id)
	return scanMarker(row)
}

// GetByMediaFileID retrieves all markers for a media file
func (r *MediaMarkerRepo) GetByMediaFileID(ctx context.Context, fileID int64) ([]model.MediaMarker, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, media_file_id, marker_type, start_sec, end_sec, source, confidence, label, created_at, updated_at
		FROM media_markers WHERE media_file_id = ? ORDER BY start_sec ASC`, fileID)
	if err != nil {
		return nil, fmt.Errorf("querying markers: %w", err)
	}
	defer rows.Close()

	return scanMarkers(rows)
}

// GetByType retrieves markers of a specific type for a media file
func (r *MediaMarkerRepo) GetByType(ctx context.Context, fileID int64, markerType string) ([]model.MediaMarker, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, media_file_id, marker_type, start_sec, end_sec, source, confidence, label, created_at, updated_at
		FROM media_markers WHERE media_file_id = ? AND marker_type = ? ORDER BY start_sec ASC`,
		fileID, markerType)
	if err != nil {
		return nil, fmt.Errorf("querying markers by type: %w", err)
	}
	defer rows.Close()

	return scanMarkers(rows)
}

// GetBestByType retrieves the best marker of a specific type for a media file
// Priority: manual > chapter > fingerprint
func (r *MediaMarkerRepo) GetBestByType(ctx context.Context, fileID int64, markerType string) (*model.MediaMarker, error) {
	// Source priority: manual (3) > chapter (2) > fingerprint (1)
	row := r.db.QueryRowContext(ctx,
		`SELECT id, media_file_id, marker_type, start_sec, end_sec, source, confidence, label, created_at, updated_at
		FROM media_markers
		WHERE media_file_id = ? AND marker_type = ?
		ORDER BY
			CASE source
				WHEN 'manual' THEN 3
				WHEN 'chapter' THEN 2
				WHEN 'fingerprint' THEN 1
				ELSE 0
			END DESC,
			confidence DESC,
			start_sec ASC
		LIMIT 1`,
		fileID, markerType)

	m, err := scanMarker(row)
	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}
	return m, err
}

// Delete removes a marker
func (r *MediaMarkerRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_markers WHERE id = ?", id)
	return err
}

// DeleteByMediaFileID removes all markers for a media file
func (r *MediaMarkerRepo) DeleteByMediaFileID(ctx context.Context, fileID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_markers WHERE media_file_id = ?", fileID)
	return err
}

// DeleteBySource removes all markers from a specific source for a media file
// Used when rebuilding markers from a particular source (e.g., re-parsing chapters)
func (r *MediaMarkerRepo) DeleteBySource(ctx context.Context, fileID int64, source string) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM media_markers WHERE media_file_id = ? AND source = ?",
		fileID, source)
	return err
}

// scanMarker scans a single marker row
func scanMarker(scanner interface{ Scan(...any) error }) (*model.MediaMarker, error) {
	var m model.MediaMarker
	err := scanner.Scan(&m.ID, &m.MediaFileID, &m.MarkerType, &m.StartSec, &m.EndSec,
		&m.Source, &m.Confidence, &m.Label, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// scanMarkers scans multiple marker rows
type rowsScanner interface {
	Next() bool
	Scan(...any) error
	Err() error
}

func scanMarkers(rows rowsScanner) ([]model.MediaMarker, error) {
	var markers []model.MediaMarker
	for rows.Next() {
		m, err := scanMarker(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning marker: %w", err)
		}
		markers = append(markers, *m)
	}
	return markers, rows.Err()
}
