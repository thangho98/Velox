package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/service"
)

type BrowseHandler struct {
	browseSvc *service.BrowseService
	mediaSvc  *service.MediaService
}

func NewBrowseHandler(browseSvc *service.BrowseService) *BrowseHandler {
	return &BrowseHandler{browseSvc: browseSvc}
}

func (h *BrowseHandler) SetMediaService(m *service.MediaService) {
	h.mediaSvc = m
}

func (h *BrowseHandler) Browse(w http.ResponseWriter, r *http.Request) {
	libraryIDStr := r.URL.Query().Get("library_id")
	relativePath := r.URL.Query().Get("path")

	if strings.Contains(relativePath, "..") {
		respondError(w, http.StatusBadRequest, "invalid path")
		return
	}

	libraryID := int64(0)
	if libraryIDStr != "" && libraryIDStr != "0" {
		parsedID, err := strconv.ParseInt(libraryIDStr, 10, 64)
		if err != nil || parsedID == 0 {
			respondError(w, http.StatusBadRequest, "invalid library_id")
			return
		}
		libraryID = parsedID
	}

	userID, isAdmin, _ := auth.UserFromContext(r.Context())
	result, err := h.browseSvc.Browse(r.Context(), userID, isAdmin, libraryID, relativePath)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "library not found")
		case errors.Is(err, service.ErrForbidden):
			respondError(w, http.StatusForbidden, "no access to this library")
		case errors.Is(err, service.ErrInvalidBrowseRoot), errors.Is(err, service.ErrLibraryHasNoRootPaths):
			respondError(w, http.StatusBadRequest, err.Error())
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	if h.mediaSvc != nil && len(result.Media) > 0 {
		_ = h.mediaSvc.AttachImageResourcesForList(r.Context(), result.Media)
	}

	respondJSON(w, http.StatusOK, result)
}
