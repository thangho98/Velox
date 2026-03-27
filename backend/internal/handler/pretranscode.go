package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

// PretranscodeHandler handles admin pre-transcode endpoints.
type PretranscodeHandler struct {
	svc          *service.PretranscodeService
	settingsRepo *repository.AppSettingsRepo
}

// NewPretranscodeHandler creates a new pre-transcode handler.
func NewPretranscodeHandler(svc *service.PretranscodeService, settingsRepo *repository.AppSettingsRepo) *PretranscodeHandler {
	return &PretranscodeHandler{svc: svc, settingsRepo: settingsRepo}
}

// GetStatus returns the current pre-transcode status.
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
	ctx := r.Context()

	// Enable in settings if not already
	if err := h.settingsRepo.Set(ctx, model.SettingPretranscodeEnabled, "true"); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to enable")
		return
	}

	n, err := h.svc.EnqueueAllLibraries(ctx)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to enqueue: "+err.Error())
		return
	}

	h.svc.Resume()
	if !h.svc.IsRunning() {
		h.svc.Start()
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
	removed, err := h.svc.CleanupAll(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "cleanup failed")
		return
	}
	_ = h.settingsRepo.Set(r.Context(), model.SettingPretranscodeEnabled, "false")
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
	ctx := r.Context()
	vals, err := h.settingsRepo.GetMulti(ctx,
		model.SettingPretranscodeEnabled,
		model.SettingPretranscodeSchedule,
		model.SettingPretranscodeConcurrency,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	schedule := vals[model.SettingPretranscodeSchedule]
	if schedule == "" {
		schedule = "always"
	}
	concurrency := vals[model.SettingPretranscodeConcurrency]
	if concurrency == "" {
		concurrency = "1"
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"enabled":     vals[model.SettingPretranscodeEnabled] == "true",
		"schedule":    schedule,
		"concurrency": concurrency,
	})
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

	ctx := r.Context()
	if req.Enabled != nil {
		val := "false"
		if *req.Enabled {
			val = "true"
		}
		if err := h.settingsRepo.Set(ctx, model.SettingPretranscodeEnabled, val); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save")
			return
		}
	}
	if req.Schedule != nil {
		if err := h.settingsRepo.Set(ctx, model.SettingPretranscodeSchedule, *req.Schedule); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save")
			return
		}
	}
	if req.Concurrency != nil {
		if err := h.settingsRepo.Set(ctx, model.SettingPretranscodeConcurrency, *req.Concurrency); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save")
			return
		}
	}

	h.GetSettings(w, r)
}
