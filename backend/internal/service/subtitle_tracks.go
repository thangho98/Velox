package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// AudioTrackService handles audio track business logic
type AudioTrackService struct {
	db             *sql.DB
	audioTrackRepo *repository.AudioTrackRepo
}

func NewAudioTrackService(db *sql.DB, audioTrackRepo *repository.AudioTrackRepo) *AudioTrackService {
	return &AudioTrackService{
		db:             db,
		audioTrackRepo: audioTrackRepo,
	}
}

// ListByMediaFile returns all audio tracks for a media file
func (s *AudioTrackService) ListByMediaFile(ctx context.Context, mediaFileID int64) ([]model.AudioTrack, error) {
	return s.audioTrackRepo.ListByMediaFileID(ctx, mediaFileID)
}

// Get returns an audio track by ID
func (s *AudioTrackService) Get(ctx context.Context, id int64) (*model.AudioTrack, error) {
	track, err := s.audioTrackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return track, nil
}

// Create creates a new audio track
func (s *AudioTrackService) Create(ctx context.Context, track *model.AudioTrack) error {
	return s.audioTrackRepo.Create(ctx, track)
}

// Update updates an audio track
func (s *AudioTrackService) Update(ctx context.Context, track *model.AudioTrack) error {
	return s.audioTrackRepo.Update(ctx, track)
}

// Delete deletes an audio track
func (s *AudioTrackService) Delete(ctx context.Context, id int64) error {
	if err := s.audioTrackRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

// SetDefault sets an audio track as default (atomic: clear all defaults + set new default in one transaction)
func (s *AudioTrackService) SetDefault(ctx context.Context, mediaFileID, trackID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Clear all defaults for this media file
	if _, err := tx.ExecContext(ctx, "UPDATE audio_tracks SET is_default = 0 WHERE media_file_id = ?", mediaFileID); err != nil {
		return fmt.Errorf("clearing default tracks for media file %d: %w", mediaFileID, err)
	}

	// Set the new default
	res, err := tx.ExecContext(ctx, "UPDATE audio_tracks SET is_default = 1 WHERE id = ? AND media_file_id = ?", trackID, mediaFileID)
	if err != nil {
		return fmt.Errorf("setting default track %d for media file %d: %w", trackID, mediaFileID, err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}
