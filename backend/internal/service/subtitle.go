package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SubtitleService handles subtitle business logic
type SubtitleService struct {
	subtitleRepo *repository.SubtitleRepo
}

func NewSubtitleService(subtitleRepo *repository.SubtitleRepo) *SubtitleService {
	return &SubtitleService{subtitleRepo: subtitleRepo}
}

// ListByMediaFile returns all subtitles for a media file
func (s *SubtitleService) ListByMediaFile(ctx context.Context, mediaFileID int64) ([]model.Subtitle, error) {
	return s.subtitleRepo.ListByMediaFileID(ctx, mediaFileID)
}

// Get returns a subtitle by ID
func (s *SubtitleService) Get(ctx context.Context, id int64) (*model.Subtitle, error) {
	sub, err := s.subtitleRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return sub, nil
}

// Create creates a new subtitle
func (s *SubtitleService) Create(ctx context.Context, subtitle *model.Subtitle) error {
	return s.subtitleRepo.Create(ctx, subtitle)
}

// Update updates a subtitle
func (s *SubtitleService) Update(ctx context.Context, subtitle *model.Subtitle) error {
	return s.subtitleRepo.Update(ctx, subtitle)
}

// Delete deletes a subtitle
func (s *SubtitleService) Delete(ctx context.Context, id int64) error {
	return s.subtitleRepo.Delete(ctx, id)
}

// SetDefault sets a subtitle as default
func (s *SubtitleService) SetDefault(ctx context.Context, mediaFileID, subtitleID int64) error {
	return s.subtitleRepo.SetDefault(ctx, mediaFileID, subtitleID)
}

// AudioTrackService handles audio track business logic
type AudioTrackService struct {
	audioTrackRepo *repository.AudioTrackRepo
}

func NewAudioTrackService(audioTrackRepo *repository.AudioTrackRepo) *AudioTrackService {
	return &AudioTrackService{audioTrackRepo: audioTrackRepo}
}

// ListByMediaFile returns all audio tracks for a media file
func (s *AudioTrackService) ListByMediaFile(ctx context.Context, mediaFileID int64) ([]model.AudioTrack, error) {
	return s.audioTrackRepo.ListByMediaFileID(ctx, mediaFileID)
}

// Get returns an audio track by ID
func (s *AudioTrackService) Get(ctx context.Context, id int64) (*model.AudioTrack, error) {
	track, err := s.audioTrackRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
	return s.audioTrackRepo.Delete(ctx, id)
}

// SetDefault sets an audio track as default
func (s *AudioTrackService) SetDefault(ctx context.Context, mediaFileID, trackID int64) error {
	return s.audioTrackRepo.SetDefault(ctx, mediaFileID, trackID)
}
