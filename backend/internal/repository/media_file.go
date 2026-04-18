package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// MediaFileRepo handles media_files (physical files) database operations
type MediaFileRepo struct {
	db DBTX
}

func NewMediaFileRepo(db DBTX) *MediaFileRepo {
	return &MediaFileRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *MediaFileRepo) WithTx(tx *sql.Tx) *MediaFileRepo {
	return &MediaFileRepo{db: tx}
}

// scanMediaFile scans a media file row into a model.MediaFile
func scanMediaFile(scanner interface{ Scan(...any) error }) (*model.MediaFile, error) {
	var mf model.MediaFile
	var isPrimary, isHDR int
	var lastVerified sql.NullString

	err := scanner.Scan(&mf.ID, &mf.MediaID, &mf.FilePath, &mf.FileSize, &mf.Duration,
		&mf.Width, &mf.Height, &mf.VideoCodec, &mf.VideoProfile, &mf.VideoLevel, &mf.VideoFPS,
		&mf.AudioCodec, &mf.Container, &mf.Bitrate,
		&isHDR, &mf.DVProfile, &mf.ColorTransfer, &mf.ColorPrimaries,
		&mf.Fingerprint, &isPrimary, &mf.AddedAt, &lastVerified)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("scanning media file: %w", err)
	}
	mf.IsPrimary = isPrimary == 1
	mf.IsHDR = isHDR == 1
	if lastVerified.Valid {
		mf.LastVerifiedAt = &lastVerified.String
	}
	return &mf, nil
}

// Create inserts a new media file
func (r *MediaFileRepo) Create(ctx context.Context, mf *model.MediaFile) error {
	query := `INSERT INTO media_files
		(media_id, file_path, file_size, duration, width, height,
		 video_codec, video_profile, video_level, video_fps,
		 audio_codec, container, bitrate,
		 is_hdr, dv_profile, color_transfer, color_primaries,
		 fingerprint, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, added_at, last_verified_at`

	isPrimary := 0
	if mf.IsPrimary {
		isPrimary = 1
	}
	isHDR := 0
	if mf.IsHDR {
		isHDR = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		mf.MediaID, mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.VideoProfile, mf.VideoLevel, mf.VideoFPS,
		mf.AudioCodec, mf.Container, mf.Bitrate,
		isHDR, mf.DVProfile, mf.ColorTransfer, mf.ColorPrimaries,
		mf.Fingerprint, isPrimary)

	var lastVerified sql.NullString
	err := row.Scan(&mf.ID, &mf.AddedAt, &lastVerified)
	if lastVerified.Valid {
		mf.LastVerifiedAt = &lastVerified.String
	}
	if err != nil {
		return fmt.Errorf("creating media file for media %d: %w", mf.MediaID, err)
	}
	return nil
}

// GetByID retrieves a media file by ID
func (r *MediaFileRepo) GetByID(ctx context.Context, id int64) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE id = ?`, id)
	return scanMediaFile(row)
}

// Update updates a media file
func (r *MediaFileRepo) Update(ctx context.Context, mf *model.MediaFile) error {
	isPrimary := 0
	if mf.IsPrimary {
		isPrimary = 1
	}

	isHDR := 0
	if mf.IsHDR {
		isHDR = 1
	}

	res, err := r.db.ExecContext(ctx, `UPDATE media_files SET
		file_path = ?, file_size = ?, duration = ?, width = ?, height = ?,
		video_codec = ?, video_profile = ?, video_level = ?, video_fps = ?,
		audio_codec = ?, container = ?, bitrate = ?,
		is_hdr = ?, dv_profile = ?, color_transfer = ?, color_primaries = ?,
		fingerprint = ?, is_primary = ?, last_verified_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.VideoProfile, mf.VideoLevel, mf.VideoFPS,
		mf.AudioCodec, mf.Container, mf.Bitrate,
		isHDR, mf.DVProfile, mf.ColorTransfer, mf.ColorPrimaries,
		mf.Fingerprint, isPrimary, mf.ID)
	if err != nil {
		return fmt.Errorf("update media file %d: %w", mf.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a media file
func (r *MediaFileRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM media_files WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete media file %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListByMediaID retrieves all files for a media item
func (r *MediaFileRepo) ListByMediaID(ctx context.Context, mediaID int64) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE media_id = ? ORDER BY is_primary DESC, id`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("listing media files: %w", err)
	}
	defer rows.Close()

	var files []model.MediaFile
	for rows.Next() {
		mf, err := scanMediaFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media file: %w", err)
		}
		files = append(files, *mf)
	}
	return files, rows.Err()
}

// GetPrimaryByMediaID retrieves the primary file for a media item
func (r *MediaFileRepo) GetPrimaryByMediaID(ctx context.Context, mediaID int64) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE media_id = ? AND is_primary = 1 LIMIT 1`, mediaID)
	return scanMediaFile(row)
}

// FindByFingerprint finds a file by its fingerprint
func (r *MediaFileRepo) FindByFingerprint(ctx context.Context, fingerprint string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE fingerprint = ?`, fingerprint)
	return scanMediaFile(row)
}

// FindByPath finds a file by its path
func (r *MediaFileRepo) FindByPath(ctx context.Context, path string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE file_path = ?`, path)
	return scanMediaFile(row)
}

// FindCloudFile finds a cloud file by its provider and native ID, ignoring query parameters
func (r *MediaFileRepo) FindCloudFile(ctx context.Context, provider, nativeID string) (*model.MediaFile, error) {
	basePath := provider + "://" + nativeID
	likePath := basePath + "?%"
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE file_path = ? OR file_path LIKE ? LIMIT 1`, basePath, likePath)
	return scanMediaFile(row)
}

// UpdatePath updates the file path (for rename detection)
func (r *MediaFileRepo) UpdatePath(ctx context.Context, id int64, newPath string) error {
	res, err := r.db.ExecContext(ctx, "UPDATE media_files SET file_path = ? WHERE id = ?", newPath, id)
	if err != nil {
		return fmt.Errorf("update path for media file %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// MarkMissing marks a file as missing (sets last_verified_at = NULL)
func (r *MediaFileRepo) MarkMissing(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET last_verified_at = NULL WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("marking media file %d missing: %w", id, err)
	}
	return nil
}

// ListByLibraryID retrieves all media files for a given library (via media table join).
// Results are returned in batches suitable for verification. Use limit/offset for pagination.
func (r *MediaFileRepo) ListByLibraryID(ctx context.Context, libraryID int64, limit, offset int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mf.id, mf.media_id, mf.file_path, mf.file_size, mf.duration,
		mf.width, mf.height, mf.video_codec, mf.video_profile, mf.video_level, mf.video_fps,
		mf.audio_codec, mf.container, mf.bitrate,
		mf.is_hdr, mf.dv_profile, mf.color_transfer, mf.color_primaries,
		mf.fingerprint, mf.is_primary, mf.added_at, mf.last_verified_at
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?
		ORDER BY mf.id
		LIMIT ? OFFSET ?`, libraryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing files by library: %w", err)
	}
	defer rows.Close()

	var files []model.MediaFile
	for rows.Next() {
		mf, err := scanMediaFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media file: %w", err)
		}
		files = append(files, *mf)
	}
	return files, rows.Err()
}

// ListAllPaginated retrieves all media files in the database, paginated.
func (r *MediaFileRepo) ListAllPaginated(ctx context.Context, limit, offset int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files ORDER BY id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing all files: %w", err)
	}
	defer rows.Close()

	var files []model.MediaFile
	for rows.Next() {
		mf, err := scanMediaFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media file: %w", err)
		}
		files = append(files, *mf)
	}
	return files, rows.Err()
}

// ListUnprobedCloud returns cloud media files that haven't been probed yet
// (video_codec is empty, meaning ffprobe hasn't run). Paginated for batch processing.
func (r *MediaFileRepo) ListUnprobedCloud(ctx context.Context, limit, offset int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		is_hdr, dv_profile, color_transfer, color_primaries,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files
		WHERE file_path LIKE '%://%' AND file_path NOT LIKE 'http%' AND video_codec = ''
		ORDER BY id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing unprobed cloud files: %w", err)
	}
	defer rows.Close()

	var files []model.MediaFile
	for rows.Next() {
		mf, err := scanMediaFile(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media file: %w", err)
		}
		files = append(files, *mf)
	}
	return files, rows.Err()
}

// UpdateLastVerified updates the last_verified_at timestamp for a file.
func (r *MediaFileRepo) UpdateLastVerified(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET last_verified_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("updating last verified for media file %d: %w", id, err)
	}
	return nil
}

// DeleteByMediaID removes all files for a media item
func (r *MediaFileRepo) DeleteByMediaID(ctx context.Context, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_files WHERE media_id = ?", mediaID)
	if err != nil {
		return fmt.Errorf("deleting files for media %d: %w", mediaID, err)
	}
	return nil
}

// SetPrimary sets a file as the primary version for its media
func (r *MediaFileRepo) SetPrimary(ctx context.Context, mediaID, fileID int64) error {
	// First clear primary for all files of this media
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET is_primary = 0 WHERE media_id = ?", mediaID)
	if err != nil {
		return fmt.Errorf("clearing primary for media %d: %w", mediaID, err)
	}
	// Then set the new primary
	_, err = r.db.ExecContext(ctx, "UPDATE media_files SET is_primary = 1 WHERE id = ? AND media_id = ?", fileID, mediaID)
	if err != nil {
		return fmt.Errorf("setting primary file %d for media %d: %w", fileID, mediaID, err)
	}
	return nil
}

// TotalSize returns the sum of all media file sizes
func (r *MediaFileRepo) TotalSize(ctx context.Context) (int64, error) {
	var size int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(file_size), 0) FROM media_files`).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("summing file sizes: %w", err)
	}
	return size, nil
}
