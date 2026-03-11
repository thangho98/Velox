package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/transcoder"
)

type StreamService struct {
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	transcoder    *transcoder.Transcoder
}

func NewStreamService(mediaRepo *repository.MediaRepo, mediaFileRepo *repository.MediaFileRepo, transcoder *transcoder.Transcoder) *StreamService {
	return &StreamService{
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		transcoder:    transcoder,
	}
}

// DirectPlayPath returns the file path for direct playback.
func (s *StreamService) DirectPlayPath(ctx context.Context, mediaID int64) (string, error) {
	mf, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return mf.FilePath, nil
}

// PrepareHLS triggers transcoding if needed and returns the master playlist path.
func (s *StreamService) PrepareHLS(ctx context.Context, mediaID int64) (string, error) {
	mf, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	if err := s.transcoder.GenerateHLS(mediaID, mf.FilePath); err != nil {
		return "", err
	}

	return s.transcoder.SegmentPath(mediaID, "master.m3u8"), nil
}

// SegmentPath returns the path to an HLS segment.
func (s *StreamService) SegmentPath(mediaID int64, segment string) string {
	return s.transcoder.SegmentPath(mediaID, segment)
}
