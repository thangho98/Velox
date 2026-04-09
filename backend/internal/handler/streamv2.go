package handler

import (
	"log/slog"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/hls"
	"github.com/thawng/velox/internal/logger"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/streamv2"
)

// pickAudioTracks returns only the audio track matching the given track ID query
// param. Encoding ALL audio streams simultaneously is fragile — a single
// corrupt/unsupported stream can crash the whole ffmpeg process. When no ID is
// provided (or it doesn't match), we fall back to the default track or the first
// one so we always encode exactly one stream.
func pickAudioTracks(all []model.AudioTrack, atParam string) []model.AudioTrack {
	if len(all) == 0 {
		return all
	}
	if atParam != "" {
		if atID, err := strconv.ParseInt(atParam, 10, 64); err == nil {
			for _, t := range all {
				if t.ID == atID {
					return []model.AudioTrack{t}
				}
			}
		}
	}
	for _, t := range all {
		if t.IsDefault {
			return []model.AudioTrack{t}
		}
	}
	return all[:1]
}

type StreamV2Handler struct {
	svc *service.StreamService
	mgr *streamv2.Manager
	log *slog.Logger
}

func NewStreamV2Handler(svc *service.StreamService, mgr *streamv2.Manager) *StreamV2Handler {
	return &StreamV2Handler{
		svc: svc,
		mgr: mgr,
		log: logger.NewWith("stream_v2"),
	}
}

// HLSMaster handles /api/stream/v2/{id}/hls/master.m3u8
func (h *StreamV2Handler) HLSMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	ss := streamSessionIDFromValues(r.URL.Query())
	if ss == "" {
		ss = newStreamSessionID()
	}

	var fileID int64
	if fid := r.URL.Query().Get("fid"); fid != "" {
		if n, err := strconv.ParseInt(fid, 10, 64); err == nil {
			fileID = n
		}
	}

	subtitleIdx := -1
	if si := r.URL.Query().Get("si"); si != "" {
		if n, err := strconv.Atoi(si); err == nil {
			subtitleIdx = n
		}
	}

	videoCopy := r.URL.Query().Get("vcopy") == "1"

	maxHeight := 0
	if mh := r.URL.Query().Get("mh"); mh != "" {
		if n, err := strconv.Atoi(mh); err == nil {
			maxHeight = n
		}
	}

	mf, err := h.svc.GetPrimaryFile(r.Context(), id, fileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "media file not found")
		return
	}

	audioTracks, _ := h.svc.ListAudioTracksForMediaFile(r.Context(), mf.ID)
	audioTracks = pickAudioTracks(audioTracks, r.URL.Query().Get("at"))

	key := hls.SessionKey{
		StreamSessionID:   ss,
		MediaID:           id,
		FileID:            mf.ID,
		SubtitleStreamIdx: subtitleIdx,
		VideoCopy:         videoCopy,
		MaxHeight:         maxHeight,
	}

	sess, err := h.mgr.GetOrCreate(r.Context(), key, mf.FilePath, mf.Duration, audioTracks)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot create session")
		return
	}

	startSeg := 0
	if so := r.URL.Query().Get("start"); so != "" {
		if n, err := strconv.ParseFloat(so, 64); err == nil && n > 0 {
			startSeg = int(math.Floor(n / sess.SegLength))
		}
	}

	if err := sess.PrimeFromSegment(r.Context(), startSeg); err != nil {
		h.log.Error("Failed to prime ffmpeg on master m3u8 request", "err", err)
		respondError(w, http.StatusServiceUnavailable, "transcoder failed to start")
		return
	}

	bw := mf.Bitrate
	if bw == 0 {
		bw = 4000000
	}

	masterStr := hls.GenerateMasterPlaylist(audioTracks, sess.Prefix(), bw)

	// Propagate auth query params (api_key, token, ssid) to all child URIs
	// in the master playlist. Without this, standalone HLS URLs (from /stream/{id}/url)
	// would lose auth on sub-playlist and segment requests.
	masterBytes := rewriteHLSPlaylist([]byte(masterStr), r.URL.Query())

	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(masterBytes)
}

// HLSMedia handles /api/stream/v2/{id}/hls/{segment}
func (h *StreamV2Handler) HLSMedia(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("segment")

	key, kind, segNum, err := hls.ParseFilename(filename)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid hls filename format")
		return
	}

	mf, err := h.svc.GetPrimaryFile(r.Context(), key.MediaID, key.FileID)
	if err != nil {
		respondError(w, http.StatusNotFound, "media file not found")
		return
	}
	audioTracks, _ := h.svc.ListAudioTracksForMediaFile(r.Context(), mf.ID)
	audioTracks = pickAudioTracks(audioTracks, r.URL.Query().Get("at"))

	sess, err := h.mgr.GetOrCreate(r.Context(), key, mf.FilePath, mf.Duration, audioTracks)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot retrieve session")
		return
	}

	if strings.HasSuffix(filename, ".m3u8") {
		// Block on first sub-playlist request until ffmpeg has produced the
		// starting segment. Without this, the first response is a playlist
		// with zero #EXTINF entries, which makes hls.js fire a
		// `levelEmptyError` and wait ~3s before retrying. Waiting on the
		// backend side is smoother and prevents player error handlers from
		// considering the stream broken.
		startSeg := sess.SessionStartSegment()
		if _, ok := sess.GetExtinfSnapshot(kind)[startSeg]; !ok {
			_ = sess.RequestSegment(r.Context(), kind, startSeg)
		}

		snapshot := sess.GetExtinfSnapshot(kind)
		playlist := hls.RenderMediaPlaylist(hls.RenderOpts{
			ExtinfMap:       snapshot,
			SessionStartSeg: startSeg,
			TotalDuration:   mf.Duration,
			SegLength:       sess.SegLength,
			Prefix:          sess.Prefix(),
			Kind:            kind,
		})

		// Propagate auth query params to segment URIs within sub-playlists
		playlistBytes := rewriteHLSPlaylist([]byte(playlist), r.URL.Query())

		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		w.Header().Set("Cache-Control", "no-store")
		w.Write(playlistBytes)
		return
	}

	if strings.HasSuffix(filename, ".mp4") && segNum == -1 {
		if err := sess.RequestInitSegment(r.Context(), kind); err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		serveStaticSegment(w, r, filepath.Join(sess.OutputDir, filename))
		return
	}

	if strings.HasSuffix(filename, ".m4s") {
		if err := sess.RequestSegment(r.Context(), kind, segNum); err != nil {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		serveStaticSegment(w, r, filepath.Join(sess.OutputDir, filename))
		return
	}

	respondError(w, http.StatusBadRequest, "unsupported extension")
}

func serveStaticSegment(w http.ResponseWriter, r *http.Request, path string) {
	if strings.HasSuffix(path, ".mp4") || strings.HasSuffix(path, ".m4s") {
		w.Header().Set("Content-Type", "video/mp4")
	} else {
		w.Header().Set("Content-Type", "video/mp2t")
	}
	w.Header().Set("Cache-Control", "public, max-age=300")
	http.ServeFile(w, r, path)
}
