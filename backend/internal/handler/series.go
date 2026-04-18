package handler

import (
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
)

// SeriesHandler handles series, season, and episode endpoints.
type SeriesHandler struct {
	seriesSvc *service.SeriesService
	mediaSvc  *service.MediaService
}

// NewSeriesHandler creates a new series handler.
func NewSeriesHandler(seriesSvc *service.SeriesService) *SeriesHandler {
	return &SeriesHandler{seriesSvc: seriesSvc}
}

func (h *SeriesHandler) SetMediaService(m *service.MediaService) {
	h.mediaSvc = m
}

// ListSeries returns a list of series with optional filtering.
// GET /api/series?library_id=&search=&genre=&year=&sort=&limit=&offset=
func (h *SeriesHandler) ListSeries(w http.ResponseWriter, r *http.Request) {
	libraryID, _ := parseInt64Query(r.URL.Query().Get("library_id"))

	// Always use ListFiltered — returns SeriesListItem[] (superset of Series[])
	filter := model.SeriesListFilter{
		LibraryID: libraryID,
		Search:    r.URL.Query().Get("search"),
		Genre:     r.URL.Query().Get("genre"),
		Year:      r.URL.Query().Get("year"),
		Sort:      r.URL.Query().Get("sort"),
		Limit:     parseIntQuery(r, "limit", 50),
		Offset:    parseIntQuery(r, "offset", 0),
		StartChar: r.URL.Query().Get("start_char"),
	}

	series, err := h.seriesSvc.ListFiltered(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		_ = h.mediaSvc.AttachImageResourcesForSeriesList(r.Context(), series)
	}

	respondJSON(w, http.StatusOK, series)
}

// GetSeries returns a single series by ID.
// GET /api/series/{id}
func (h *SeriesHandler) GetSeries(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	series, err := h.seriesSvc.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "series not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		_ = h.mediaSvc.AttachImageResourcesForSeries(r.Context(), series)
	}

	respondJSON(w, http.StatusOK, series)
}

// SearchSeries searches for series by title.
// GET /api/series/search?q=&limit=
func (h *SeriesHandler) SearchSeries(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		respondError(w, http.StatusBadRequest, "query required")
		return
	}

	limit := parseIntQuery(r, "limit", 20)

	results, err := h.seriesSvc.Search(r.Context(), q, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		_ = h.mediaSvc.AttachImageResourcesForSeriesBatch(r.Context(), results)
	}

	respondJSON(w, http.StatusOK, results)
}

// ListSeasons returns all seasons for a series.
// GET /api/series/{id}/seasons
func (h *SeriesHandler) ListSeasons(w http.ResponseWriter, r *http.Request) {
	seriesID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	seasons, err := h.seriesSvc.ListSeasons(r.Context(), seriesID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		_ = h.mediaSvc.AttachImageResourcesForSeasons(r.Context(), seasons)
	}

	respondJSON(w, http.StatusOK, seasons)
}

// ListEpisodes returns all episodes for a season within a series.
// GET /api/series/{id}/seasons/{seasonId}/episodes
func (h *SeriesHandler) ListEpisodes(w http.ResponseWriter, r *http.Request) {
	seasonID, err := parseID(r, "seasonId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid season id")
		return
	}

	episodes, err := h.seriesSvc.ListEpisodes(r.Context(), seasonID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if h.mediaSvc != nil {
		_ = h.mediaSvc.AttachImageResourcesForEpisodes(r.Context(), episodes)
	}

	respondJSON(w, http.StatusOK, episodes)
}

// GetAlphabet returns the A-Z grouping count for the series library.
// GET /api/series/alphabet
func (h *SeriesHandler) GetAlphabet(w http.ResponseWriter, r *http.Request) {
	libraryID, _ := parseInt64Query(r.URL.Query().Get("library_id"))
	filter := model.SeriesListFilter{
		LibraryID: libraryID,
		Genre:     r.URL.Query().Get("genre"),
		Year:      r.URL.Query().Get("year"),
	}

	alphabet, err := h.seriesSvc.GetAlphabet(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, alphabet)
}
