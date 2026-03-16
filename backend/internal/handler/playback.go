package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/transcoder"
)

var supportsSubtitleBurnIn = transcoder.SupportsSubtitleBurnIn

// PlaybackHandler provides playback decision and info endpoints
type PlaybackHandler struct {
	mediaSvc        *service.MediaService
	streamSvc       *service.StreamService
	userDataSvc     *service.UserDataService
	subtitleSvc     *service.SubtitleService
	audioTrackSvc   *service.AudioTrackService
	markerSvc       *service.MarkerService // NEW: marker service for skip segments
	prefRepo        *repository.UserPreferencesRepo
	appSettingsRepo *repository.AppSettingsRepo
}

// NewPlaybackHandler creates a new playback handler
func NewPlaybackHandler(
	mediaSvc *service.MediaService,
	streamSvc *service.StreamService,
	userDataSvc *service.UserDataService,
	subtitleSvc *service.SubtitleService,
	audioTrackSvc *service.AudioTrackService,
	markerSvc *service.MarkerService,
	prefRepo *repository.UserPreferencesRepo,
	appSettingsRepo *repository.AppSettingsRepo,
) *PlaybackHandler {
	return &PlaybackHandler{
		mediaSvc:        mediaSvc,
		streamSvc:       streamSvc,
		userDataSvc:     userDataSvc,
		subtitleSvc:     subtitleSvc,
		audioTrackSvc:   audioTrackSvc,
		markerSvc:       markerSvc,
		prefRepo:        prefRepo,
		appSettingsRepo: appSettingsRepo,
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
	SelectedSubtitleID int      `json:"selected_subtitle_id,omitempty"` // exact subtitle track ID
}

// PlaybackInfoResponse represents playback decision response
type PlaybackInfoResponse struct {
	MediaID          int                 `json:"media_id"`
	PrimaryFileID    int64               `json:"primary_file_id,omitempty"` // file ID used for this decision
	Method           string              `json:"method"`                    // DirectPlay, DirectStream, TranscodeAudio, FullTranscode
	StreamURL        string              `json:"stream_url"`
	AbrURL           string              `json:"abr_url,omitempty"` // adaptive bitrate HLS (multi-quality); only set when transcoding
	VideoCodec       string              `json:"video_codec"`
	VideoProfile     string              `json:"video_profile,omitempty"`
	VideoLevel       int                 `json:"video_level,omitempty"`
	VideoFPS         float64             `json:"video_fps,omitempty"`
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
	Position         float64             `json:"position,omitempty"`      // Resume position
	SkipSegments     []model.SkipSegment `json:"skip_segments,omitempty"` // NEW: intro/credits skip markers
}

// AudioTrackInfo represents an audio track
type AudioTrackInfo struct {
	ID         int    `json:"id"`
	Language   string `json:"language"`
	Label      string `json:"label"`
	Codec      string `json:"codec,omitempty"`
	Channels   int    `json:"channels,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
	IsDefault  bool   `json:"is_default"`
	Selected   bool   `json:"selected"`
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

	// Detect client profile from User-Agent, then tighten it using actual browser
	// capability probes sent by the frontend.
	profile := applyClientCapabilityOverrides(playback.DetectClient(r.UserAgent()), clientCaps)

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
	subtitles = filterPlayableSubtitles(subtitles, supportsSubtitleBurnIn())
	hasSubtitles := subtitleErr == nil && len(subtitles) > 0
	audioTracks, audioTrackErr := h.audioTrackSvc.ListByMediaFile(ctx, primaryFile.ID)
	effectiveAudioTrackID := resolveSelectedAudioTrackID(prefs.SelectedAudioTrack, audioTracks)
	selectedSubtitle := findSubtitleByID(subtitles, clientCaps.SelectedSubtitleID)
	if selectedSubtitle == nil {
		selectedSubtitle = findSubtitleByLanguage(subtitles, prefs.SelectedSubtitle)
	}

	// subType: use the selected subtitle's codec (not always the first one)
	// Priority: language match for selected subtitle → default subtitle → first subtitle
	var subType string
	if hasSubtitles {
		if selectedSubtitle != nil {
			subType = playback.NormalizeSubtitleCodec(selectedSubtitle.Codec)
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

	decision = applyAdminPlaybackPolicy(ctx, h.appSettingsRepo, decision, profile, mediaInfo)

	// Find the subtitle stream index for burn-in (needed to build the HLS URL with ?si=N)
	subtitleStreamIndex := -1
	if decision.SubtitleAction == playback.SubtitleBurnIn && selectedSubtitle != nil {
		subtitleStreamIndex = selectedSubtitle.StreamIndex
	}

	// Build response
	resp := PlaybackInfoResponse{
		MediaID:          int(mediaID),
		PrimaryFileID:    primaryFile.ID,
		Method:           string(decision.Method),
		VideoCodec:       primaryFile.VideoCodec,
		VideoProfile:     primaryFile.VideoProfile,
		VideoLevel:       primaryFile.VideoLevel,
		VideoFPS:         primaryFile.VideoFPS,
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
		resp.StreamURL = baseURL + "?fid=" + fid + "&pm=" + playbackModeQuery(decision.Method)
		if effectiveAudioTrackID > 0 {
			resp.StreamURL += "&at=" + strconv.Itoa(effectiveAudioTrackID)
		}
		if prefs.SelectedSubtitle != "" && prefs.SelectedSubtitle != "off" {
			resp.StreamURL += "&sub=" + prefs.SelectedSubtitle
		}
	case playback.MethodTranscodeAudio, playback.MethodFullTranscode:
		resp.StreamURL = baseURL + "/hls/master.m3u8?fid=" + fid
		// ABR pipeline encodes audio once (default track only) and has no subtitle filter.
		// Only offer ABR when neither subtitle burn-in nor a non-default audio track is needed.
		if subtitleStreamIndex < 0 && effectiveAudioTrackID == 0 {
			if h.streamSvc.ABRCached(mediaID, primaryFile.ID) {
				// Already transcoded — serve it immediately.
				resp.AbrURL = baseURL + "/hls/abr.m3u8?fid=" + fid
			} else {
				// Not cached yet: kick off background generation; client will use
				// regular HLS this time and get ABR on the next visit.
				h.streamSvc.StartABRBackground(mediaID, primaryFile.ID, primaryFile.FilePath, primaryFile.Height)
			}
		}
		if subtitleStreamIndex >= 0 {
			resp.StreamURL += "&si=" + strconv.Itoa(subtitleStreamIndex)
		}
		if effectiveAudioTrackID > 0 {
			resp.StreamURL += "&at=" + strconv.Itoa(effectiveAudioTrackID)
		}
	default:
		resp.StreamURL = baseURL + "/hls/master.m3u8?fid=" + fid
	}

	// Populate audio tracks
	if audioTrackErr == nil {
		for _, track := range audioTracks {
			selected := track.IsDefault
			if effectiveAudioTrackID > 0 {
				selected = int(track.ID) == effectiveAudioTrackID
			} else if defaultAudioLanguage != "" {
				selected = track.Language == defaultAudioLanguage
			}
			resp.AudioTracks = append(resp.AudioTracks, AudioTrackInfo{
				ID:         int(track.ID),
				Language:   track.Language,
				Label:      track.Title,
				Codec:      track.Codec,
				Channels:   track.Channels,
				Bitrate:    track.Bitrate,
				SampleRate: track.SampleRate,
				IsDefault:  track.IsDefault,
				Selected:   selected,
			})
		}
	}

	// Populate subtitle tracks (reuse subtitles already fetched above)
	// When forced direct play, skip image-based subtitles (PGS/VobSub) since
	// they require server-side burn-in and the client cannot render them.
	forcedDirectPlay := strings.Contains(decision.Reason, "admin policy")
	for _, sub := range subtitles {
		normalized := playback.NormalizeSubtitleCodec(sub.Codec)
		isImage := normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub
		if forcedDirectPlay && isImage {
			continue
		}
		resp.SubtitleTracks = append(resp.SubtitleTracks, SubtitleTrackInfo{
			ID:        int(sub.ID),
			Language:  sub.Language,
			Label:     sub.Title,
			Format:    normalized,
			IsDefault: sub.IsDefault,
			IsImage:   isImage,
		})
	}

	// NEW: Load skip segments (intro/credits markers) for the primary file
	if h.markerSvc != nil {
		skipSegments, _ := h.markerSvc.GetSkipSegments(ctx, primaryFile.ID)
		resp.SkipSegments = skipSegments
	}

	respondJSON(w, http.StatusOK, resp)
}

func resolveSelectedAudioTrackID(requestedID int, audioTracks []model.AudioTrack) int {
	if requestedID <= 0 {
		return 0
	}

	for _, track := range audioTracks {
		if int(track.ID) != requestedID {
			continue
		}
		if track.IsDefault {
			return 0
		}
		return requestedID
	}

	return 0
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

func applyClientCapabilityOverrides(profile *playback.DeviceProfile, clientCaps PlaybackInfoRequest) *playback.DeviceProfile {
	if profile == nil {
		return nil
	}

	clone := *profile
	if len(clientCaps.VideoCodecs) > 0 {
		clone.SupportedVideoCodecs = normalizeCapabilityValues(clientCaps.VideoCodecs)
	}
	if len(clientCaps.AudioCodecs) > 0 {
		clone.SupportedAudioCodecs = normalizeCapabilityValues(clientCaps.AudioCodecs)
	}
	if len(clientCaps.Containers) > 0 {
		clone.SupportedContainers = normalizeCapabilityValues(clientCaps.Containers)
	}
	if clientCaps.MaxHeight > 0 {
		clone.MaxHeight = clientCaps.MaxHeight
	}

	return &clone
}

func normalizeCapabilityValues(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		normalized = append(normalized, value)
	}
	return normalized
}

func normalizeLanguageCode(language string) string {
	language = strings.TrimSpace(strings.ToLower(language))
	switch language {
	case "en", "eng":
		return "eng"
	case "vi", "vie":
		return "vie"
	case "zh", "zho", "chi":
		return "zho"
	case "ja", "jpn":
		return "jpn"
	case "ko", "kor":
		return "kor"
	case "fr", "fra", "fre":
		return "fra"
	case "de", "deu", "ger", "dut":
		return "deu"
	case "es", "spa":
		return "spa"
	case "pt", "por":
		return "por"
	default:
		return language
	}
}

func languageMatches(lhs, rhs string) bool {
	if lhs == "" || rhs == "" {
		return false
	}
	return normalizeLanguageCode(lhs) == normalizeLanguageCode(rhs)
}

func findSubtitleByLanguage(subtitles []model.Subtitle, language string) *model.Subtitle {
	if language == "" || language == "off" {
		return nil
	}

	var imageMatch *model.Subtitle
	for i := range subtitles {
		if !languageMatches(subtitles[i].Language, language) {
			continue
		}

		normalized := playback.NormalizeSubtitleCodec(subtitles[i].Codec)
		if normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub {
			if imageMatch == nil {
				imageMatch = &subtitles[i]
			}
			continue
		}

		return &subtitles[i]
	}

	if imageMatch != nil {
		return imageMatch
	}

	for i := range subtitles {
		if languageMatches(subtitles[i].Language, language) {
			return &subtitles[i]
		}
	}

	return nil
}

func findSubtitleByID(subtitles []model.Subtitle, subtitleID int) *model.Subtitle {
	if subtitleID <= 0 {
		return nil
	}
	for i := range subtitles {
		if int(subtitles[i].ID) == subtitleID {
			return &subtitles[i]
		}
	}
	return nil
}

func filterPlayableSubtitles(subtitles []model.Subtitle, burnInSupported bool) []model.Subtitle {
	if burnInSupported {
		return subtitles
	}

	filtered := make([]model.Subtitle, 0, len(subtitles))
	for _, sub := range subtitles {
		normalized := playback.NormalizeSubtitleCodec(sub.Codec)
		if normalized == playback.SubtitlePGS || normalized == playback.SubtitleVobSub {
			continue
		}
		filtered = append(filtered, sub)
	}
	return filtered
}

func playbackModeQuery(method playback.PlaybackMethod) string {
	switch method {
	case playback.MethodDirectStream:
		return "directstream"
	default:
		return "direct"
	}
}

func normalizeContainerValue(container string) string {
	container = strings.TrimSpace(strings.ToLower(container))
	switch container {
	case "mp4", "mpeg4", "m4v":
		return playback.ContainerMP4
	case "webm":
		return playback.ContainerWebM
	case "mkv", "matroska", "matroska,webm":
		return playback.ContainerMKV
	case "mov", "qt":
		return playback.ContainerMOV
	default:
		return container
	}
}

func applyAdminPlaybackPolicy(
	ctx context.Context,
	appSettingsRepo *repository.AppSettingsRepo,
	decision playback.PlaybackDecision,
	profile *playback.DeviceProfile,
	mediaInfo playback.MediaFileInfo,
) playback.PlaybackDecision {
	if appSettingsRepo == nil {
		return decision
	}

	playbackMode, _ := appSettingsRepo.Get(ctx, model.SettingPlaybackMode)
	if playbackMode != "direct_play" {
		return decision
	}

	// Force direct play: never transcode video, but allow audio transcode
	// when the browser can't decode the audio codec (DTS, TrueHD, etc.).
	// This mirrors Emby/Jellyfin behavior: video direct play + audio fallback.
	decision.VideoAction = playback.VideoCopy
	decision.SubtitleAction = playback.SubtitleCopy

	if profile != nil && !profile.SupportsAudioCodec(mediaInfo.AudioCodec) {
		// Audio incompatible — transcode audio only (lightweight)
		decision.Method = playback.MethodTranscodeAudio
		decision.AudioAction = playback.AudioTranscode
		decision.Reason = "forced direct play + audio transcode (admin policy)"
	} else {
		decision.Method = playback.MethodDirectPlay
		decision.AudioAction = playback.AudioCopy
		decision.Reason = "forced direct play (admin policy)"
	}
	return decision
}
