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

func setupMediaFileTestDB(t *testing.T) (*sql.DB, *MediaFileRepo) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	require.NoError(t, err)

	schema := `
	CREATE TABLE media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		duration REAL DEFAULT 0,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		video_codec TEXT DEFAULT '',
		video_profile TEXT DEFAULT '',
		video_level TEXT DEFAULT '',
		video_fps REAL DEFAULT 0,
		audio_codec TEXT DEFAULT '',
		container TEXT DEFAULT '',
		bitrate INTEGER DEFAULT 0,
		is_hdr INTEGER DEFAULT 0,
		dv_profile TEXT DEFAULT '',
		color_transfer TEXT DEFAULT '',
		color_primaries TEXT DEFAULT '',
		fingerprint TEXT DEFAULT '',
		is_primary INTEGER DEFAULT 0,
		added_at TEXT NOT NULL DEFAULT (datetime('now')),
		last_verified_at TEXT
	);

	CREATE TABLE media (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		library_id INTEGER NOT NULL,
		title TEXT NOT NULL
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db, NewMediaFileRepo(db)
}

func TestMediaFileRepo_Create(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{
		MediaID:    1,
		FilePath:   "/movies/test.mp4",
		FileSize:   2000000000,
		Duration:   7200.5,
		Width:      1920,
		Height:     1080,
		VideoCodec: "h264",
		AudioCodec: "aac",
		Container:  "mp4",
		Bitrate:    5000000,
		IsPrimary:  true,
	}

	err := repo.Create(ctx, mf)
	require.NoError(t, err)
	assert.Greater(t, mf.ID, int64(0))
	assert.NotEmpty(t, mf.AddedAt)
}

func TestMediaFileRepo_GetByID(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/test.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	got, err := repo.GetByID(ctx, mf.ID)
	require.NoError(t, err)
	assert.Equal(t, "/test.mp4", got.FilePath)
	assert.Equal(t, int64(1), got.MediaID)
}

func TestMediaFileRepo_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetByID(ctx, 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_Update(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/original.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	mf.FilePath = "/updated.mp4"
	mf.FileSize = 2000
	err := repo.Update(ctx, mf)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, mf.ID)
	require.NoError(t, err)
	assert.Equal(t, "/updated.mp4", got.FilePath)
	assert.Equal(t, int64(2000), got.FileSize)
}

func TestMediaFileRepo_Update_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{ID: 9999, FilePath: "/ghost.mp4"}
	err := repo.Update(ctx, mf)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_Delete(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/delete.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	err := repo.Delete(ctx, mf.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, mf.ID)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_Delete_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.Delete(ctx, 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_ListByMediaID(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		mf := &model.MediaFile{MediaID: 1, FilePath: "/movie_" + string(rune('a'+i)) + ".mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
		require.NoError(t, repo.Create(ctx, mf))
	}

	files, err := repo.ListByMediaID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, files, 3)
}

func TestMediaFileRepo_ListByMediaID_Empty(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	files, err := repo.ListByMediaID(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestMediaFileRepo_GetPrimaryByMediaID(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf1 := &model.MediaFile{MediaID: 1, FilePath: "/primary.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264", IsPrimary: true}
	require.NoError(t, repo.Create(ctx, mf1))
	mf2 := &model.MediaFile{MediaID: 1, FilePath: "/secondary.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264", IsPrimary: false}
	require.NoError(t, repo.Create(ctx, mf2))

	primary, err := repo.GetPrimaryByMediaID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, "/primary.mp4", primary.FilePath)
	assert.True(t, primary.IsPrimary)
}

func TestMediaFileRepo_GetPrimaryByMediaID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetPrimaryByMediaID(ctx, 9999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_FindByFingerprint(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/fingerprinted.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264", Fingerprint: "abc123"}
	require.NoError(t, repo.Create(ctx, mf))

	found, err := repo.FindByFingerprint(ctx, "abc123")
	require.NoError(t, err)
	assert.Equal(t, mf.ID, found.ID)
}

func TestMediaFileRepo_FindByFingerprint_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.FindByFingerprint(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_FindByPath(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/unique/path.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	found, err := repo.FindByPath(ctx, "/unique/path.mp4")
	require.NoError(t, err)
	assert.Equal(t, mf.ID, found.ID)
}

func TestMediaFileRepo_FindByPath_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.FindByPath(ctx, "/nonexistent.mp4")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_FindCloudFile(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "gdrive://abc123/movie.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	found, err := repo.FindCloudFile(ctx, "gdrive", "abc123/movie.mp4")
	require.NoError(t, err)
	assert.Equal(t, mf.ID, found.ID)
}

func TestMediaFileRepo_UpdatePath(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/old/path.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	err := repo.UpdatePath(ctx, mf.ID, "/new/path.mp4")
	require.NoError(t, err)

	found, err := repo.FindByPath(ctx, "/new/path.mp4")
	require.NoError(t, err)
	assert.Equal(t, mf.ID, found.ID)
}

func TestMediaFileRepo_UpdatePath_NotFound(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.UpdatePath(ctx, 9999, "/new.mp4")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestMediaFileRepo_MarkMissing(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/missing.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	err := repo.MarkMissing(ctx, mf.ID)
	require.NoError(t, err)

	var lastVerified sql.NullString
	db.QueryRowContext(ctx, "SELECT last_verified_at FROM media_files WHERE id = ?", mf.ID).Scan(&lastVerified)
	assert.False(t, lastVerified.Valid)
}

func TestMediaFileRepo_ListByLibraryID(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, "INSERT INTO media (id, library_id, title) VALUES (1, 10, 'Library 10 Media')")
	require.NoError(t, err)

	mf := &model.MediaFile{MediaID: 1, FilePath: "/library10.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	files, err := repo.ListByLibraryID(ctx, 10, 10, 0)
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestMediaFileRepo_ListByLibraryID_Pagination(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := db.ExecContext(ctx, "INSERT INTO media (id, library_id, title) VALUES (1, 10, 'Media')")
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		mf := &model.MediaFile{MediaID: 1, FilePath: "/file" + string(rune('0'+i)) + ".mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
		require.NoError(t, repo.Create(ctx, mf))
	}

	first, err := repo.ListByLibraryID(ctx, 10, 2, 0)
	require.NoError(t, err)
	assert.Len(t, first, 2)

	second, err := repo.ListByLibraryID(ctx, 10, 2, 2)
	require.NoError(t, err)
	assert.Len(t, second, 2)
}

func TestMediaFileRepo_ListAllPaginated(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		mf := &model.MediaFile{MediaID: 1, FilePath: "/paginated" + string(rune('0'+i)) + ".mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
		require.NoError(t, repo.Create(ctx, mf))
	}

	files, err := repo.ListAllPaginated(ctx, 3, 0)
	require.NoError(t, err)
	assert.Len(t, files, 3)
}

func TestMediaFileRepo_DeleteByMediaID(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf1 := &model.MediaFile{MediaID: 1, FilePath: "/media1_a.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	mf2 := &model.MediaFile{MediaID: 1, FilePath: "/media1_b.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf1))
	require.NoError(t, repo.Create(ctx, mf2))

	err := repo.DeleteByMediaID(ctx, 1)
	require.NoError(t, err)

	count, _ := repo.ListByMediaID(ctx, 1)
	assert.Len(t, count, 0)
}

func TestMediaFileRepo_SetPrimary(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf1 := &model.MediaFile{MediaID: 1, FilePath: "/old_primary.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264", IsPrimary: true}
	mf2 := &model.MediaFile{MediaID: 1, FilePath: "/new_primary.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264", IsPrimary: false}
	require.NoError(t, repo.Create(ctx, mf1))
	require.NoError(t, repo.Create(ctx, mf2))

	err := repo.SetPrimary(ctx, 1, mf2.ID)
	require.NoError(t, err)

	// mf2 should now be primary
	primary, err := repo.GetPrimaryByMediaID(ctx, 1)
	require.NoError(t, err)
	assert.Equal(t, mf2.ID, primary.ID)
}

func TestMediaFileRepo_TotalSize(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf1 := &model.MediaFile{MediaID: 1, FilePath: "/a.mp4", FileSize: 1000000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	mf2 := &model.MediaFile{MediaID: 1, FilePath: "/b.mp4", FileSize: 2000000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf1))
	require.NoError(t, repo.Create(ctx, mf2))

	size, err := repo.TotalSize(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(3000000), size)
}

func TestMediaFileRepo_UpdateLastVerified(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	mf := &model.MediaFile{MediaID: 1, FilePath: "/verify.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	require.NoError(t, repo.Create(ctx, mf))

	// Set last_verified_at to old value first
	_, err := db.ExecContext(ctx, "UPDATE media_files SET last_verified_at = '2020-01-01' WHERE id = ?", mf.ID)
	require.NoError(t, err)

	err = repo.UpdateLastVerified(ctx, mf.ID)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, mf.ID)
	require.NoError(t, err)
	assert.NotNil(t, got.LastVerifiedAt)
	assert.NotEqual(t, "2020-01-01", *got.LastVerifiedAt)
}

func TestMediaFileRepo_WithTx(t *testing.T) {
	t.Parallel()
	db, repo := setupMediaFileTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	mf := &model.MediaFile{MediaID: 1, FilePath: "/tx.mp4", FileSize: 1000, Width: 1920, Height: 1080, VideoCodec: "h264"}
	err = txRepo.Create(ctx, mf)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, mf.ID)
	require.NoError(t, err)
	assert.Equal(t, "/tx.mp4", got.FilePath)
}
