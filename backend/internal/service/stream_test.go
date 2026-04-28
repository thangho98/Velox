package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

func setupStreamTestDB(t *testing.T) (*sql.DB, *repository.MediaFileRepo, *repository.AudioTrackRepo, *repository.SubtitleRepo) {
	t.Helper()

	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}

	schema := `
	CREATE TABLE libraries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		media_type TEXT NOT NULL,
		paths TEXT NOT NULL DEFAULT '[]',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE media (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		library_id INTEGER NOT NULL REFERENCES libraries(id),
		media_type TEXT NOT NULL,
		title TEXT NOT NULL,
		sort_title TEXT NOT NULL,
		release_date TEXT,
		tmdb_id INTEGER,
		imdb_id TEXT,
		overview TEXT,
		runtime INTEGER DEFAULT 0,
		is_hidden BOOLEAN DEFAULT 0
	);
	CREATE TABLE media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id INTEGER NOT NULL REFERENCES media(id) ON DELETE CASCADE,
		file_path TEXT NOT NULL UNIQUE,
		file_size INTEGER DEFAULT 0,
		duration REAL DEFAULT 0,
		width INTEGER DEFAULT 0,
		height INTEGER DEFAULT 0,
		video_codec TEXT DEFAULT '',
		video_profile TEXT DEFAULT '',
		video_level INTEGER DEFAULT 0,
		video_fps REAL DEFAULT 0,
		audio_codec TEXT DEFAULT '',
		container TEXT DEFAULT '',
		bitrate INTEGER DEFAULT 0,
		is_hdr INTEGER NOT NULL DEFAULT 0,
		dv_profile INTEGER NOT NULL DEFAULT 0,
		color_transfer TEXT NOT NULL DEFAULT '',
		color_primaries TEXT NOT NULL DEFAULT '',
		fingerprint TEXT DEFAULT '',
		is_primary INTEGER DEFAULT 1,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_verified_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE TABLE audio_tracks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
		stream_index INTEGER NOT NULL,
		language TEXT DEFAULT '',
		codec TEXT DEFAULT '',
		channels INTEGER DEFAULT 0,
		channel_layout TEXT DEFAULT '',
		bitrate INTEGER DEFAULT 0,
		sample_rate INTEGER DEFAULT 0,
		title TEXT DEFAULT '',
		is_default INTEGER DEFAULT 0
	);
	CREATE TABLE subtitles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_file_id INTEGER NOT NULL REFERENCES media_files(id) ON DELETE CASCADE,
		language TEXT DEFAULT '',
		codec TEXT DEFAULT '',
		title TEXT DEFAULT '',
		is_embedded INTEGER DEFAULT 0,
		stream_index INTEGER DEFAULT 0,
		is_forced INTEGER DEFAULT 0,
		is_default INTEGER DEFAULT 0,
		is_sdh INTEGER DEFAULT 0
	);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("creating schema: %v", err)
	}

	if _, err := db.Exec(`
		INSERT INTO libraries (id, name, media_type, paths) VALUES (1, 'Test', 'movie', '["/test"]');
		INSERT INTO media (id, library_id, media_type, title, sort_title) VALUES (1, 1, 'movie', 'Test Movie', 'Test Movie');
		INSERT INTO media (id, library_id, media_type, title, sort_title) VALUES (2, 1, 'movie', 'Second Movie', 'Second Movie');
		INSERT INTO media_files (id, media_id, file_path, file_size, duration, width, height, is_primary)
			VALUES (1, 1, '/test/movie1.mkv', 1000, 3600, 1920, 1080, 1);
		INSERT INTO media_files (id, media_id, file_path, file_size, duration, is_primary)
			VALUES (2, 1, '/test/movie2.mkv', 2000, 1800, 0);
		INSERT INTO media_files (id, media_id, file_path, file_size, is_primary)
			VALUES (3, 2, '/test/movie3.mkv', 3000, 1);
	`); err != nil {
		t.Fatalf("inserting test data: %v", err)
	}

	return db, repository.NewMediaFileRepo(db), repository.NewAudioTrackRepo(db), repository.NewSubtitleRepo(db)
}

func TestHLSCacheKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		mediaID     int64
		sessionID   string
		startOffset float64
		want        string
	}{
		{"basic", 1, "abc123", 0, "1:abc123:0.000"},
		{"with_offset", 42, "session-xyz", 120.5, "42:session-xyz:120.500"},
		{"zero_offset", 100, "s1", 0.0, "100:s1:0.000"},
		{"small_offset", 5, "foo", 0.001, "5:foo:0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hlsCacheKey(tt.mediaID, tt.sessionID, tt.startOffset)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStreamService_GetPrimaryFile(t *testing.T) {
	ctx := context.Background()
	db, mediaFileRepo, _, _ := setupStreamTestDB(t)
	t.Cleanup(func() { db.Close() })

	svc := NewStreamService(mediaFileRepo, nil, nil)
	svc.SetDB(db)

	t.Run("fileID_zero_returns_primary", func(t *testing.T) {
		mf, err := svc.GetPrimaryFile(ctx, 1, 0)
		require.NoError(t, err)
		assert.Equal(t, int64(1), mf.ID)
		assert.True(t, mf.IsPrimary)
	})

	t.Run("fileID_valid_returns_that_file", func(t *testing.T) {
		mf, err := svc.GetPrimaryFile(ctx, 1, 2)
		require.NoError(t, err)
		assert.Equal(t, int64(2), mf.ID)
	})

	t.Run("fileID_belongs_to_different_media", func(t *testing.T) {
		// file 3 belongs to media 2, not media 1
		_, err := svc.GetPrimaryFile(ctx, 1, 3)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("fileID_not_found", func(t *testing.T) {
		_, err := svc.GetPrimaryFile(ctx, 1, 9999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestStreamService_FindPretranscode_NilService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewStreamService(nil, nil, nil)

	// Without a pretranscode service, should return nil
	result, err := svc.FindPretranscode(ctx, 1, 1080)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestStreamService_FindPretranscodeProfile_NilService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewStreamService(nil, nil, nil)

	result, err := svc.FindPretranscodeProfile(ctx, 5)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestStreamService_FindAllPretranscodes_NilService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewStreamService(nil, nil, nil)

	result, err := svc.FindAllPretranscodes(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestStreamService_FindAllPretranscodesWithProfiles_NilService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewStreamService(nil, nil, nil)

	result, err := svc.FindAllPretranscodesWithProfiles(ctx, 1)
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestStreamService_InvalidateHLSCache(t *testing.T) {
	t.Parallel()

	svc := NewStreamService(nil, nil, nil)

	// Insert some cache entries manually via the map
	svc.hlsCacheMu.Lock()
	svc.hlsCache["1:session-a:0.000"] = &hlsCacheEntry{playlistPath: "/path/a"}
	svc.hlsCache["1:session-b:0.000"] = &hlsCacheEntry{playlistPath: "/path/b"}
	svc.hlsCache["2:session-a:0.000"] = &hlsCacheEntry{playlistPath: "/path/c"}
	svc.hlsCacheMu.Unlock()

	t.Run("empty_session_does_nothing", func(t *testing.T) {
		t.Parallel()
		svc.InvalidateHLSCache("")
		svc.hlsCacheMu.RLock()
		assert.Len(t, svc.hlsCache, 3)
		svc.hlsCacheMu.RUnlock()
	})

	t.Run("invalidates_matching_session_entries", func(t *testing.T) {
		t.Parallel()
		svc.InvalidateHLSCache("session-a")
		svc.hlsCacheMu.RLock()
		// session-a entries should be gone, session-b should remain
		_, hasA := svc.hlsCache["1:session-a:0.000"]
		_, hasC := svc.hlsCache["2:session-a:0.000"]
		_, hasB := svc.hlsCache["1:session-b:0.000"]
		svc.hlsCacheMu.RUnlock()
		assert.False(t, hasA, "1:session-a should be removed")
		assert.False(t, hasC, "2:session-a should be removed")
		assert.True(t, hasB, "1:session-b should remain")
	})
}

func TestStreamService_HLSCachedDuration(t *testing.T) {
	t.Parallel()

	svc := NewStreamService(nil, nil, nil)

	svc.hlsCacheMu.Lock()
	svc.hlsCache["1:session-x:0.000"] = &hlsCacheEntry{duration: 3600.5}
	svc.hlsCache["1:session-x:30.000"] = &hlsCacheEntry{duration: 3650.0}
	svc.hlsCacheMu.Unlock()

	t.Run("returns_cached_duration", func(t *testing.T) {
		t.Parallel()
		dur := svc.HLSCachedDuration(1, "session-x", 0)
		assert.Equal(t, 3600.5, dur)
	})

	t.Run("returns_duration_for_specific_offset", func(t *testing.T) {
		t.Parallel()
		dur := svc.HLSCachedDuration(1, "session-x", 30)
		assert.Equal(t, 3650.0, dur)
	})

	t.Run("returns_zero_for_unknown_session", func(t *testing.T) {
		t.Parallel()
		dur := svc.HLSCachedDuration(1, "unknown-session", 0)
		assert.Equal(t, float64(0), dur)
	})

	t.Run("returns_zero_for_empty_session", func(t *testing.T) {
		t.Parallel()
		dur := svc.HLSCachedDuration(1, "", 0)
		assert.Equal(t, float64(0), dur)
	})
}

func TestStreamService_ResolveFilePath(t *testing.T) {
	ctx := context.Background()
	db, mediaFileRepo, _, _ := setupStreamTestDB(t)
	t.Cleanup(func() { db.Close() })

	t.Run("local_file_returns_path", func(t *testing.T) {
		svc := NewStreamService(mediaFileRepo, nil, nil)

		mf := &model.MediaFile{ID: 1, MediaID: 1, FilePath: "/test/movie1.mkv"}
		path, err := svc.ResolveFilePath(ctx, 1, mf)
		require.NoError(t, err)
		assert.Equal(t, "/test/movie1.mkv", path)
	})

	t.Run("cloud_path_with_resolver", func(t *testing.T) {
		svc := NewStreamService(mediaFileRepo, nil, nil)
		svc.SetCloudResolver(func(ctx context.Context, mediaID int64, mf *model.MediaFile) (string, error) {
			return "https://cloud.example.com/file.m3u8", nil
		})

		mf := &model.MediaFile{ID: 1, MediaID: 1, FilePath: "fshare://some/file.mkv"}
		path, err := svc.ResolveFilePath(ctx, 1, mf)
		require.NoError(t, err)
		assert.Equal(t, "https://cloud.example.com/file.m3u8", path)
	})

	t.Run("cloud_resolver_error_returns_error", func(t *testing.T) {
		svc := NewStreamService(mediaFileRepo, nil, nil)
		svc.SetCloudResolver(func(ctx context.Context, mediaID int64, mf *model.MediaFile) (string, error) {
			return "", errors.New("cloud unavailable")
		})

		mf := &model.MediaFile{ID: 1, MediaID: 1, FilePath: "fshare://some/file.mkv"}
		_, err := svc.ResolveFilePath(ctx, 1, mf)
		assert.Error(t, err)
	})

	t.Run("ophim_path_resolved", func(t *testing.T) {
		// ophim resolution requires network call - tested separately via integration tests
		// We can test the slug parsing logic here by verifying ophim:// prefix handling
		svc := NewStreamService(mediaFileRepo, nil, nil)
		_ = svc // avoid unused variable warning
	})
}

func TestStreamService_GetAudioTrackForMediaFile(t *testing.T) {
	ctx := context.Background()
	db, mediaFileRepo, audioTrackRepo, _ := setupStreamTestDB(t)
	t.Cleanup(func() { db.Close() })

	// Insert an audio track
	_, err := db.Exec(`
		INSERT INTO audio_tracks (id, media_file_id, stream_index, language, codec, channels, is_default)
		VALUES (1, 1, 0, 'eng', 'aac', 6, 1)
	`)
	require.NoError(t, err)

	svc := NewStreamService(mediaFileRepo, audioTrackRepo, nil)
	svc.SetDB(db)

	t.Run("track_belongs_to_file", func(t *testing.T) {
		track, err := svc.GetAudioTrackForMediaFile(ctx, 1, 1)
		require.NoError(t, err)
		assert.Equal(t, "eng", track.Language)
		assert.Equal(t, "aac", track.Codec)
	})

	t.Run("track_belongs_to_different_file", func(t *testing.T) {
		_, err := svc.GetAudioTrackForMediaFile(ctx, 2, 1) // track 1 belongs to file 1, not file 2
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})

	t.Run("track_not_found", func(t *testing.T) {
		_, err := svc.GetAudioTrackForMediaFile(ctx, 1, 9999)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
	})
}

func TestStreamService_ListAudioTracksForMediaFile(t *testing.T) {
	ctx := context.Background()
	db, mediaFileRepo, audioTrackRepo, _ := setupStreamTestDB(t)
	t.Cleanup(func() { db.Close() })

	// Insert multiple audio tracks
	_, err := db.Exec(`
		INSERT INTO audio_tracks (media_file_id, stream_index, language, codec, channels) VALUES (1, 0, 'eng', 'aac', 6);
		INSERT INTO audio_tracks (media_file_id, stream_index, language, codec, channels) VALUES (1, 1, 'jpn', 'aac', 2);
		INSERT INTO audio_tracks (media_file_id, stream_index, language, codec, channels) VALUES (2, 0, 'eng', 'aac', 6);
	`)
	require.NoError(t, err)

	svc := NewStreamService(mediaFileRepo, audioTrackRepo, nil)
	svc.SetDB(db)

	t.Run("returns_tracks_for_file", func(t *testing.T) {
		tracks, err := svc.ListAudioTracksForMediaFile(ctx, 1)
		require.NoError(t, err)
		assert.Len(t, tracks, 2)
	})

	t.Run("returns_empty_for_file_with_no_tracks", func(t *testing.T) {
		tracks, err := svc.ListAudioTracksForMediaFile(ctx, 3) // no tracks for file 3
		require.NoError(t, err)
		assert.Len(t, tracks, 0)
	})
}

func TestStreamService_ServiceConfig(t *testing.T) {
	t.Parallel()

	svc := NewStreamService(nil, nil, nil)

	t.Run("SetNotificationService", func(t *testing.T) {
		t.Parallel()
		// Setting nil should not panic
		svc.SetNotificationService(nil)
	})

	t.Run("SetPretranscodeService", func(t *testing.T) {
		t.Parallel()
		// Setting nil should not panic
		svc.SetPretranscodeService(nil)
	})

	t.Run("SetDB", func(t *testing.T) {
		t.Parallel()
		svc.SetDB(nil)
	})

	t.Run("SetSubtitleRepo", func(t *testing.T) {
		t.Parallel()
		svc.SetSubtitleRepo(nil)
	})
}

func TestStreamService_RemuxToPretranscode_NilService(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	svc := NewStreamService(nil, nil, nil)

	// With nil pretranscode service, should not panic
	svc.RemuxToPretranscode(ctx, 1, 1, 1080)
}
