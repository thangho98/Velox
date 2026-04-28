package handler

import (
	"net/http"
	"strconv"

	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

type DownloadHandler struct {
	downloadSvc *service.DownloadService
	episodeRepo *repository.EpisodeRepo
}

func NewDownloadHandler(downloadSvc *service.DownloadService, episodeRepo *repository.EpisodeRepo) *DownloadHandler {
	return &DownloadHandler{
		downloadSvc: downloadSvc,
		episodeRepo: episodeRepo,
	}
}

// StartDownload initiates a download for a specific media item.
// POST /api/media/{id}/download
func (h *DownloadHandler) StartDownload(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	mediaID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	task, err := h.downloadSvc.Enqueue(r.Context(), mediaID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, task)
}

// StartSeriesDownload initiates downloads for all episodes of a series.
// POST /api/series/{id}/download
func (h *DownloadHandler) StartSeriesDownload(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	seriesID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	episodes, err := h.episodeRepo.ListBySeriesID(r.Context(), seriesID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	enqueued := 0
	for _, ep := range episodes {
		// Ignore errors (e.g. if already local or already downloading)
		_, err := h.downloadSvc.Enqueue(r.Context(), ep.MediaID)
		if err == nil {
			enqueued++
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Series download initiated",
		"enqueued": enqueued,
		"total":    len(episodes),
	})
}

// GetTasks returns the list of all active or completed download tasks.
// GET /api/downloads
func (h *DownloadHandler) GetTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.downloadSvc.ListTasks()
	respondJSON(w, http.StatusOK, tasks)
}

// DeleteDownload deletes the local download data for a media.
// DELETE /api/media/{id}/download
func (h *DownloadHandler) DeleteDownload(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	mediaID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	if err := h.downloadSvc.DeleteDownloadedFile(r.Context(), mediaID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Local download deleted successfully"})
}

// DeleteSeriesDownload deletes local downloaded files for all episodes of a series.
// DELETE /api/series/{id}/download
func (h *DownloadHandler) DeleteSeriesDownload(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	seriesID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid series id")
		return
	}

	episodes, err := h.episodeRepo.ListBySeriesID(r.Context(), seriesID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	deletedCount := 0
	for _, ep := range episodes {
		// Ignore errors (e.g. if not downloaded)
		if err := h.downloadSvc.DeleteDownloadedFile(r.Context(), ep.MediaID); err == nil {
			deletedCount++
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Local downloads deleted successfully",
		"deleted": deletedCount,
	})
}
