package repository

import (
	"context"
	"database/sql"

	"github.com/thawng/velox/internal/model"
)

// PretranscodeRepo handles CRUD for pre-transcode tables.
type PretranscodeRepo struct {
	db DBTX
}

// NewPretranscodeRepo creates a new pre-transcode repository.
func NewPretranscodeRepo(db DBTX) *PretranscodeRepo {
	return &PretranscodeRepo{db: db}
}

// --- Profiles ---

// ListProfiles returns all pre-transcode quality profiles.
func (r *PretranscodeRepo) ListProfiles(ctx context.Context) ([]model.PretranscodeProfile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled, created_at
		FROM pretranscode_profiles ORDER BY height ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.PretranscodeProfile
	for rows.Next() {
		var p model.PretranscodeProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.Height, &p.VideoBitrate, &p.AudioBitrate,
			&p.VideoCodec, &p.AudioCodec, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// GetProfile returns a single profile by ID.
func (r *PretranscodeRepo) GetProfile(ctx context.Context, id int64) (*model.PretranscodeProfile, error) {
	var p model.PretranscodeProfile
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled, created_at
		FROM pretranscode_profiles WHERE id = ?`, id).Scan(
		&p.ID, &p.Name, &p.Height, &p.VideoBitrate, &p.AudioBitrate,
		&p.VideoCodec, &p.AudioCodec, &p.Enabled, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	return &p, err
}

// GetProfileByHeight returns a profile matching the given height, or nil.
func (r *PretranscodeRepo) GetProfileByHeight(ctx context.Context, height int) (*model.PretranscodeProfile, error) {
	var p model.PretranscodeProfile
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled, created_at
		FROM pretranscode_profiles WHERE height = ?`, height).Scan(
		&p.ID, &p.Name, &p.Height, &p.VideoBitrate, &p.AudioBitrate,
		&p.VideoCodec, &p.AudioCodec, &p.Enabled, &p.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// ListEnabledProfiles returns only enabled profiles.
func (r *PretranscodeRepo) ListEnabledProfiles(ctx context.Context) ([]model.PretranscodeProfile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled, created_at
		FROM pretranscode_profiles WHERE enabled = 1 ORDER BY height ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []model.PretranscodeProfile
	for rows.Next() {
		var p model.PretranscodeProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.Height, &p.VideoBitrate, &p.AudioBitrate,
			&p.VideoCodec, &p.AudioCodec, &p.Enabled, &p.CreatedAt); err != nil {
			return nil, err
		}
		profiles = append(profiles, p)
	}
	return profiles, rows.Err()
}

// SetProfileEnabled updates the enabled flag for a profile.
func (r *PretranscodeRepo) SetProfileEnabled(ctx context.Context, id int64, enabled bool) error {
	val := 0
	if enabled {
		val = 1
	}
	res, err := r.db.ExecContext(ctx, `UPDATE pretranscode_profiles SET enabled = ? WHERE id = ?`, val, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// --- Files ---

// GetFileByMediaAndProfile returns a pre-transcode file for the given media file + profile.
func (r *PretranscodeRepo) GetFileByMediaAndProfile(ctx context.Context, mediaFileID, profileID int64) (*model.PretranscodeFile, error) {
	var f model.PretranscodeFile
	var errMsg, startedAt, completedAt sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, media_file_id, profile_id, file_path, file_size, duration_secs, status,
		       error_message, started_at, completed_at, created_at
		FROM pretranscode_files WHERE media_file_id = ? AND profile_id = ?`, mediaFileID, profileID).Scan(
		&f.ID, &f.MediaFileID, &f.ProfileID, &f.FilePath, &f.FileSize, &f.DurationSecs, &f.Status,
		&errMsg, &startedAt, &completedAt, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	f.ErrorMessage = errMsg.String
	f.StartedAt = startedAt.String
	f.CompletedAt = completedAt.String
	return &f, nil
}

// ListReadyFilesByMedia returns all ready pre-transcode files for a media_file_id.
func (r *PretranscodeRepo) ListReadyFilesByMedia(ctx context.Context, mediaFileID int64) ([]model.PretranscodeFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pf.id, pf.media_file_id, pf.profile_id, pf.file_path, pf.file_size, pf.duration_secs,
		       pf.status, pf.error_message, pf.started_at, pf.completed_at, pf.created_at
		FROM pretranscode_files pf
		JOIN pretranscode_profiles pp ON pp.id = pf.profile_id
		WHERE pf.media_file_id = ? AND pf.status = 'ready'
		ORDER BY pp.height DESC`, mediaFileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.PretranscodeFile
	for rows.Next() {
		var f model.PretranscodeFile
		var errMsg, startedAt, completedAt sql.NullString
		if err := rows.Scan(&f.ID, &f.MediaFileID, &f.ProfileID, &f.FilePath, &f.FileSize, &f.DurationSecs,
			&f.Status, &errMsg, &startedAt, &completedAt, &f.CreatedAt); err != nil {
			return nil, err
		}
		f.ErrorMessage = errMsg.String
		f.StartedAt = startedAt.String
		f.CompletedAt = completedAt.String
		files = append(files, f)
	}
	return files, rows.Err()
}

// ReadyFileWithProfile pairs a pre-transcode file with its profile metadata.
type ReadyFileWithProfile struct {
	File    model.PretranscodeFile
	Profile model.PretranscodeProfile
}

// ListReadyFilesWithProfiles returns ready files joined with their profiles in a single query.
// Avoids N+1 queries when building quality options.
func (r *PretranscodeRepo) ListReadyFilesWithProfiles(ctx context.Context, mediaFileID int64) ([]ReadyFileWithProfile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pf.id, pf.media_file_id, pf.profile_id, pf.file_path, pf.file_size, pf.duration_secs,
		       pf.status, pf.error_message, pf.started_at, pf.completed_at, pf.created_at,
		       pp.id, pp.name, pp.height, pp.video_bitrate, pp.audio_bitrate, pp.video_codec, pp.audio_codec, pp.enabled, pp.created_at
		FROM pretranscode_files pf
		JOIN pretranscode_profiles pp ON pp.id = pf.profile_id
		WHERE pf.media_file_id = ? AND pf.status = 'ready'
		ORDER BY pp.height DESC`, mediaFileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ReadyFileWithProfile
	for rows.Next() {
		var f model.PretranscodeFile
		var p model.PretranscodeProfile
		var errMsg, startedAt, completedAt sql.NullString
		if err := rows.Scan(
			&f.ID, &f.MediaFileID, &f.ProfileID, &f.FilePath, &f.FileSize, &f.DurationSecs,
			&f.Status, &errMsg, &startedAt, &completedAt, &f.CreatedAt,
			&p.ID, &p.Name, &p.Height, &p.VideoBitrate, &p.AudioBitrate, &p.VideoCodec, &p.AudioCodec, &p.Enabled, &p.CreatedAt,
		); err != nil {
			return nil, err
		}
		f.ErrorMessage = errMsg.String
		f.StartedAt = startedAt.String
		f.CompletedAt = completedAt.String
		results = append(results, ReadyFileWithProfile{File: f, Profile: p})
	}
	return results, rows.Err()
}

// UpsertFile inserts or updates a pre-transcode file record.
func (r *PretranscodeRepo) UpsertFile(ctx context.Context, f *model.PretranscodeFile) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, file_size, duration_secs, status, error_message, started_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (media_file_id, profile_id) DO UPDATE SET
			file_path = excluded.file_path,
			file_size = excluded.file_size,
			duration_secs = excluded.duration_secs,
			status = excluded.status,
			error_message = excluded.error_message,
			started_at = excluded.started_at,
			completed_at = excluded.completed_at`,
		f.MediaFileID, f.ProfileID, f.FilePath, f.FileSize, f.DurationSecs, f.Status,
		nullStr(f.ErrorMessage), nullStr(f.StartedAt), nullStr(f.CompletedAt))
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return id, nil
}

// UpdateFileStatus updates the status (and optional error) of a pre-transcode file.
func (r *PretranscodeRepo) UpdateFileStatus(ctx context.Context, id int64, status, errMsg, startedAt, completedAt string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE pretranscode_files SET status = ?, error_message = ?, started_at = COALESCE(?, started_at), completed_at = ?
		WHERE id = ?`, status, nullStr(errMsg), nullStr(startedAt), nullStr(completedAt), id)
	return err
}

// DeleteFilesByProfile deletes all pre-transcode files for a given profile.
func (r *PretranscodeRepo) DeleteFilesByProfile(ctx context.Context, profileID int64) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT file_path FROM pretranscode_files WHERE profile_id = ?`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM pretranscode_files WHERE profile_id = ?`, profileID)
	return paths, err
}

// DeleteAllFiles deletes all pre-transcode file records, returning paths for cleanup.
func (r *PretranscodeRepo) DeleteAllFiles(ctx context.Context) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT file_path FROM pretranscode_files`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, `DELETE FROM pretranscode_files`)
	return paths, err
}

// TotalDiskUsed returns the sum of file_size for all ready pre-transcode files.
func (r *PretranscodeRepo) TotalDiskUsed(ctx context.Context) (int64, error) {
	var total sql.NullInt64
	err := r.db.QueryRowContext(ctx, `SELECT SUM(file_size) FROM pretranscode_files WHERE status = 'ready'`).Scan(&total)
	return total.Int64, err
}

// --- Queue ---

// EnqueueJob adds a job to the queue. Silently ignores duplicates.
func (r *PretranscodeRepo) EnqueueJob(ctx context.Context, mediaFileID, profileID int64, priority int) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT OR IGNORE INTO pretranscode_queue (media_file_id, profile_id, priority)
		VALUES (?, ?, ?)`, mediaFileID, profileID, priority)
	return err
}

// PickNextJob picks the highest-priority queued job and marks it encoding.
// Returns nil if no jobs available.
func (r *PretranscodeRepo) PickNextJob(ctx context.Context) (*model.PretranscodeQueueItem, error) {
	var q model.PretranscodeQueueItem
	err := r.db.QueryRowContext(ctx, `
		SELECT id, media_file_id, profile_id, priority, status, created_at
		FROM pretranscode_queue
		WHERE status = 'queued'
		ORDER BY priority DESC, created_at ASC
		LIMIT 1`).Scan(&q.ID, &q.MediaFileID, &q.ProfileID, &q.Priority, &q.Status, &q.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	_, err = r.db.ExecContext(ctx, `UPDATE pretranscode_queue SET status = 'encoding' WHERE id = ?`, q.ID)
	if err != nil {
		return nil, err
	}
	q.Status = "encoding"
	return &q, nil
}

// CompleteJob marks a queue job as done or failed.
func (r *PretranscodeRepo) CompleteJob(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE pretranscode_queue SET status = ? WHERE id = ?`, status, id)
	return err
}

// CancelAllQueued cancels all queued (not encoding) jobs.
func (r *PretranscodeRepo) CancelAllQueued(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE pretranscode_queue SET status = 'cancelled' WHERE status = 'queued'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// QueueStats returns counts by status.
func (r *PretranscodeRepo) QueueStats(ctx context.Context) (total, queued, encoding, done, failed int, err error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM pretranscode_queue GROUP BY status`)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s string
		var c int
		if err = rows.Scan(&s, &c); err != nil {
			return
		}
		total += c
		switch s {
		case "queued":
			queued = c
		case "encoding":
			encoding = c
		case "done":
			done = c
		case "failed":
			failed = c
		}
	}
	err = rows.Err()
	return
}

// ResetEncodingJobs resets interrupted 'encoding' queue items back to 'queued' (startup recovery).
func (r *PretranscodeRepo) ResetEncodingJobs(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `UPDATE pretranscode_queue SET status = 'queued' WHERE status = 'encoding'`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ResetEncodingFiles resets interrupted 'encoding' file records back to 'pending' (startup recovery).
func (r *PretranscodeRepo) ResetEncodingFiles(ctx context.Context) {
	_, _ = r.db.ExecContext(ctx, `UPDATE pretranscode_files SET status = 'pending' WHERE status = 'encoding'`)
}

// ClearQueue deletes all queue entries.
func (r *PretranscodeRepo) ClearQueue(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM pretranscode_queue`)
	return err
}

// CountMediaFilesInLibrary returns the number of media files in a library.
func (r *PretranscodeRepo) CountMediaFilesInLibrary(ctx context.Context, libraryID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?`, libraryID).Scan(&count)
	return count, err
}

// ListMediaFilesForEnqueue returns media files in a library that don't have a ready
// pre-transcode file (or queued job) for the given profile.
func (r *PretranscodeRepo) ListMediaFilesForEnqueue(ctx context.Context, libraryID, profileID int64, profileHeight int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT mf.id, mf.media_id, mf.file_path, mf.file_size, mf.duration, mf.width, mf.height,
		       mf.video_codec, mf.audio_codec, mf.container, mf.bitrate
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?
		  AND mf.height >= ?
		  AND NOT EXISTS (
			SELECT 1 FROM pretranscode_files pf
			WHERE pf.media_file_id = mf.id AND pf.profile_id = ? AND pf.status = 'ready'
		  )
		  AND NOT EXISTS (
			SELECT 1 FROM pretranscode_queue pq
			WHERE pq.media_file_id = mf.id AND pq.profile_id = ? AND pq.status IN ('queued','encoding')
		  )
		ORDER BY m.title ASC`, libraryID, profileHeight, profileID, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []model.MediaFile
	for rows.Next() {
		var f model.MediaFile
		if err := rows.Scan(&f.ID, &f.MediaID, &f.FilePath, &f.FileSize, &f.Duration,
			&f.Width, &f.Height, &f.VideoCodec, &f.AudioCodec, &f.Container, &f.Bitrate); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

// AvgDurationByLibrary returns the average duration in seconds for media files in a library.
func (r *PretranscodeRepo) AvgDurationByLibrary(ctx context.Context, libraryID int64) (float64, error) {
	var avg sql.NullFloat64
	err := r.db.QueryRowContext(ctx, `
		SELECT AVG(mf.duration) FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ? AND mf.duration > 0`, libraryID).Scan(&avg)
	return avg.Float64, err
}

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}
