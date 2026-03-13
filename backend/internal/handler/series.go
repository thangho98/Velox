package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/repository"
)

// SeriesHandler handles series, season, and episode endpoints.
type SeriesHandler struct {
	seasonRepo  *repository.SeasonRepo
	episodeRepo *repository.EpisodeRepo
}

// NewSeriesHandler creates a new series handler.
func NewSeriesHandler(seasonRepo *repository.SeasonRepo, episodeRepo *repository.EpisodeRepo) *SeriesHandler {
	return &SeriesHandler{seasonRepo: seasonRepo, episodeRepo: episodeRepo}
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
