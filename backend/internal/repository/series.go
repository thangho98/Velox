package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/thawng/velox/internal/model"
)

// SeriesRepo handles series database operations
type SeriesRepo struct {
	db DBTX
}

func NewSeriesRepo(db DBTX) *SeriesRepo {
	return &SeriesRepo{db: db}
}

// Create inserts a new series
func (r *SeriesRepo) Create(ctx context.Context, s *model.Series) error {
	query := `INSERT INTO series
		(library_id, title, sort_title, tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date,
		 poster_path, backdrop_path, logo_path, thumb_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		s.LibraryID, s.Title, s.SortTitle, s.TmdbID, s.ImdbID, s.TvdbID, s.AnilistID, s.RomajiTitle, s.Studio,
		s.Overview, s.Status, s.Network, s.FirstAirDate, s.PosterPath, s.BackdropPath,
		s.LogoPath, s.ThumbPath)

	return row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

// GetByID retrieves a series by ID.
// Returns ErrNotFound if the series ID does not exist.
func (r *SeriesRepo) GetByID(ctx context.Context, id int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE id = ?`, id).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("getting series %d: %w", id, err)
	}
	return &s, nil
}

// Update updates a series (full update — used by metadata enrichment pipeline).
func (r *SeriesRepo) Update(ctx context.Context, s *model.Series) error {
	res, err := r.db.ExecContext(ctx, `UPDATE series SET
		title = ?, sort_title = ?, tmdb_id = ?, imdb_id = ?, tvdb_id = ?, anilist_id = ?, romaji_title = ?, studio = ?,
		overview = ?, status = ?, network = ?, first_air_date = ?,
		poster_path = ?, backdrop_path = ?, logo_path = ?, thumb_path = ?,
		metadata_locked = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		s.Title, s.SortTitle, s.TmdbID, s.ImdbID, s.TvdbID, s.AnilistID, s.RomajiTitle, s.Studio,
		s.Overview, s.Status, s.Network, s.FirstAirDate,
		s.PosterPath, s.BackdropPath, s.LogoPath, s.ThumbPath,
		s.MetadataLocked, s.ID)
	if err != nil {
		return fmt.Errorf("update series %d: %w", s.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateMetadata performs a partial metadata update for a series.
// Returns ErrNotFound if the series ID does not exist.
func (r *SeriesRepo) UpdateMetadata(ctx context.Context, id int64, req model.SeriesMetadataEditRequest) error {
	setClauses := []string{}
	args := []any{}

	if req.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *req.Title)
	}
	if req.SortTitle != nil {
		setClauses = append(setClauses, "sort_title = ?")
		args = append(args, *req.SortTitle)
	}
	if req.Overview != nil {
		setClauses = append(setClauses, "overview = ?")
		args = append(args, *req.Overview)
	}
	if req.Status != nil {
		setClauses = append(setClauses, "status = ?")
		args = append(args, *req.Status)
	}
	if req.Network != nil {
		setClauses = append(setClauses, "network = ?")
		args = append(args, *req.Network)
	}
	if req.FirstAirDate != nil {
		setClauses = append(setClauses, "first_air_date = ?")
		args = append(args, *req.FirstAirDate)
	}
	if req.MetadataLocked != nil {
		setClauses = append(setClauses, "metadata_locked = ?")
		args = append(args, *req.MetadataLocked)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	query := fmt.Sprintf("UPDATE series SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("updating series %d: %w", id, err)
	}
	return checkRowsAffected(res)
}

// UpdateImagePath updates poster_path or backdrop_path for a series.
// Returns ErrNotFound if the series ID does not exist.
func (r *SeriesRepo) UpdateImagePath(ctx context.Context, id int64, imageType, path string) error {
	col := "poster_path"
	if imageType == "backdrop" {
		col = "backdrop_path"
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf("UPDATE series SET %s = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", col), path, id)
	if err != nil {
		return fmt.Errorf("updating image path for series %d: %w", id, err)
	}
	return checkRowsAffected(res)
}

// SetMetadataLocked sets the metadata_locked flag for a series.
func (r *SeriesRepo) SetMetadataLocked(ctx context.Context, id int64, locked bool) error {
	res, err := r.db.ExecContext(ctx, "UPDATE series SET metadata_locked = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", locked, id)
	if err != nil {
		return fmt.Errorf("setting metadata locked for series %d: %w", id, err)
	}
	return checkRowsAffected(res)
}

// checkRowsAffected returns ErrNotFound when an UPDATE hit zero rows.
func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a series and its seasons/episodes (CASCADE)
func (r *SeriesRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM series WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete series %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves series with optional filters
func (r *SeriesRepo) List(ctx context.Context, libraryID int64, limit, offset int) ([]model.Series, error) {
	query := `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE 1=1`
	args := []any{}

	if libraryID > 0 {
		query += " AND library_id = ?"
		args = append(args, libraryID)
	}

	query += " ORDER BY sort_title"

	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing series: %w", err)
	}
	defer rows.Close()

	items := []model.Series{}
	for rows.Next() {
		var s model.Series
		if err := rows.Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning series: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// GetByTmdbID retrieves series by TMDb ID and Library ID
func (r *SeriesRepo) GetByTmdbID(ctx context.Context, libraryID int64, tmdbID int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE library_id = ? AND tmdb_id = ?`, libraryID, tmdbID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByTvdbID retrieves series by TheTVDB ID and Library ID
func (r *SeriesRepo) GetByTvdbID(ctx context.Context, libraryID int64, tvdbID int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE library_id = ? AND tvdb_id = ?`, libraryID, tvdbID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByAnilistID retrieves series by AniList ID and Library ID
func (r *SeriesRepo) GetByAnilistID(ctx context.Context, libraryID int64, anilistID int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE library_id = ? AND anilist_id = ?`, libraryID, anilistID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByImdbID retrieves series by IMDb ID and Library ID
func (r *SeriesRepo) GetByImdbID(ctx context.Context, libraryID int64, imdbID string) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE library_id = ? AND imdb_id = ?`, libraryID, imdbID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Search searches series by title
func (r *SeriesRepo) Search(ctx context.Context, query string, limit int) ([]model.Series, error) {
	q := `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, tvdb_id, anilist_id, romaji_title, studio, overview, status, network, first_air_date, poster_path, backdrop_path, logo_path, thumb_path,
		metadata_locked, created_at, updated_at
		FROM series WHERE title LIKE ? OR sort_title LIKE ?
		ORDER BY sort_title LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("searching series: %w", err)
	}
	defer rows.Close()

	items := []model.Series{}
	for rows.Next() {
		var s model.Series
		if err := rows.Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.TvdbID, &s.AnilistID, &s.RomajiTitle, &s.Studio, &s.Overview, &s.Status, &s.Network, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.LogoPath, &s.ThumbPath, &s.MetadataLocked, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning series: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// ListFiltered retrieves series with advanced filtering, sorting, and pagination.
// Supports filtering by library, search query, genre, and year.
func (r *SeriesRepo) ListFiltered(ctx context.Context, f model.SeriesListFilter) ([]model.SeriesListItem, error) {
	query := `SELECT s.id, s.library_id, s.title, s.sort_title,
		s.tmdb_id, s.imdb_id, s.tvdb_id, s.anilist_id, s.romaji_title, s.studio,
		s.overview, s.status, s.network, s.first_air_date,
		s.poster_path, s.backdrop_path, s.logo_path, s.thumb_path,
		s.metadata_locked, s.created_at, s.updated_at,
		GROUP_CONCAT(DISTINCT g.name) as genre_names,
		COUNT(DISTINCT sea.id) as season_count,
		COALESCE(SUM(sea.episode_count), 0) as episode_count
		FROM series s
		LEFT JOIN media_genres mg ON mg.series_id = s.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN seasons sea ON sea.series_id = s.id
		WHERE 1=1`
	args := []any{}

	// Library filter
	if f.LibraryID > 0 {
		query += " AND s.library_id = ?"
		args = append(args, f.LibraryID)
	}

	// Search filter (LIKE on title OR sort_title)
	if f.Search != "" {
		query += " AND (s.title LIKE ? OR s.sort_title LIKE ?)"
		pattern := "%" + f.Search + "%"
		args = append(args, pattern, pattern)
	}

	// Genre filter using EXISTS subquery for exact match
	if f.Genre != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM media_genres mg2
			JOIN genres g2 ON g2.id = mg2.genre_id
			WHERE mg2.series_id = s.id AND g2.name = ?
		)`
		args = append(args, f.Genre)
	}

	// Year filter (extract year from first_air_date)
	if f.Year != "" {
		query += " AND s.first_air_date LIKE ?"
		args = append(args, f.Year+"%")
	}

	// StartChar filter
	if f.StartChar != "" && f.StartChar != "#" {
		query += " AND UPPER(s.sort_title) >= ?"
		args = append(args, strings.ToUpper(f.StartChar))
	}

	query += " GROUP BY s.id"

	// Sort order
	switch f.Sort {
	case "newest":
		query += " ORDER BY s.first_air_date DESC, s.sort_title ASC"
	case "oldest":
		query += " ORDER BY s.first_air_date ASC, s.sort_title ASC"
	case "rating":
		// series table has no rating column — fall back to newest
		query += " ORDER BY s.first_air_date DESC, s.sort_title ASC"
	case "title":
		query += " ORDER BY s.sort_title ASC"
	default:
		query += " ORDER BY s.sort_title ASC"
	}

	// Pagination
	if f.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, f.Limit)
	}
	if f.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, f.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing filtered series: %w", err)
	}
	defer rows.Close()

	results := []model.SeriesListItem{}
	for rows.Next() {
		var item model.SeriesListItem
		var genreNames sql.NullString
		if err := rows.Scan(
			&item.ID, &item.LibraryID, &item.Title, &item.SortTitle,
			&item.TmdbID, &item.ImdbID, &item.TvdbID, &item.AnilistID, &item.RomajiTitle, &item.Studio,
			&item.Overview, &item.Status, &item.Network, &item.FirstAirDate,
			&item.PosterPath, &item.BackdropPath, &item.LogoPath, &item.ThumbPath,
			&item.MetadataLocked, &item.CreatedAt, &item.UpdatedAt,
			&genreNames, &item.SeasonCount, &item.EpisodeCount,
		); err != nil {
			return nil, fmt.Errorf("scanning filtered series: %w", err)
		}

		// Handle NULL or empty genre list
		if genreNames.Valid && genreNames.String != "" {
			item.Genres = strings.Split(genreNames.String, ",")
		}

		results = append(results, item)
	}
	return results, rows.Err()
}

// GetAlphabet returns the count of series items for each starting letter
func (r *SeriesRepo) GetAlphabet(ctx context.Context, f model.SeriesListFilter) ([]model.AlphabetCount, error) {
	query := `SELECT 
		(CASE 
			WHEN UPPER(SUBSTR(s.sort_title, 1, 1)) BETWEEN 'A' AND 'Z' 
			THEN UPPER(SUBSTR(s.sort_title, 1, 1)) 
			ELSE '#' END) as letter, 
		COUNT(DISTINCT s.id) as count 
		FROM series s
		LEFT JOIN media_genres mg ON mg.series_id = s.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		WHERE 1=1`
	args := []any{}

	if f.LibraryID > 0 {
		query += " AND s.library_id = ?"
		args = append(args, f.LibraryID)
	}
	if f.Genre != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM media_genres mg2
			JOIN genres g2 ON g2.id = mg2.genre_id
			WHERE mg2.series_id = s.id AND g2.name = ?
		)`
		args = append(args, f.Genre)
	}
	if f.Year != "" {
		query += " AND s.first_air_date LIKE ?"
		args = append(args, f.Year+"%")
	}

	query += " GROUP BY letter ORDER BY letter"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting series alphabet: %w", err)
	}
	defer rows.Close()

	var results []model.AlphabetCount
	for rows.Next() {
		var item model.AlphabetCount
		if err := rows.Scan(&item.Letter, &item.Count); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	return results, rows.Err()
}

// SeasonRepo handles seasons database operations
type SeasonRepo struct {
	db DBTX
}

func NewSeasonRepo(db DBTX) *SeasonRepo {
	return &SeasonRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *SeasonRepo) WithTx(tx *sql.Tx) *SeasonRepo {
	return &SeasonRepo{db: tx}
}

// Create inserts a new season
func (r *SeasonRepo) Create(ctx context.Context, s *model.Season) error {
	query := `INSERT INTO seasons
		(series_id, season_number, title, overview, poster_path, episode_count)
		VALUES (?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query,
		s.SeriesID, s.SeasonNumber, s.Title, s.Overview, s.PosterPath, s.EpisodeCount)

	return row.Scan(&s.ID, &s.CreatedAt)
}

// GetByID retrieves a season by ID
func (r *SeasonRepo) GetByID(ctx context.Context, id int64) (*model.Season, error) {
	var s model.Season
	err := r.db.QueryRowContext(ctx, `SELECT id, series_id, season_number, title,
		overview, poster_path, episode_count, created_at
		FROM seasons WHERE id = ?`, id).
		Scan(&s.ID, &s.SeriesID, &s.SeasonNumber, &s.Title,
			&s.Overview, &s.PosterPath, &s.EpisodeCount, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get season by id %d: %w", id, err)
	}
	return &s, nil
}

// GetBySeriesAndNumber retrieves a season by series ID and season number
func (r *SeasonRepo) GetBySeriesAndNumber(ctx context.Context, seriesID int64, seasonNumber int) (*model.Season, error) {
	var s model.Season
	err := r.db.QueryRowContext(ctx, `SELECT id, series_id, season_number, title,
		overview, poster_path, episode_count, created_at
		FROM seasons WHERE series_id = ? AND season_number = ?`, seriesID, seasonNumber).
		Scan(&s.ID, &s.SeriesID, &s.SeasonNumber, &s.Title,
			&s.Overview, &s.PosterPath, &s.EpisodeCount, &s.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get season by series %d number %d: %w", seriesID, seasonNumber, err)
	}
	return &s, nil
}

// Update updates a season
func (r *SeasonRepo) Update(ctx context.Context, s *model.Season) error {
	res, err := r.db.ExecContext(ctx, `UPDATE seasons SET
		season_number = ?, title = ?, overview = ?, poster_path = ?, episode_count = ?
		WHERE id = ?`,
		s.SeasonNumber, s.Title, s.Overview, s.PosterPath, s.EpisodeCount, s.ID)
	if err != nil {
		return fmt.Errorf("update season %d: %w", s.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a season (episodes will be deleted by CASCADE)
func (r *SeasonRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM seasons WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete season %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListBySeriesID retrieves all seasons for a series
func (r *SeasonRepo) ListBySeriesID(ctx context.Context, seriesID int64) ([]model.Season, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, series_id, season_number, title,
		overview, poster_path, episode_count, created_at
		FROM seasons WHERE series_id = ? ORDER BY season_number`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing seasons: %w", err)
	}
	defer rows.Close()

	items := []model.Season{}
	for rows.Next() {
		var s model.Season
		if err := rows.Scan(&s.ID, &s.SeriesID, &s.SeasonNumber, &s.Title,
			&s.Overview, &s.PosterPath, &s.EpisodeCount, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning season: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// EpisodeRepo handles episodes database operations
type EpisodeRepo struct {
	db DBTX
}

func NewEpisodeRepo(db DBTX) *EpisodeRepo {
	return &EpisodeRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *EpisodeRepo) WithTx(tx *sql.Tx) *EpisodeRepo {
	return &EpisodeRepo{db: tx}
}

// Create inserts a new episode
func (r *EpisodeRepo) Create(ctx context.Context, e *model.Episode) error {
	query := `INSERT INTO episodes
		(series_id, season_id, media_id, episode_number, title, overview, still_path, air_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at`

	row := r.db.QueryRowContext(ctx, query,
		e.SeriesID, e.SeasonID, e.MediaID, e.EpisodeNumber, e.Title, e.Overview, e.StillPath, e.AirDate)

	return row.Scan(&e.ID, &e.CreatedAt)
}

// GetByID retrieves an episode by ID
func (r *EpisodeRepo) GetByID(ctx context.Context, id int64) (*model.Episode, error) {
	var e model.Episode
	err := r.db.QueryRowContext(ctx, `SELECT id, series_id, season_id, media_id,
		episode_number, title, overview, still_path, air_date, created_at
		FROM episodes WHERE id = ?`, id).
		Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get episode by id %d: %w", id, err)
	}
	return &e, nil
}

// GetByMediaID retrieves an episode by its media ID
func (r *EpisodeRepo) GetByMediaID(ctx context.Context, mediaID int64) (*model.Episode, error) {
	var e model.Episode
	err := r.db.QueryRowContext(ctx, `SELECT id, series_id, season_id, media_id,
		episode_number, title, overview, still_path, air_date, created_at
		FROM episodes WHERE media_id = ?`, mediaID).
		Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get episode by media_id %d: %w", mediaID, err)
	}
	return &e, nil
}

// GetBySeasonAndNumber retrieves an episode by season ID and episode number
func (r *EpisodeRepo) GetBySeasonAndNumber(ctx context.Context, seasonID int64, episodeNumber int) (*model.Episode, error) {
	var e model.Episode
	err := r.db.QueryRowContext(ctx, `SELECT id, series_id, season_id, media_id,
		episode_number, title, overview, still_path, air_date, created_at
		FROM episodes WHERE season_id = ? AND episode_number = ?`, seasonID, episodeNumber).
		Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get episode by season %d number %d: %w", seasonID, episodeNumber, err)
	}
	return &e, nil
}

// Update updates an episode
func (r *EpisodeRepo) Update(ctx context.Context, e *model.Episode) error {
	res, err := r.db.ExecContext(ctx, `UPDATE episodes SET
		episode_number = ?, title = ?, overview = ?, still_path = ?, air_date = ?
		WHERE id = ?`,
		e.EpisodeNumber, e.Title, e.Overview, e.StillPath, e.AirDate, e.ID)
	if err != nil {
		return fmt.Errorf("update episode %d: %w", e.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateSeasonLink updates the season and episode number for an episode record.
func (r *EpisodeRepo) UpdateSeasonLink(ctx context.Context, id, seasonID int64, episodeNumber int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE episodes SET season_id = ?, episode_number = ? WHERE id = ?`,
		seasonID, episodeNumber, id)
	if err != nil {
		return fmt.Errorf("update season link for episode %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an episode
func (r *EpisodeRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM episodes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete episode %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListBySeasonID retrieves all episodes for a season, with duration from primary media_file
func (r *EpisodeRepo) ListBySeasonID(ctx context.Context, seasonID int64) ([]model.Episode, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id, e.series_id, e.season_id, e.media_id,
		e.episode_number, e.title, e.overview, e.still_path, e.air_date, e.created_at,
		COALESCE(mf.duration, 0)
		FROM episodes e
		LEFT JOIN media_files mf ON mf.media_id = e.media_id AND mf.is_primary = 1
		WHERE e.season_id = ? ORDER BY e.episode_number`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("listing episodes: %w", err)
	}
	defer rows.Close()

	items := []model.Episode{}
	for rows.Next() {
		var e model.Episode
		if err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt,
			&e.Duration); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// ListBySeriesID retrieves all episodes for a series, with duration from primary media_file
func (r *EpisodeRepo) ListBySeriesID(ctx context.Context, seriesID int64) ([]model.Episode, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id, e.series_id, e.season_id, e.media_id,
		e.episode_number, e.title, e.overview, e.still_path, e.air_date, e.created_at,
		COALESCE(mf.duration, 0)
		FROM episodes e
		LEFT JOIN media_files mf ON mf.media_id = e.media_id AND mf.is_primary = 1
		WHERE e.series_id = ? ORDER BY e.season_id, e.episode_number`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing episodes by series: %w", err)
	}
	defer rows.Close()

	items := []model.Episode{}
	for rows.Next() {
		var e model.Episode
		if err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt,
			&e.Duration); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// cache bust 1
