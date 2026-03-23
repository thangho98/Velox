package service

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/transcoder"
)

type StreamService struct {
	mediaFileRepo   *repository.MediaFileRepo
	audioTrackRepo  *repository.AudioTrackRepo
	transcoder      *transcoder.Transcoder
	notificationSvc *NotificationService
}

func NewStreamService(mediaFileRepo *repository.MediaFileRepo, audioTrackRepo *repository.AudioTrackRepo, transcoder *transcoder.Transcoder) *StreamService {
	return &StreamService{
		mediaFileRepo:  mediaFileRepo,
		audioTrackRepo: audioTrackRepo,
		transcoder:     transcoder,
	}
}

// SetNotificationService sets the optional notification service for transcode events.
func (s *StreamService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

// PrepareHLS triggers transcoding if needed and returns the master playlist path.
// If multiple audio tracks exist, generates HLS with #EXT-X-MEDIA support.
// fileID: if > 0, transcode that specific file; otherwise use the primary file for mediaID.
// subtitleStreamIndex: if >= 0, burn-in that subtitle stream into the video.
// videoCopy: if true, copy the video stream unchanged (only transcode audio).
func (s *StreamService) PrepareHLS(ctx context.Context, mediaID int64, fileID int64, subtitleStreamIndex int, videoCopy bool) (string, error) {
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
		log.Printf("stream: list audio tracks for file %d: %v — falling back to simple HLS", mf.ID, err)
		audioTracks = nil
	}

	// Use multi-audio HLS if more than one track.
	// Pass mf.ID so the cache key is (mediaID, fileID, subtitleStreamIndex) — avoids
	// version collisions when multiple file versions exist for the same media.
	if len(audioTracks) > 1 {
		if err := s.transcoder.GenerateHLSWithAudio(mediaID, mf.FilePath, audioTracks, mf.ID, subtitleStreamIndex, videoCopy); err != nil {
			return "", err
		}
	} else {
		if err := s.transcoder.GenerateHLS(mediaID, mf.FilePath, mf.ID, subtitleStreamIndex, videoCopy); err != nil {
			return "", err
		}
	}

	return s.transcoder.MasterPlaylistPath(mediaID, mf.ID, subtitleStreamIndex, videoCopy), nil
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

// PrepareABRHLS triggers multi-quality adaptive bitrate HLS transcoding and
// returns the ABR master playlist path.
// If fileID > 0, uses that specific file (verified to belong to mediaID).
func (s *StreamService) PrepareABRHLS(ctx context.Context, mediaID, fileID int64) (string, error) {
	mf, err := s.GetPrimaryFile(ctx, mediaID, fileID)
	if err != nil {
		return "", err
	}
	if err := s.transcoder.GenerateABRHLS(mediaID, mf.FilePath, mf.Height, mf.ID); err != nil {
		return "", err
	}
	return s.transcoder.ABRMasterPath(mediaID, mf.ID), nil
}

// ABRCached reports whether the ABR master playlist for (mediaID, fileID) already exists on disk.
func (s *StreamService) ABRCached(mediaID, fileID int64) bool {
	return s.transcoder.ABRCached(mediaID, fileID)
}

// StartABRBackground triggers ABR HLS generation for the given media file
// asynchronously. The caller does not wait for completion.
// userID and mediaTitle are used to notify the requesting user when done.
func (s *StreamService) StartABRBackground(userID, mediaID, fileID int64, inputPath, mediaTitle string, sourceHeight int) {
	go func() {
		err := s.transcoder.GenerateABRHLS(mediaID, inputPath, sourceHeight, fileID)
		if err != nil {
			log.Printf("stream: background ABR generation for media %d file %d: %v", mediaID, fileID, err)
		}
		if s.notificationSvc != nil {
			ctx := context.Background()
			_ = s.notificationSvc.NotifyTranscodeComplete(ctx, userID, mediaID, mediaTitle, err == nil, "ABR", 0)
		}
	}()
}

// RemuxToWriter remuxes the file to fragmented MP4 and writes to w.
// Used for DirectStream: container-only remux, no codec transcoding.
func (s *StreamService) RemuxToWriter(inputPath string, w io.Writer) error {
	return s.transcoder.RemuxToWriter(inputPath, w)
}
