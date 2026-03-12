package handler

import (
	"errors"
	"net/http"
	"os"
	"strconv"

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
		hlsURL := "/api/stream/" + strconv.FormatInt(id, 10) + "/hls/master.m3u8?fid=" + strconv.FormatInt(mf.ID, 10) + "&at=" + at
		http.Redirect(w, r, hlsURL, http.StatusTemporaryRedirect)
		return
	}

	// Use permissive defaults: quality limits are enforced by the frontend via
	// GetPlaybackInfo, which returns the HLS URL when transcoding is required.
	profile := playback.DetectClient(r.UserAgent())
	prefs := playback.UserPreferences{
		MaxStreamingQuality: "original",
		PreferDirectPlay:    true,
	}
	decision := playback.Decide(mediaInfo, profile, prefs)

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
			return
		}

	default:
		// TranscodeAudio or FullTranscode: redirect client to HLS endpoint.
		http.Redirect(w, r,
			"/api/stream/"+strconv.FormatInt(id, 10)+"/hls/master.m3u8?fid="+strconv.FormatInt(mf.ID, 10),
			http.StatusTemporaryRedirect)
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

	playlistPath, err := h.svc.PrepareHLS(r.Context(), id, hlsFileID, subtitleStreamIndex)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "transcoding failed: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	http.ServeFile(w, r, playlistPath)
}

// HLSSegment serves individual HLS .ts segments.
func (h *StreamHandler) HLSSegment(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	segment := r.PathValue("segment")
	path := h.svc.SegmentPath(id, segment)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		respondError(w, http.StatusNotFound, "segment not found")
		return
	}

	w.Header().Set("Content-Type", "video/mp2t")
	http.ServeFile(w, r, path)
}
