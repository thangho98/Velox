package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// MetadataHandler handles metadata identify/refresh endpoints.
type MetadataHandler struct {
	mediaSvc    *service.MediaService
	metadataSvc *service.MetadataService
}

// NewMetadataHandler creates a new metadata handler. Returns nil if metadataSvc is nil.
func NewMetadataHandler(mediaSvc *service.MediaService, metadataSvc *service.MetadataService) *MetadataHandler {
	if metadataSvc == nil {
		return nil
	}
	return &MetadataHandler{mediaSvc: mediaSvc, metadataSvc: metadataSvc}
}

// identifyRequest is the body for PUT /api/media/{id}/identify
type identifyRequest struct {
	TmdbID    int    `json:"tmdb_id"`
	MediaType string `json:"media_type"` // "movie" or "tv"
}

// Identify overrides auto-match with a specific TMDb ID.
// PUT /api/media/{id}/identify
func (h *MetadataHandler) Identify(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req identifyRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.TmdbID <= 0 {
		respondError(w, http.StatusBadRequest, "tmdb_id is required")
		return
	}

	media, err := h.mediaSvc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if err := h.metadataSvc.IdentifyByTmdbID(r.Context(), media, req.TmdbID, req.MediaType); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Re-fetch updated media
	updated, _ := h.mediaSvc.Get(r.Context(), id)
	respondJSON(w, http.StatusOK, updated)
}

// Refresh re-fetches metadata from TMDb for a media item.
// POST /api/media/{id}/refresh
func (h *MetadataHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	media, err := h.mediaSvc.Get(r.Context(), id)
	if errors.Is(err, service.ErrNotFound) {
		respondError(w, http.StatusNotFound, "media not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if media.TmdbID == nil {
		// No TMDb match yet — try auto-matching from file name
		if err := h.metadataSvc.AutoMatchAndRefresh(r.Context(), media); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		if err := h.metadataSvc.RefreshMetadata(r.Context(), media); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	updated, _ := h.mediaSvc.Get(r.Context(), id)
	respondJSON(w, http.StatusOK, updated)
}

// BulkRefreshRatings auto-matches all unmatched media and fetches OMDb ratings.
// POST /api/admin/metadata/refresh-ratings
func (h *MetadataHandler) BulkRefreshRatings(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting bulk metadata refresh (TMDb match + OMDb ratings)...")
	updated, err := h.metadataSvc.BulkRefreshAllMetadata(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("Bulk metadata refresh complete: %d items updated", updated)
	respondJSON(w, http.StatusOK, map[string]int{"updated": updated})
}
