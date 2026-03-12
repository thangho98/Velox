package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/trickplay"
)

// validSpriteName matches exactly "sprite_N.jpg" where N is one or more digits.
var validSpriteName = regexp.MustCompile(`^sprite_\d+\.jpg$`)

// TrickplayHandler serves trickplay sprite sheets and VTT manifests.
// If the generator is nil (trickplay disabled), all endpoints return 404.
type TrickplayHandler struct {
	gen    *trickplay.Generator // nil when trickplay is disabled
	stream *service.StreamService
}

func NewTrickplayHandler(gen *trickplay.Generator, stream *service.StreamService) *TrickplayHandler {
	return &TrickplayHandler{gen: gen, stream: stream}
}

// ServeVTT serves the WebVTT manifest for a media item's trickplay thumbnails.
// If sprites haven't been generated yet, triggers async generation and returns 202.
func (h *TrickplayHandler) ServeVTT(w http.ResponseWriter, r *http.Request) {
	if h.gen == nil {
		respondError(w, http.StatusNotFound, "trickplay not enabled")
		return
	}

	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	vttPath := h.gen.VTTPath(id)
	if _, err := os.Stat(vttPath); os.IsNotExist(err) {
		// Trigger async generation; client should poll until 200.
		mf, err := h.stream.GetPrimaryFile(r.Context(), id, 0)
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "media not found")
			return
		}
		if err != nil {
			respondError(w, http.StatusInternalServerError, "internal error")
			return
		}
		h.gen.GenerateAsync(id, mf.FilePath, int(mf.Duration))
		w.WriteHeader(http.StatusAccepted) // 202: generation started, try again later
		return
	}

	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, vttPath)
}

// ServeSprite serves a trickplay sprite sheet image.
// sprite path value must be "sprite_N.jpg" (1-based index).
func (h *TrickplayHandler) ServeSprite(w http.ResponseWriter, r *http.Request) {
	if h.gen == nil {
		respondError(w, http.StatusNotFound, "trickplay not enabled")
		return
	}

	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	sprite := r.PathValue("sprite")
	// Sanitize: only allow "sprite_N.jpg" to prevent path traversal.
	if !validSpriteName.MatchString(sprite) {
		respondError(w, http.StatusBadRequest, "invalid sprite name")
		return
	}

	spritePath := filepath.Join(h.gen.MediaDir(id), sprite)
	if _, err := os.Stat(spritePath); os.IsNotExist(err) {
		respondError(w, http.StatusNotFound, "sprite not found")
		return
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	http.ServeFile(w, r, spritePath)
}
