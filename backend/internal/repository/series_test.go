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

func setupSeriesTestDB(t *testing.T) (*sql.DB, *SeriesRepo, *SeasonRepo, *EpisodeRepo) {
	t.Helper()
	db, err := sql.Open("sqlite3", "file::memory:?_foreign_keys=on")
	require.NoError(t, err)

	schema := `
	CREATE TABLE series (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		library_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		sort_title TEXT DEFAULT '',
		tmdb_id INTEGER,
		imdb_id TEXT,
		tvdb_id INTEGER,
		anilist_id INTEGER,
		romaji_title TEXT DEFAULT '',
		studio TEXT DEFAULT '',
		overview TEXT DEFAULT '',
		status TEXT DEFAULT '',
		network TEXT DEFAULT '',
		first_air_date TEXT DEFAULT '',
		poster_path TEXT DEFAULT '',
		backdrop_path TEXT DEFAULT '',
		logo_path TEXT DEFAULT '',
		thumb_path TEXT DEFAULT '',
		metadata_locked INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE seasons (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
		season_number INTEGER NOT NULL,
		title TEXT DEFAULT '',
		overview TEXT DEFAULT '',
		poster_path TEXT DEFAULT '',
		episode_count INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE episodes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		series_id INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
		season_id INTEGER NOT NULL REFERENCES seasons(id) ON DELETE CASCADE,
		media_id INTEGER DEFAULT 0,
		episode_number INTEGER NOT NULL,
		title TEXT NOT NULL,
		overview TEXT DEFAULT '',
		still_path TEXT DEFAULT '',
		air_date TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE genres (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);

	CREATE TABLE media_genres (
		series_id INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
		genre_id INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
		PRIMARY KEY (series_id, genre_id)
	);

	CREATE TABLE media_files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		media_id INTEGER NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER DEFAULT 0,
		duration REAL DEFAULT 0,
		is_primary INTEGER DEFAULT 0
	);
	`
	_, err = db.Exec(schema)
	require.NoError(t, err)

	return db, NewSeriesRepo(db), NewSeasonRepo(db), NewEpisodeRepo(db)
}

// SeriesRepo Tests

func TestSeriesRepo_Create(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{
		LibraryID:    1,
		Title:        "Test Series",
		SortTitle:    "Test Series",
		Status:       "Returning Series",
		FirstAirDate: "2024-01-15",
	}

	err := repo.Create(ctx, s)
	require.NoError(t, err)
	assert.Greater(t, s.ID, int64(0))
	assert.NotEmpty(t, s.CreatedAt)
	assert.NotEmpty(t, s.UpdatedAt)
}

func TestSeriesRepo_GetByID(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Get By ID Test"}
	require.NoError(t, repo.Create(ctx, s))

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Get By ID Test", got.Title)
	assert.Equal(t, s.ID, got.ID)
	assert.Equal(t, int64(1), got.LibraryID)
}

func TestSeriesRepo_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetByID(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_Update(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Original Title", Status: "Ended"}
	require.NoError(t, repo.Create(ctx, s))

	s.Title = "Updated Title"
	s.Status = "Returning Series"
	err := repo.Update(ctx, s)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", got.Title)
	assert.Equal(t, "Returning Series", got.Status)
}

func TestSeriesRepo_Update_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{ID: 9999, Title: "Ghost"}
	err := repo.Update(ctx, s)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_UpdateMetadata(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Original"}
	require.NoError(t, repo.Create(ctx, s))

	newTitle := "Updated via Metadata"
	req := model.SeriesMetadataEditRequest{Title: &newTitle}
	err := repo.UpdateMetadata(ctx, s.ID, req)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated via Metadata", got.Title)
}

func TestSeriesRepo_UpdateMetadata_Partial(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Original", Status: "Ended", Overview: "Original overview"}
	require.NoError(t, repo.Create(ctx, s))

	newStatus := "Returning Series"
	req := model.SeriesMetadataEditRequest{Status: &newStatus}
	err := repo.UpdateMetadata(ctx, s.ID, req)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Original", got.Title)             // unchanged
	assert.Equal(t, "Returning Series", got.Status)    // changed
	assert.Equal(t, "Original overview", got.Overview) // unchanged
}

func TestSeriesRepo_UpdateMetadata_NoFields(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Test"}
	require.NoError(t, repo.Create(ctx, s))

	err := repo.UpdateMetadata(ctx, s.ID, model.SeriesMetadataEditRequest{})
	require.NoError(t, err) // no-op, should not error
}

func TestSeriesRepo_UpdateMetadata_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	newTitle := "X"
	err := repo.UpdateMetadata(ctx, 9999, model.SeriesMetadataEditRequest{Title: &newTitle})
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_UpdateImagePath(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Image Test"}
	require.NoError(t, repo.Create(ctx, s))

	err := repo.UpdateImagePath(ctx, s.ID, "poster", "/new/poster.jpg")
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "/new/poster.jpg", string(got.PosterPath))
}

func TestSeriesRepo_UpdateImagePath_Backdrop(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Backdrop Test"}
	require.NoError(t, repo.Create(ctx, s))

	err := repo.UpdateImagePath(ctx, s.ID, "backdrop", "/new/backdrop.jpg")
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "/new/backdrop.jpg", string(got.BackdropPath))
}

func TestSeriesRepo_UpdateImagePath_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.UpdateImagePath(ctx, 9999, "poster", "/path")
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_SetMetadataLocked(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Lock Test"}
	require.NoError(t, repo.Create(ctx, s))

	err := repo.SetMetadataLocked(ctx, s.ID, true)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.True(t, got.MetadataLocked)
}

func TestSeriesRepo_SetMetadataLocked_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.SetMetadataLocked(ctx, 9999, true)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_Delete(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Delete Test"}
	require.NoError(t, repo.Create(ctx, s))

	err := repo.Delete(ctx, s.ID)
	require.NoError(t, err)

	_, err = repo.GetByID(ctx, s.ID)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_Delete_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := repo.Delete(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_List(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		s := &model.Series{LibraryID: 1, Title: "Series " + string(rune('A'+i))}
		require.NoError(t, repo.Create(ctx, s))
	}

	all, err := repo.List(ctx, 0, 0, 0)
	require.NoError(t, err)
	assert.Len(t, all, 3)

	byLib, err := repo.List(ctx, 1, 10, 0)
	require.NoError(t, err)
	assert.Len(t, byLib, 3)

	empty, err := repo.List(ctx, 999, 10, 0)
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestSeriesRepo_List_Pagination(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		s := &model.Series{LibraryID: 1, Title: "Series"}
		require.NoError(t, repo.Create(ctx, s))
	}

	first, err := repo.List(ctx, 0, 2, 0)
	require.NoError(t, err)
	assert.Len(t, first, 2)

	second, err := repo.List(ctx, 0, 2, 2)
	require.NoError(t, err)
	assert.Len(t, second, 2)

	third, err := repo.List(ctx, 0, 2, 4)
	require.NoError(t, err)
	assert.Len(t, third, 1)
}

func TestSeriesRepo_GetByTmdbID(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	tmdbID := int64(12345)
	s := &model.Series{LibraryID: 1, Title: "TMDB Test"}
	require.NoError(t, repo.Create(ctx, s))

	// Update to set tmdb_id
	_, err := db.ExecContext(ctx, "UPDATE series SET tmdb_id = ? WHERE id = ?", tmdbID, s.ID)
	require.NoError(t, err)

	got, err := repo.GetByTmdbID(ctx, 1, tmdbID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestSeriesRepo_GetByTmdbID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetByTmdbID(ctx, 1, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_GetByTvdbID(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	tvdbID := int64(67890)
	s := &model.Series{LibraryID: 1, Title: "TVDB Test"}
	require.NoError(t, repo.Create(ctx, s))

	_, err := db.ExecContext(ctx, "UPDATE series SET tvdb_id = ? WHERE id = ?", tvdbID, s.ID)
	require.NoError(t, err)

	got, err := repo.GetByTvdbID(ctx, 1, tvdbID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestSeriesRepo_GetByAnilistID(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	anilistID := int64(11111)
	s := &model.Series{LibraryID: 1, Title: "Anilist Test"}
	require.NoError(t, repo.Create(ctx, s))

	_, err := db.ExecContext(ctx, "UPDATE series SET anilist_id = ? WHERE id = ?", anilistID, s.ID)
	require.NoError(t, err)

	got, err := repo.GetByAnilistID(ctx, 1, anilistID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestSeriesRepo_GetByAnilistID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetByAnilistID(ctx, 1, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_GetByImdbID(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	imdbID := "tt1234567"
	s := &model.Series{LibraryID: 1, Title: "IMDb Test"}
	require.NoError(t, repo.Create(ctx, s))

	_, err := db.ExecContext(ctx, "UPDATE series SET imdb_id = ? WHERE id = ?", imdbID, s.ID)
	require.NoError(t, err)

	got, err := repo.GetByImdbID(ctx, 1, imdbID)
	require.NoError(t, err)
	assert.Equal(t, s.ID, got.ID)
}

func TestSeriesRepo_GetByImdbID_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := repo.GetByImdbID(ctx, 1, "tt9999999")
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeriesRepo_Search(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "Breaking Bad"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "Better Call Saul"}
	require.NoError(t, repo.Create(ctx, s2))
	s3 := &model.Series{LibraryID: 1, Title: "The Office"}
	require.NoError(t, repo.Create(ctx, s3))

	results, err := repo.Search(ctx, "Bad", 10)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Breaking Bad", results[0].Title)
}

func TestSeriesRepo_Search_NoResults(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Breaking Bad"}
	require.NoError(t, repo.Create(ctx, s))

	results, err := repo.Search(ctx, " nonexistent ", 10)
	require.NoError(t, err)
	assert.Len(t, results, 0)
}

func TestSeriesRepo_Search_Limit(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		s := &model.Series{LibraryID: 1, Title: "Series " + string(rune('A'+i))}
		require.NoError(t, repo.Create(ctx, s))
	}

	results, err := repo.Search(ctx, "Series", 3)
	require.NoError(t, err)
	assert.Len(t, results, 3)
}

func TestSeriesRepo_ListFiltered(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Filtered Series", SortTitle: "Filtered Series"}
	require.NoError(t, repo.Create(ctx, s))

	// Add genre
	_, err := db.ExecContext(ctx, "INSERT INTO genres (name) VALUES ('Drama')")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO media_genres (series_id, genre_id) VALUES (?, 1)", s.ID)
	require.NoError(t, err)

	filter := model.SeriesListFilter{LibraryID: 1, Limit: 10}
	results, err := repo.ListFiltered(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Filtered Series", results[0].Title)
	assert.Contains(t, results[0].Genres, "Drama")
}

func TestSeriesRepo_ListFiltered_WithSearch(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "Breaking Bad", SortTitle: "Breaking Bad"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "Better Call Saul", SortTitle: "Better Call Saul"}
	require.NoError(t, repo.Create(ctx, s2))

	filter := model.SeriesListFilter{Search: "Breaking", Limit: 10}
	results, err := repo.ListFiltered(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Breaking Bad", results[0].Title)
}

func TestSeriesRepo_ListFiltered_WithGenre(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "Drama Show", SortTitle: "Drama Show"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "Comedy Show", SortTitle: "Comedy Show"}
	require.NoError(t, repo.Create(ctx, s2))

	_, err := db.ExecContext(ctx, "INSERT INTO genres (id, name) VALUES (1, 'Drama'), (2, 'Comedy')")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO media_genres (series_id, genre_id) VALUES (1, 1), (2, 2)")
	require.NoError(t, err)

	filter := model.SeriesListFilter{Genre: "Drama", Limit: 10}
	results, err := repo.ListFiltered(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "Drama Show", results[0].Title)
}

func TestSeriesRepo_ListFiltered_WithYear(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "Old Show", SortTitle: "Old Show", FirstAirDate: "2020-01-01"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "New Show", SortTitle: "New Show", FirstAirDate: "2024-06-15"}
	require.NoError(t, repo.Create(ctx, s2))

	filter := model.SeriesListFilter{Year: "2024", Limit: 10}
	results, err := repo.ListFiltered(ctx, filter)
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "New Show", results[0].Title)
}

func TestSeriesRepo_ListFiltered_SortOrders(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "First", SortTitle: "First", FirstAirDate: "2020-01-01"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "Second", SortTitle: "Second", FirstAirDate: "2021-06-01"}
	require.NoError(t, repo.Create(ctx, s2))

	// Test newest
	newest, err := repo.ListFiltered(ctx, model.SeriesListFilter{Limit: 10, Sort: "newest"})
	require.NoError(t, err)
	assert.Equal(t, "Second", newest[0].Title)

	// Test oldest
	oldest, err := repo.ListFiltered(ctx, model.SeriesListFilter{Limit: 10, Sort: "oldest"})
	require.NoError(t, err)
	assert.Equal(t, "First", oldest[0].Title)

	// Test title sort
	byTitle, err := repo.ListFiltered(ctx, model.SeriesListFilter{Limit: 10, Sort: "title"})
	require.NoError(t, err)
	assert.Equal(t, "First", byTitle[0].Title)
}

func TestSeriesRepo_GetAlphabet(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s1 := &model.Series{LibraryID: 1, Title: "Alpha", SortTitle: "Alpha"}
	require.NoError(t, repo.Create(ctx, s1))
	s2 := &model.Series{LibraryID: 1, Title: "Beta", SortTitle: "Beta"}
	require.NoError(t, repo.Create(ctx, s2))
	s3 := &model.Series{LibraryID: 1, Title: "123 Numbers", SortTitle: "123 Numbers"}
	require.NoError(t, repo.Create(ctx, s3))

	counts, err := repo.GetAlphabet(ctx, model.SeriesListFilter{LibraryID: 1})
	require.NoError(t, err)

	found := false
	for _, c := range counts {
		if c.Letter == "A" && c.Count == 1 {
			found = true
			break
		}
	}
	assert.True(t, found, "should have A=1")
}

func TestSeriesRepo_WithTx(t *testing.T) {
	t.Parallel()
	db, repo, _, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	txRepo := repo.WithTx(tx)
	s := &model.Series{LibraryID: 1, Title: "Transaction Test"}
	err = txRepo.Create(ctx, s)
	require.NoError(t, err)

	err = tx.Commit()
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, s.ID)
	require.NoError(t, err)
	assert.Equal(t, "Transaction Test", got.Title)
}

// SeasonRepo Tests

func TestSeasonRepo_Create(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series for Season"}
	require.NoError(t, repo.Create(ctx, s))

	season := &model.Season{
		SeriesID:     s.ID,
		SeasonNumber: 1,
		Title:        "Season 1",
		EpisodeCount: 10,
	}
	err := seasonRepo.Create(ctx, season)
	require.NoError(t, err)
	assert.Greater(t, season.ID, int64(0))
	assert.NotEmpty(t, season.CreatedAt)
}

func TestSeasonRepo_GetByID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1, Title: "First Season"}
	require.NoError(t, seasonRepo.Create(ctx, season))

	got, err := seasonRepo.GetByID(ctx, season.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Season", got.Title)
	assert.Equal(t, 1, got.SeasonNumber)
}

func TestSeasonRepo_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, _, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := seasonRepo.GetByID(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeasonRepo_GetBySeriesAndNumber(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	season := &model.Season{SeriesID: s.ID, SeasonNumber: 2, Title: "Season 2"}
	require.NoError(t, seasonRepo.Create(ctx, season))

	got, err := seasonRepo.GetBySeriesAndNumber(ctx, s.ID, 2)
	require.NoError(t, err)
	assert.Equal(t, season.ID, got.ID)
	assert.Equal(t, 2, got.SeasonNumber)
}

func TestSeasonRepo_GetBySeriesAndNumber_NotFound(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	_, err := seasonRepo.GetBySeriesAndNumber(ctx, s.ID, 99)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeasonRepo_Update(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1, Title: "Old Title", EpisodeCount: 5}
	require.NoError(t, seasonRepo.Create(ctx, season))

	season.Title = "Updated Title"
	season.EpisodeCount = 12
	err := seasonRepo.Update(ctx, season)
	require.NoError(t, err)

	got, err := seasonRepo.GetByID(ctx, season.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", got.Title)
	assert.Equal(t, 12, got.EpisodeCount)
}

func TestSeasonRepo_Update_NotFound(t *testing.T) {
	t.Parallel()
	db, _, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	season := &model.Season{ID: 9999, SeasonNumber: 1, Title: "Ghost"}
	err := seasonRepo.Update(ctx, season)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeasonRepo_Delete(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1, Title: "To Delete"}
	require.NoError(t, seasonRepo.Create(ctx, season))

	err := seasonRepo.Delete(ctx, season.ID)
	require.NoError(t, err)

	_, err = seasonRepo.GetByID(ctx, season.ID)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeasonRepo_Delete_NotFound(t *testing.T) {
	t.Parallel()
	db, _, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := seasonRepo.Delete(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestSeasonRepo_ListBySeriesID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	for i := 1; i <= 3; i++ {
		season := &model.Season{SeriesID: s.ID, SeasonNumber: i, Title: "Season"}
		require.NoError(t, seasonRepo.Create(ctx, season))
	}

	seasons, err := seasonRepo.ListBySeriesID(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, seasons, 3)
	// Ordered by season_number
	assert.Equal(t, 1, seasons[0].SeasonNumber)
	assert.Equal(t, 2, seasons[1].SeasonNumber)
}

func TestSeasonRepo_ListBySeriesID_Empty(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, _ := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))

	seasons, err := seasonRepo.ListBySeriesID(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, seasons, 0)
}

// EpisodeRepo Tests

func TestEpisodeRepo_Create(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))

	ep := &model.Episode{
		SeriesID:      s.ID,
		SeasonID:      season.ID,
		EpisodeNumber: 1,
		Title:         "Pilot",
		AirDate:       "2024-01-15",
	}
	err := episodeRepo.Create(ctx, ep)
	require.NoError(t, err)
	assert.Greater(t, ep.ID, int64(0))
	assert.NotEmpty(t, ep.CreatedAt)
}

func TestEpisodeRepo_GetByID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))
	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 1, Title: "First Episode"}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	got, err := episodeRepo.GetByID(ctx, ep.ID)
	require.NoError(t, err)
	assert.Equal(t, "First Episode", got.Title)
	assert.Equal(t, 1, got.EpisodeNumber)
}

func TestEpisodeRepo_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := episodeRepo.GetByID(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_GetByMediaID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))
	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 1, Title: "Ep", MediaID: 100}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	got, err := episodeRepo.GetByMediaID(ctx, 100)
	require.NoError(t, err)
	assert.Equal(t, ep.ID, got.ID)
}

func TestEpisodeRepo_GetByMediaID_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := episodeRepo.GetByMediaID(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_GetBySeasonAndNumber(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))
	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 5, Title: "Ep 5"}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	got, err := episodeRepo.GetBySeasonAndNumber(ctx, season.ID, 5)
	require.NoError(t, err)
	assert.Equal(t, ep.ID, got.ID)
}

func TestEpisodeRepo_GetBySeasonAndNumber_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	_, err := episodeRepo.GetBySeasonAndNumber(ctx, 1, 99)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_Update(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))
	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 1, Title: "Old Title", Overview: "Old"}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	ep.Title = "New Title"
	ep.Overview = "New overview"
	err := episodeRepo.Update(ctx, ep)
	require.NoError(t, err)

	got, err := episodeRepo.GetByID(ctx, ep.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Title", got.Title)
	assert.Equal(t, "New overview", got.Overview)
}

func TestEpisodeRepo_Update_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	ep := &model.Episode{ID: 9999, EpisodeNumber: 1, Title: "Ghost"}
	err := episodeRepo.Update(ctx, ep)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_UpdateSeasonLink(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season1 := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season1))
	season2 := &model.Season{SeriesID: s.ID, SeasonNumber: 2}
	require.NoError(t, seasonRepo.Create(ctx, season2))

	ep := &model.Episode{SeriesID: s.ID, SeasonID: season1.ID, EpisodeNumber: 1, Title: "Ep"}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	err := episodeRepo.UpdateSeasonLink(ctx, ep.ID, season2.ID, 2)
	require.NoError(t, err)

	got, err := episodeRepo.GetByID(ctx, ep.ID)
	require.NoError(t, err)
	assert.Equal(t, season2.ID, got.SeasonID)
	assert.Equal(t, 2, got.EpisodeNumber)
}

func TestEpisodeRepo_UpdateSeasonLink_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := episodeRepo.UpdateSeasonLink(ctx, 9999, 1, 1)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_Delete(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))
	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 1, Title: "To Delete"}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	err := episodeRepo.Delete(ctx, ep.ID)
	require.NoError(t, err)

	_, err = episodeRepo.GetByID(ctx, ep.ID)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_Delete_NotFound(t *testing.T) {
	t.Parallel()
	db, _, _, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	err := episodeRepo.Delete(ctx, 9999)
	assert.Error(t, err) // method does not wrap to ErrNotFound
}

func TestEpisodeRepo_ListBySeasonID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))

	for i := 1; i <= 5; i++ {
		ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: i, Title: "Ep"}
		require.NoError(t, episodeRepo.Create(ctx, ep))
	}

	episodes, err := episodeRepo.ListBySeasonID(ctx, season.ID)
	require.NoError(t, err)
	assert.Len(t, episodes, 5)
	assert.Equal(t, 1, episodes[0].EpisodeNumber)
}

func TestEpisodeRepo_ListBySeriesID(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))

	for i := 1; i <= 3; i++ {
		ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: i, Title: "Ep"}
		require.NoError(t, episodeRepo.Create(ctx, ep))
	}

	episodes, err := episodeRepo.ListBySeriesID(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, episodes, 3)
}

func TestEpisodeRepo_ListBySeriesID_WithDuration(t *testing.T) {
	t.Parallel()
	db, repo, seasonRepo, episodeRepo := setupSeriesTestDB(t)
	t.Cleanup(func() { db.Close() })

	ctx := context.Background()
	s := &model.Series{LibraryID: 1, Title: "Series"}
	require.NoError(t, repo.Create(ctx, s))
	season := &model.Season{SeriesID: s.ID, SeasonNumber: 1}
	require.NoError(t, seasonRepo.Create(ctx, season))

	ep := &model.Episode{SeriesID: s.ID, SeasonID: season.ID, EpisodeNumber: 1, Title: "Ep", MediaID: 100}
	require.NoError(t, episodeRepo.Create(ctx, ep))

	// Add primary media file with duration
	_, err := db.ExecContext(ctx,
		"INSERT INTO media_files (media_id, file_path, duration, is_primary) VALUES (100, '/path.mp4', 3600.5, 1)")
	require.NoError(t, err)

	episodes, err := episodeRepo.ListBySeriesID(ctx, s.ID)
	require.NoError(t, err)
	assert.Len(t, episodes, 1)
	assert.Equal(t, 3600.5, episodes[0].Duration)
}
