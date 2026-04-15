package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/service"
)

type CatalogHandler struct {
	catalogSvc *service.CatalogService
	mediaSvc   *service.MediaService
}

func NewCatalogHandler(catalogSvc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{catalogSvc: catalogSvc}
}

func (h *CatalogHandler) SetMediaService(m *service.MediaService) {
	h.mediaSvc = m
}

func (h *CatalogHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")
	genres, err := h.catalogSvc.ListGenres(r.Context(), typeFilter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, genres)
}

func (h *CatalogHandler) ListMediaGenres(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	genres, err := h.catalogSvc.ListMediaGenres(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, genres)
}

func (h *CatalogHandler) ListMediaCredits(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	credits, err := h.catalogSvc.ListMediaCredits(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, credits)
}

func (h *CatalogHandler) ListSeriesGenres(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	genres, err := h.catalogSvc.ListSeriesGenres(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, genres)
}

func (h *CatalogHandler) ListSeriesCredits(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	credits, err := h.catalogSvc.ListSeriesCredits(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, credits)
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		respondError(w, http.StatusBadRequest, "query required (use ?q=search_term)")
		return
	}

	limit := parseIntQuery(r, "limit", 20)

	result, err := h.catalogSvc.Search(r.Context(), query, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		if len(result.Movies) > 0 {
			_ = h.mediaSvc.AttachImageResourcesForList(r.Context(), result.Movies)
		}
		if len(result.Series) > 0 {
			_ = h.mediaSvc.AttachImageResourcesForSeriesList(r.Context(), result.Series)
		}
	}

	respondJSON(w, http.StatusOK, result)
}
