package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SeriesHandler handles series, season, and episode endpoints.
type SeriesHandler struct {
	seriesRepo  *repository.SeriesRepo
	seasonRepo  *repository.SeasonRepo
	episodeRepo *repository.EpisodeRepo
}

// NewSeriesHandler creates a new series handler.
func NewSeriesHandler(seriesRepo *repository.SeriesRepo, seasonRepo *repository.SeasonRepo, episodeRepo *repository.EpisodeRepo) *SeriesHandler {
	return &SeriesHandler{seriesRepo: seriesRepo, seasonRepo: seasonRepo, episodeRepo: episodeRepo}
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
	}

	series, err := h.seriesRepo.ListFiltered(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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

	series, err := h.seriesRepo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondError(w, http.StatusNotFound, "series not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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

	results, err := h.seriesRepo.Search(r.Context(), q, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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

	seasons, err := h.seasonRepo.ListBySeriesID(r.Context(), seriesID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
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

	episodes, err := h.episodeRepo.ListBySeasonID(r.Context(), seasonID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, episodes)
}
