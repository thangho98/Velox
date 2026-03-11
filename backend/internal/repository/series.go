package repository

import (
	"context"
	"database/sql"
	"fmt"

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
		(library_id, title, sort_title, tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		s.LibraryID, s.Title, s.SortTitle, s.TmdbID, s.ImdbID,
		s.Overview, s.Status, s.FirstAirDate, s.PosterPath, s.BackdropPath)

	return row.Scan(&s.ID, &s.CreatedAt, &s.UpdatedAt)
}

// GetByID retrieves a series by ID
func (r *SeriesRepo) GetByID(ctx context.Context, id int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path,
		created_at, updated_at
		FROM series WHERE id = ?`, id).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.Overview, &s.Status, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates a series
func (r *SeriesRepo) Update(ctx context.Context, s *model.Series) error {
	_, err := r.db.ExecContext(ctx, `UPDATE series SET
		title = ?, sort_title = ?, tmdb_id = ?, imdb_id = ?,
		overview = ?, status = ?, first_air_date = ?, poster_path = ?, backdrop_path = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		s.Title, s.SortTitle, s.TmdbID, s.ImdbID,
		s.Overview, s.Status, s.FirstAirDate, s.PosterPath, s.BackdropPath, s.ID)
	return err
}

// Delete removes a series and its seasons/episodes (CASCADE)
func (r *SeriesRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM series WHERE id = ?", id)
	return err
}

// List retrieves series with optional filters
func (r *SeriesRepo) List(ctx context.Context, libraryID int64, limit, offset int) ([]model.Series, error) {
	query := `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path,
		created_at, updated_at
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

	var items []model.Series
	for rows.Next() {
		var s model.Series
		if err := rows.Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.Overview, &s.Status, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning series: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

// GetByTmdbID retrieves series by TMDb ID
func (r *SeriesRepo) GetByTmdbID(ctx context.Context, tmdbID int64) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path,
		created_at, updated_at
		FROM series WHERE tmdb_id = ?`, tmdbID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.Overview, &s.Status, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// GetByImdbID retrieves series by IMDb ID
func (r *SeriesRepo) GetByImdbID(ctx context.Context, imdbID string) (*model.Series, error) {
	var s model.Series
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path,
		created_at, updated_at
		FROM series WHERE imdb_id = ?`, imdbID).
		Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.Overview, &s.Status, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Search searches series by title
func (r *SeriesRepo) Search(ctx context.Context, query string, limit int) ([]model.Series, error) {
	q := `SELECT id, library_id, title, sort_title,
		tmdb_id, imdb_id, overview, status, first_air_date, poster_path, backdrop_path,
		created_at, updated_at
		FROM series WHERE title LIKE ? OR sort_title LIKE ?
		ORDER BY sort_title LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("searching series: %w", err)
	}
	defer rows.Close()

	var items []model.Series
	for rows.Next() {
		var s model.Series
		if err := rows.Scan(&s.ID, &s.LibraryID, &s.Title, &s.SortTitle,
			&s.TmdbID, &s.ImdbID, &s.Overview, &s.Status, &s.FirstAirDate,
			&s.PosterPath, &s.BackdropPath, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning series: %w", err)
		}
		items = append(items, s)
	}
	return items, rows.Err()
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
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates a season
func (r *SeasonRepo) Update(ctx context.Context, s *model.Season) error {
	_, err := r.db.ExecContext(ctx, `UPDATE seasons SET
		season_number = ?, title = ?, overview = ?, poster_path = ?, episode_count = ?
		WHERE id = ?`,
		s.SeasonNumber, s.Title, s.Overview, s.PosterPath, s.EpisodeCount, s.ID)
	return err
}

// Delete removes a season (episodes will be deleted by CASCADE)
func (r *SeasonRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM seasons WHERE id = ?", id)
	return err
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

	var items []model.Season
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
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Update updates an episode
func (r *EpisodeRepo) Update(ctx context.Context, e *model.Episode) error {
	_, err := r.db.ExecContext(ctx, `UPDATE episodes SET
		episode_number = ?, title = ?, overview = ?, still_path = ?, air_date = ?
		WHERE id = ?`,
		e.EpisodeNumber, e.Title, e.Overview, e.StillPath, e.AirDate, e.ID)
	return err
}

// Delete removes an episode
func (r *EpisodeRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM episodes WHERE id = ?", id)
	return err
}

// ListBySeasonID retrieves all episodes for a season
func (r *EpisodeRepo) ListBySeasonID(ctx context.Context, seasonID int64) ([]model.Episode, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, series_id, season_id, media_id,
		episode_number, title, overview, still_path, air_date, created_at
		FROM episodes WHERE season_id = ? ORDER BY episode_number`, seasonID)
	if err != nil {
		return nil, fmt.Errorf("listing episodes: %w", err)
	}
	defer rows.Close()

	var items []model.Episode
	for rows.Next() {
		var e model.Episode
		if err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

// ListBySeriesID retrieves all episodes for a series
func (r *EpisodeRepo) ListBySeriesID(ctx context.Context, seriesID int64) ([]model.Episode, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, series_id, season_id, media_id,
		episode_number, title, overview, still_path, air_date, created_at
		FROM episodes WHERE series_id = ? ORDER BY season_id, episode_number`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing episodes by series: %w", err)
	}
	defer rows.Close()

	var items []model.Episode
	for rows.Next() {
		var e model.Episode
		if err := rows.Scan(&e.ID, &e.SeriesID, &e.SeasonID, &e.MediaID,
			&e.EpisodeNumber, &e.Title, &e.Overview, &e.StillPath, &e.AirDate, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning episode: %w", err)
		}
		items = append(items, e)
	}
	return items, rows.Err()
}
