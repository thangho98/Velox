package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

// PlaybackHandler provides playback decision and info endpoints
type PlaybackHandler struct {
	mediaSvc      *service.MediaService
	streamSvc     *service.StreamService
	userDataSvc   *service.UserDataService
	subtitleSvc   *service.SubtitleService
	audioTrackSvc *service.AudioTrackService
	prefRepo      *repository.UserPreferencesRepo
}

// NewPlaybackHandler creates a new playback handler
func NewPlaybackHandler(
	mediaSvc *service.MediaService,
	streamSvc *service.StreamService,
	userDataSvc *service.UserDataService,
	subtitleSvc *service.SubtitleService,
	audioTrackSvc *service.AudioTrackService,
	prefRepo *repository.UserPreferencesRepo,
) *PlaybackHandler {
	return &PlaybackHandler{
		mediaSvc:      mediaSvc,
		streamSvc:     streamSvc,
		userDataSvc:   userDataSvc,
		subtitleSvc:   subtitleSvc,
		audioTrackSvc: audioTrackSvc,
		prefRepo:      prefRepo,
	}
}

// PlaybackInfoRequest represents client-sent capabilities
type PlaybackInfoRequest struct {
	VideoCodecs        []string `json:"video_codecs,omitempty"`
	AudioCodecs        []string `json:"audio_codecs,omitempty"`
	Containers         []string `json:"containers,omitempty"`
	MaxHeight          int      `json:"max_height,omitempty"`
	PreferDirectPlay   bool     `json:"prefer_direct_play,omitempty"`
	MediaFileID        int64    `json:"media_file_id,omitempty"`        // specific file version
	SelectedAudioTrack int      `json:"selected_audio_track,omitempty"` // 0 = default
	SelectedSubtitle   string   `json:"selected_subtitle,omitempty"`    // language code or "off"
}

// PlaybackInfoResponse represents playback decision response
type PlaybackInfoResponse struct {
	MediaID          int                 `json:"media_id"`
	PrimaryFileID    int64               `json:"primary_file_id,omitempty"` // file ID used for this decision
	Method           string              `json:"method"`                    // DirectPlay, DirectStream, TranscodeAudio, FullTranscode
	StreamURL        string              `json:"stream_url"`
	VideoCodec       string              `json:"video_codec"`
	AudioCodec       string              `json:"audio_codec"`
	Container        string              `json:"container"`
	FileSize         int64               `json:"file_size,omitempty"`
	Bitrate          int                 `json:"bitrate,omitempty"`
	Duration         int                 `json:"duration,omitempty"`
	Width            int                 `json:"width,omitempty"`
	Height           int                 `json:"height,omitempty"`
	AudioTracks      []AudioTrackInfo    `json:"audio_tracks,omitempty"`
	SubtitleTracks   []SubtitleTrackInfo `json:"subtitle_tracks,omitempty"`
	DecisionReason   string              `json:"decision_reason"`
	EstimatedBitrate int                 `json:"estimated_bitrate,omitempty"`
	Position         float64             `json:"position,omitempty"` // Resume position
}

// AudioTrackInfo represents an audio track
type AudioTrackInfo struct {
	ID        int    `json:"id"`
	Language  string `json:"language"`
	Label     string `json:"label"`
	Codec     string `json:"codec,omitempty"`
	Channels  int    `json:"channels,omitempty"`
	IsDefault bool   `json:"is_default"`
	Selected  bool   `json:"selected"`
}

// SubtitleTrackInfo represents a subtitle track
type SubtitleTrackInfo struct {
	ID        int    `json:"id"`
	Language  string `json:"language"`
	Label     string `json:"label"`
	Format    string `json:"format"` // srt, vtt, pgs, etc.
	IsDefault bool   `json:"is_default"`
	IsImage   bool   `json:"is_image"` // PGS/VobSub require burn-in
}

// GetPlaybackInfo returns playback decision for a media item
func (h *PlaybackHandler) GetPlaybackInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, _, ok := auth.UserFromContext(ctx)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	// Get media file info
	media, err := h.mediaSvc.GetWithFiles(ctx, mediaID)
	if err != nil {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}

	if len(media.Files) == 0 {
		respondError(w, http.StatusNotFound, "no media files found")
		return
	}

	// Parse client capabilities from request body (optional)
	var clientCaps PlaybackInfoRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&clientCaps); err != nil {
			// Ignore decode errors, use UA detection
		}
	}

	// Select specific file version if requested, otherwise use primary (first)
	primaryFile := media.Files[0]
	if clientCaps.MediaFileID > 0 {
		for _, f := range media.Files {
			if f.ID == clientCaps.MediaFileID {
				primaryFile = f
				break
			}
		}
	}

	// Detect client profile from User-Agent
	profile := playback.DetectClient(r.UserAgent())

	// Override with client-provided capabilities if available
	if len(clientCaps.VideoCodecs) > 0 {
		profile.SupportedVideoCodecs = clientCaps.VideoCodecs
	}
	if len(clientCaps.AudioCodecs) > 0 {
		profile.SupportedAudioCodecs = clientCaps.AudioCodecs
	}
	if clientCaps.MaxHeight > 0 {
		profile.MaxHeight = clientCaps.MaxHeight
	}

	// Get user preferences from DB, then let client request override
	prefs := playback.UserPreferences{
		MaxStreamingQuality: "original",
		PreferDirectPlay:    true,
	}
	var defaultAudioLanguage string
	if dbPrefs, err := h.prefRepo.Get(ctx, userID); err == nil {
		prefs.MaxStreamingQuality = dbPrefs.MaxStreamingQuality
		// Use DB language prefs as defaults (client request overrides below)
		prefs.SelectedSubtitle = dbPrefs.SubtitleLanguage
		defaultAudioLanguage = dbPrefs.AudioLanguage
	}
	// Client-provided values take precedence over DB defaults
	if clientCaps.SelectedSubtitle != "" {
		prefs.SelectedSubtitle = clientCaps.SelectedSubtitle
	}
	if clientCaps.SelectedAudioTrack > 0 {
		prefs.SelectedAudioTrack = clientCaps.SelectedAudioTrack
	}

	// Get user progress
	progress, _ := h.userDataSvc.GetProgress(ctx, userID, mediaID)
	var resumePosition float64
	if progress != nil && !progress.Completed {
		resumePosition = progress.Position
	}

	// Load subtitles before building mediaInfo so we can derive subType for the correct track
	subtitles, subtitleErr := h.subtitleSvc.ListByMediaFile(ctx, primaryFile.ID)
	hasSubtitles := subtitleErr == nil && len(subtitles) > 0

	// subType: use the selected subtitle's codec (not always the first one)
	// Priority: language match for selected subtitle → default subtitle → first subtitle
	var subType string
	if hasSubtitles {
		if prefs.SelectedSubtitle != "" && prefs.SelectedSubtitle != "off" {
			for _, sub := range subtitles {
				if sub.Language == prefs.SelectedSubtitle {
					subType = playback.NormalizeSubtitleCodec(sub.Codec)
					break
				}
			}
		}
		if subType == "" {
			// Fall back to default subtitle
			for _, sub := range subtitles {
				if sub.IsDefault {
					subType = playback.NormalizeSubtitleCodec(sub.Codec)
					break
				}
			}
		}
		if subType == "" {
			subType = playback.NormalizeSubtitleCodec(subtitles[0].Codec)
		}
	}

	// Create media file info for decision engine
	mediaInfo := playback.MediaFileInfo{
		ID:           int(primaryFile.ID),
		Path:         primaryFile.FilePath,
		VideoCodec:   primaryFile.VideoCodec,
		AudioCodec:   primaryFile.AudioCodec,
		Container:    primaryFile.Container,
		Width:        primaryFile.Width,
		Height:       primaryFile.Height,
		Duration:     int(primaryFile.Duration),
		Bitrate:      primaryFile.Bitrate / 1000, // Convert to kbps
		HasSubtitles: hasSubtitles,
		SubType:      subType,
	}

	// Make playback decision
	decision := playback.Decide(mediaInfo, profile, prefs)

	// Find the subtitle stream index for burn-in (needed to build the HLS URL with ?si=N)
	subtitleStreamIndex := -1
	if decision.SubtitleAction == playback.SubtitleBurnIn {
		for _, sub := range subtitles {
			if sub.Language == prefs.SelectedSubtitle {
				subtitleStreamIndex = sub.StreamIndex
				break
			}
		}
	}

	// Build response
	resp := PlaybackInfoResponse{
		MediaID:          int(mediaID),
		PrimaryFileID:    primaryFile.ID,
		Method:           string(decision.Method),
		VideoCodec:       primaryFile.VideoCodec,
		AudioCodec:       primaryFile.AudioCodec,
		Container:        primaryFile.Container,
		FileSize:         primaryFile.FileSize,
		Bitrate:          primaryFile.Bitrate / 1000,
		Duration:         int(primaryFile.Duration),
		Width:            primaryFile.Width,
		Height:           primaryFile.Height,
		DecisionReason:   decision.Reason,
		EstimatedBitrate: decision.EstimatedBitrate,
		Position:         resumePosition,
	}

	// Determine stream URL based on decision.
	// ?fid= is always included so stream handlers serve the exact file used for this decision.
	// Other user selections (audio track, subtitle) are forwarded as query params.
	baseURL := "/api/stream/" + strconv.FormatInt(mediaID, 10)
	fid := strconv.FormatInt(primaryFile.ID, 10)
	switch decision.Method {
	case playback.MethodDirectPlay, playback.MethodDirectStream:
		resp.StreamURL = baseURL + "?fid=" + fid
		if prefs.SelectedAudioTrack > 0 {
			resp.StreamURL += "&at=" + strconv.Itoa(prefs.SelectedAudioTrack)
		}
		if prefs.SelectedSubtitle != "" && prefs.SelectedSubtitle != "off" {
			resp.StreamURL += "&sub=" + prefs.SelectedSubtitle
		}
	case playback.MethodTranscodeAudio, playback.MethodFullTranscode:
		resp.StreamURL = baseURL + "/hls/master.m3u8?fid=" + fid
		if subtitleStreamIndex >= 0 {
			resp.StreamURL += "&si=" + strconv.Itoa(subtitleStreamIndex)
		}
		if prefs.SelectedAudioTrack > 0 {
			resp.StreamURL += "&at=" + strconv.Itoa(prefs.SelectedAudioTrack)
		}
	default:
		resp.StreamURL = baseURL + "/hls/master.m3u8?fid=" + fid
	}

	// Populate audio tracks
	audioTracks, err := h.audioTrackSvc.ListByMediaFile(ctx, primaryFile.ID)
	if err == nil {
		for _, track := range audioTracks {
			selected := track.IsDefault
			if prefs.SelectedAudioTrack > 0 {
				selected = int(track.ID) == prefs.SelectedAudioTrack
			} else if defaultAudioLanguage != "" {
				selected = track.Language == defaultAudioLanguage
			}
			resp.AudioTracks = append(resp.AudioTracks, AudioTrackInfo{
				ID:        int(track.ID),
				Language:  track.Language,
				Label:     track.Title,
				Codec:     track.Codec,
				Channels:  track.Channels,
				IsDefault: track.IsDefault,
				Selected:  selected,
			})
		}
	}

	// Populate subtitle tracks (reuse subtitles already fetched above)
	for _, sub := range subtitles {
		normalized := playback.NormalizeSubtitleCodec(sub.Codec)
		isImage := normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub
		resp.SubtitleTracks = append(resp.SubtitleTracks, SubtitleTrackInfo{
			ID:        int(sub.ID),
			Language:  sub.Language,
			Label:     sub.Title,
			Format:    normalized,
			IsDefault: sub.IsDefault,
			IsImage:   isImage,
		})
	}

	respondJSON(w, http.StatusOK, resp)
}

// GetClientCapabilities returns detected client capabilities
func (h *PlaybackHandler) GetClientCapabilities(w http.ResponseWriter, r *http.Request) {
	info := playback.GetClientInfo(r.UserAgent())

	resp := map[string]interface{}{
		"browser":   info.Browser,
		"platform":  info.Platform,
		"is_mobile": info.IsMobile,
	}

	if info.Profile != nil {
		resp["video_codecs"] = info.Profile.SupportedVideoCodecs
		resp["audio_codecs"] = info.Profile.SupportedAudioCodecs
		resp["containers"] = info.Profile.SupportedContainers
		resp["max_resolution"] = info.Profile.MaxHeight
		resp["max_bitrate"] = info.Profile.MaxBitrate
		resp["supports_hls"] = info.Profile.SupportsHLS
		resp["supports_webm"] = info.Profile.SupportsWebM
	}

	respondJSON(w, http.StatusOK, resp)
}
