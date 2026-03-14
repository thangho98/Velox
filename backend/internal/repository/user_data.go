package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// UserDataRepo handles user data operations (progress, favorites, ratings)
type UserDataRepo struct {
	db DBTX
}

// NewUserDataRepo creates a new user data repository
func NewUserDataRepo(db DBTX) *UserDataRepo {
	return &UserDataRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction
func (r *UserDataRepo) WithTx(tx *sql.Tx) *UserDataRepo {
	return &UserDataRepo{db: tx}
}

// GetProgress returns user data for a media item
func (r *UserDataRepo) GetProgress(ctx context.Context, userID, mediaID int64) (*model.UserData, error) {
	var d model.UserData
	var completed, isFavorite int
	var rating sql.NullFloat64
	var lastPlayedAt sql.NullString

	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, media_id, position, completed, is_favorite, rating, play_count, last_played_at, updated_at
		FROM user_data
		WHERE user_id = ? AND media_id = ?`,
		userID, mediaID).
		Scan(&d.UserID, &d.MediaID, &d.Position, &completed, &isFavorite, &rating, &d.PlayCount, &lastPlayedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	d.Completed = completed == 1
	d.IsFavorite = isFavorite == 1
	if rating.Valid {
		d.Rating = &rating.Float64
	}
	if lastPlayedAt.Valid {
		d.LastPlayedAt = &lastPlayedAt.String
	}
	return &d, nil
}

// UpsertProgress creates or updates watch progress
// Also updates last_played_at and increments play_count if completed
func (r *UserDataRepo) UpsertProgress(ctx context.Context, userID, mediaID int64, position float64, completed bool) error {
	completedInt := 0
	if completed {
		completedInt = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_data (user_id, media_id, position, completed, last_played_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, media_id) DO UPDATE SET
			position = excluded.position,
			completed = excluded.completed,
			last_played_at = CURRENT_TIMESTAMP,
			play_count = CASE WHEN excluded.completed = 1 AND user_data.completed = 0
							THEN user_data.play_count + 1
							ELSE user_data.play_count END,
			updated_at = CURRENT_TIMESTAMP`,
		userID, mediaID, position, completedInt)
	return err
}

// ToggleFavorite flips is_favorite (UPSERT: INSERT if not exists, UPDATE if exists)
func (r *UserDataRepo) ToggleFavorite(ctx context.Context, userID, mediaID int64) (isFavorite bool, err error) {
	var result int
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO user_data (user_id, media_id, is_favorite)
		VALUES (?, ?, 1)
		ON CONFLICT(user_id, media_id) DO UPDATE SET
			is_favorite = CASE WHEN user_data.is_favorite = 1 THEN 0 ELSE 1 END,
			updated_at = CURRENT_TIMESTAMP
		RETURNING is_favorite`,
		userID, mediaID).Scan(&result)
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// SetRating sets user rating (nil = remove rating). UPSERT.
func (r *UserDataRepo) SetRating(ctx context.Context, userID, mediaID int64, rating *float64) error {
	var ratingValue interface{}
	if rating != nil {
		ratingValue = *rating
	} else {
		ratingValue = nil
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_data (user_id, media_id, rating)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, media_id) DO UPDATE SET
			rating = excluded.rating,
			updated_at = CURRENT_TIMESTAMP`,
		userID, mediaID, ratingValue)
	return err
}

// ListFavorites returns items where is_favorite = 1, JOIN media for title/poster
func (r *UserDataRepo) ListFavorites(ctx context.Context, userID int64, limit, offset int) ([]*model.UserData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ud.user_id, ud.media_id, ud.position, ud.completed, ud.is_favorite, ud.rating, ud.play_count, ud.last_played_at, ud.updated_at,
			m.title, m.poster_path, COALESCE(mf.duration, 0)
		FROM user_data ud
		JOIN media m ON ud.media_id = m.id
		LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
		WHERE ud.user_id = ? AND ud.is_favorite = 1
		ORDER BY ud.updated_at DESC
		LIMIT ? OFFSET ?`,
		userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing favorites: %w", err)
	}
	defer rows.Close()

	var items []*model.UserData
	for rows.Next() {
		item := &model.UserData{}
		var completed, isFavorite int
		var rating sql.NullFloat64
		var lastPlayedAt sql.NullString

		if err := rows.Scan(&item.UserID, &item.MediaID, &item.Position, &completed, &isFavorite, &rating, &item.PlayCount, &lastPlayedAt, &item.UpdatedAt,
			&item.MediaTitle, &item.MediaPoster, &item.MediaDuration); err != nil {
			return nil, fmt.Errorf("scanning favorite: %w", err)
		}
		item.Completed = completed == 1
		item.IsFavorite = isFavorite == 1
		if rating.Valid {
			item.Rating = &rating.Float64
		}
		if lastPlayedAt.Valid {
			item.LastPlayedAt = &lastPlayedAt.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListRecentlyWatched returns items ordered by last_played_at DESC, JOIN media
func (r *UserDataRepo) ListRecentlyWatched(ctx context.Context, userID int64, limit int) ([]*model.UserData, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ud.user_id, ud.media_id, ud.position, ud.completed, ud.is_favorite, ud.rating, ud.play_count, ud.last_played_at, ud.updated_at,
			m.title, m.poster_path, COALESCE(mf.duration, 0)
		FROM user_data ud
		JOIN media m ON ud.media_id = m.id
		LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
		WHERE ud.user_id = ? AND ud.last_played_at IS NOT NULL
		ORDER BY ud.last_played_at DESC
		LIMIT ?`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing recently watched: %w", err)
	}
	defer rows.Close()

	var items []*model.UserData
	for rows.Next() {
		item := &model.UserData{}
		var completed, isFavorite int
		var rating sql.NullFloat64
		var lastPlayedAt sql.NullString

		if err := rows.Scan(&item.UserID, &item.MediaID, &item.Position, &completed, &isFavorite, &rating, &item.PlayCount, &lastPlayedAt, &item.UpdatedAt,
			&item.MediaTitle, &item.MediaPoster, &item.MediaDuration); err != nil {
			return nil, fmt.Errorf("scanning recently watched: %w", err)
		}
		item.Completed = completed == 1
		item.IsFavorite = isFavorite == 1
		if rating.Valid {
			item.Rating = &rating.Float64
		}
		if lastPlayedAt.Valid {
			item.LastPlayedAt = &lastPlayedAt.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// DeleteProgress removes progress for a user and media
func (r *UserDataRepo) DeleteProgress(ctx context.Context, userID, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM user_data WHERE user_id = ? AND media_id = ?`,
		userID, mediaID)
	return err
}

// DeleteAllUserData removes all data for a user (useful when deleting user)
func (r *UserDataRepo) DeleteAllUserData(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_data WHERE user_id = ?`, userID)
	return err
}

// DismissProgress resets progress fields while preserving favorite/rating/play_count
func (r *UserDataRepo) DismissProgress(ctx context.Context, userID, mediaID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE user_data
		SET position = 0, completed = 0, last_played_at = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND media_id = ?`,
		userID, mediaID)
	return err
}

// ListContinueWatching returns in-progress items (position > 0 AND completed = 0)
func (r *UserDataRepo) ListContinueWatching(ctx context.Context, userID int64, limit int) ([]*model.ContinueWatchingItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT ud.media_id, COALESCE(e.series_id, 0) as series_id, ud.position, ud.completed, ud.last_played_at,
			   m.title, m.poster_path, m.backdrop_path, m.media_type,
			   COALESCE(mf.duration, 0) as media_duration,
			   e.episode_number,
			   s.title as series_title,
			   se.season_number
		FROM user_data ud
		JOIN media m ON ud.media_id = m.id
		LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
		LEFT JOIN episodes e ON e.media_id = m.id
		LEFT JOIN series s ON e.series_id = s.id
		LEFT JOIN seasons se ON e.season_id = se.id
		WHERE ud.user_id = ?
		  AND ud.position > 0
		  AND ud.completed = 0
		ORDER BY ud.last_played_at DESC
		LIMIT ?`,
		userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing continue watching: %w", err)
	}
	defer rows.Close()

	var items []*model.ContinueWatchingItem
	for rows.Next() {
		item := &model.ContinueWatchingItem{}
		var completed int
		var lastPlayedAt sql.NullString
		var episodeNumber sql.NullInt64
		var seriesTitle sql.NullString
		var seasonNumber sql.NullInt64

		if err := rows.Scan(
			&item.MediaID, &item.SeriesID, &item.Position, &completed, &lastPlayedAt,
			&item.Title, &item.PosterPath, &item.BackdropPath, &item.MediaType,
			&item.MediaDuration,
			&episodeNumber,
			&seriesTitle, &seasonNumber); err != nil {
			return nil, fmt.Errorf("scanning continue watching: %w", err)
		}
		item.Completed = completed == 1
		if lastPlayedAt.Valid {
			item.LastPlayedAt = &lastPlayedAt.String
		}
		if episodeNumber.Valid {
			item.EpisodeNumber = int(episodeNumber.Int64)
		}
		if seriesTitle.Valid {
			item.SeriesTitle = seriesTitle.String
		}
		if seasonNumber.Valid {
			item.SeasonNumber = int(seasonNumber.Int64)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ListNextUp returns the first unwatched episode for each series being followed
func (r *UserDataRepo) ListNextUp(ctx context.Context, userID int64, limit int) ([]*model.NextUpItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		WITH user_series AS (
			-- Series where user has completed at least one episode
			SELECT e.series_id,
				   MAX(ud.last_played_at) as last_watched
			FROM user_data ud
			JOIN episodes e ON e.media_id = ud.media_id
			WHERE ud.user_id = ? AND ud.completed = 1
			GROUP BY e.series_id
		),
		candidates AS (
			-- All episodes not completed AND not in-progress, sorted by order
			SELECT us.series_id, us.last_watched,
				   e.media_id,
				   se.season_number, e.episode_number,
				   ROW_NUMBER() OVER (
					   PARTITION BY us.series_id
					   ORDER BY se.season_number, e.episode_number
				   ) as rn
			FROM user_series us
			JOIN episodes e ON e.series_id = us.series_id
			JOIN seasons se ON e.season_id = se.id
			LEFT JOIN user_data ud ON ud.media_id = e.media_id AND ud.user_id = ?
			WHERE COALESCE(ud.completed, 0) = 0
			  AND COALESCE(ud.position, 0) = 0  -- Exclude in-progress (already in Continue Watching)
		)
		SELECT m.id as media_id, m.title, m.media_type, m.backdrop_path,
			   c.series_id,
			   e.episode_number, e.title as episode_title, e.still_path,
			   c.season_number, s.title as series_title, s.poster_path as series_poster,
			   c.last_watched,
			   COALESCE(mf.duration, 0) as media_duration
		FROM candidates c
		JOIN episodes e ON e.media_id = c.media_id
		JOIN media m ON m.id = c.media_id
		JOIN series s ON e.series_id = s.id
		LEFT JOIN media_files mf ON m.id = mf.media_id AND mf.is_primary = 1
		WHERE c.rn = 1  -- Only first unwatched episode per series
		ORDER BY c.last_watched DESC
		LIMIT ?`,
		userID, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing next up: %w", err)
	}
	defer rows.Close()

	var items []*model.NextUpItem
	for rows.Next() {
		item := &model.NextUpItem{}
		var lastWatched sql.NullString

		if err := rows.Scan(
			&item.MediaID, &item.Title, &item.MediaType, &item.BackdropPath,
			&item.SeriesID,
			&item.EpisodeNumber, &item.EpisodeTitle, &item.StillPath,
			&item.SeasonNumber, &item.SeriesTitle, &item.SeriesPoster,
			&lastWatched, &item.Duration); err != nil {
			return nil, fmt.Errorf("scanning next up: %w", err)
		}
		if lastWatched.Valid {
			item.LastWatchedAt = &lastWatched.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
