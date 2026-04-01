package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

type CinemaHandler struct {
	cinemaSvc *service.CinemaService
}

func NewCinemaHandler(cinemaSvc *service.CinemaService) *CinemaHandler {
	return &CinemaHandler{cinemaSvc: cinemaSvc}
}

func (h *CinemaHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.cinemaSvc.GetSettings(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load cinema setting")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

func (h *CinemaHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled     *bool   `json:"enabled"`
		MaxTrailers *string `json:"max_trailers"`
	}

	if err := parseJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if err := h.cinemaSvc.UpdateSettings(r.Context(), service.CinemaSettingsUpdate{
		Enabled:     body.Enabled,
		MaxTrailers: body.MaxTrailers,
	}); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save cinema setting")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *CinemaHandler) MediaCinema(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	items, err := h.cinemaSvc.MediaItems(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CinemaHandler) SeriesCinema(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	items, err := h.cinemaSvc.SeriesItems(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CinemaHandler) ServeIntro(w http.ResponseWriter, r *http.Request) {
	introPath, err := h.cinemaSvc.GetIntroPath(r.Context())
	if errors.Is(err, service.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load cinema intro")
		return
	}

	http.ServeFile(w, r, introPath)
}

func (h *CinemaHandler) UploadIntro(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "file too large (max 100MB)")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()

	introPath, err := h.cinemaSvc.SaveIntro(r.Context(), file)
	if err != nil {
		if errors.Is(err, io.EOF) {
			respondError(w, http.StatusBadRequest, "empty file")
			return
		}
		respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"path": introPath})
}
