package service

import (
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/transcoder"
)

type StreamService struct {
	mediaRepo  *repository.MediaRepo
	transcoder *transcoder.Transcoder
}

func NewStreamService(mediaRepo *repository.MediaRepo, transcoder *transcoder.Transcoder) *StreamService {
	return &StreamService{mediaRepo: mediaRepo, transcoder: transcoder}
}

// DirectPlayPath returns the file path for direct playback.
func (s *StreamService) DirectPlayPath(mediaID int64) (string, error) {
	m, err := s.mediaRepo.GetByID(mediaID)
	if err != nil {
		return "", err
	}
	return m.FilePath, nil
}

// PrepareHLS triggers transcoding if needed and returns the master playlist path.
func (s *StreamService) PrepareHLS(mediaID int64) (string, error) {
	m, err := s.mediaRepo.GetByID(mediaID)
	if err != nil {
		return "", err
	}

	if err := s.transcoder.GenerateHLS(mediaID, m.FilePath); err != nil {
		return "", err
	}

	return s.transcoder.SegmentPath(mediaID, "master.m3u8"), nil
}

// SegmentPath returns the path to an HLS segment.
func (s *StreamService) SegmentPath(mediaID int64, segment string) string {
	return s.transcoder.SegmentPath(mediaID, segment)
}
