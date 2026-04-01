package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
)

// GetPlaybackInfo returns playback decision for a media item
func (h *PlaybackHandler) GetPlaybackInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID, isAdmin, ok := auth.UserFromContext(ctx)
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

	// Fetch 4 independent data sources in parallel (prefs, progress, subtitles, audio).
	// SQLite with WAL mode supports concurrent reads — this cuts latency by ~3-4x.
	prefs := playback.UserPreferences{
		MaxStreamingQuality: "original",
		PreferDirectPlay:    true,
	}
	var defaultAudioLanguage string
	var resumePosition float64
	var subtitles []model.Subtitle
	var subtitleErr error
	var audioTracks []model.AudioTrack
	var audioTrackErr error

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		if dbPrefs, err := h.configSvc.GetUserPreferences(ctx, userID); err == nil {
			prefs.MaxStreamingQuality = dbPrefs.MaxStreamingQuality
			prefs.SelectedSubtitle = dbPrefs.SubtitleLanguage
			defaultAudioLanguage = dbPrefs.AudioLanguage
		}
	}()

	go func() {
		defer wg.Done()
		if progress, err := h.userDataSvc.GetProgress(ctx, userID, mediaID); err == nil && progress != nil && !progress.Completed {
			resumePosition = progress.Position
		}
	}()

	go func() {
		defer wg.Done()
		subtitles, subtitleErr = h.subtitleSvc.ListByMediaFile(ctx, primaryFile.ID)
		subtitles = filterPlayableSubtitles(subtitles, supportsSubtitleBurnIn())
	}()

	go func() {
		defer wg.Done()
		audioTracks, audioTrackErr = h.audioTrackSvc.ListByMediaFile(ctx, primaryFile.ID)
	}()

	wg.Wait()

	// Client-provided values take precedence over DB defaults
	if clientCaps.MaxHeight > 0 {
		prefs.MaxStreamingQuality = fmt.Sprintf("%dp", clientCaps.MaxHeight)
	}
	if clientCaps.SelectedSubtitle != "" {
		prefs.SelectedSubtitle = clientCaps.SelectedSubtitle
	}
	if clientCaps.SelectedAudioTrack > 0 {
		prefs.SelectedAudioTrack = clientCaps.SelectedAudioTrack
	}

	hasSubtitles := subtitleErr == nil && len(subtitles) > 0
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

	playbackMode, _ := h.configSvc.GetPlaybackMode(ctx)
	decision = applyAdminPlaybackPolicy(playbackMode, decision, profile, mediaInfo)

	// Check for pre-transcoded file (Plan P: instant playback)
	// Use pretranscode when: (a) normal decision requires transcoding, OR
	// (b) user explicitly selected a lower resolution via max_height
	userRequestedLower := clientCaps.MaxHeight > 0 && clientCaps.MaxHeight < primaryFile.Height
	if decision.Method != playback.MethodDirectPlay || userRequestedLower {
		maxHeight := clientCaps.MaxHeight
		if maxHeight <= 0 {
			maxHeight = playback.ParseQuality(prefs.MaxStreamingQuality)
		}
		if ptFile, err := h.streamSvc.FindPretranscode(ctx, primaryFile.ID, maxHeight); err == nil && ptFile != nil {
			ptProfile, _ := h.streamSvc.FindPretranscodeProfile(ctx, ptFile.ProfileID)
			// Validate: file must exist on disk and its video codec must be browser-compatible.
			// Audio-remux pretranscodes (video_codec='copy') preserve the original video codec,
			// so HEVC sources stay HEVC — unusable by browsers that don't support it.
			ptVideoCodec := primaryFile.VideoCodec
			if ptProfile != nil && ptProfile.VideoCodec != "copy" {
				ptVideoCodec = ptProfile.VideoCodec
			}
			ptUsable := fileExists(ptFile.FilePath) && profile.SupportsVideoCodec(playback.NormalizeCodec(ptVideoCodec))
			if ptUsable {
				decision = playback.PlaybackDecision{
					Method:           playback.MethodPreTranscode,
					VideoAction:      playback.VideoCopy,
					AudioAction:      playback.AudioCopy,
					SubtitleAction:   playback.SubtitleNone,
					Container:        "mp4",
					Reason:           "Pre-transcoded file available",
					PreTranscodePath: ptFile.FilePath,
				}
				if ptProfile != nil {
					decision.PreTranscodeQuality = ptProfile.Name
					decision.EstimatedBitrate = ptProfile.VideoBitrate + ptProfile.AudioBitrate
				}
			}
		}
	}

	// Fetch quality options + skip segments in parallel (both independent, happen after decision).
	var availableQualities []QualityOption
	var skipSegments []model.SkipSegment

	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		availableQualities = h.buildQualityOptions(ctx, primaryFile)
	}()
	go func() {
		defer wg2.Done()
		if h.markerSvc != nil {
			skipSegments, _ = h.markerSvc.GetSkipSegments(ctx, primaryFile.ID)
		}
	}()
	wg2.Wait()

	// Find the subtitle stream index for burn-in (needed to build the HLS URL with ?si=N)
	subtitleStreamIndex := -1
	if decision.SubtitleAction == playback.SubtitleBurnIn && selectedSubtitle != nil {
		subtitleStreamIndex = selectedSubtitle.StreamIndex
	}

	// Build response
	resp := PlaybackInfoResponse{
		MediaID:            int(mediaID),
		PrimaryFileID:      primaryFile.ID,
		Method:             string(decision.Method),
		VideoCodec:         primaryFile.VideoCodec,
		VideoProfile:       primaryFile.VideoProfile,
		VideoLevel:         primaryFile.VideoLevel,
		VideoFPS:           primaryFile.VideoFPS,
		AudioCodec:         primaryFile.AudioCodec,
		Container:          primaryFile.Container,
		FileSize:           primaryFile.FileSize,
		Bitrate:            primaryFile.Bitrate / 1000,
		Duration:           int(primaryFile.Duration),
		Width:              primaryFile.Width,
		Height:             primaryFile.Height,
		DecisionReason:     decision.Reason,
		EstimatedBitrate:   decision.EstimatedBitrate,
		Position:           resumePosition,
		AvailableQualities: availableQualities,
	}

	// Determine stream URL based on decision.
	// ?fid= is always included so stream handlers serve the exact file used for this decision.
	// Other user selections (audio track, subtitle) are forwarded as query params.
	baseURL := "/api/stream/" + strconv.FormatInt(mediaID, 10)
	apiKey := ""
	if h.apiKeyStore != nil {
		apiKey = h.apiKeyStore.Generate(userID, isAdmin)
	}
	baseQuery := url.Values{}
	baseQuery.Set("fid", strconv.FormatInt(primaryFile.ID, 10))
	if apiKey != "" {
		baseQuery.Set("api_key", apiKey)
	}
	switch decision.Method {
	case playback.MethodPreTranscode:
		// Pre-transcoded MP4 — serve via same direct play endpoint (pre-transcode check runs first)
		resp.StreamURL = buildURLWithQuery(baseURL, cloneValues(baseQuery))
	case playback.MethodDirectPlay, playback.MethodDirectStream:
		query := cloneValues(baseQuery)
		query.Set("pm", playbackModeQuery(decision.Method))
		if effectiveAudioTrackID > 0 {
			query.Set("at", strconv.Itoa(effectiveAudioTrackID))
		}
		if prefs.SelectedSubtitle != "" && prefs.SelectedSubtitle != "off" {
			query.Set("sub", prefs.SelectedSubtitle)
		}
		resp.StreamURL = buildURLWithQuery(baseURL, query)
	case playback.MethodTranscodeAudio, playback.MethodFullTranscode:
		query := cloneValues(baseQuery)
		if decision.VideoAction == playback.VideoCopy {
			query.Set("vcopy", "1")
		}
		resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", query)
		// ABR pipeline encodes audio once (default track only) and has no subtitle filter.
		// Only offer ABR when neither subtitle burn-in nor a non-default audio track is needed.
		// Only serve ABR if already cached — don't start background generation here
		// because it consumes a transcode slot and can block realtime playback for
		// other media. ABR will be generated opportunistically when slots are idle.
		if subtitleStreamIndex < 0 && effectiveAudioTrackID == 0 {
			if h.streamSvc.ABRCached(mediaID, primaryFile.ID) {
				resp.AbrURL = buildURLWithQuery(baseURL+"/hls/abr.m3u8", cloneValues(baseQuery))
			}
		}
		if subtitleStreamIndex >= 0 {
			query.Set("si", strconv.Itoa(subtitleStreamIndex))
		}
		if effectiveAudioTrackID > 0 {
			query.Set("at", strconv.Itoa(effectiveAudioTrackID))
		}
		resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", query)
		// On-demand pretranscode: remux HLS → MP4 for next time ⚡
		if clientCaps.MaxHeight > 0 {
			go h.streamSvc.RemuxToPretranscode(context.Background(), primaryFile.ID, mediaID, clientCaps.MaxHeight)
		}
	default:
		resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", cloneValues(baseQuery))
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

	resp.SkipSegments = skipSegments

	// Log play start activity
	if h.activitySvc != nil {
		h.activitySvc.Log(&userID, "play_start", r.RemoteAddr, &mediaID, "")
	}

	w.Header().Set("Cache-Control", "no-store")
	respondJSON(w, http.StatusOK, resp)
}

func applyAdminPlaybackPolicy(
	playbackMode string,
	decision playback.PlaybackDecision,
	profile *playback.DeviceProfile,
	mediaInfo playback.MediaFileInfo,
) playback.PlaybackDecision {
	if playbackMode != "direct_play" {
		return decision
	}

	// Force direct play: never transcode video, but allow audio transcode
	// when the browser can't decode the audio codec (DTS, TrueHD, etc.).
	// Browser HLS audio-only transcode uses video copy segments that seek
	// unreliably on some sources, so HLS clients must fully transcode instead.
	decision.VideoAction = playback.VideoCopy
	decision.SubtitleAction = playback.SubtitleCopy

	if profile != nil && !profile.SupportsAudioCodec(mediaInfo.AudioCodec) {
		if playback.RequiresFullTranscodeForAudioMismatch(profile) {
			decision.Method = playback.MethodFullTranscode
			decision.VideoAction = playback.VideoTranscode
			decision.VideoCodec = playback.CodecH264
			decision.Reason = "forced direct play + audio mismatch uses full transcode for reliable HLS seeking (admin policy)"
		} else {
			// Video copy + audio transcode: video stream passed through unchanged,
			// only audio re-encoded. Fast and avoids full transcode CPU overhead.
			decision.Method = playback.MethodTranscodeAudio
			decision.VideoAction = playback.VideoCopy
			decision.Reason = "forced direct play + audio transcode with video copy (admin policy)"
		}
		decision.AudioAction = playback.AudioTranscode
	} else {
		decision.Method = playback.MethodDirectPlay
		decision.AudioAction = playback.AudioCopy
		decision.Reason = "forced direct play (admin policy)"
	}
	return decision
}
