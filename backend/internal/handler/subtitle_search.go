package handler

import (
	"log"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// SubtitleSearchHandler handles external subtitle search endpoints.
type SubtitleSearchHandler struct {
	svc *service.SubtitleSearchService
}

// NewSubtitleSearchHandler creates a new subtitle search handler.
func NewSubtitleSearchHandler(svc *service.SubtitleSearchService) *SubtitleSearchHandler {
	return &SubtitleSearchHandler{svc: svc}
}

// Search queries external providers for subtitles.
// GET /api/media/{id}/subtitles/search?lang=en
func (h *SubtitleSearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	lang := r.URL.Query().Get("lang")

	results, err := h.svc.Search(r.Context(), mediaID, lang)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "subtitle search failed")
		return
	}

	respondJSON(w, http.StatusOK, results)
}

// downloadRequest is the body for POST /api/media/{id}/subtitles/download
type downloadRequest struct {
	Provider   string `json:"provider"`
	ExternalID string `json:"external_id"`
	Language   string `json:"language"`
}

// Download fetches a subtitle from an external provider and saves it locally.
// POST /api/media/{id}/subtitles/download
func (h *SubtitleSearchHandler) Download(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	var req downloadRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.Provider == "" || req.ExternalID == "" {
		respondError(w, http.StatusBadRequest, "provider and external_id are required")
		return
	}

	sub, err := h.svc.Download(r.Context(), mediaID, req.Provider, req.ExternalID, req.Language)
	if err != nil {
		log.Printf("subtitle download failed for media %d provider %s: %v", mediaID, req.Provider, err)
		respondError(w, http.StatusInternalServerError, "subtitle download failed")
		return
	}

	respondJSON(w, http.StatusOK, sub)
}
