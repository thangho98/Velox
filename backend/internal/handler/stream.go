package handler

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/internal/logger"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/service"
)

type StreamHandler struct {
	svc *service.StreamService
	log *slog.Logger
}

func NewStreamHandler(svc *service.StreamService) *StreamHandler {
	return &StreamHandler{svc: svc, log: logger.New("stream")}
}

// DirectPlay routes the request to the appropriate streaming method based on the
// playback decision engine:
//   - DirectPlay  → HTTP range request (full range support)
//   - DirectStream → fragmented MP4 remux pipe (no codec transcode)
//   - TranscodeAudio / FullTranscode → redirect to HLS master playlist
func (h *StreamHandler) DirectPlay(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	streamSessionID := streamSessionIDFromValues(r.URL.Query())

	// Parse optional file ID — GetPlaybackInfo embeds ?fid= so we serve the exact same file.
	var fileID int64
	if fid := r.URL.Query().Get("fid"); fid != "" {
		if n, err := strconv.ParseInt(fid, 10, 64); err == nil {
			fileID = n
		}
	}

	mf, err := h.svc.GetPrimaryFile(r.Context(), id, fileID)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error")
		return
	}

	mediaInfo := playback.MediaFileInfo{
		ID:         int(mf.ID),
		Path:       mf.FilePath,
		VideoCodec: mf.VideoCodec,
		AudioCodec: mf.AudioCodec,
		Container:  mf.Container,
		Width:      mf.Width,
		Height:     mf.Height,
		Duration:   int(mf.Duration),
		Bitrate:    mf.Bitrate / 1000,
	}

	decision := playback.PlaybackDecision{Method: explicitPlaybackMethod(r.URL.Query().Get("pm"))}
	h.log.Debug("DirectPlay request",
		"media_id", id,
		"file_id", fileID,
		"explicit_method", r.URL.Query().Get("pm"),
		"codec", mf.VideoCodec+"/"+mf.AudioCodec,
		"resolution", fmt.Sprintf("%dx%d", mf.Width, mf.Height),
	)
	if decision.Method == "" {
		// Fallback for callers that hit /api/stream directly without going through
		// POST /api/playback/{id}/info first.
		profile := playback.DetectClient(r.UserAgent())
		prefs := playback.UserPreferences{
			MaxStreamingQuality: "original",
			PreferDirectPlay:    true,
		}
		decision = playback.Decide(mediaInfo, profile, prefs)
	}

	// If a non-default audio track is selected (forwarded by GetPlaybackInfo via ?at=N),
	// keep DirectStream lightweight when possible by remuxing the requested audio
	// track; otherwise fall back to HLS.
	if at := r.URL.Query().Get("at"); at != "" {
		trackID, parseErr := strconv.ParseInt(at, 10, 64)
		if parseErr == nil && (decision.Method == playback.MethodDirectPlay || decision.Method == playback.MethodDirectStream) {
			track, trackErr := h.svc.GetAudioTrackForMediaFile(r.Context(), mf.ID, trackID)
			if trackErr == nil {
				w.Header().Set("Content-Type", "video/mp4")
				if err := h.svc.RemuxSelectedAudioToWriter(mf.FilePath, track.StreamIndex, w); err != nil {
					log.Printf("stream: selected-audio remux failed for media %d track %d (%s): %v", id, trackID, mf.FilePath, err)
				}
				return
			}
		}

		http.Redirect(w, r, buildHLSRedirectURL(id, mf.ID, streamSessionID, r.URL.Query()), http.StatusTemporaryRedirect)
		return
	}

	// Check for pre-transcoded file (Plan P: instant playback).
	// Only when no explicit playback method is set — explicit direct play always serves the original.
	if r.URL.Query().Get("pm") == "" {
		if ptFile, err := h.svc.FindPretranscode(r.Context(), mf.ID, 0); err == nil && ptFile != nil && ptFile.Status == "ready" {
			f, err := os.Open(ptFile.FilePath)
			if err == nil {
				defer f.Close()
				stat, err := f.Stat()
				if err == nil {
					w.Header().Set("Cache-Control", "no-store")
					http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
					return
				}
			}
			// Pre-transcode file missing/corrupt — redirect to HLS transcode instead of
			// falling through to Direct Play (which may serve an incompatible original file).
			log.Printf("stream: pretranscode file missing for media %d (%s), falling back to HLS", id, ptFile.FilePath)
			w.Header().Set("Cache-Control", "no-store")
			http.Redirect(w, r, buildHLSRedirectURL(id, mf.ID, streamSessionID, r.URL.Query()), http.StatusTemporaryRedirect)
			return
		}
	}

	switch decision.Method {
	case playback.MethodDirectPlay:
		f, err := os.Open(mf.FilePath)
		if err != nil {
			respondError(w, http.StatusInternalServerError, "cannot open file")
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			respondError(w, http.StatusInternalServerError, "cannot stat file")
			return
		}
		// ServeContent handles Range headers, Content-Type, ETag, etc.
		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)

	case playback.MethodDirectStream:
		// Remux to fragmented MP4 — no codec change, only container repack.
		// Range requests are not supported on a live pipe.
		w.Header().Set("Content-Type", "video/mp4")
		if err := h.svc.RemuxToWriter(mf.FilePath, w); err != nil {
			// Headers already sent; cannot write an error response.
			log.Printf("stream: directstream remux failed for media %d (%s): %v", id, mf.FilePath, err)
			return
		}

	default:
		// TranscodeAudio or FullTranscode: redirect client to HLS endpoint.
		w.Header().Set("Cache-Control", "no-store")
		http.Redirect(w, r, buildHLSRedirectURL(id, mf.ID, streamSessionID, r.URL.Query()), http.StatusTemporaryRedirect)
	}
}

// HLSMaster serves the HLS master playlist. Triggers transcoding if needed.
func (h *StreamHandler) HLSMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// ?fid=N: specific file ID (forwarded from GetPlaybackInfo)
	var hlsFileID int64
	if fid := r.URL.Query().Get("fid"); fid != "" {
		if n, err := strconv.ParseInt(fid, 10, 64); err == nil {
			hlsFileID = n
		}
	}

	// ?si=N: subtitle stream index for burn-in (set by GetPlaybackInfo for PGS/VobSub)
	subtitleStreamIndex := -1
	if si := r.URL.Query().Get("si"); si != "" {
		if n, err := strconv.Atoi(si); err == nil {
			subtitleStreamIndex = n
		}
	}

	videoCopy := r.URL.Query().Get("vcopy") == "1"
	streamSessionID := streamSessionIDFromValues(r.URL.Query())
	if streamSessionID == "" {
		streamSessionID = newStreamSessionID()
	}
	startOffset := parseStartOffset(r.URL.Query().Get("start"))
	maxHeight := 0
	if mh := r.URL.Query().Get("mh"); mh != "" {
		if n, err := strconv.Atoi(mh); err == nil {
			maxHeight = n
		}
	}
	h.log.Debug("HLSMaster request",
		"media_id", id,
		"file_id", hlsFileID,
		"session", streamSessionID,
		"start_offset", startOffset,
		"video_copy", videoCopy,
		"max_height", maxHeight,
		"subtitle_si", subtitleStreamIndex,
	)
	hlsStart := time.Now()
	playlistPath, err := h.svc.PrepareHLS(
		r.Context(),
		id,
		streamSessionID,
		hlsFileID,
		subtitleStreamIndex,
		videoCopy,
		startOffset,
		maxHeight,
	)
	h.log.Debug("PrepareHLS completed",
		"media_id", id,
		"session", streamSessionID,
		"duration", time.Since(hlsStart),
		"playlist", playlistPath,
		"error", err,
	)
	if err != nil {
		// Transcode cancelled (superseded by a seek) — tell hls.js to retry
		// instead of treating it as a fatal error that triggers an error loop.
		if r.Context().Err() != nil || strings.Contains(err.Error(), "cancel") {
			w.Header().Set("Retry-After", "1")
			respondError(w, http.StatusServiceUnavailable, "transcode restarting")
			return
		}
		respondError(w, http.StatusInternalServerError, "transcoding failed: "+err.Error())
		return
	}

	// For video copy: project remaining segments into a VOD playlist so the
	// player can seek anywhere without polling. Duration comes from the HLS
	// cache (populated by PrepareHLS, no extra DB query).
	if videoCopy {
		totalDuration := h.svc.HLSCachedDuration(id, streamSessionID, startOffset)
		if totalDuration > 0 {
			serveProjectedHLSPlaylist(w, r, playlistPath, totalDuration, startOffset)
			return
		}
	}
	serveHLSPlaylist(w, r, playlistPath)
}

// HLSABRMaster serves the adaptive bitrate HLS master playlist.
// Triggers multi-quality transcoding (480p/720p/1080p) if not yet cached.
// The FE quality picker (hls.js levels API) uses the resulting variant streams.
func (h *StreamHandler) HLSABRMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if streamSessionIDFromValues(r.URL.Query()) != "" {
		respondError(w, http.StatusConflict, "adaptive bitrate playback is disabled for isolated stream sessions")
		return
	}

	var fileID int64
	if fid := r.URL.Query().Get("fid"); fid != "" {
		if n, err := strconv.ParseInt(fid, 10, 64); err == nil {
			fileID = n
		}
	}

	playlistPath, err := h.svc.PrepareABRHLS(r.Context(), id, fileID)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "abr transcoding failed: "+err.Error())
		return
	}
	serveHLSPlaylist(w, r, playlistPath)
}

// HLSSegment serves individual HLS .ts segments.
func (h *StreamHandler) HLSSegment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	segment := r.PathValue("segment")
	if !isValidSegmentName(segment) {
		respondError(w, http.StatusBadRequest, "invalid segment name")
		return
	}
	path := h.svc.SegmentPath(id, segment)
	streamSessionID := streamSessionIDFromValues(r.URL.Query())
	if streamSessionID == "" {
		streamSessionID = extractStreamSessionIDFromSegmentName(segment)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		h.log.Debug("segment not ready, waiting",
			"media_id", id,
			"segment", segment,
			"session", streamSessionID,
		)
		waitStart := time.Now()
		if !h.svc.WaitForSegment(streamSessionID, path, 90*time.Second) {
			h.log.Warn("segment wait timeout",
				"media_id", id,
				"segment", segment,
				"waited", time.Since(waitStart),
			)
			respondError(w, http.StatusNotFound, "segment not found")
			return
		}
		h.log.Debug("segment ready",
			"media_id", id,
			"segment", segment,
			"waited", time.Since(waitStart),
		)
	}

	if strings.HasSuffix(segment, ".m3u8") {
		serveHLSPlaylist(w, r, path)
		return
	}

	// Update last activity so abandoned transcode sessions are killed after a timeout.
	h.svc.TouchTranscodeActivity(id, streamSessionID)

	if strings.HasSuffix(segment, ".mp4") || strings.HasSuffix(segment, ".m4s") {
		w.Header().Set("Content-Type", "video/mp4")
	} else {
		w.Header().Set("Content-Type", "video/mp2t")
	}
	// Allow browser to cache segments for 5 minutes to avoid re-downloading
	// the same segment on seek/rebuffer. Segments are immutable once written.
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, path)
}

// StopTranscode kills active FFmpeg transcode processes for a media item.
// Called by the client when leaving the watch page or switching media/quality.
func (h *StreamHandler) StopTranscode(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}
	killed := h.svc.StopTranscode(id, streamSessionIDFromValues(r.URL.Query()))
	respondJSON(w, http.StatusOK, map[string]int{"killed": killed})
}

// StopTranscodeBySession kills active FFmpeg processes for a specific
// viewer-scoped stream session without affecting other viewers of the same media.
func (h *StreamHandler) StopTranscodeBySession(w http.ResponseWriter, r *http.Request) {
	streamSessionID := sanitizeStreamSessionID(r.PathValue("ssid"))
	if streamSessionID == "" {
		respondError(w, http.StatusBadRequest, "invalid stream session id")
		return
	}
	killed := h.svc.StopTranscode(0, streamSessionID)
	respondJSON(w, http.StatusOK, map[string]int{"killed": killed})
}

// isValidSegmentName validates that a segment filename is safe to serve.
// Only alphanumeric, underscore, hyphen, and dot are allowed; no path separators.
// Accepts .ts (MPEG-TS segments), .m3u8 (audio sub-playlists), and .mp4 (fMP4 init segments).
func isValidSegmentName(name string) bool {
	if name == "" {
		return false
	}
	if !strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".m4s") && !strings.HasSuffix(name, ".m3u8") && !strings.HasSuffix(name, ".mp4") {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

func buildHLSRedirectURL(mediaID, fileID int64, streamSessionID string, original url.Values) string {
	query := make(url.Values)
	query.Set("fid", strconv.FormatInt(fileID, 10))

	for _, key := range []string{"api_key", "token", "at", "si", "vcopy", "start", streamSessionQueryKey} {
		if value := original.Get(key); value != "" {
			query.Set(key, value)
		}
	}
	if streamSessionID != "" {
		query.Set(streamSessionQueryKey, streamSessionID)
	}

	return "/api/stream/" + strconv.FormatInt(mediaID, 10) + "/hls/master.m3u8?" + query.Encode()
}

func explicitPlaybackMethod(raw string) playback.PlaybackMethod {
	switch raw {
	case "direct":
		return playback.MethodDirectPlay
	case "directstream":
		return playback.MethodDirectStream
	default:
		return ""
	}
}

// serveProjectedHLSPlaylist reads the FFmpeg-generated EVENT playlist, projects
// remaining segments based on total duration, adds #EXT-X-ENDLIST, and serves
// it as a VOD playlist. This lets hls.js seek to any position without polling
// and without restarting FFmpeg (Jellyfin-style projected playlist).
//
// If the playlist is already complete (has ENDLIST), it's served as-is.
func serveProjectedHLSPlaylist(w http.ResponseWriter, r *http.Request, playlistPath string, totalDuration, startOffset float64) {
	content, err := os.ReadFile(playlistPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot read playlist")
		return
	}

	projected := projectPlaylistToVOD(content, totalDuration-startOffset)

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(rewriteHLSPlaylist(projected, r.URL.Query()))
}

// projectPlaylistToVOD takes an EVENT playlist (possibly incomplete) and appends
// projected segments + #EXT-X-ENDLIST so hls.js treats it as VOD.
// If the playlist already has ENDLIST, returns it unchanged.
func projectPlaylistToVOD(content []byte, adjustedDuration float64) []byte {
	text := string(content)

	// Already complete — nothing to project.
	if strings.Contains(text, "#EXT-X-ENDLIST") {
		return content
	}

	lines := strings.Split(text, "\n")

	// Parse existing segments: count them, sum their durations, extract naming prefix.
	var coveredTime float64
	var lastSegNum int
	var segPrefix string // e.g. "ss12345_vcf163_off5000_"
	var segExt string    // ".ts" or ".m4s" — detected from first segment
	const defaultSegDur = 6.0

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#EXTINF:") {
			durStr := strings.TrimSuffix(strings.TrimPrefix(trimmed, "#EXTINF:"), ",")
			if d, err := strconv.ParseFloat(durStr, 64); err == nil {
				coveredTime += d
			}
		}
		// Accept both .ts (MPEG-TS) and .m4s (fMP4) segment extensions.
		if strings.HasSuffix(trimmed, ".ts") || strings.HasSuffix(trimmed, ".m4s") {
			if segExt == "" {
				if strings.HasSuffix(trimmed, ".m4s") {
					segExt = ".m4s"
				} else {
					segExt = ".ts"
				}
			}
			if idx := strings.Index(trimmed, "seg_"); idx >= 0 {
				segPrefix = trimmed[:idx]
				numStr := trimmed[idx+4 : len(trimmed)-len(segExt)] // between "seg_" and extension
				if n, err := strconv.Atoi(numStr); err == nil {
					lastSegNum = n
				}
			}
			// Remove trailing empty lines after last segment for clean append.
			_ = i
		}
	}

	if segPrefix == "" || adjustedDuration <= 0 {
		return content
	}

	remaining := adjustedDuration - coveredTime
	if remaining < 1.0 {
		// Nearly done — just close the playlist.
		return []byte(text + "#EXT-X-ENDLIST\n")
	}

	// Project remaining segments with estimated 6-second durations.
	var sb strings.Builder
	sb.WriteString(text)
	nextSeg := lastSegNum + 1
	for remaining > 0 {
		dur := defaultSegDur
		if remaining < defaultSegDur {
			dur = remaining
		}
		sb.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", dur))
		sb.WriteString(fmt.Sprintf("%sseg_%04d%s\n", segPrefix, nextSeg, segExt))
		remaining -= dur
		nextSeg++
	}
	sb.WriteString("#EXT-X-ENDLIST\n")

	// Change playlist type from EVENT to VOD so hls.js doesn't poll.
	result := strings.Replace(sb.String(), "#EXT-X-PLAYLIST-TYPE:EVENT", "#EXT-X-PLAYLIST-TYPE:VOD", 1)

	return []byte(result)
}

func serveHLSPlaylist(w http.ResponseWriter, r *http.Request, playlistPath string) {
	content, err := os.ReadFile(playlistPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot read playlist")
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(rewriteHLSPlaylist(content, r.URL.Query()))
}

func rewriteHLSPlaylist(content []byte, original url.Values) []byte {
	apiKey := original.Get("api_key")
	token := original.Get("token")
	at := original.Get("at")
	si := original.Get("si")
	start := original.Get("start")
	streamSessionID := original.Get(streamSessionQueryKey)

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "#EXT-X-MEDIA:") {
			lines[i] = rewriteExtXMediaURI(line, apiKey, token, at, si, start, streamSessionID)
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		lines[i] = appendQueryToPlaylistURI(line, apiKey, token, at, si, start, streamSessionID)
	}

	return []byte(strings.Join(lines, "\n"))
}

func rewriteExtXMediaURI(line, apiKey, token, at, si, startOffset, streamSessionID string) string {
	const marker = `URI="`
	start := strings.Index(line, marker)
	if start < 0 {
		return line
	}
	start += len(marker)
	end := strings.Index(line[start:], `"`)
	if end < 0 {
		return line
	}
	end += start

	return line[:start] + appendQueryToPlaylistURI(line[start:end], apiKey, token, at, si, startOffset, streamSessionID) + line[end:]
}

func appendQueryToPlaylistURI(rawURI, apiKey, token, at, si, start, streamSessionID string) string {
	uri, err := url.Parse(rawURI)
	if err != nil || uri == nil {
		return rawURI
	}

	query := uri.Query()
	if apiKey != "" {
		query.Set("api_key", apiKey)
	}
	if token != "" {
		query.Set("token", token)
	}
	if streamSessionID != "" {
		query.Set(streamSessionQueryKey, streamSessionID)
	}
	if at != "" && strings.HasSuffix(uri.Path, ".m3u8") {
		query.Set("at", at)
	}
	if si != "" && strings.HasSuffix(uri.Path, ".m3u8") {
		query.Set("si", si)
	}
	if start != "" && strings.HasSuffix(uri.Path, ".m3u8") {
		query.Set("start", start)
	}
	uri.RawQuery = query.Encode()
	return uri.String()
}

func parseStartOffset(raw string) float64 {
	if raw == "" {
		return 0
	}
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil || n < 0 {
		return 0
	}
	return n
}
