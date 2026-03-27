package repository

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sort"
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

// mediaColumns is the shared column list for media queries.
const mediaColumns = `id, library_id, media_type, title, sort_title,
	tmdb_id, imdb_id, tvdb_id, overview, tagline, release_date, rating,
	imdb_rating, rt_score, metacritic_score,
	poster_path, backdrop_path, logo_path, thumb_path, metadata_locked, created_at, updated_at`

// scanMedia scans a row into a model.Media using the standard column order.
func scanMedia(scanner interface{ Scan(...any) error }) (*model.Media, error) {
	var m model.Media
	var locked int
	err := scanner.Scan(&m.ID, &m.LibraryID, &m.MediaType, &m.Title, &m.SortTitle,
		&m.TmdbID, &m.ImdbID, &m.TvdbID, &m.Overview, &m.Tagline, &m.ReleaseDate, &m.Rating,
		&m.IMDbRating, &m.RTScore, &m.MetacriticScore,
		&m.PosterPath, &m.BackdropPath, &m.LogoPath, &m.ThumbPath, &locked, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	m.MetadataLocked = locked == 1
	return &m, nil
}

// Create inserts a new media item
func (r *MediaRepo) Create(ctx context.Context, m *model.Media) error {
	query := `INSERT INTO media
		(library_id, media_type, title, sort_title, tmdb_id, imdb_id, tvdb_id,
		 overview, tagline, release_date, rating, imdb_rating, rt_score, metacritic_score,
		 poster_path, backdrop_path, logo_path, thumb_path, metadata_locked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at`

	locked := 0
	if m.MetadataLocked {
		locked = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		m.LibraryID, m.MediaType, m.Title, m.SortTitle, m.TmdbID, m.ImdbID, m.TvdbID,
		m.Overview, m.Tagline, m.ReleaseDate, m.Rating, m.IMDbRating, m.RTScore, m.MetacriticScore,
		m.PosterPath, m.BackdropPath, m.LogoPath, m.ThumbPath, locked)

	return row.Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

// GetByID retrieves a media item by ID
func (r *MediaRepo) GetByID(ctx context.Context, id int64) (*model.Media, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+mediaColumns+` FROM media WHERE id = ?`, id)
	return scanMedia(row)
}

// Update updates a media item (full update — used by metadata enrichment pipeline).
func (r *MediaRepo) Update(ctx context.Context, m *model.Media) error {
	locked := 0
	if m.MetadataLocked {
		locked = 1
	}
	_, err := r.db.ExecContext(ctx, `UPDATE media SET
		media_type = ?, title = ?, sort_title = ?, tmdb_id = ?, imdb_id = ?, tvdb_id = ?,
		overview = ?, tagline = ?, release_date = ?, rating = ?,
		imdb_rating = ?, rt_score = ?, metacritic_score = ?,
		poster_path = ?, backdrop_path = ?, logo_path = ?, thumb_path = ?,
		metadata_locked = ?,
		updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		m.MediaType, m.Title, m.SortTitle, m.TmdbID, m.ImdbID, m.TvdbID,
		m.Overview, m.Tagline, m.ReleaseDate, m.Rating,
		m.IMDbRating, m.RTScore, m.MetacriticScore,
		m.PosterPath, m.BackdropPath, m.LogoPath, m.ThumbPath,
		locked, m.ID)
	return err
}

// UpdateMetadata performs a partial metadata update — only SET fields present in the request.
func (r *MediaRepo) UpdateMetadata(ctx context.Context, id int64, req model.MetadataEditRequest) error {
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
	if req.Tagline != nil {
		setClauses = append(setClauses, "tagline = ?")
		args = append(args, *req.Tagline)
	}
	if req.ReleaseDate != nil {
		setClauses = append(setClauses, "release_date = ?")
		args = append(args, *req.ReleaseDate)
	}
	if req.Rating != nil {
		setClauses = append(setClauses, "rating = ?")
		args = append(args, *req.Rating)
	}
	if req.MetadataLocked != nil {
		locked := 0
		if *req.MetadataLocked {
			locked = 1
		}
		setClauses = append(setClauses, "metadata_locked = ?")
		args = append(args, locked)
	}

	if len(setClauses) == 0 {
		return nil
	}

	setClauses = append(setClauses, "updated_at = CURRENT_TIMESTAMP")
	query := fmt.Sprintf("UPDATE media SET %s WHERE id = ?", strings.Join(setClauses, ", "))
	args = append(args, id)
	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateImagePath updates poster_path or backdrop_path for a media item.
// Returns ErrNotFound if the media ID does not exist.
func (r *MediaRepo) UpdateImagePath(ctx context.Context, id int64, imageType, path string) error {
	col := "poster_path"
	if imageType == "backdrop" {
		col = "backdrop_path"
	}
	res, err := r.db.ExecContext(ctx, fmt.Sprintf("UPDATE media SET %s = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", col), path, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetMetadataLocked sets the metadata_locked flag for a media item.
func (r *MediaRepo) SetMetadataLocked(ctx context.Context, id int64, locked bool) error {
	v := 0
	if locked {
		v = 1
	}
	res, err := r.db.ExecContext(ctx, "UPDATE media SET metadata_locked = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", v, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// UpdateOMDbRatings updates only the OMDb rating fields.
func (r *MediaRepo) UpdateOMDbRatings(ctx context.Context, id int64, imdbRating float64, rtScore, metacriticScore int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE media SET imdb_rating = ?, rt_score = ?, metacritic_score = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		imdbRating, rtScore, metacriticScore, id)
	return err
}

// UpdateTitle updates only the title and sort_title of a media item.
func (r *MediaRepo) UpdateTitle(ctx context.Context, id int64, title string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE media SET title = ?, sort_title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`, title, title, id)
	return err
}

// Delete removes a media item and its files (CASCADE)
func (r *MediaRepo) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM media WHERE id = ?", id)
	return err
}

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

	var items []model.Media
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

	var items []model.Media
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

	var items []model.Media
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

	var results []model.MediaListItem
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
		COALESCE(e.series_id, 0) as series_id
		FROM media m
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN episodes e ON e.media_id = m.id
		WHERE 1=1`
	args := []any{}

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

	var results []model.MediaListItem
	for rows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		if err := rows.Scan(&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType,
			&item.ReleaseDate, &item.Rating, &item.Overview, &genreNames, &item.SeriesID); err != nil {
			return nil, fmt.Errorf("scanning filtered media: %w", err)
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
		&mf.Width, &mf.Height, &mf.VideoCodec, &mf.VideoProfile, &mf.VideoLevel, &mf.VideoFPS,
		&mf.AudioCodec, &mf.Container, &mf.Bitrate,
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
		 video_codec, video_profile, video_level, video_fps,
		 audio_codec, container, bitrate, fingerprint, is_primary)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, added_at, last_verified_at`

	isPrimary := 0
	if mf.IsPrimary {
		isPrimary = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		mf.MediaID, mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.VideoProfile, mf.VideoLevel, mf.VideoFPS,
		mf.AudioCodec, mf.Container, mf.Bitrate, mf.Fingerprint, isPrimary)

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
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
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
		video_codec = ?, video_profile = ?, video_level = ?, video_fps = ?,
		audio_codec = ?, container = ?, bitrate = ?,
		fingerprint = ?, is_primary = ?, last_verified_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		mf.FilePath, mf.FileSize, mf.Duration, mf.Width, mf.Height,
		mf.VideoCodec, mf.VideoProfile, mf.VideoLevel, mf.VideoFPS,
		mf.AudioCodec, mf.Container, mf.Bitrate,
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
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
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
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE media_id = ? AND is_primary = 1 LIMIT 1`, mediaID)
	return scanMediaFile(row)
}

// FindByFingerprint finds a file by its fingerprint
func (r *MediaFileRepo) FindByFingerprint(ctx context.Context, fingerprint string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files WHERE fingerprint = ?`, fingerprint)
	return scanMediaFile(row)
}

// FindByPath finds a file by its path
func (r *MediaFileRepo) FindByPath(ctx context.Context, path string) (*model.MediaFile, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
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

// ListByLibraryID retrieves all media files for a given library (via media table join).
// Results are returned in batches suitable for verification. Use limit/offset for pagination.
func (r *MediaFileRepo) ListByLibraryID(ctx context.Context, libraryID int64, limit, offset int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT mf.id, mf.media_id, mf.file_path, mf.file_size, mf.duration,
		mf.width, mf.height, mf.video_codec, mf.video_profile, mf.video_level, mf.video_fps,
		mf.audio_codec, mf.container, mf.bitrate,
		mf.fingerprint, mf.is_primary, mf.added_at, mf.last_verified_at
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?
		ORDER BY mf.id
		LIMIT ? OFFSET ?`, libraryID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing files by library: %w", err)
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

// ListAllPaginated retrieves all media files in the database, paginated.
func (r *MediaFileRepo) ListAllPaginated(ctx context.Context, limit, offset int) ([]model.MediaFile, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_id, file_path, file_size, duration,
		width, height, video_codec, video_profile, video_level, video_fps,
		audio_codec, container, bitrate,
		fingerprint, is_primary, added_at, last_verified_at
		FROM media_files ORDER BY id LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing all files: %w", err)
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

// UpdateLastVerified updates the last_verified_at timestamp for a file.
func (r *MediaFileRepo) UpdateLastVerified(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "UPDATE media_files SET last_verified_at = CURRENT_TIMESTAMP WHERE id = ?", id)
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

// BrowseFolderItem represents a subfolder in browse results
type BrowseFolderItem struct {
	Name       string `json:"name"`
	Path       string `json:"path"`             // library-relative path for navigation
	MediaCount int    `json:"media_count"`      // number of media files under this folder
	Poster     string `json:"poster,omitempty"` // poster from first media in folder (Emby-style)
}

// BrowseResult represents the result of a folder browse operation.
// Includes both subfolders and media items directly in the current folder.
type BrowseResult struct {
	LibraryID int64                 `json:"library_id"`
	Path      string                `json:"path"`   // current library-relative path
	Parent    string                `json:"parent"` // parent relative path, "" if root
	Folders   []BrowseFolderItem    `json:"folders"`
	Media     []model.MediaListItem `json:"media"`
}

// BrowseFolders returns subfolders + media at a given path within a library.
// absDir is the resolved absolute directory path on disk.
// relativePath is the library-relative path for the response.
func (r *MediaFileRepo) BrowseFolders(ctx context.Context, libraryID int64, absDir, relativePath string) (*BrowseResult, error) {
	// Ensure absDir ends without trailing slash for consistent LIKE matching
	absDir = strings.TrimRight(absDir, "/")
	prefix := absDir + "/"

	result := &BrowseResult{
		LibraryID: libraryID,
		Path:      relativePath,
		Folders:   []BrowseFolderItem{},
		Media:     []model.MediaListItem{},
	}

	// Compute parent path
	if relativePath != "" {
		parent := filepath.Dir(relativePath)
		if parent == "." {
			parent = ""
		}
		result.Parent = parent
	}

	// Step 1: Get all file paths under this directory to extract subdirectories
	rows, err := r.db.QueryContext(ctx, `
		SELECT mf.file_path
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ?
		  AND mf.file_path LIKE ? || '%'`,
		libraryID, prefix)
	if err != nil {
		return nil, fmt.Errorf("browse listing files: %w", err)
	}
	defer rows.Close()

	// Extract immediate subdirectory names and count media per subdir
	dirCounts := map[string]int{}
	for rows.Next() {
		var fp string
		if err := rows.Scan(&fp); err != nil {
			return nil, fmt.Errorf("browse scanning path: %w", err)
		}
		// Strip prefix to get relative remainder: "SubFolder/file.mkv" or "file.mkv"
		rel := strings.TrimPrefix(fp, prefix)
		if slashIdx := strings.Index(rel, "/"); slashIdx > 0 {
			// Has subdirectory — count it
			dirName := rel[:slashIdx]
			dirCounts[dirName]++
		}
		// Files directly in this folder (no slash) are handled in Step 2
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("browse iterating paths: %w", err)
	}

	// Sort folder names alphabetically
	dirNames := make([]string, 0, len(dirCounts))
	for name := range dirCounts {
		dirNames = append(dirNames, name)
	}
	sort.Strings(dirNames)

	for _, name := range dirNames {
		folderRelPath := name
		if relativePath != "" {
			folderRelPath = relativePath + "/" + name
		}
		// Fetch poster from first media in this subfolder
		subPrefix := prefix + name + "/"
		var poster sql.NullString
		_ = r.db.QueryRowContext(ctx, `
			SELECT m.poster_path FROM media_files mf
			JOIN media m ON m.id = mf.media_id
			WHERE m.library_id = ? AND mf.file_path LIKE ? || '%' AND m.poster_path != ''
			ORDER BY m.sort_title LIMIT 1`,
			libraryID, subPrefix).Scan(&poster)

		result.Folders = append(result.Folders, BrowseFolderItem{
			Name:       name,
			Path:       folderRelPath,
			MediaCount: dirCounts[name],
			Poster:     poster.String,
		})
	}

	// Step 2: Get media items directly in this folder (not in subdirectories)
	mediaRows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT m.id, m.title, m.sort_title, m.poster_path, m.media_type,
		       m.release_date, m.rating, m.overview,
		       GROUP_CONCAT(DISTINCT g.name) as genre_names,
		       COALESCE(e.series_id, 0) as series_id
		FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		LEFT JOIN media_genres mg ON mg.media_id = m.id
		LEFT JOIN genres g ON g.id = mg.genre_id
		LEFT JOIN episodes e ON e.media_id = m.id
		WHERE m.library_id = ?
		  AND mf.file_path LIKE ? || '%'
		  AND mf.file_path NOT LIKE ? || '%/%'
		GROUP BY m.id
		ORDER BY m.sort_title`,
		libraryID, prefix, prefix)
	if err != nil {
		return nil, fmt.Errorf("browse listing media: %w", err)
	}
	defer mediaRows.Close()

	for mediaRows.Next() {
		var item model.MediaListItem
		var genreNames sql.NullString
		if err := mediaRows.Scan(
			&item.ID, &item.Title, &item.SortTitle, &item.PosterPath, &item.MediaType,
			&item.ReleaseDate, &item.Rating, &item.Overview,
			&genreNames, &item.SeriesID,
		); err != nil {
			return nil, fmt.Errorf("browse scanning media: %w", err)
		}
		if genreNames.Valid && genreNames.String != "" {
			item.Genres = strings.Split(genreNames.String, ",")
		}
		result.Media = append(result.Media, item)
	}

	return result, mediaRows.Err()
}
