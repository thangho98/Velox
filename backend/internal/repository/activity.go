package repository

import (
	"context"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// ActivityRepo handles activity_log database operations
type ActivityRepo struct {
	db DBTX
}

func NewActivityRepo(db DBTX) *ActivityRepo {
	return &ActivityRepo{db: db}
}

// Insert adds a single activity log entry.
func (r *ActivityRepo) Insert(ctx context.Context, userID *int64, action string, mediaID *int64, details, ip string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO activity_log (user_id, action, media_id, details_json, ip_address)
		 VALUES (?, ?, ?, ?, ?)`,
		userID, action, mediaID, details, ip)
	if err != nil {
		return fmt.Errorf("inserting activity log: %w", err)
	}
	return nil
}

// InsertBatch inserts multiple activity entries in a single batch.
func (r *ActivityRepo) InsertBatch(ctx context.Context, entries []ActivityEntry) error {
	for _, e := range entries {
		if err := r.Insert(ctx, e.UserID, e.Action, e.MediaID, e.Details, e.IP); err != nil {
			return err
		}
	}
	return nil
}

// ActivityEntry represents a single activity for batch insertion.
type ActivityEntry struct {
	UserID  *int64
	Action  string
	MediaID *int64
	Details string
	IP      string
}

// List retrieves activity log entries with optional filters and joined user/media info.
func (r *ActivityRepo) List(ctx context.Context, filter model.ActivityFilter) ([]model.ActivityLog, error) {
	query := `SELECT
		a.id, a.user_id, COALESCE(u.username, ''), a.action,
		a.media_id, COALESCE(m.title, ''), a.details_json, a.ip_address, a.created_at
		FROM activity_log a
		LEFT JOIN users u ON a.user_id = u.id
		LEFT JOIN media m ON a.media_id = m.id
		WHERE 1=1`
	args := []any{}

	if filter.UserID != nil {
		query += " AND a.user_id = ?"
		args = append(args, *filter.UserID)
	}
	if filter.Action != "" {
		query += " AND a.action = ?"
		args = append(args, filter.Action)
	}
	if filter.From != "" {
		query += " AND a.created_at >= ?"
		args = append(args, filter.From)
	}
	if filter.To != "" {
		query += " AND a.created_at <= ?"
		args = append(args, filter.To)
	}

	query += " ORDER BY a.created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing activity: %w", err)
	}
	defer rows.Close()

	var items []model.ActivityLog
	for rows.Next() {
		var a model.ActivityLog
		if err := rows.Scan(&a.ID, &a.UserID, &a.Username, &a.Action,
			&a.MediaID, &a.MediaTitle, &a.DetailsJSON, &a.IPAddress, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning activity: %w", err)
		}
		items = append(items, a)
	}
	return items, rows.Err()
}

// CountByAction returns action counts within a time range.
func (r *ActivityRepo) CountByAction(ctx context.Context, from, to string) (map[string]int, error) {
	query := `SELECT action, COUNT(*) FROM activity_log WHERE 1=1`
	args := []any{}

	if from != "" {
		query += " AND created_at >= ?"
		args = append(args, from)
	}
	if to != "" {
		query += " AND created_at <= ?"
		args = append(args, to)
	}

	query += " GROUP BY action"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("counting activity by action: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var action string
		var count int
		if err := rows.Scan(&action, &count); err != nil {
			return nil, fmt.Errorf("scanning action count: %w", err)
		}
		counts[action] = count
	}
	return counts, rows.Err()
}

// PlaybackStats returns aggregated playback statistics.
func (r *ActivityRepo) PlaybackStats(ctx context.Context) (*model.PlaybackStatsResult, error) {
	result := &model.PlaybackStatsResult{}

	// Total plays
	row := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM activity_log WHERE action = 'playback_start'`)
	if err := row.Scan(&result.TotalPlays); err != nil {
		return nil, fmt.Errorf("counting total plays: %w", err)
	}

	// Unique titles
	row = r.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT media_id) FROM activity_log WHERE action = 'playback_start' AND media_id IS NOT NULL`)
	if err := row.Scan(&result.UniqueTitles); err != nil {
		return nil, fmt.Errorf("counting unique titles: %w", err)
	}

	// Most watched (top 10)
	rows, err := r.db.QueryContext(ctx,
		`SELECT a.media_id, COALESCE(m.title, 'Unknown'), COUNT(*) as play_count, COALESCE(m.poster_path, '')
		 FROM activity_log a
		 LEFT JOIN media m ON a.media_id = m.id
		 WHERE a.action = 'playback_start' AND a.media_id IS NOT NULL
		 GROUP BY a.media_id
		 ORDER BY play_count DESC
		 LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("querying most watched: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item model.MostWatchedItem
		if err := rows.Scan(&item.MediaID, &item.Title, &item.PlayCount, &item.PosterPath); err != nil {
			return nil, fmt.Errorf("scanning most watched: %w", err)
		}
		result.MostWatched = append(result.MostWatched, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Most active users (top 10)
	rows2, err := r.db.QueryContext(ctx,
		`SELECT a.user_id, COALESCE(u.username, 'Unknown'), COUNT(*) as action_count
		 FROM activity_log a
		 LEFT JOIN users u ON a.user_id = u.id
		 WHERE a.user_id IS NOT NULL
		 GROUP BY a.user_id
		 ORDER BY action_count DESC
		 LIMIT 10`)
	if err != nil {
		return nil, fmt.Errorf("querying most active users: %w", err)
	}
	defer rows2.Close()

	for rows2.Next() {
		var user model.MostActiveUser
		if err := rows2.Scan(&user.UserID, &user.Username, &user.ActionCount); err != nil {
			return nil, fmt.Errorf("scanning most active user: %w", err)
		}
		result.MostActive = append(result.MostActive, user)
	}
	return result, rows2.Err()
}
