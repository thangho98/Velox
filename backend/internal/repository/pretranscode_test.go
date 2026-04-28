package repository

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/model"
)

func setupPretranscodeTestDB(t *testing.T) (*sql.DB, *PretranscodeRepo) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	require.NoError(t, err)

	schema := `
	CREATE TABLE pretranscode_profiles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		height INTEGER NOT NULL,
		video_bitrate INTEGER NOT NULL,
		audio_bitrate INTEGER NOT NULL,
		video_codec TEXT NOT NULL DEFAULT 'h264',
		audio_codec TEXT NOT NULL DEFAULT 'aac',
		enabled INTEGER NOT NULL DEFAULT 0,
		created_at TEXT NOT NULL DEFAULT (datetime('now'))
	);

	CREATE TABLE pretranscode_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_file_id INTEGER NOT NULL,
		profile_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER NOT NULL DEFAULT 0,
		duration_secs REAL,
		status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','encoding','ready','failed')),
		error_message TEXT,
		started_at TEXT,
		completed_at TEXT,
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		audio_codec TEXT DEFAULT '',
		audio_channels INTEGER DEFAULT 0,
		audio_bitrate INTEGER DEFAULT 0,
		audio_sample_rate INTEGER DEFAULT 0,
		UNIQUE(media_file_id, profile_id)
	);

	CREATE TABLE pretranscode_queue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_file_id INTEGER NOT NULL,
		profile_id INTEGER NOT NULL,
		priority INTEGER NOT NULL DEFAULT 0,
		status TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','encoding','done','failed','cancelled')),
		created_at TEXT NOT NULL DEFAULT (datetime('now')),
		UNIQUE(media_file_id, profile_id)
	);

	CREATE TABLE media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		duration REAL DEFAULT 0,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		video_codec TEXT DEFAULT '',
		audio_codec TEXT DEFAULT '',
		container TEXT DEFAULT '',
		bitrate INTEGER DEFAULT 0
	);

	CREATE TABLE media (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		library_id INTEGER NOT NULL,
		title TEXT NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db, NewPretranscodeRepo(db)
}

// Profile Tests

func TestPretranscodeRepo_ListProfiles(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled) VALUES
		('720p', 720, 3000, 128, 'h264', 'aac', 1),
		('1080p', 1080, 5000, 192, 'h264', 'aac', 1),
		('4K', 2160, 15000, 256, 'h265', 'aac', 0)`)
	require.NoError(t, err)

	profiles, err := repo.ListProfiles(ctx)
	require.NoError(t, err)
	assert.Len(t, profiles, 3) // excludes 'copy' audio-remux which we don't have here
}

func TestPretranscodeRepo_GetProfile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		('720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	var id int64
	row := db.QueryRowContext(ctx, "SELECT id FROM pretranscode_profiles WHERE name = '720p'")
	row.Scan(&id)

	p, err := repo.GetProfile(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "720p", p.Name)
	assert.Equal(t, 720, p.Height)
}

func TestPretranscodeRepo_GetProfile_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetProfile(ctx, 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPretranscodeRepo_GetProfileByHeight(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		('1080p', 1080, 5000, 192, 'h264', 'aac')`)
	require.NoError(t, err)

	p, err := repo.GetProfileByHeight(ctx, 1080)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "1080p", p.Name)
}

func TestPretranscodeRepo_GetProfileByHeight_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	p, err := repo.GetProfileByHeight(ctx, 9999)
	require.NoError(t, err)
	assert.Nil(t, p)
}

func TestPretranscodeRepo_ListEnabledProfiles(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled) VALUES
		('720p', 720, 3000, 128, 'h264', 'aac', 1),
		('4K', 2160, 15000, 256, 'h265', 'aac', 0)`)
	require.NoError(t, err)

	profiles, err := repo.ListEnabledProfiles(ctx)
	require.NoError(t, err)
	assert.Len(t, profiles, 1)
	assert.Equal(t, "720p", profiles[0].Name)
}

func TestPretranscodeRepo_SetProfileEnabled(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec, enabled) VALUES
		('720p', 720, 3000, 128, 'h264', 'aac', 1)`)
	require.NoError(t, err)

	var id int64
	db.QueryRowContext(ctx, "SELECT id FROM pretranscode_profiles").Scan(&id)

	err = repo.SetProfileEnabled(ctx, id, false)
	require.NoError(t, err)

	var enabled int
	db.QueryRowContext(ctx, "SELECT enabled FROM pretranscode_profiles WHERE id = ?", id).Scan(&enabled)
	assert.Equal(t, 0, enabled)
}

func TestPretranscodeRepo_SetProfileEnabled_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.SetProfileEnabled(ctx, 9999, true)
	assert.ErrorIs(t, err, ErrNotFound)
}

// File Tests

func TestPretranscodeRepo_GetFileByMediaAndProfile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	// Setup profile
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	// Setup file
	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, file_size, duration_secs, status) VALUES
		(100, 1, '/output/100_720p.mp4', 500000000, 3600.5, 'ready')`)
	require.NoError(t, err)

	f, err := repo.GetFileByMediaAndProfile(ctx, 100, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(100), f.MediaFileID)
	assert.Equal(t, int64(1), f.ProfileID)
	assert.Equal(t, "/output/100_720p.mp4", f.FilePath)
	assert.Equal(t, "ready", f.Status)
}

func TestPretranscodeRepo_GetFileByMediaAndProfile_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetFileByMediaAndProfile(ctx, 9999, 1)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPretranscodeRepo_ListReadyFilesByMedia(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac'),
		(2, '1080p', 1080, 5000, 192, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, status, duration_secs) VALUES
		(100, 1, '/720p.mp4', 'ready', 3600.0),
		(100, 2, '/1080p.mp4', 'ready', 3600.0)`)
	require.NoError(t, err)

	files, err := repo.ListReadyFilesByMedia(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, files, 2)
}

func TestPretranscodeRepo_ListReadyFilesByMedia_NotReady(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, status, duration_secs) VALUES
		(100, 1, '/720p.mp4', 'encoding', 3600.0)`)
	require.NoError(t, err)

	files, err := repo.ListReadyFilesByMedia(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestPretranscodeRepo_ListReadyFilesWithProfiles(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, file_size, duration_secs, status) VALUES
		(100, 1, '/720p.mp4', 500000000, 3600, 'ready')`)
	require.NoError(t, err)

	results, err := repo.ListReadyFilesWithProfiles(ctx, 100)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "720p", results[0].Profile.Name)
	assert.Equal(t, 720, results[0].Profile.Height)
}

func TestPretranscodeRepo_UpsertFile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	f := &model.PretranscodeFile{
		MediaFileID:  100,
		ProfileID:    1,
		FilePath:     "/output/100_720p.mp4",
		FileSize:     500000000,
		DurationSecs: 3600,
		Status:       "ready",
	}
	id, err := repo.UpsertFile(ctx, f)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Upsert again - should update
	f.FileSize = 600000000
	id2, err := repo.UpsertFile(ctx, f)
	require.NoError(t, err)
	assert.Equal(t, id, id2)
}

func TestPretranscodeRepo_UpdateFileStatus(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (id, media_file_id, profile_id, file_path, status) VALUES
		(1, 100, 1, '/out.mp4', 'pending')`)
	require.NoError(t, err)

	err = repo.UpdateFileStatus(ctx, 1, "encoding", "", "2024-01-01 10:00:00", "")
	require.NoError(t, err)

	var status, startedAt string
	db.QueryRowContext(ctx, "SELECT status, started_at FROM pretranscode_files WHERE id = 1").Scan(&status, &startedAt)
	assert.Equal(t, "encoding", status)
	assert.NotEmpty(t, startedAt)
}

func TestPretranscodeRepo_UpdateFileAudioMeta(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (id, media_file_id, profile_id, file_path, status) VALUES
		(1, 100, 1, '/out.mp4', 'encoding')`)
	require.NoError(t, err)

	err = repo.UpdateFileAudioMeta(ctx, 1, "ac3", 6, 640000, 48000)
	require.NoError(t, err)

	var codec string
	var channels, bitrate, sampleRate int
	db.QueryRowContext(ctx, "SELECT audio_codec, audio_channels, audio_bitrate, audio_sample_rate FROM pretranscode_files WHERE id = 1").
		Scan(&codec, &channels, &bitrate, &sampleRate)
	assert.Equal(t, "ac3", codec)
	assert.Equal(t, 6, channels)
}

func TestPretranscodeRepo_DeleteFilesByProfile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, status) VALUES
		(100, 1, '/720p_a.mp4', 'ready'),
		(101, 1, '/720p_b.mp4', 'ready')`)
	require.NoError(t, err)

	paths, err := repo.DeleteFilesByProfile(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, paths, 2)

	var count int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pretranscode_files WHERE profile_id = 1").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestPretranscodeRepo_DeleteAllFiles(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, status) VALUES
		(100, 1, '/a.mp4', 'ready'),
		(101, 1, '/b.mp4', 'ready')`)
	require.NoError(t, err)

	paths, err := repo.DeleteAllFiles(ctx)
	require.NoError(t, err)
	assert.Len(t, paths, 2)

	var count int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pretranscode_files").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestPretranscodeRepo_TotalDiskUsed(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, file_size, status) VALUES
		(100, 1, '/a.mp4', 1000000, 'ready'),
		(101, 1, '/b.mp4', 2000000, 'ready')`)
	require.NoError(t, err)

	total, err := repo.TotalDiskUsed(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3000000), total)
}

// Queue Tests

func TestPretranscodeRepo_EnqueueJob(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	err = repo.EnqueueJob(ctx, 100, 1, 10)
	require.NoError(t, err)

	var status string
	db.QueryRowContext(ctx, "SELECT status FROM pretranscode_queue WHERE media_file_id = 100").Scan(&status)
	assert.Equal(t, "queued", status)
}

func TestPretranscodeRepo_PickNextJob(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac'),
		(2, 'audio', 0, 0, 128, 'copy', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, priority, status) VALUES
		(100, 1, 10, 'queued'),
		(101, 1, 5, 'queued'),
		(102, 2, 20, 'queued')`)
	require.NoError(t, err)

	job, err := repo.PickNextJob(ctx)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, int64(100), job.MediaFileID) // highest priority
	assert.Equal(t, "encoding", job.Status)

	// Verify audio-remux profile (video_codec='copy') is excluded
	var status string
	db.QueryRowContext(ctx, "SELECT status FROM pretranscode_queue WHERE media_file_id = 102").Scan(&status)
	assert.Equal(t, "queued", status) // not picked
}

func TestPretranscodeRepo_PickNextJob_Empty(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	job, err := repo.PickNextJob(ctx)
	require.NoError(t, err)
	assert.Nil(t, job)
}

func TestPretranscodeRepo_PickNextJobForProfile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac'),
		(2, '1080p', 1080, 5000, 192, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, priority, status) VALUES
		(100, 1, 10, 'queued'),
		(101, 2, 20, 'queued')`)
	require.NoError(t, err)

	job, err := repo.PickNextJobForProfile(ctx, 1)
	require.NoError(t, err)
	assert.NotNil(t, job)
	assert.Equal(t, int64(1), job.ProfileID)
	assert.Equal(t, int64(100), job.MediaFileID)
}

func TestPretranscodeRepo_CompleteJob(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (id, media_file_id, profile_id, status) VALUES
		(1, 100, 1, 'encoding')`)
	require.NoError(t, err)

	err = repo.CompleteJob(ctx, 1, "done")
	require.NoError(t, err)

	var status string
	db.QueryRowContext(ctx, "SELECT status FROM pretranscode_queue WHERE id = 1").Scan(&status)
	assert.Equal(t, "done", status)
}

func TestPretranscodeRepo_CancelAllQueued(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, status) VALUES
		(100, 1, 'queued'),
		(101, 1, 'queued'),
		(102, 1, 'encoding')`)
	require.NoError(t, err)

	cancelled, err := repo.CancelAllQueued(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), cancelled)

	var queuedCount int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pretranscode_queue WHERE status = 'queued'").Scan(&queuedCount)
	assert.Equal(t, 0, queuedCount)
}

func TestPretranscodeRepo_QueueStats(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac'),
		(2, 'audio', 0, 0, 128, 'copy', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, status) VALUES
		(100, 1, 'queued'),
		(101, 1, 'queued'),
		(102, 1, 'encoding'),
		(103, 1, 'done'),
		(104, 1, 'failed'),
		(105, 2, 'queued')`) // audio-remux excluded
	require.NoError(t, err)

	total, queued, encoding, done, failed, err := repo.QueueStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, total) // all profile 1 entries (audio-remux profile 2 excluded)
	assert.Equal(t, 2, queued)
	assert.Equal(t, 1, encoding)
	assert.Equal(t, 1, done)
	assert.Equal(t, 1, failed) // one failed entry for profile 1
}

func TestPretranscodeRepo_ResetEncodingJobs(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, status) VALUES
		(100, 1, 'encoding'),
		(101, 1, 'encoding')`)
	require.NoError(t, err)

	reset, err := repo.ResetEncodingJobs(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(2), reset)

	var status string
	db.QueryRowContext(ctx, "SELECT status FROM pretranscode_queue WHERE media_file_id = 100").Scan(&status)
	assert.Equal(t, "queued", status)
}

func TestPretranscodeRepo_ResetEncodingFiles(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_files (media_file_id, profile_id, file_path, status) VALUES
		(100, 1, '/a.mp4', 'encoding'),
		(101, 1, '/b.mp4', 'encoding')`)
	require.NoError(t, err)

	repo.ResetEncodingFiles(ctx)

	var status string
	db.QueryRowContext(ctx, "SELECT status FROM pretranscode_files WHERE media_file_id = 100").Scan(&status)
	assert.Equal(t, "pending", status)
}

func TestPretranscodeRepo_ClearQueue(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (id, name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		(1, '720p', 720, 3000, 128, 'h264', 'aac')`)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, `
		INSERT INTO pretranscode_queue (media_file_id, profile_id, status) VALUES
		(100, 1, 'queued'),
		(101, 1, 'queued')`)
	require.NoError(t, err)

	err = repo.ClearQueue(ctx)
	require.NoError(t, err)

	var count int
	db.QueryRowContext(ctx, "SELECT COUNT(*) FROM pretranscode_queue").Scan(&count)
	assert.Equal(t, 0, count)
}

func TestPretranscodeRepo_GetAudioRemuxProfile(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO pretranscode_profiles (name, height, video_bitrate, audio_bitrate, video_codec, audio_codec) VALUES
		('audio-remux', 0, 0, 256, 'copy', 'aac')`)
	require.NoError(t, err)

	p, err := repo.GetAudioRemuxProfile(ctx)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Equal(t, "copy", p.VideoCodec)
}

func TestPretranscodeRepo_GetAudioRemuxProfile_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupPretranscodeTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	p, err := repo.GetAudioRemuxProfile(ctx)
	require.NoError(t, err)
	assert.Nil(t, p)
}
