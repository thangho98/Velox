package handler

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/service"
)

type StreamHandler struct {
	svc *service.StreamService
}

func NewStreamHandler(svc *service.StreamService) *StreamHandler {
	return &StreamHandler{svc: svc}
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

	// If a non-default audio track is selected (forwarded by GetPlaybackInfo via ?at=N),
	// HLS is required. Preserve both ?fid and ?at in the redirect so HLSMaster uses
	// the same file and the frontend knows which audio track was selected.
	if at := r.URL.Query().Get("at"); at != "" {
		http.Redirect(w, r, buildHLSRedirectURL(id, mf.ID, r.URL.Query()), http.StatusTemporaryRedirect)
		return
	}

	// Check for pre-transcoded file first (Plan P: instant playback)
	if ptFile, err := h.svc.FindPretranscode(r.Context(), mf.ID, 0); err == nil && ptFile != nil {
		f, err := os.Open(ptFile.FilePath)
		if err == nil {
			defer f.Close()
			stat, err := f.Stat()
			if err == nil {
				http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
				return
			}
		}
		// Pre-transcode file missing/corrupt — fall through to normal playback
	}

	decision := playback.PlaybackDecision{Method: explicitPlaybackMethod(r.URL.Query().Get("pm"))}
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
		http.Redirect(w, r, buildHLSRedirectURL(id, mf.ID, r.URL.Query()), http.StatusTemporaryRedirect)
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
	playlistPath, err := h.svc.PrepareHLS(r.Context(), id, hlsFileID, subtitleStreamIndex, videoCopy)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "transcoding failed: "+err.Error())
		return
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

	if _, err := os.Stat(path); os.IsNotExist(err) {
		// If a transcode is in progress, wait up to 10s for the segment to appear
		// instead of returning 404 immediately (which causes client retry spam).
		if !h.svc.WaitForSegment(path, 10*time.Second) {
			respondError(w, http.StatusNotFound, "segment not found")
			return
		}
	}

	if strings.HasSuffix(segment, ".m3u8") {
		serveHLSPlaylist(w, r, path)
		return
	}

	w.Header().Set("Content-Type", "video/mp2t")
	http.ServeFile(w, r, path)
}

// isValidSegmentName validates that a segment filename is safe to serve.
// Only alphanumeric, underscore, hyphen, and dot are allowed; no path separators.
// Accepts .ts (video/audio segments) and .m3u8 (audio sub-playlists).
func isValidSegmentName(name string) bool {
	if name == "" {
		return false
	}
	if !strings.HasSuffix(name, ".ts") && !strings.HasSuffix(name, ".m3u8") {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}

func buildHLSRedirectURL(mediaID, fileID int64, original url.Values) string {
	query := make(url.Values)
	query.Set("fid", strconv.FormatInt(fileID, 10))

	for _, key := range []string{"token", "at", "si", "vcopy"} {
		if value := original.Get(key); value != "" {
			query.Set(key, value)
		}
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

func serveHLSPlaylist(w http.ResponseWriter, r *http.Request, playlistPath string) {
	content, err := os.ReadFile(playlistPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot read playlist")
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Write(rewriteHLSPlaylist(content, r.URL.Query()))
}

func rewriteHLSPlaylist(content []byte, original url.Values) []byte {
	token := original.Get("token")
	at := original.Get("at")
	si := original.Get("si")

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if strings.HasPrefix(trimmed, "#EXT-X-MEDIA:") {
			lines[i] = rewriteExtXMediaURI(line, token, at, si)
			continue
		}

		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		lines[i] = appendQueryToPlaylistURI(line, token, at, si)
	}

	return []byte(strings.Join(lines, "\n"))
}

func rewriteExtXMediaURI(line, token, at, si string) string {
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

	return line[:start] + appendQueryToPlaylistURI(line[start:end], token, at, si) + line[end:]
}

func appendQueryToPlaylistURI(rawURI, token, at, si string) string {
	uri, err := url.Parse(rawURI)
	if err != nil || uri == nil {
		return rawURI
	}

	query := uri.Query()
	if token != "" {
		query.Set("token", token)
	}
	if at != "" && strings.HasSuffix(uri.Path, ".m3u8") {
		query.Set("at", at)
	}
	if si != "" && strings.HasSuffix(uri.Path, ".m3u8") {
		query.Set("si", si)
	}
	uri.RawQuery = query.Encode()
	return uri.String()
}
