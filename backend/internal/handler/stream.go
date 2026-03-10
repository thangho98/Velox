package handler

import (
	"net/http"
	"os"

	"github.com/thawng/velox/internal/service"
)

type StreamHandler struct {
	svc *service.StreamService
}

func NewStreamHandler(svc *service.StreamService) *StreamHandler {
	return &StreamHandler{svc: svc}
}

// DirectPlay serves the video file directly with range request support.
func (h *StreamHandler) DirectPlay(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	path, err := h.svc.DirectPlayPath(id)
	if err != nil {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}

	f, err := os.Open(path)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cannot open file")
		return
	}
	defer f.Close()

	stat, _ := f.Stat()

	// ServeContent handles Range headers, Content-Type, etc.
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

// HLSMaster serves the HLS master playlist. Triggers transcoding if needed.
func (h *StreamHandler) HLSMaster(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	playlistPath, err := h.svc.PrepareHLS(id)
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
