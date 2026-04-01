package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type CatalogHandler struct {
	mediaRepo  *repository.MediaRepo
	seriesRepo *repository.SeriesRepo
	genreRepo  *repository.GenreRepo
	personRepo *repository.PersonRepo
}

func NewCatalogHandler(
	mediaRepo *repository.MediaRepo,
	seriesRepo *repository.SeriesRepo,
	genreRepo *repository.GenreRepo,
	personRepo *repository.PersonRepo,
) *CatalogHandler {
	return &CatalogHandler{
		mediaRepo:  mediaRepo,
		seriesRepo: seriesRepo,
		genreRepo:  genreRepo,
		personRepo: personRepo,
	}
}

func (h *CatalogHandler) ListGenres(w http.ResponseWriter, r *http.Request) {
	typeFilter := r.URL.Query().Get("type")

	var (
		genres []model.Genre
		err    error
	)

	if typeFilter != "" {
		genres, err = h.genreRepo.ListWithFilter(r.Context(), typeFilter)
	} else {
		genres, err = h.genreRepo.List(r.Context())
	}

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

	genres, err := h.genreRepo.ListByMediaID(r.Context(), id)
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

	credits, err := h.personRepo.ListCreditsByMedia(r.Context(), id)
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

	genres, err := h.genreRepo.ListBySeriesID(r.Context(), id)
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

	credits, err := h.personRepo.ListCreditsBySeries(r.Context(), id)
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

	movieResults, err := h.mediaRepo.ListFiltered(r.Context(), model.MediaListFilter{
		Search:    query,
		MediaType: "movie",
		Limit:     limit,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	seriesResults, err := h.seriesRepo.ListFiltered(r.Context(), model.SeriesListFilter{
		Search: query,
		Limit:  limit,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := model.SearchResult{
		Movies: movieResults,
		Series: seriesResults,
	}
	if result.Movies == nil {
		result.Movies = []model.MediaListItem{}
	}
	if result.Series == nil {
		result.Series = []model.SeriesListItem{}
	}

	respondJSON(w, http.StatusOK, result)
}
