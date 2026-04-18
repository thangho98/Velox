package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/thawng/velox/internal/model"
)

// List retrieves media items with optional filters
func (r *MediaRepo) List(ctx context.Context, libraryID int64, mediaType string, limit, offset int) ([]model.Media, error) {
	query := `SELECT ` + mediaColumns + ` FROM media WHERE 1=1`
	args := []any{}

	if libraryID > 0 {
		query += " AND library_id = ?"
		args = append(args, libraryID)
	}
	if mediaType != "" {
		query += " AND media_type = ?"
		args = append(args, mediaType)
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
		return nil, fmt.Errorf("listing media: %w", err)
	}
	defer rows.Close()

	items := []model.Media{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}
		items = append(items, *m)
	}
	return items, rows.Err()
}

// Search searches media by title
func (r *MediaRepo) Search(ctx context.Context, query string, limit int) ([]model.Media, error) {
	q := `SELECT ` + mediaColumns + ` FROM media WHERE title LIKE ? OR sort_title LIKE ?
		ORDER BY sort_title LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("searching media: %w", err)
	}
	defer rows.Close()

	items := []model.Media{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}
		items = append(items, *m)
	}
	return items, rows.Err()
}

// GetByTmdbID retrieves media by TMDb ID
func (r *MediaRepo) GetByTmdbID(ctx context.Context, tmdbID int64) (*model.Media, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+mediaColumns+` FROM media WHERE tmdb_id = ?`, tmdbID)
	return scanMedia(row)
}

// GetByImdbID retrieves media by IMDb ID
func (r *MediaRepo) GetByImdbID(ctx context.Context, imdbID string) (*model.Media, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+mediaColumns+` FROM media WHERE imdb_id = ?`, imdbID)
	return scanMedia(row)
}

// ListWithIMDbID retrieves all media items that have an imdb_id set.
func (r *MediaRepo) ListWithIMDbID(ctx context.Context) ([]model.Media, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+mediaColumns+` FROM media WHERE imdb_id IS NOT NULL AND imdb_id != ''`)
	if err != nil {
		return nil, fmt.Errorf("listing media with imdb: %w", err)
	}
	defer rows.Close()

	items := []model.Media{}
	for rows.Next() {
		m, err := scanMedia(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}
		items = append(items, *m)
	}
	return items, rows.Err()
}

// ListWithGenres retrieves media items with their genres
func (r *MediaRepo) ListWithGenres(ctx context.Context, libraryID int64, mediaType string) ([]model.MediaListItem, error) {
	query := `SELECT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
		GROUP_CONCAT(g.name, ',') as genre_names,
		COALESCE(e.series_id, 0) as series_id
		FROM media m
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN episodes e ON e.media_id = m.id
		WHERE 1=1`
	args := []any{}

	if libraryID > 0 {
		query += " AND m.library_id = ?"
		args = append(args, libraryID)
	}
	if mediaType != "" {
		query += " AND m.media_type = ?"
		args = append(args, mediaType)
	}

	query += " GROUP BY m.id ORDER BY m.sort_title"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing media with genres: %w", err)
	}
	defer rows.Close()

	results := []model.MediaListItem{}
	for rows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType, &genreNames, &item.SeriesID); err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}

		// Handle NULL or empty genre list
		if genreNames.Valid && genreNames.String != "" {
			item.Genres = strings.Split(genreNames.String, ",")
		}

		results = append(results, item)
	}
	return results, rows.Err()
}

// ListFiltered retrieves media items with advanced filtering, sorting, and pagination.
// Supports filtering by library, media type, search query, genre, and year.
func (r *MediaRepo) ListFiltered(ctx context.Context, f model.MediaListFilter) ([]model.MediaListItem, error) {
	query := `SELECT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
		m.release_date, m.rating, m.overview,
		GROUP_CONCAT(DISTINCT g.name) as genre_names,
		COALESCE(e.series_id, 0) as series_id,
		ud.position, mf.duration, ud.completed
		FROM media m
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN episodes e ON e.media_id = m.id
		LEFT JOIN user_data ud ON ud.media_id = m.id AND ud.user_id = ?
		LEFT JOIN media_files mf ON mf.media_id = m.id AND mf.is_primary = 1
		WHERE 1=1`
	args := []any{f.UserID}

	// Library filter
	if f.LibraryID > 0 {
		query += " AND m.library_id = ?"
		args = append(args, f.LibraryID)
	}

	// Media type filter ("movie" | "episode")
	if f.MediaType != "" {
		query += " AND m.media_type = ?"
		args = append(args, f.MediaType)
	}

	// Search filter (LIKE on title OR sort_title)
	if f.Search != "" {
		query += " AND (m.title LIKE ? OR m.sort_title LIKE ?)"
		pattern := "%" + f.Search + "%"
		args = append(args, pattern, pattern)
	}

	// Genre filter using EXISTS subquery for exact match
	// This avoids false positives like "Action" matching "Live Action"
	if f.Genre != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM media_genres mg2
			JOIN genres g2 ON g2.id = mg2.genre_id
			WHERE mg2.media_id = m.id AND g2.name = ?
		)`
		args = append(args, f.Genre)
	}

	// Year filter (extract year from release_date)
	if f.Year != "" {
		query += " AND m.release_date LIKE ?"
		args = append(args, f.Year+"%")
	}

	// StartChar filter
	if f.StartChar != "" && f.StartChar != "#" {
		query += " AND UPPER(m.sort_title) >= ?"
		args = append(args, strings.ToUpper(f.StartChar))
	}

	query += " GROUP BY m.id"

	// Sort order
	switch f.Sort {
	case "newest":
		query += " ORDER BY m.release_date DESC, m.sort_title ASC"
	case "oldest":
		query += " ORDER BY m.release_date ASC, m.sort_title ASC"
	case "rating":
		query += " ORDER BY m.rating DESC, m.sort_title ASC"
	case "title":
		query += " ORDER BY m.sort_title ASC"
	default:
		query += " ORDER BY m.sort_title ASC"
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
		return nil, fmt.Errorf("listing filtered media: %w", err)
	}
	defer rows.Close()

	results := []model.MediaListItem{}
	for rows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		var position, duration sql.NullFloat64
		var completed sql.NullBool
		if err := rows.Scan(&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType,
			&item.ReleaseDate, &item.Rating, &item.Overview, &genreNames, &item.SeriesID,
			&position, &duration, &completed); err != nil {
			return nil, fmt.Errorf("scanning filtered media: %w", err)
		}

		// Handle NULL or empty genre list
		if genreNames.Valid && genreNames.String != "" {
			item.Genres = strings.Split(genreNames.String, ",")
		}

		// Progress fields
		if position.Valid {
			item.Position = &position.Float64
		}
		if duration.Valid {
			item.Duration = &duration.Float64
		}
		if completed.Valid {
			item.Completed = &completed.Bool
		}

		results = append(results, item)
	}
	return results, rows.Err()
}

// GetAlphabet returns the count of media items for each starting letter
func (r *MediaRepo) GetAlphabet(ctx context.Context, f model.MediaListFilter) ([]model.AlphabetCount, error) {
	query := `SELECT 
		(CASE 
			WHEN UPPER(SUBSTR(m.sort_title, 1, 1)) BETWEEN 'A' AND 'Z' 
			THEN UPPER(SUBSTR(m.sort_title, 1, 1)) 
			ELSE '#' END) as letter, 
		COUNT(DISTINCT m.id) as count 
		FROM media m
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		WHERE 1=1`
	args := []any{}

	if f.LibraryID > 0 {
		query += " AND m.library_id = ?"
		args = append(args, f.LibraryID)
	}
	if f.MediaType != "" {
		query += " AND m.media_type = ?"
		args = append(args, f.MediaType)
	}
	if f.Genre != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM media_genres mg2
			JOIN genres g2 ON g2.id = mg2.genre_id
			WHERE mg2.media_id = m.id AND g2.name = ?
		)`
		args = append(args, f.Genre)
	}
	if f.Year != "" {
		query += " AND m.release_date LIKE ?"
		args = append(args, f.Year+"%")
	}

	query += " GROUP BY letter ORDER BY letter"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("getting alphabet: %w", err)
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
