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

// FirstPosterByLibrary returns the poster_path of the first media item (by sort_title) in a library.
func (r *MediaRepo) FirstPosterByLibrary(ctx context.Context, libraryID int64) string {
	var poster sql.NullString
	_ = r.db.QueryRowContext(ctx, `
		SELECT poster_path FROM media
		WHERE library_id = ? AND poster_path != ''
		ORDER BY sort_title LIMIT 1`, libraryID).Scan(&poster)
	return poster.String
}

// FirstPosterByLibraryPath returns the poster_path of the first media item
// whose file_path starts with the given root path prefix.
func (r *MediaRepo) FirstPosterByLibraryPath(ctx context.Context, libraryID int64, pathPrefix string) string {
	var poster sql.NullString
	_ = r.db.QueryRowContext(ctx, `
		SELECT m.poster_path FROM media_files mf
		JOIN media m ON m.id = mf.media_id
		WHERE m.library_id = ? AND mf.file_path LIKE ? || '%' AND m.poster_path != ''
		ORDER BY m.sort_title LIMIT 1`, libraryID, pathPrefix+"/").Scan(&poster)
	return poster.String
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
