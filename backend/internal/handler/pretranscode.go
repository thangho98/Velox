package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// PretranscodeHandler handles admin pre-transcode endpoints.
type PretranscodeHandler struct {
	svc *service.PretranscodeService
}

// NewPretranscodeHandler creates a new pre-transcode handler.
func NewPretranscodeHandler(svc *service.PretranscodeService) *PretranscodeHandler {
	return &PretranscodeHandler{svc: svc}
}

// GetStatus returns the current pre-transcode status.
// Uses main DB (not background DB) to avoid starving the scheduler goroutine.
// GET /api/admin/pretranscode/status
func (h *PretranscodeHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.svc.GetStatus(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get status")
		return
	}
	respondJSON(w, http.StatusOK, status)
}

// Start enqueues all libraries and starts encoding.
// POST /api/admin/pretranscode/start
func (h *PretranscodeHandler) Start(w http.ResponseWriter, r *http.Request) {
	n, err := h.svc.StartAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to enqueue: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]int{"enqueued": n})
}

// Stop cancels queued jobs and pauses the scheduler.
// POST /api/admin/pretranscode/stop
func (h *PretranscodeHandler) Stop(w http.ResponseWriter, r *http.Request) {
	cancelled, err := h.svc.CancelAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to stop")
		return
	}
	respondJSON(w, http.StatusOK, map[string]int64{"cancelled": cancelled})
}

// Resume resumes the paused scheduler.
// POST /api/admin/pretranscode/resume
func (h *PretranscodeHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.svc.Resume()
	respondJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}

// Estimate returns storage estimates for pre-transcoding.
// GET /api/admin/pretranscode/estimate?library_id=1
func (h *PretranscodeHandler) Estimate(w http.ResponseWriter, r *http.Request) {
	libraryID := int64(parseIntQuery(r, "library_id", 0))
	if libraryID <= 0 {
		respondError(w, http.StatusBadRequest, "library_id required")
		return
	}
	estimate, err := h.svc.EstimateStorage(r.Context(), libraryID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "estimation failed: "+err.Error())
		return
	}
	respondJSON(w, http.StatusOK, estimate)
}

// Cleanup deletes all pre-transcode files.
// DELETE /api/admin/pretranscode/files
func (h *PretranscodeHandler) Cleanup(w http.ResponseWriter, r *http.Request) {
	removed, err := h.svc.CleanupAndDisable(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}
	respondJSON(w, http.StatusOK, map[string]int{"removed": removed})
}

// ListProfiles returns all quality profiles.
// GET /api/admin/pretranscode/profiles
func (h *PretranscodeHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.svc.ListProfiles(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list profiles")
		return
	}
	respondJSON(w, http.StatusOK, profiles)
}

// ToggleProfile enables or disables a quality profile.
// PUT /api/admin/pretranscode/profiles/{id}
func (h *PretranscodeHandler) ToggleProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid profile id")
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := h.svc.SetProfileEnabled(r.Context(), id, req.Enabled); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"enabled": req.Enabled})
}

// GetSettings returns the pre-transcode settings.
// GET /api/admin/settings/pretranscode
func (h *PretranscodeHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.svc.GetSettings(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateSettings saves pre-transcode settings.
// PUT /api/admin/settings/pretranscode
func (h *PretranscodeHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled     *bool   `json:"enabled"`
		Schedule    *string `json:"schedule"`
		Concurrency *string `json:"concurrency"`
	}
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	settings, err := h.svc.UpdateSettings(r.Context(), service.PretranscodeSettingsUpdate{
		Enabled:     req.Enabled,
		Schedule:    req.Schedule,
		Concurrency: req.Concurrency,
	})
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}
