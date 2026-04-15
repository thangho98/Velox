package repository

import (
	"context"
	"database/sql"
	"strings"

	"github.com/thawng/velox/internal/model"
)

type ImageMetadataRepo struct {
	db DBTX
}

func NewImageMetadataRepo(db DBTX) *ImageMetadataRepo {
	return &ImageMetadataRepo{db: db}
}

func (r *ImageMetadataRepo) Get(ctx context.Context, path string) (*model.ImageMetadata, error) {
	var m model.ImageMetadata
	err := r.db.QueryRowContext(ctx, `
		SELECT path, blurhash, width, height, computed_at 
		FROM image_metadata 
		WHERE path = ?
	`, path).Scan(&m.Path, &m.Blurhash, &m.Width, &m.Height, &m.ComputedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *ImageMetadataRepo) Upsert(ctx context.Context, m *model.ImageMetadata) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO image_metadata (path, blurhash, width, height)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			blurhash = excluded.blurhash,
			width = excluded.width,
			height = excluded.height,
			computed_at = CURRENT_TIMESTAMP
	`, m.Path, m.Blurhash, m.Width, m.Height)
	return err
}

// DeleteBatch removes metadata rows for the given paths. Used by force-rescan
// to invalidate cached blurhash so the worker recomputes on next enqueue.
func (r *ImageMetadataRepo) DeleteBatch(ctx context.Context, paths []string) error {
	if len(paths) == 0 {
		return nil
	}
	args := make([]interface{}, len(paths))
	for i, p := range paths {
		args[i] = p
	}
	placeholders := strings.Repeat("?,", len(paths)-1) + "?"
	query := `DELETE FROM image_metadata WHERE path IN (` + placeholders + `)`
	_, err := r.db.ExecContext(ctx, query, args...)
	return err
}

func (r *ImageMetadataRepo) GetBatch(ctx context.Context, paths []string) (map[string]*model.ImageMetadata, error) {
	if len(paths) == 0 {
		return make(map[string]*model.ImageMetadata), nil
	}

	set := make(map[string]struct{})
	for _, p := range paths {
		if p != "" {
			set[p] = struct{}{}
		}
	}
	if len(set) == 0 {
		return make(map[string]*model.ImageMetadata), nil
	}

	args := make([]interface{}, 0, len(set))
	for p := range set {
		args = append(args, p)
	}

	placeholders := strings.Repeat("?,", len(args)-1) + "?"
	query := `SELECT path, blurhash, width, height, computed_at FROM image_metadata WHERE path IN (` + placeholders + `)`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]*model.ImageMetadata)
	for rows.Next() {
		var m model.ImageMetadata
		if err := rows.Scan(&m.Path, &m.Blurhash, &m.Width, &m.Height, &m.ComputedAt); err != nil {
			return nil, err
		}
		res[m.Path] = &m
	}
	return res, rows.Err()
}
