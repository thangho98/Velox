package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// GenreRepo handles genres database operations
type GenreRepo struct {
	db DBTX
}

func NewGenreRepo(db DBTX) *GenreRepo {
	return &GenreRepo{db: db}
}

// Create inserts a new genre
func (r *GenreRepo) Create(ctx context.Context, g *model.Genre) error {
	query := `INSERT INTO genres (name, tmdb_id) VALUES (?, ?) RETURNING id`
	row := r.db.QueryRowContext(ctx, query, g.Name, g.TmdbID)
	return row.Scan(&g.ID)
}

// GetByID retrieves a genre by ID
func (r *GenreRepo) GetByID(ctx context.Context, id int64) (*model.Genre, error) {
	var g model.Genre
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id FROM genres WHERE id = ?", id).
		Scan(&g.ID, &g.Name, &g.TmdbID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genre by id %d: %w", id, err)
	}
	return &g, nil
}

// GetByName retrieves a genre by name
func (r *GenreRepo) GetByName(ctx context.Context, name string) (*model.Genre, error) {
	var g model.Genre
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id FROM genres WHERE name = ?", name).
		Scan(&g.ID, &g.Name, &g.TmdbID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genre by name %q: %w", name, err)
	}
	return &g, nil
}

// GetByTmdbID retrieves a genre by TMDb ID
func (r *GenreRepo) GetByTmdbID(ctx context.Context, tmdbID int64) (*model.Genre, error) {
	var g model.Genre
	err := r.db.QueryRowContext(ctx, "SELECT id, name, tmdb_id FROM genres WHERE tmdb_id = ?", tmdbID).
		Scan(&g.ID, &g.Name, &g.TmdbID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get genre by tmdb_id %d: %w", tmdbID, err)
	}
	return &g, nil
}

// Update updates a genre
func (r *GenreRepo) Update(ctx context.Context, g *model.Genre) error {
	res, err := r.db.ExecContext(ctx, "UPDATE genres SET name = ?, tmdb_id = ? WHERE id = ?",
		g.Name, g.TmdbID, g.ID)
	if err != nil {
		return fmt.Errorf("update genre %d: %w", g.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes a genre
func (r *GenreRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM genres WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete genre %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// List retrieves all genres
func (r *GenreRepo) List(ctx context.Context) ([]model.Genre, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, tmdb_id FROM genres ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("listing genres: %w", err)
	}
	defer rows.Close()

	var items []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.TmdbID); err != nil {
			return nil, fmt.Errorf("scanning genre: %w", err)
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

// LinkToMedia links a genre to a media item
func (r *GenreRepo) LinkToMedia(ctx context.Context, mediaID, genreID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO media_genres (media_id, genre_id) VALUES (?, ?)",
		mediaID, genreID)
	if err != nil {
		return fmt.Errorf("linking genre %d to media %d: %w", genreID, mediaID, err)
	}
	return nil
}

// LinkToSeries links a genre to a series
func (r *GenreRepo) LinkToSeries(ctx context.Context, seriesID, genreID int64) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO media_genres (series_id, genre_id) VALUES (?, ?)",
		seriesID, genreID)
	if err != nil {
		return fmt.Errorf("linking genre %d to series %d: %w", genreID, seriesID, err)
	}
	return nil
}

// UnlinkFromMedia removes a genre link from a media item
func (r *GenreRepo) UnlinkFromMedia(ctx context.Context, mediaID, genreID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM media_genres WHERE media_id = ? AND genre_id = ?",
		mediaID, genreID)
	if err != nil {
		return fmt.Errorf("unlinking genre %d from media %d: %w", genreID, mediaID, err)
	}
	return nil
}

// UnlinkFromSeries removes a genre link from a series
func (r *GenreRepo) UnlinkFromSeries(ctx context.Context, seriesID, genreID int64) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM media_genres WHERE series_id = ? AND genre_id = ?",
		seriesID, genreID)
	if err != nil {
		return fmt.Errorf("unlinking genre %d from series %d: %w", genreID, seriesID, err)
	}
	return nil
}

// ListByMediaID retrieves all genres for a media item
func (r *GenreRepo) ListByMediaID(ctx context.Context, mediaID int64) ([]model.Genre, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT g.id, g.name, g.tmdb_id
		FROM genres g
		JOIN media_genres mg ON mg.genre_id = g.id
		WHERE mg.media_id = ?
		ORDER BY g.name`, mediaID)
	if err != nil {
		return nil, fmt.Errorf("listing genres by media: %w", err)
	}
	defer rows.Close()

	var items []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.TmdbID); err != nil {
			return nil, fmt.Errorf("scanning genre: %w", err)
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

// ListBySeriesID retrieves all genres for a series
func (r *GenreRepo) ListBySeriesID(ctx context.Context, seriesID int64) ([]model.Genre, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT g.id, g.name, g.tmdb_id
		FROM genres g
		JOIN media_genres mg ON mg.genre_id = g.id
		WHERE mg.series_id = ?
		ORDER BY g.name`, seriesID)
	if err != nil {
		return nil, fmt.Errorf("listing genres by series: %w", err)
	}
	defer rows.Close()

	var items []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.TmdbID); err != nil {
			return nil, fmt.Errorf("scanning genre: %w", err)
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

// ClearMediaGenres removes all genre links for a media item
func (r *GenreRepo) ClearMediaGenres(ctx context.Context, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_genres WHERE media_id = ?", mediaID)
	if err != nil {
		return fmt.Errorf("clearing genres for media %d: %w", mediaID, err)
	}
	return nil
}

// ClearSeriesGenres removes all genre links for a series
func (r *GenreRepo) ClearSeriesGenres(ctx context.Context, seriesID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_genres WHERE series_id = ?", seriesID)
	if err != nil {
		return fmt.Errorf("clearing genres for series %d: %w", seriesID, err)
	}
	return nil
}

// ListWithFilter retrieves genres with optional type filtering.
// typeFilter: "movie" | "series" | "" (empty = all)
// Returns genres that have at least one linked item of the specified type.
func (r *GenreRepo) ListWithFilter(ctx context.Context, typeFilter string) ([]model.Genre, error) {
	query := "SELECT DISTINCT g.id, g.name, g.tmdb_id FROM genres g"
	args := []any{}

	switch typeFilter {
	case "movie":
		query += ` JOIN media_genres mg ON mg.genre_id = g.id
			JOIN media m ON m.id = mg.media_id
			WHERE m.media_type = 'movie'`
	case "series":
		query += ` JOIN media_genres mg ON mg.genre_id = g.id
			WHERE mg.series_id IS NOT NULL`
	}

	query += " ORDER BY g.name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing genres with filter: %w", err)
	}
	defer rows.Close()

	var items []model.Genre
	for rows.Next() {
		var g model.Genre
		if err := rows.Scan(&g.ID, &g.Name, &g.TmdbID); err != nil {
			return nil, fmt.Errorf("scanning genre: %w", err)
		}
		items = append(items, g)
	}
	return items, rows.Err()
}
