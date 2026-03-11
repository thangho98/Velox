package handler

import (
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

type MediaHandler struct {
	svc *service.MediaService
}

func NewMediaHandler(svc *service.MediaService) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	libraryID := int64(parseIntQuery(r, "library_id", 0))
	mediaType := r.URL.Query().Get("type") // "movie" or "episode"
	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)

	items, err := h.svc.List(r.Context(), libraryID, mediaType, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, items)
}

func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	m, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}

func (h *MediaHandler) GetWithFiles(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	m, err := h.svc.GetWithFiles(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, m)
}
