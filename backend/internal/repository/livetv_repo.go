package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/thawng/velox/internal/model"
)

type LiveTVRepo struct {
	db DBTX
}

func NewLiveTVRepo(db DBTX) *LiveTVRepo {
	return &LiveTVRepo{db: db}
}

// CreatePlaylist adds a new Live TV playlist source
func (r *LiveTVRepo) CreatePlaylist(ctx context.Context, p *model.LivePlaylist) error {
	query := `
		INSERT INTO live_playlists (name, url, epg_url, sync_interval_hours, is_active)
		VALUES (?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query, p.Name, p.URL, p.EpgURL, p.SyncIntervalHours, p.IsActive).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create live playlist: %w", err)
	}
	return nil
}

// GetPlaylist returns a single playlist by ID
func (r *LiveTVRepo) GetPlaylist(ctx context.Context, id int64) (*model.LivePlaylist, error) {
	query := `SELECT id, name, url, epg_url, last_synced_at, sync_interval_hours, is_active, created_at, updated_at FROM live_playlists WHERE id = ?`
	p := &model.LivePlaylist{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.URL, &p.EpgURL, &p.LastSyncedAt, &p.SyncIntervalHours, &p.IsActive, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get playlist by id: %w", err)
	}
	return p, nil
}

// GetPlaylists returns all playlists for admin
func (r *LiveTVRepo) GetPlaylists(ctx context.Context) ([]*model.LivePlaylist, error) {
	query := `SELECT id, name, url, epg_url, last_synced_at, sync_interval_hours, is_active, created_at, updated_at FROM live_playlists`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pList []*model.LivePlaylist
	for rows.Next() {
		p := &model.LivePlaylist{}
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.EpgURL, &p.LastSyncedAt, &p.SyncIntervalHours, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pList = append(pList, p)
	}
	return pList, rows.Err()
}

// GetActivePlaylists returns all active playlists for background sync
func (r *LiveTVRepo) GetActivePlaylists(ctx context.Context) ([]*model.LivePlaylist, error) {
	query := `SELECT id, name, url, epg_url, last_synced_at, sync_interval_hours, is_active, created_at, updated_at FROM live_playlists WHERE is_active = 1`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pList []*model.LivePlaylist
	for rows.Next() {
		p := &model.LivePlaylist{}
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.EpgURL, &p.LastSyncedAt, &p.SyncIntervalHours, &p.IsActive, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		pList = append(pList, p)
	}
	return pList, rows.Err()
}

// SyncChannelsTransaction safely syncs the channels for a playlist within a transaction:
// 1. Deactivates all existing channels for the playlist.
// 2. Upserts the newly fetched channels using batch sizes.
func (r *LiveTVRepo) SyncChannelsTransaction(ctx context.Context, playlistID int64, channels []*model.LiveChannel) error {
	sqlDB, ok := r.db.(*sql.DB)
	if !ok {
		// Fallback to non-transactional if db is already a Tx or unsupported
		return r.syncChannelsNonTx(ctx, playlistID, channels)
	}

	tx, err := sqlDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Deactivate
	_, err = tx.ExecContext(ctx, `UPDATE live_channels SET is_active = 0 WHERE playlist_id = ?`, playlistID)
	if err != nil {
		return fmt.Errorf("deactivate channels: %w", err)
	}

	if len(channels) > 0 {
		batchSize := 100
		for i := 0; i < len(channels); i += batchSize {
			end := i + batchSize
			if end > len(channels) {
				end = len(channels)
			}
			chunk := channels[i:end]

			query := `INSERT INTO live_channels (playlist_id, channel_id, name, logo, group_title, stream_url, country, user_agent, referer, origin, is_active) VALUES `
			var args []any
			var placeholders []string

			for _, ch := range chunk {
				placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
				args = append(args, playlistID, ch.ChannelID, ch.Name, ch.Logo, ch.GroupTitle, ch.StreamURL, ch.Country, ch.UserAgent, ch.Referer, ch.Origin, ch.IsActive)
			}

			query += strings.Join(placeholders, ",")

			query += ` ON CONFLICT(playlist_id, name) DO UPDATE SET
				channel_id = excluded.channel_id,
				logo = excluded.logo,
				group_title = excluded.group_title,
				stream_url = excluded.stream_url,
				country = excluded.country,
				user_agent = excluded.user_agent,
				referer = excluded.referer,
				origin = excluded.origin,
				is_active = excluded.is_active,
				updated_at = CURRENT_TIMESTAMP
			`

			_, err = tx.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("bulk insert live channels (batch): %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// syncChannelsNonTx does the same as SyncChannelsTransaction but without standard transaction wrap
func (r *LiveTVRepo) syncChannelsNonTx(ctx context.Context, playlistID int64, channels []*model.LiveChannel) error {
	_, err := r.db.ExecContext(ctx, `UPDATE live_channels SET is_active = 0 WHERE playlist_id = ?`, playlistID)
	if err != nil {
		return fmt.Errorf("deactivate channels: %w", err)
	}

	if len(channels) > 0 {
		batchSize := 100
		for i := 0; i < len(channels); i += batchSize {
			end := i + batchSize
			if end > len(channels) {
				end = len(channels)
			}
			chunk := channels[i:end]

			query := `INSERT INTO live_channels (playlist_id, channel_id, name, logo, group_title, stream_url, country, user_agent, referer, origin, is_active) VALUES `
			var args []any
			var placeholders []string

			for _, ch := range chunk {
				placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
				args = append(args, playlistID, ch.ChannelID, ch.Name, ch.Logo, ch.GroupTitle, ch.StreamURL, ch.Country, ch.UserAgent, ch.Referer, ch.Origin, ch.IsActive)
			}

			query += strings.Join(placeholders, ",")

			query += ` ON CONFLICT(playlist_id, name) DO UPDATE SET
				channel_id = excluded.channel_id,
				logo = excluded.logo,
				group_title = excluded.group_title,
				stream_url = excluded.stream_url,
				country = excluded.country,
				user_agent = excluded.user_agent,
				referer = excluded.referer,
				origin = excluded.origin,
				is_active = excluded.is_active,
				updated_at = CURRENT_TIMESTAMP
			`

			_, err = r.db.ExecContext(ctx, query, args...)
			if err != nil {
				return fmt.Errorf("bulk insert live channels (batch): %w", err)
			}
		}
	}
	return nil
}

// UpdatePlaylist updates a playlist's properties
func (r *LiveTVRepo) UpdatePlaylist(ctx context.Context, p *model.LivePlaylist) error {
	query := `
		UPDATE live_playlists
		SET name = ?, url = ?, epg_url = ?, sync_interval_hours = ?, is_active = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`
	res, err := r.db.ExecContext(ctx, query, p.Name, p.URL, p.EpgURL, p.SyncIntervalHours, p.IsActive, p.ID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// DeletePlaylist removes a playlist and cascades to remove its channels
func (r *LiveTVRepo) DeletePlaylist(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM live_playlists WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// GetChannels fetches a paginated list of active channels from active playlists, optionally filtered by group_title
func (r *LiveTVRepo) GetChannels(ctx context.Context, groupTitle string, limit, offset int) ([]*model.LiveChannel, error) {
	query := `
		SELECT c.id, c.playlist_id, c.channel_id, c.name, c.logo, c.group_title, c.stream_url, c.country, c.user_agent, c.referer, c.origin, c.is_active, c.is_hidden, c.created_at, c.updated_at
		FROM live_channels c
		JOIN live_playlists p ON c.playlist_id = p.id
		WHERE c.is_active = 1 AND p.is_active = 1 AND c.is_hidden = 0
	`
	args := []any{}
	if groupTitle != "" {
		if strings.ToLower(groupTitle) == "hidden" {
			query = strings.Replace(query, "AND c.is_hidden = 0", "AND c.is_hidden = 1", 1)
		} else if groupTitle == "Others" {
			query += ` AND (';' || c.group_title || ';' LIKE ? OR c.group_title = '' OR LOWER(c.group_title) = 'undefined')`
			args = append(args, "%;"+groupTitle+";%")
		} else {
			query += ` AND ';' || c.group_title || ';' LIKE ?`
			args = append(args, "%;"+groupTitle+";%")
		}
	}

	query += ` ORDER BY c.playlist_id, c.group_title, c.name LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get channels: %w", err)
	}
	defer rows.Close()

	var channels []*model.LiveChannel
	for rows.Next() {
		c := &model.LiveChannel{}
		if err := rows.Scan(
			&c.ID, &c.PlaylistID, &c.ChannelID, &c.Name, &c.Logo, &c.GroupTitle, &c.StreamURL, &c.Country, &c.UserAgent, &c.Referer, &c.Origin, &c.IsActive, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan live channel: %w", err)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

// GetChannelByID fetches a single active channel by its primary key.
func (r *LiveTVRepo) GetChannelByID(ctx context.Context, id int64) (*model.LiveChannel, error) {
	query := `
		SELECT c.id, c.playlist_id, c.channel_id, c.name, c.logo, c.group_title, c.stream_url, c.country, c.user_agent, c.referer, c.origin, c.is_active, c.is_hidden, c.created_at, c.updated_at
		FROM live_channels c
		JOIN live_playlists p ON c.playlist_id = p.id
		WHERE c.id = ? AND c.is_active = 1 AND p.is_active = 1 AND c.is_hidden = 0
	`
	c := &model.LiveChannel{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.PlaylistID, &c.ChannelID, &c.Name, &c.Logo, &c.GroupTitle, &c.StreamURL, &c.Country, &c.UserAgent, &c.Referer, &c.Origin, &c.IsActive, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get channel by id: %w", err)
	}
	return c, nil
}

// GetChannelGroups returns a unique list of active channel group titles from active playlists
func (r *LiveTVRepo) GetChannelGroups(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT c.group_title 
		FROM live_channels c
		JOIN live_playlists p ON c.playlist_id = p.id
		WHERE c.is_active = 1 AND p.is_active = 1 AND c.is_hidden = 0
		ORDER BY c.group_title
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("get channel groups: %w", err)
	}
	defer rows.Close()

	var groups []string
	seen := make(map[string]bool)
	for rows.Next() {
		var rawGroup string
		if err := rows.Scan(&rawGroup); err != nil {
			return nil, fmt.Errorf("scan group title: %w", err)
		}

		// Unpack semicolon-separated groups (like "Animation;Classic")
		parts := strings.Split(rawGroup, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)

			if part == "" || strings.ToLower(part) == "undefined" {
				part = "Others"
			}

			if part != "" && !seen[part] {
				seen[part] = true
				groups = append(groups, part)
			}
		}
	}

	// Ensure groups are returned alphabetically
	sort.Strings(groups)
	return groups, rows.Err()
}

// UpdatePlaylistSyncedAt updates the last_synced_at timestamp
func (r *LiveTVRepo) UpdatePlaylistSyncedAt(ctx context.Context, playlistID int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE live_playlists SET last_synced_at = ? WHERE id = ?`, time.Now().UTC(), playlistID)
	return err
}

// SearchChannels searches active channels by name from active playlists
func (r *LiveTVRepo) SearchChannels(ctx context.Context, query string, limit int) ([]*model.LiveChannel, error) {
	searchQuery := `
		SELECT c.id, c.playlist_id, c.channel_id, c.name, c.logo, c.group_title, c.stream_url, c.country, c.user_agent, c.referer, c.origin, c.is_active, c.is_hidden, c.created_at, c.updated_at
		FROM live_channels c
		JOIN live_playlists p ON c.playlist_id = p.id
		WHERE c.is_active = 1 AND p.is_active = 1 AND c.is_hidden = 0 AND c.name LIKE ?
		ORDER BY c.name ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, searchQuery, "%"+query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("search channels: %w", err)
	}
	defer rows.Close()

	var channels []*model.LiveChannel
	for rows.Next() {
		c := &model.LiveChannel{}
		if err := rows.Scan(
			&c.ID, &c.PlaylistID, &c.ChannelID, &c.Name, &c.Logo, &c.GroupTitle, &c.StreamURL, &c.Country, &c.UserAgent, &c.Referer, &c.Origin, &c.IsActive, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan live channel: %w", err)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

// MarkChannelWatched upserts a row in user_live_channel_data recording that
// the user just played (or resumed) the given channel. Safe to call repeatedly.
func (r *LiveTVRepo) MarkChannelWatched(ctx context.Context, userID, channelID int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO user_live_channel_data (user_id, channel_id, play_count, last_watched_at, updated_at)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id, channel_id) DO UPDATE SET
			play_count = play_count + 1,
			last_watched_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`, userID, channelID)
	if err != nil {
		return fmt.Errorf("mark channel watched: %w", err)
	}
	return nil
}

// ListRecentChannels returns the user's most-recently-watched channels.
// Filters out channels that have become inactive/hidden or whose playlist was deactivated.
func (r *LiveTVRepo) ListRecentChannels(ctx context.Context, userID int64, limit int) ([]*model.LiveChannel, error) {
	query := `
		SELECT c.id, c.playlist_id, c.channel_id, c.name, c.logo, c.group_title, c.stream_url, c.country, c.user_agent, c.referer, c.origin, c.is_active, c.is_hidden, c.created_at, c.updated_at, u.last_watched_at
		FROM user_live_channel_data u
		JOIN live_channels c ON c.id = u.channel_id
		JOIN live_playlists p ON p.id = c.playlist_id
		WHERE u.user_id = ? AND c.is_active = 1 AND c.is_hidden = 0 AND p.is_active = 1
		ORDER BY u.last_watched_at DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent channels: %w", err)
	}
	defer rows.Close()

	var channels []*model.LiveChannel
	for rows.Next() {
		c := &model.LiveChannel{}
		var lastWatched sql.NullTime
		if err := rows.Scan(
			&c.ID, &c.PlaylistID, &c.ChannelID, &c.Name, &c.Logo, &c.GroupTitle, &c.StreamURL, &c.Country, &c.UserAgent, &c.Referer, &c.Origin, &c.IsActive, &c.IsHidden, &c.CreatedAt, &c.UpdatedAt, &lastWatched,
		); err != nil {
			return nil, fmt.Errorf("scan recent channel: %w", err)
		}
		if lastWatched.Valid {
			t := lastWatched.Time
			c.LastWatchedAt = &t
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

// ToggleChannelHidden toggles a channel's hidden status and returns the new state
func (r *LiveTVRepo) ToggleChannelHidden(ctx context.Context, id int64) (bool, error) {
	var isHidden int
	err := r.db.QueryRowContext(ctx, `
		UPDATE live_channels 
		SET is_hidden = CASE WHEN is_hidden = 1 THEN 0 ELSE 1 END,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
		RETURNING is_hidden`, id).Scan(&isHidden)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, ErrNotFound
		}
		return false, fmt.Errorf("toggle channel hidden: %w", err)
	}
	return isHidden == 1, nil
}
