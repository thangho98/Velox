package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/logger"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/transcoder"
	"github.com/thawng/velox/pkg/ffprobe"
	"github.com/thawng/velox/pkg/ophim"
)

// hlsCacheEntry caches PrepareHLS results so subsequent playlist polls
// skip DB queries entirely (Jellyfin-style: master playlist is cheap to serve).
type hlsCacheEntry struct {
	playlistPath string
	duration     float64 // total media duration (seconds) for VOD projection
}

type StreamService struct {
	db              *sql.DB
	mediaFileRepo   *repository.MediaFileRepo
	audioTrackRepo  *repository.AudioTrackRepo
	subtitleRepo    *repository.SubtitleRepo
	transcoder      *transcoder.Transcoder
	notificationSvc *NotificationService
	pretranscodeSvc *PretranscodeService
	log             *slog.Logger

	// hlsCache maps "mediaID:streamSessionID:startOffset" → cached playlist info.
	// Avoids 2 DB queries per playlist poll (~150-250ms savings on NAS HDD).
	hlsCache   map[string]*hlsCacheEntry
	hlsCacheMu sync.RWMutex

	cloudResolver func(context.Context, int64, *model.MediaFile) (string, error)
}

func NewStreamService(mediaFileRepo *repository.MediaFileRepo, audioTrackRepo *repository.AudioTrackRepo, transcoder *transcoder.Transcoder) *StreamService {
	return &StreamService{
		mediaFileRepo:  mediaFileRepo,
		audioTrackRepo: audioTrackRepo,
		transcoder:     transcoder,
		log:            logger.New("stream-svc"),
		hlsCache:       make(map[string]*hlsCacheEntry),
	}
}

// SetNotificationService sets the optional notification service for transcode events.
func (s *StreamService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

// SetPretranscodeService sets the optional pre-transcode service for pre-encoded file lookup.
func (s *StreamService) SetPretranscodeService(svc *PretranscodeService) {
	s.pretranscodeSvc = svc
}

// SetDB injects the database handle for transactional operations.
func (s *StreamService) SetDB(db *sql.DB) {
	s.db = db
}

// SetSubtitleRepo injects the subtitle repository for cloud probe subtitle persistence.
func (s *StreamService) SetSubtitleRepo(repo *repository.SubtitleRepo) {
	s.subtitleRepo = repo
}

// SetCloudResolver injects a function that translates cloud native paths into HTTP URLs.
func (s *StreamService) SetCloudResolver(resolver func(context.Context, int64, *model.MediaFile) (string, error)) {
	s.cloudResolver = resolver
}

// ResolveFilePath converts mf.FilePath to a playable HTTP URL if it is a cloud path.
func (s *StreamService) ResolveFilePath(ctx context.Context, mediaID int64, mf *model.MediaFile) (string, error) {
	if strings.HasPrefix(mf.FilePath, "ophim://") {
		parts := strings.Split(strings.TrimPrefix(mf.FilePath, "ophim://"), "/")
		slug := parts[0]
		client := ophim.New()
		movie, err := client.GetMovieDetails(ctx, slug)
		if err == nil && len(movie.Episodes) > 0 {
			if len(parts) > 1 {
				epSlug := parts[1]
				for _, ep := range movie.Episodes[0].ServerData {
					if ep.Slug == epSlug {
						return ep.LinkM3U8, nil
					}
				}
			} else if len(movie.Episodes[0].ServerData) > 0 {
				return movie.Episodes[0].ServerData[0].LinkM3U8, nil
			}
		}
		return "", fmt.Errorf("failed to fetch ophim stream link for slug: %s", slug)
	}

	if s.cloudResolver == nil || !scanner.IsCloudPath(mf.FilePath) {
		return mf.FilePath, nil
	}

	var resolved string
	var err error
	for attempt := 1; attempt <= 3; attempt++ {
		resolved, err = s.cloudResolver(ctx, mediaID, mf)
		if err == nil {
			return resolved, nil
		}
		s.log.Warn("failed to resolve cloud path", "attempt", attempt, "media_file_id", mf.ID, "error", err)
		if attempt < 3 {
			// Backoff: 1s, then 2s. If FShare API is 502ing, a short wait might help.
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}
	}
	return "", err
}

// FindPretranscode looks up a ready pre-transcode file for a media file.
// maxHeight: 0 means no limit (pick best available). Returns nil if none found.
func (s *StreamService) FindPretranscode(ctx context.Context, mediaFileID int64, maxHeight int) (*model.PretranscodeFile, error) {
	if s.pretranscodeSvc == nil {
		return nil, nil
	}
	return s.pretranscodeSvc.FindBestPretranscode(ctx, mediaFileID, maxHeight)
}

// FindPretranscodeProfile returns a pre-transcode profile by ID.
func (s *StreamService) FindPretranscodeProfile(ctx context.Context, profileID int64) (*model.PretranscodeProfile, error) {
	if s.pretranscodeSvc == nil {
		return nil, nil
	}
	return s.pretranscodeSvc.GetProfile(ctx, profileID)
}

// StopTranscode kills active FFmpeg transcode processes for a viewer stream
// session when available, otherwise falls back to media-wide cancellation.
func (s *StreamService) StopTranscode(mediaID int64, streamSessionID string) int {
	s.InvalidateHLSCache(streamSessionID)
	if streamSessionID != "" {
		return s.transcoder.CancelTranscodeByStreamSessionID(streamSessionID)
	}
	return s.transcoder.CancelTranscode(mediaID)
}

// TouchTranscodeActivity updates the last activity time for all active transcode
// jobs for a viewer stream session when available, otherwise falls back to
// media-wide activity updates for legacy callers.
func (s *StreamService) TouchTranscodeActivity(mediaID int64, streamSessionID string) {
	if streamSessionID != "" {
		s.transcoder.TouchJobByStreamSessionID(streamSessionID)
		return
	}
	s.transcoder.TouchJobByMediaID(mediaID)
}

// RemuxToPretranscode copies existing HLS transcode output into pretranscode MP4.
// Called after realtime transcode — "transcode once, instant forever".
func (s *StreamService) RemuxToPretranscode(ctx context.Context, mediaFileID int64, mediaID int64, height int) {
	if s.pretranscodeSvc == nil {
		return
	}
	// Find the HLS playlist for this quality level
	hlsDir := s.transcoder.HLSDir(mediaID)
	// ABR playlists are named: f{fileID}_q{height}.m3u8
	playlist := filepath.Join(hlsDir, fmt.Sprintf("f%d_q%d.m3u8", mediaFileID, height))
	if _, err := os.Stat(playlist); err != nil {
		return // no HLS segments for this quality
	}
	s.pretranscodeSvc.RemuxFromHLS(ctx, mediaFileID, height, playlist)
}

// FindAllPretranscodes returns all ready pre-transcode files for a media file.
func (s *StreamService) FindAllPretranscodes(ctx context.Context, mediaFileID int64) ([]model.PretranscodeFile, error) {
	if s.pretranscodeSvc == nil {
		return nil, nil
	}
	return s.pretranscodeSvc.ListReadyFiles(ctx, mediaFileID)
}

// FindAllPretranscodesWithProfiles returns ready files joined with profiles (1 query, no N+1).
func (s *StreamService) FindAllPretranscodesWithProfiles(ctx context.Context, mediaFileID int64) ([]repository.ReadyFileWithProfile, error) {
	if s.pretranscodeSvc == nil {
		return nil, nil
	}
	return s.pretranscodeSvc.ListReadyFilesWithProfiles(ctx, mediaFileID)
}

// PrepareHLS triggers transcoding if needed and returns the master playlist path.
// If multiple audio tracks exist, generates HLS with #EXT-X-MEDIA support.
// fileID: if > 0, transcode that specific file; otherwise use the primary file for mediaID.
// subtitleStreamIndex: if >= 0, burn-in that subtitle stream into the video.
// videoCopy: if true, copy the video stream unchanged (only transcode audio).
// startOffset: if > 0, begin the HLS session at that global timeline position.
func (s *StreamService) PrepareHLS(ctx context.Context, mediaID int64, streamSessionID string, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64, maxHeight int) (string, error) {
	// Fast path: return cached playlist path (skips 2 DB queries per poll).
	if streamSessionID != "" {
		cacheKey := hlsCacheKey(mediaID, streamSessionID, startOffset)
		s.hlsCacheMu.RLock()
		entry, ok := s.hlsCache[cacheKey]
		s.hlsCacheMu.RUnlock()
		if ok {
			s.log.Debug("PrepareHLS cache hit", "media_id", mediaID, "session", streamSessionID)
			return entry.playlistPath, nil
		}
	}
	s.log.Debug("PrepareHLS starting", "media_id", mediaID, "session", streamSessionID, "start_offset", startOffset, "video_copy", videoCopy)

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
	if errors.Is(err, repository.ErrNotFound) {
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
	s.log.Debug("PrepareHLS transcoding",
		"media_id", mediaID,
		"file", mf.FilePath,
		"audio_tracks", len(audioTracks),
		"duration", mf.Duration,
	)
	resolvedPath, resErr := s.ResolveFilePath(ctx, mediaID, mf)
	if resErr != nil {
		return "", resErr
	}

	transcodeStart := time.Now()
	if len(audioTracks) > 1 {
		if err := s.transcoder.GenerateHLSWithAudio(mediaID, streamSessionID, resolvedPath, audioTracks, mf.ID, subtitleStreamIndex, videoCopy, startOffset, maxHeight); err != nil {
			return "", err
		}
	} else {
		if err := s.transcoder.GenerateHLS(mediaID, streamSessionID, resolvedPath, mf.ID, subtitleStreamIndex, videoCopy, startOffset, maxHeight); err != nil {
			return "", err
		}
	}
	s.log.Debug("PrepareHLS transcode ready", "media_id", mediaID, "session", streamSessionID, "took", time.Since(transcodeStart))

	playlistPath := s.transcoder.MasterPlaylistPath(mediaID, streamSessionID, mf.ID, subtitleStreamIndex, videoCopy, startOffset)

	// Cache for subsequent polls (same session+offset = same playlist).
	if streamSessionID != "" {
		s.hlsCacheMu.Lock()
		s.hlsCache[hlsCacheKey(mediaID, streamSessionID, startOffset)] = &hlsCacheEntry{
			playlistPath: playlistPath,
			duration:     mf.Duration,
		}
		s.hlsCacheMu.Unlock()
	}

	return playlistPath, nil
}

// HLSCachedDuration returns the cached total duration for a session's HLS playlist.
// Used by the handler to project remaining segments into a VOD playlist.
func (s *StreamService) HLSCachedDuration(mediaID int64, streamSessionID string, startOffset float64) float64 {
	if streamSessionID == "" {
		return 0
	}
	s.hlsCacheMu.RLock()
	defer s.hlsCacheMu.RUnlock()
	if entry, ok := s.hlsCache[hlsCacheKey(mediaID, streamSessionID, startOffset)]; ok {
		return entry.duration
	}
	return 0
}

// InvalidateHLSCache removes all cached playlist paths for a stream session.
// Called when a session is stopped or replaced by a seek.
func (s *StreamService) InvalidateHLSCache(streamSessionID string) {
	if streamSessionID == "" {
		return
	}
	needle := ":" + streamSessionID + ":"
	s.hlsCacheMu.Lock()
	for k := range s.hlsCache {
		if strings.Contains(k, needle) {
			delete(s.hlsCache, k)
		}
	}
	s.hlsCacheMu.Unlock()
}

func hlsCacheKey(mediaID int64, streamSessionID string, startOffset float64) string {
	return fmt.Sprintf("%d:%s:%.3f", mediaID, streamSessionID, startOffset)
}

// SegmentPath returns the path to an HLS segment.
func (s *StreamService) SegmentPath(mediaID int64, segment string) string {
	return s.transcoder.SegmentPath(mediaID, segment)
}

// WaitForSegment waits for a segment to appear on disk if a transcode is active.
func (s *StreamService) WaitForSegment(streamSessionID, path string, timeout time.Duration) bool {
	return s.transcoder.WaitForSegment(streamSessionID, path, timeout)
}

// GetPrimaryFile returns the media file for streaming.
// If fileID > 0, returns that specific file (verified to belong to mediaID);
// otherwise returns the primary file for mediaID.
func (s *StreamService) GetPrimaryFile(ctx context.Context, mediaID, fileID int64) (*model.MediaFile, error) {
	if fileID > 0 {
		mf, err := s.mediaFileRepo.GetByID(ctx, fileID)
		if errors.Is(err, repository.ErrNotFound) || (err == nil && mf.MediaID != mediaID) {
			return nil, ErrNotFound
		}
		return mf, err
	}
	mf, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return mf, err
}

// ProbeAndUpdateCloudMetadata fetches and saves missing metadata for cloud media on the fly.
func (s *StreamService) ProbeAndUpdateCloudMetadata(ctx context.Context, mf *model.MediaFile) error {
	resolvedPath, err := s.ResolveFilePath(ctx, mf.MediaID, mf)
	if err != nil {
		return fmt.Errorf("failed to resolve path for probe: %w", err)
	}
	if resolvedPath == "" {
		return fmt.Errorf("resolved empty path for media file %d", mf.ID)
	}

	s.log.Info("Probing un-indexed cloud media on the fly", "media_file_id", mf.ID)

	probeResult, err := ffprobe.Probe(resolvedPath)
	if err != nil {
		return fmt.Errorf("ffprobe failed: %w", err)
	}

	mf.VideoCodec = probeResult.VideoCodec
	mf.VideoProfile = probeResult.VideoProfile
	mf.VideoLevel = probeResult.VideoLevel
	mf.VideoFPS = probeResult.VideoFPS
	mf.AudioCodec = probeResult.AudioCodec
	mf.Container = probeResult.Container
	// Some fast probe might not get width/height exactly right if relying on format only,
	// but ffprobe usually extracts it from streams.
	if probeResult.Width > 0 && probeResult.Height > 0 {
		mf.Width = probeResult.Width
		mf.Height = probeResult.Height
	}
	if probeResult.Duration > 0 {
		mf.Duration = probeResult.Duration
	}
	mf.Bitrate = int(probeResult.Bitrate)
	mf.IsHDR = probeResult.IsHDR
	mf.DVProfile = probeResult.DVProfile
	mf.ColorTransfer = probeResult.ColorTransfer
	mf.ColorPrimaries = probeResult.ColorPrimaries

	if err := s.mediaFileRepo.Update(ctx, mf); err != nil {
		return fmt.Errorf("failed to update media file: %w", err)
	}

	// Persist audio tracks and subtitles atomically when db is available.
	if s.db != nil {
		if err := s.persistProbedTracks(ctx, mf.ID, probeResult); err != nil {
			s.log.Warn("cloud probe: persist tracks", "media_file_id", mf.ID, "error", err)
		}
	} else {
		// Fallback: no transaction (background scan context without db handle)
		s.persistProbedTracksNoTx(ctx, mf.ID, probeResult)
	}

	return nil
}

// persistProbedTracks deletes and recreates audio tracks + subtitles inside a transaction.
func (s *StreamService) persistProbedTracks(ctx context.Context, mediaFileID int64, probe *ffprobe.ProbeResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	audioRepo := s.audioTrackRepo.WithTx(tx)
	if err := audioRepo.DeleteByMediaFileID(ctx, mediaFileID); err != nil {
		return fmt.Errorf("delete audio tracks: %w", err)
	}
	for _, t := range probe.AudioTracks {
		track := &model.AudioTrack{
			MediaFileID:   mediaFileID,
			StreamIndex:   t.StreamIndex,
			Language:      t.Language,
			Codec:         t.Codec,
			Channels:      t.Channels,
			ChannelLayout: t.ChannelLayout,
			Bitrate:       int(t.Bitrate),
			SampleRate:    int(t.SampleRate),
			Title:         t.Title,
			IsDefault:     t.IsDefault,
		}
		if err := audioRepo.Create(ctx, track); err != nil {
			return fmt.Errorf("create audio track %d: %w", t.StreamIndex, err)
		}
	}

	if s.subtitleRepo != nil && len(probe.Subtitles) > 0 {
		subRepo := s.subtitleRepo.WithTx(tx)
		if err := subRepo.DeleteByMediaFileID(ctx, mediaFileID); err != nil {
			return fmt.Errorf("delete subtitles: %w", err)
		}
		for _, sub := range probe.Subtitles {
			subtitle := &model.Subtitle{
				MediaFileID: mediaFileID,
				Language:    sub.Language,
				Codec:       sub.Codec,
				Title:       sub.Title,
				IsEmbedded:  true,
				StreamIndex: sub.StreamIndex,
				IsForced:    sub.IsForced,
				IsDefault:   sub.IsDefault,
				IsSDH:       sub.IsSDH,
			}
			if err := subRepo.Create(ctx, subtitle); err != nil {
				return fmt.Errorf("create subtitle %d: %w", sub.StreamIndex, err)
			}
		}
		s.log.Info("Persisted cloud embedded subtitles", "media_file_id", mediaFileID, "count", len(probe.Subtitles))
	}

	return tx.Commit()
}

// persistProbedTracksNoTx is the non-transactional fallback (used by scanner where db may not be injected).
func (s *StreamService) persistProbedTracksNoTx(ctx context.Context, mediaFileID int64, probe *ffprobe.ProbeResult) {
	if err := s.audioTrackRepo.DeleteByMediaFileID(ctx, mediaFileID); err != nil {
		s.log.Warn("cloud probe: delete audio tracks", "media_file_id", mediaFileID, "error", err)
	}
	for _, t := range probe.AudioTracks {
		track := &model.AudioTrack{
			MediaFileID:   mediaFileID,
			StreamIndex:   t.StreamIndex,
			Language:      t.Language,
			Codec:         t.Codec,
			Channels:      t.Channels,
			ChannelLayout: t.ChannelLayout,
			Bitrate:       int(t.Bitrate),
			SampleRate:    int(t.SampleRate),
			Title:         t.Title,
			IsDefault:     t.IsDefault,
		}
		if err := s.audioTrackRepo.Create(ctx, track); err != nil {
			s.log.Warn("cloud probe: create audio track", "media_file_id", mediaFileID, "stream", t.StreamIndex, "error", err)
		}
	}

	if s.subtitleRepo != nil && len(probe.Subtitles) > 0 {
		if err := s.subtitleRepo.DeleteByMediaFileID(ctx, mediaFileID); err != nil {
			s.log.Warn("cloud probe: delete subtitles", "media_file_id", mediaFileID, "error", err)
		}
		for _, sub := range probe.Subtitles {
			subtitle := &model.Subtitle{
				MediaFileID: mediaFileID,
				Language:    sub.Language,
				Codec:       sub.Codec,
				Title:       sub.Title,
				IsEmbedded:  true,
				StreamIndex: sub.StreamIndex,
				IsForced:    sub.IsForced,
				IsDefault:   sub.IsDefault,
				IsSDH:       sub.IsSDH,
			}
			if err := s.subtitleRepo.Create(ctx, subtitle); err != nil {
				s.log.Warn("cloud probe: create subtitle", "media_file_id", mediaFileID, "stream", sub.StreamIndex, "error", err)
			}
		}
		s.log.Info("Persisted cloud embedded subtitles", "media_file_id", mediaFileID, "count", len(probe.Subtitles))
	}
}

// GetAudioTrackForMediaFile returns an audio track if it belongs to the given media file.
func (s *StreamService) GetAudioTrackForMediaFile(ctx context.Context, mediaFileID, trackID int64) (*model.AudioTrack, error) {
	track, err := s.audioTrackRepo.GetByID(ctx, trackID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if track.MediaFileID != mediaFileID {
		return nil, ErrNotFound
	}
	return track, nil
}

// ListAudioTracksForMediaFile returns all audio tracks for a given media file.
func (s *StreamService) ListAudioTracksForMediaFile(ctx context.Context, mediaFileID int64) ([]model.AudioTrack, error) {
	return s.audioTrackRepo.ListByMediaFileID(ctx, mediaFileID)
}

// PrepareABRHLS triggers multi-quality adaptive bitrate HLS transcoding and
// returns the ABR master playlist path.
// If fileID > 0, uses that specific file (verified to belong to mediaID).
func (s *StreamService) PrepareABRHLS(ctx context.Context, mediaID, fileID int64) (string, error) {
	mf, err := s.GetPrimaryFile(ctx, mediaID, fileID)
	if err != nil {
		return "", err
	}
	resolvedPath, err := s.ResolveFilePath(ctx, mediaID, mf)
	if err != nil {
		return "", err
	}
	if err := s.transcoder.GenerateABRHLS(mediaID, resolvedPath, mf.Height, mf.ID); err != nil {
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
		ctx := context.Background()
		resolvedPath := inputPath

		if mf, err := s.mediaFileRepo.GetByID(ctx, fileID); err == nil {
			if path, idxErr := s.ResolveFilePath(ctx, mediaID, mf); idxErr == nil {
				resolvedPath = path
			}
		}

		err := s.transcoder.GenerateABRHLS(mediaID, resolvedPath, sourceHeight, fileID)
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
func (s *StreamService) RemuxToWriter(ctx context.Context, mediaID int64, mf *model.MediaFile, w io.Writer) error {
	resolvedPath, err := s.ResolveFilePath(ctx, mediaID, mf)
	if err != nil {
		return err
	}
	return s.transcoder.RemuxToWriter(resolvedPath, w)
}

// RemuxSelectedAudioToWriter remuxes the file while selecting a specific audio stream.
func (s *StreamService) RemuxSelectedAudioToWriter(ctx context.Context, mediaID int64, mf *model.MediaFile, audioStreamIndex int, w io.Writer) error {
	resolvedPath, err := s.ResolveFilePath(ctx, mediaID, mf)
	if err != nil {
		return err
	}
	return s.transcoder.RemuxSelectedAudioToWriter(resolvedPath, audioStreamIndex, w)
}
