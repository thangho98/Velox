package handler

import (
	"encoding/json"
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
)

type ProgressHandler struct {
	svc *service.ProgressService
}

func NewProgressHandler(svc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) Get(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "mediaID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}
	p, err := h.svc.Get(mediaID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}

func (h *ProgressHandler) Update(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "mediaID")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	var p model.Progress
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	p.MediaID = mediaID

	if err := h.svc.Update(&p); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, p)
}
