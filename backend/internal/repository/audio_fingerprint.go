package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// AudioFingerprintRepo handles audio_fingerprints database operations.
type AudioFingerprintRepo struct {
	db DBTX
}

func NewAudioFingerprintRepo(db DBTX) *AudioFingerprintRepo {
	return &AudioFingerprintRepo{db: db}
}

func (r *AudioFingerprintRepo) WithTx(tx *sql.Tx) *AudioFingerprintRepo {
	return &AudioFingerprintRepo{db: tx}
}

// Upsert inserts or replaces a fingerprint for a (media_file_id, region) pair.
func (r *AudioFingerprintRepo) Upsert(ctx context.Context, fp *model.AudioFingerprint) error {
	query := `INSERT INTO audio_fingerprints
		(media_file_id, region, fingerprint, duration_sec, sample_count)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(media_file_id, region) DO UPDATE SET
			fingerprint = excluded.fingerprint,
			duration_sec = excluded.duration_sec,
			sample_count = excluded.sample_count,
			created_at = CURRENT_TIMESTAMP
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query,
		fp.MediaFileID, fp.Region, fp.Fingerprint, fp.DurationSec, fp.SampleCount,
	).Scan(&fp.ID, &fp.CreatedAt)
}

// GetByMediaFileID retrieves a fingerprint for a specific file and region.
func (r *AudioFingerprintRepo) GetByMediaFileID(ctx context.Context, fileID int64, region string) (*model.AudioFingerprint, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, media_file_id, region, fingerprint, duration_sec, sample_count, created_at
		FROM audio_fingerprints WHERE media_file_id = ? AND region = ?`,
		fileID, region)

	var fp model.AudioFingerprint
	err := row.Scan(&fp.ID, &fp.MediaFileID, &fp.Region, &fp.Fingerprint, &fp.DurationSec, &fp.SampleCount, &fp.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &fp, nil
}

// GetByMediaFileIDs retrieves fingerprints for multiple files and a specific region.
func (r *AudioFingerprintRepo) GetByMediaFileIDs(ctx context.Context, fileIDs []int64, region string) ([]model.AudioFingerprint, error) {
	if len(fileIDs) == 0 {
		return nil, nil
	}

	// Build IN clause
	placeholders := make([]string, len(fileIDs))
	args := make([]interface{}, 0, len(fileIDs)+1)
	for i, id := range fileIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}
	args = append(args, region)

	query := fmt.Sprintf(
		`SELECT id, media_file_id, region, fingerprint, duration_sec, sample_count, created_at
		FROM audio_fingerprints WHERE media_file_id IN (%s) AND region = ?`,
		joinStrings(placeholders, ","))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying fingerprints: %w", err)
	}
	defer rows.Close()

	var results []model.AudioFingerprint
	for rows.Next() {
		var fp model.AudioFingerprint
		if err := rows.Scan(&fp.ID, &fp.MediaFileID, &fp.Region, &fp.Fingerprint, &fp.DurationSec, &fp.SampleCount, &fp.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning fingerprint: %w", err)
		}
		results = append(results, fp)
	}
	return results, rows.Err()
}

// DeleteByMediaFileID removes all fingerprints for a media file.
func (r *AudioFingerprintRepo) DeleteByMediaFileID(ctx context.Context, fileID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM audio_fingerprints WHERE media_file_id = ?", fileID)
	return err
}

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
