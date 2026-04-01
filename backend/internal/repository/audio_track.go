package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
)

// AudioTrackRepo handles audio_tracks database operations
type AudioTrackRepo struct {
	db DBTX
}

func NewAudioTrackRepo(db DBTX) *AudioTrackRepo {
	return &AudioTrackRepo{db: db}
}

// WithTx returns a copy of the repo that uses the given transaction.
func (r *AudioTrackRepo) WithTx(tx *sql.Tx) *AudioTrackRepo {
	return &AudioTrackRepo{db: tx}
}

// Create inserts a new audio track
func (r *AudioTrackRepo) Create(ctx context.Context, at *model.AudioTrack) error {
	query := `INSERT INTO audio_tracks
		(media_file_id, stream_index, codec, language, channels, channel_layout, bitrate, sample_rate, title, is_default)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id`

	isDefault := 0
	if at.IsDefault {
		isDefault = 1
	}

	row := r.db.QueryRowContext(ctx, query,
		at.MediaFileID, at.StreamIndex, at.Codec, at.Language, at.Channels,
		at.ChannelLayout, at.Bitrate, at.SampleRate, at.Title, isDefault)
	return row.Scan(&at.ID)
}

// GetByID retrieves an audio track by ID
func (r *AudioTrackRepo) GetByID(ctx context.Context, id int64) (*model.AudioTrack, error) {
	var at model.AudioTrack
	var isDefault int

	err := r.db.QueryRowContext(ctx, `SELECT id, media_file_id, stream_index, codec,
		language, channels, channel_layout, bitrate, sample_rate, title, is_default
		FROM audio_tracks WHERE id = ?`, id).
		Scan(&at.ID, &at.MediaFileID, &at.StreamIndex, &at.Codec,
			&at.Language, &at.Channels, &at.ChannelLayout, &at.Bitrate, &at.SampleRate, &at.Title, &isDefault)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get audio track by id %d: %w", id, err)
	}
	at.IsDefault = isDefault == 1
	return &at, nil
}

// ListByMediaFileID retrieves all audio tracks for a media file
func (r *AudioTrackRepo) ListByMediaFileID(ctx context.Context, mediaFileID int64) ([]model.AudioTrack, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, media_file_id, stream_index, codec,
		language, channels, channel_layout, bitrate, sample_rate, title, is_default
		FROM audio_tracks WHERE media_file_id = ?
		ORDER BY is_default DESC, stream_index`, mediaFileID)
	if err != nil {
		return nil, fmt.Errorf("listing audio tracks: %w", err)
	}
	defer rows.Close()

	var items []model.AudioTrack
	for rows.Next() {
		var at model.AudioTrack
		var isDefault int

		if err := rows.Scan(&at.ID, &at.MediaFileID, &at.StreamIndex, &at.Codec,
			&at.Language, &at.Channels, &at.ChannelLayout, &at.Bitrate, &at.SampleRate, &at.Title, &isDefault); err != nil {
			return nil, fmt.Errorf("scanning audio track: %w", err)
		}
		at.IsDefault = isDefault == 1
		items = append(items, at)
	}
	return items, rows.Err()
}

// DeleteByMediaFileID removes all audio tracks for a media file (for rescan)
func (r *AudioTrackRepo) DeleteByMediaFileID(ctx context.Context, mediaFileID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM audio_tracks WHERE media_file_id = ?", mediaFileID)
	if err != nil {
		return fmt.Errorf("deleting audio tracks for media file %d: %w", mediaFileID, err)
	}
	return nil
}

// Delete removes an audio track
func (r *AudioTrackRepo) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM audio_tracks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete audio track %d: %w", id, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Update updates an audio track
func (r *AudioTrackRepo) Update(ctx context.Context, at *model.AudioTrack) error {
	isDefault := 0
	if at.IsDefault {
		isDefault = 1
	}

	res, err := r.db.ExecContext(ctx, `UPDATE audio_tracks SET
		codec = ?, language = ?, channels = ?, channel_layout = ?, bitrate = ?, title = ?, is_default = ?
		WHERE id = ?`,
		at.Codec, at.Language, at.Channels, at.ChannelLayout, at.Bitrate, at.Title, isDefault, at.ID)
	if err != nil {
		return fmt.Errorf("update audio track %d: %w", at.ID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetDefault sets an audio track as the default for its media file.
// Returns ErrNotFound if the track does not exist or does not belong to the media file.
func (r *AudioTrackRepo) SetDefault(ctx context.Context, mediaFileID, trackID int64) error {
	// First validate the track exists and belongs to this media file
	// This prevents partial state if the track is stale
	var exists int
	err := r.db.QueryRowContext(ctx,
		"SELECT 1 FROM audio_tracks WHERE id = ? AND media_file_id = ?",
		trackID, mediaFileID).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("checking audio track existence: %w", err)
	}

	// Track validated, now clear default for all tracks of this media file
	_, err = r.db.ExecContext(ctx, "UPDATE audio_tracks SET is_default = 0 WHERE media_file_id = ?", mediaFileID)
	if err != nil {
		return fmt.Errorf("clearing default track for media file %d: %w", mediaFileID, err)
	}

	// Set the new default
	_, err = r.db.ExecContext(ctx, "UPDATE audio_tracks SET is_default = 1 WHERE id = ? AND media_file_id = ?", trackID, mediaFileID)
	if err != nil {
		return fmt.Errorf("setting default track %d for media file %d: %w", trackID, mediaFileID, err)
	}
	return nil
}
