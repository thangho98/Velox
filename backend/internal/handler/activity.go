package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
)

// ActivityHandler handles activity log and stats endpoints.
type ActivityHandler struct {
	svc *service.ActivityService
}

func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

// List returns activity log entries with optional filters.
// GET /api/admin/activity?limit=50&user_id=&action=&from=&to=&offset=
func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := model.ActivityFilter{
		Action: r.URL.Query().Get("action"),
		From:   r.URL.Query().Get("from"),
		To:     r.URL.Query().Get("to"),
		Limit:  parseIntQuery(r, "limit", 50),
		Offset: parseIntQuery(r, "offset", 0),
	}

	if uid := r.URL.Query().Get("user_id"); uid != "" {
		id, err := parseInt64Query(uid)
		if err == nil {
			filter.UserID = &id
		}
	}

	logs, err := h.svc.ListActivity(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, logs)
}

// GetPlaybackStats returns aggregated playback statistics.
// GET /api/admin/stats/playback
func (h *ActivityHandler) GetPlaybackStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetPlaybackStats(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, stats)
}
