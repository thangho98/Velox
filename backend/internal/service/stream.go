package service

import (
	"context"
	"database/sql"
	"errors"
	"io"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/transcoder"
)

type StreamService struct {
	mediaFileRepo  *repository.MediaFileRepo
	audioTrackRepo *repository.AudioTrackRepo
	transcoder     *transcoder.Transcoder
}

func NewStreamService(mediaFileRepo *repository.MediaFileRepo, audioTrackRepo *repository.AudioTrackRepo, transcoder *transcoder.Transcoder) *StreamService {
	return &StreamService{
		mediaFileRepo:  mediaFileRepo,
		audioTrackRepo: audioTrackRepo,
		transcoder:     transcoder,
	}
}

// PrepareHLS triggers transcoding if needed and returns the master playlist path.
// If multiple audio tracks exist, generates HLS with #EXT-X-MEDIA support.
// fileID: if > 0, transcode that specific file; otherwise use the primary file for mediaID.
// subtitleStreamIndex: if >= 0, burn-in that subtitle stream into the video.
func (s *StreamService) PrepareHLS(ctx context.Context, mediaID int64, fileID int64, subtitleStreamIndex int) (string, error) {
	var mf *model.MediaFile
	var err error
	if fileID > 0 {
		mf, err = s.mediaFileRepo.GetByID(ctx, fileID)
		if err == nil && mf.MediaID != mediaID {
			return "", ErrNotFound
		}
	} else {
		mf, err = s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}

	// Check for multiple audio tracks
	audioTracks, err := s.audioTrackRepo.ListByMediaFileID(ctx, mf.ID)
	if err != nil {
		// Fall back to simple HLS if can't get audio tracks
		audioTracks = nil
	}

	// Use multi-audio HLS if more than one track.
	// Pass mf.ID so the cache key is (mediaID, fileID, subtitleStreamIndex) — avoids
	// version collisions when multiple file versions exist for the same media.
	if len(audioTracks) > 1 {
		if err := s.transcoder.GenerateHLSWithAudio(mediaID, mf.FilePath, audioTracks, mf.ID, subtitleStreamIndex); err != nil {
			return "", err
		}
	} else {
		if err := s.transcoder.GenerateHLS(mediaID, mf.FilePath, mf.ID, subtitleStreamIndex); err != nil {
			return "", err
		}
	}

	return s.transcoder.MasterPlaylistPath(mediaID, mf.ID, subtitleStreamIndex), nil
}

// SegmentPath returns the path to an HLS segment.
func (s *StreamService) SegmentPath(mediaID int64, segment string) string {
	return s.transcoder.SegmentPath(mediaID, segment)
}

// GetPrimaryFile returns the media file for streaming.
// If fileID > 0, returns that specific file (verified to belong to mediaID);
// otherwise returns the primary file for mediaID.
func (s *StreamService) GetPrimaryFile(ctx context.Context, mediaID, fileID int64) (*model.MediaFile, error) {
	if fileID > 0 {
		mf, err := s.mediaFileRepo.GetByID(ctx, fileID)
		if errors.Is(err, sql.ErrNoRows) || (err == nil && mf.MediaID != mediaID) {
			return nil, ErrNotFound
		}
		return mf, err
	}
	mf, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return mf, err
}

// RemuxToWriter remuxes the file to fragmented MP4 and writes to w.
// Used for DirectStream: container-only remux, no codec transcoding.
func (s *StreamService) RemuxToWriter(inputPath string, w io.Writer) error {
	return s.transcoder.RemuxToWriter(inputPath, w)
}
