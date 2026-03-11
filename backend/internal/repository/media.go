package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/thawng/velox/internal/model"
)

// MediaRepo handles media (logical items) database operations
type MediaRepo struct {
	db DBTX
}

func NewMediaRepo(db DBTX) *MediaRepo {
	return &MediaRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *MediaRepo) WithTx(tx *sql.Tx) *MediaRepo {
	return &MediaRepo{db: tx}
}

// Create inserts a new media item
func (r *MediaRepo) Create(ctx context.Context, m *model.Media) error {
	query := `INSERT INTO media
		(library_id, media_type, title, sort_title, tmdb_id, imdb_id,
		 overview, release_date, rating, poster_path, backdrop_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	row := r.db.QueryRowContext(ctx, query,
		m.LibraryID, m.MediaType, m.Title, m.SortTitle, m.TmdbID, m.ImdbID,
		m.Overview, m.ReleaseDate, m.Rating, m.PosterPath, m.BackdropPath)

	return row.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// GetByID retrieves a media item by ID
func (r *MediaRepo) GetByID(ctx context.Context, id int64) (*model.Media, error) {
	var m model.Media
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, media_type, title, sort_title,
		tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
		created_at, updated_at
		FROM media WHERE id = ?`, id).
		Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
			&m.TmdbID, &m.ImdbID, &m.Overview, &m.ReleaseDate, &m.Rating,
			&m.PosterPath, &m.BackdropPath, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Update updates a media item
func (r *MediaRepo) Update(ctx context.Context, m *model.Media) error {
	_, err := r.db.ExecContext(ctx, `UPDATE media SET
		media_type = ?, title = ?, sort_title = ?, tmdb_id = ?, imdb_id = ?,
		overview = ?, release_date = ?, rating = ?, poster_path = ?, backdrop_path = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		m.MediaType, m.Title, m.SortTitle, m.TmdbID, m.ImdbID,
		m.Overview, m.ReleaseDate, m.Rating, m.PosterPath, m.BackdropPath, m.ID)
	return err
}

// Delete removes a media item and its files (CASCADE)
func (r *MediaRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media WHERE id = ?", id)
	return err
}

// List retrieves media items with optional filters
func (r *MediaRepo) List(ctx context.Context, libraryID int64, mediaType string, limit, offset int) ([]model.Media, error) {
	query := `SELECT id, library_id, media_type, title, sort_title,
		tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
		created_at, updated_at
		FROM media WHERE 1=1`
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

	var items []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
			&m.TmdbID, &m.ImdbID, &m.Overview, &m.ReleaseDate, &m.Rating,
			&m.PosterPath, &m.BackdropPath, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// Search searches media by title
func (r *MediaRepo) Search(ctx context.Context, query string, limit int) ([]model.Media, error) {
	q := `SELECT id, library_id, media_type, title, sort_title,
		tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
		created_at, updated_at
		FROM media WHERE title LIKE ? OR sort_title LIKE ?
		ORDER BY sort_title LIMIT ?`

	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, q, pattern, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("searching media: %w", err)
	}
	defer rows.Close()

	var items []model.Media
	for rows.Next() {
		var m model.Media
		if err := rows.Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
			&m.TmdbID, &m.ImdbID, &m.Overview, &m.ReleaseDate, &m.Rating,
			&m.PosterPath, &m.BackdropPath, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning media: %w", err)
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// GetByTmdbID retrieves media by TMDb ID
func (r *MediaRepo) GetByTmdbID(ctx context.Context, tmdbID int64) (*model.Media, error) {
	var m model.Media
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, media_type, title, sort_title,
		tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
		created_at, updated_at
		FROM media WHERE tmdb_id = ?`, tmdbID).
		Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
			&m.TmdbID, &m.ImdbID, &m.Overview, &m.ReleaseDate, &m.Rating,
			&m.PosterPath, &m.BackdropPath, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetByImdbID retrieves media by IMDb ID
func (r *MediaRepo) GetByImdbID(ctx context.Context, imdbID string) (*model.Media, error) {
	var m model.Media
	err := r.db.QueryRowContext(ctx, `SELECT id, library_id, media_type, title, sort_title,
		tmdb_id, imdb_id, overview, release_date, rating, poster_path, backdrop_path,
		created_at, updated_at
		FROM media WHERE imdb_id = ?`, imdbID).
		Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
			&m.TmdbID, &m.ImdbID, &m.Overview, &m.ReleaseDate, &m.Rating,
			&m.PosterPath, &m.BackdropPath, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListWithGenres retrieves media items with their genres
func (r *MediaRepo) ListWithGenres(ctx context.Context, libraryID int64, mediaType string) ([]model.MediaListItem, error) {
	query := `SELECT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
		GROUP_CONCAT(g.name, ',') as genre_names
		FROM media m
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
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

	var results []model.MediaListItem
	for rows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType, &genreNames); err != nil {
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
	var isPrimary int
	var lastVerified sql.NullString

	err := scanner.Scan(&mf.ID, &mf.MediaID, &mf.FilePath, &mf.FileSize, &mf.Duration,
		&mf.Width, &mf.Height, &mf.VideoCodec, &mf.AudioCodec, &mf.Container, &mf.Bitrate,
		&mf.Fingerprint, &isPrimary, &mf.AddedAt, &lastVerified)
	if err != nil {
		return nil, err
	}
	mf.IsPrimary = isPrimary == 1
	if lastVerified.Valid {
		mf.LastVerifiedAt = &lastVerified.String
	}
	return &mf, nil
}

// Create inserts a new media file
func (r *MediaFileRepo) Create(ctx context.Context, mf *model.MediaFile) error {
	query := `INSERT INTO media_files
		(media_id, file_path, file_size, duration, width, height,
		 video_codec, audio_codec, container, bitrate, fingerprint, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, added_at, last_verified_at`

	isPrimary := 0
	if mf.IsPrimary {
		isPrimary = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		mf.MediaID, mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.AudioCodec, mf.Container, mf.Bitrate, mf.Fingerprint, isPrimary)

	var lastVerified sql.NullString
	err := row.Scan(&mf.ID, &mf.AddedAt, &lastVerified)
	if lastVerified.Valid {
		mf.LastVerifiedAt = &lastVerified.String
	}
	return err
}

// GetByID retrieves a media file by ID
func (r *MediaFileRepo) GetByID(ctx context.Context, id int64) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, audio_codec, container, bitrate,
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

	_, err := r.db.ExecContext(ctx, `UPDATE media_files SET
		file_path = ?, file_size = ?, duration = ?, width = ?, height = ?,
		video_codec = ?, audio_codec = ?, container = ?, bitrate = ?,
		fingerprint = ?, is_primary = ?, last_verified_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.AudioCodec, mf.Container, mf.Bitrate,
		mf.Fingerprint, isPrimary, mf.ID)
	return err
}

// Delete removes a media file
func (r *MediaFileRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_files WHERE id = ?", id)
	return err
}

// ListByMediaID retrieves all files for a media item
func (r *MediaFileRepo) ListByMediaID(ctx context.Context, mediaID int64) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, audio_codec, container, bitrate,
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
		width, height, video_codec, audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE media_id = ? AND is_primary = 1 LIMIT 1`, mediaID)
	return scanMediaFile(row)
}

// FindByFingerprint finds a file by its fingerprint
func (r *MediaFileRepo) FindByFingerprint(ctx context.Context, fingerprint string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE fingerprint = ?`, fingerprint)
	return scanMediaFile(row)
}

// FindByPath finds a file by its path
func (r *MediaFileRepo) FindByPath(ctx context.Context, path string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE file_path = ?`, path)
	return scanMediaFile(row)
}

// UpdatePath updates the file path (for rename detection)
func (r *MediaFileRepo) UpdatePath(ctx context.Context, id int64, newPath string) error {
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET file_path = ? WHERE id = ?", newPath, id)
	return err
}

// MarkMissing marks a file as missing (sets last_verified_at = NULL)
func (r *MediaFileRepo) MarkMissing(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET last_verified_at = NULL WHERE id = ?", id)
	return err
}

// DeleteByMediaID removes all files for a media item
func (r *MediaFileRepo) DeleteByMediaID(ctx context.Context, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media_files WHERE media_id = ?", mediaID)
	return err
}

// SetPrimary sets a file as the primary version for its media
func (r *MediaFileRepo) SetPrimary(ctx context.Context, mediaID, fileID int64) error {
	// First clear primary for all files of this media
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET is_primary = 0 WHERE media_id = ?", mediaID)
	if err != nil {
		return err
	}
	// Then set the new primary
	_, err = r.db.ExecContext(ctx, "UPDATE media_files SET is_primary = 1 WHERE id = ? AND media_id = ?", fileID, mediaID)
	return err
}
