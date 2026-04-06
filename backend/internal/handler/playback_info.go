package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slices"
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
	selectedAudioTrack, effectiveAudioTrackID, autoSelectedCompatibleAudio := resolvePlaybackAudioTrack(
		prefs.SelectedAudioTrack,
		defaultAudioLanguage,
		audioTracks,
		profile,
	)
	selectedAudioCodec := primaryFile.AudioCodec
	if selectedAudioTrack != nil && selectedAudioTrack.Codec != "" {
		selectedAudioCodec = selectedAudioTrack.Codec
	}
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
		AudioCodec:   selectedAudioCodec,
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
	decision = adjustPlaybackDecisionForSelectedAudioTrack(
		decision,
		selectedAudioTrack,
		effectiveAudioTrackID,
		autoSelectedCompatibleAudio,
	)

	playbackMode, _ := h.configSvc.GetPlaybackMode(ctx)
	decision = applyAdminPlaybackPolicy(playbackMode, decision, profile, mediaInfo)

	// Check for pre-transcoded file (Plan P: instant playback)
	// Only use pretranscode when height EXACTLY matches what user requested,
	// or when user didn't request a specific height (auto mode).
	// This ensures selecting "1440p" gives real 1440p transcode, not 1080p pretranscode.
	var selectedPtProfile *model.PretranscodeProfile
	userRequestedLower := clientCaps.MaxHeight > 0 && clientCaps.MaxHeight < primaryFile.Height
	if decision.Method != playback.MethodDirectPlay || userRequestedLower {
		maxHeight := clientCaps.MaxHeight
		if maxHeight <= 0 {
			maxHeight = playback.ParseQuality(prefs.MaxStreamingQuality)
		}
		if ptFile, err := h.streamSvc.FindPretranscode(ctx, primaryFile.ID, maxHeight); err == nil && ptFile != nil {
			ptProfile, _ := h.streamSvc.FindPretranscodeProfile(ctx, ptFile.ProfileID)

			// Pretranscode is usable when:
			// - Auto mode: user didn't request a specific height, so any quality is acceptable
			// - Explicit height: pretranscode matches the requested height exactly
			autoMode := clientCaps.MaxHeight <= 0
			heightMatches := autoMode || (ptProfile != nil && ptProfile.Height == clientCaps.MaxHeight)

			ptVideoCodec := primaryFile.VideoCodec
			if ptProfile != nil && ptProfile.VideoCodec != "copy" {
				ptVideoCodec = ptProfile.VideoCodec
			}
			ptUsable := heightMatches && fileExists(ptFile.FilePath) && profile.SupportsVideoCodec(playback.NormalizeCodec(ptVideoCodec))

			// Pre-transcodes always carry the file's default audio track (0:a:0).
			// Skip them when a compatible non-default track was selected for playback.
			if effectiveAudioTrackID > 0 {
				ptUsable = false
				decision.Reason += " (pretranscode skipped: selected audio track is non-default)"
			}

			// Don't downgrade video quality for audio-only issues.
			// If the original decision only needs audio transcode (video is compatible),
			// only use pretranscode when it doesn't reduce resolution (e.g. 4K → 1080p).
			audioOnlyIssue := decision.Method == playback.MethodTranscodeAudio
			if audioOnlyIssue && ptProfile != nil && ptProfile.Height < primaryFile.Height {
				ptUsable = false
				decision.Reason += fmt.Sprintf(" (pretranscode %dp skipped: video compatible at %dp, only audio needs transcode)",
					ptProfile.Height, primaryFile.Height)
			}

			if !ptUsable && !heightMatches && ptProfile != nil {
				decision.Reason += fmt.Sprintf(" (pretranscode %dp skipped: requested %dp)",
					ptProfile.Height, clientCaps.MaxHeight)
			}

			if ptUsable {
				selectedPtProfile = ptProfile
				decision = playback.PlaybackDecision{
					Method:           playback.MethodPreTranscode,
					VideoAction:      playback.VideoCopy,
					AudioAction:      playback.AudioCopy,
					SubtitleAction:   playback.SubtitleNone,
					Container:        "mp4",
					Reason:           fmt.Sprintf("Pre-transcoded file available (%dp)", ptProfile.Height),
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
		AudioCodec:         selectedAudioCodec,
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

	// Populate pretranscode details so the stats overlay shows actual playback info
	if selectedPtProfile != nil {
		resp.PtVideoCodec = selectedPtProfile.VideoCodec
		resp.PtAudioCodec = selectedPtProfile.AudioCodec
		resp.PtHeight = selectedPtProfile.Height
		resp.PtVideoBitrate = selectedPtProfile.VideoBitrate
		resp.PtAudioBitrate = selectedPtProfile.AudioBitrate
	}

	// Determine stream URLs based on decision.
	// ?fid= is always included so stream handlers serve the exact file used for this decision.
	// Other user selections (audio track, subtitle) are forwarded as query params.
	baseURL := "/api/stream/" + strconv.FormatInt(mediaID, 10)
	apiKey := ""
	if h.apiKeyStore != nil {
		apiKey = h.apiKeyStore.Generate(userID, isAdmin)
	}
	streamSessionID := newStreamSessionID()
	baseQuery := url.Values{}
	baseQuery.Set("fid", strconv.FormatInt(primaryFile.ID, 10))
	baseQuery.Set(streamSessionQueryKey, streamSessionID)
	if apiKey != "" {
		baseQuery.Set("api_key", apiKey)
	}

	// direct_url: always points to original file (pm=direct) — client tries this first
	directQuery := cloneValues(baseQuery)
	directQuery.Set("pm", "direct")
	resp.DirectURL = buildURLWithQuery(baseURL, directQuery)
	resp.StreamSessionID = streamSessionID

	// Build HLS query with all necessary params
	hlsQuery := cloneValues(baseQuery)
	if decision.VideoAction == playback.VideoCopy {
		hlsQuery.Set("vcopy", "1")
	}
	if clientCaps.MaxHeight > 0 {
		hlsQuery.Set("mh", strconv.Itoa(clientCaps.MaxHeight))
	}
	if subtitleStreamIndex >= 0 {
		hlsQuery.Set("si", strconv.Itoa(subtitleStreamIndex))
	}
	if effectiveAudioTrackID > 0 {
		hlsQuery.Set("at", strconv.Itoa(effectiveAudioTrackID))
	}

	switch decision.Method {
	case playback.MethodPreTranscode:
		// Pre-transcoded file — serve via direct endpoint for browsers,
		// or HLS for clients that don't support direct MP4 (e.g. mobile)
		if !slices.Contains(clientCaps.Containers, "mp4") && slices.Contains(clientCaps.Containers, "hls") {
			query := cloneValues(baseQuery)
			query.Set("vcopy", "1")
			resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", query)
		} else {
			resp.StreamURL = buildURLWithQuery(baseURL, cloneValues(baseQuery))
		}
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
		if effectiveAudioTrackID > 0 {
			resp.DirectURL = resp.StreamURL
		}
	case playback.MethodTranscodeAudio, playback.MethodFullTranscode:
		resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", hlsQuery)
		// Session-scoped playback must remain isolated per viewer. The current
		// ABR cache is shared by mediaID+fileID, so advertising abr_url here
		// would let clients bypass the per-session HLS transcode path.
		// On-demand pretranscode: remux HLS → MP4 for next time ⚡
		if clientCaps.MaxHeight > 0 {
			go h.streamSvc.RemuxToPretranscode(context.Background(), primaryFile.ID, mediaID, clientCaps.MaxHeight)
		}
	default:
		resp.StreamURL = buildURLWithQuery(baseURL+"/hls/master.m3u8", hlsQuery)
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
		h.activitySvc.Log(
			&userID,
			"play_start",
			r.RemoteAddr,
			&mediaID,
			fmt.Sprintf(`{"stream_session_id":"%s","method":"%s"}`, streamSessionID, resp.Method),
		)
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
	// Video copy + audio transcode: video stream passed through unchanged,
	// only audio re-encoded. Fast and avoids full transcode CPU overhead.
	// Matches Jellyfin/Emby behavior: audio mismatch never triggers video transcode.
	decision.VideoAction = playback.VideoCopy
	decision.SubtitleAction = playback.SubtitleCopy

	if profile != nil && !profile.SupportsAudioCodec(mediaInfo.AudioCodec) {
		decision.Method = playback.MethodTranscodeAudio
		decision.AudioAction = playback.AudioTranscode
		decision.Reason = "forced direct play + audio transcode with video copy (admin policy)"
	} else {
		decision.Method = playback.MethodDirectPlay
		decision.AudioAction = playback.AudioCopy
		decision.Reason = "forced direct play (admin policy)"
	}
	return decision
}
