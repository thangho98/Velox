package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
)

// NotificationHandler handles notification API endpoints
type NotificationHandler struct {
	svc *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

// List handles GET /api/notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	limit := parseIntQuery(r, "limit", 20)
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := parseIntQuery(r, "offset", 0)

	notifications, err := h.svc.GetByUser(r.Context(), userID, unreadOnly, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	unreadCount, _ := h.svc.CountUnread(r.Context(), userID)

	respondJSON(w, http.StatusOK, map[string]any{
		"notifications": notifications,
		"unread_count":  unreadCount,
	})
}

// MarkAsRead handles PATCH /api/notifications/read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if len(req.IDs) == 0 {
		respondError(w, http.StatusBadRequest, "no notification ids provided")
		return
	}

	if err := h.svc.MarkAsRead(r.Context(), userID, req.IDs); err != nil {
		respondError(w, http.StatusNotFound, "notifications not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true})
}

// MarkAllAsRead handles PATCH /api/notifications/read-all
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.svc.MarkAllAsRead(r.Context(), userID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to mark all as read")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true})
}

// Delete handles POST /api/notifications/delete — bulk delete by IDs in JSON body
func (h *NotificationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		IDs []int64 `json:"ids"`
	}
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if len(req.IDs) == 0 {
		respondError(w, http.StatusBadRequest, "no notification ids provided")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, req.IDs); err != nil {
		respondError(w, http.StatusNotFound, "notifications not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]any{"success": true})
}

// DeleteOne handles DELETE /api/notifications/{id} — single delete by path param
func (h *NotificationHandler) DeleteOne(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), userID, []int64{id}); err != nil {
		respondError(w, http.StatusNotFound, "notification not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UnreadCount handles GET /api/notifications/unread-count
func (h *NotificationHandler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	count, err := h.svc.CountUnread(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get unread count")
		return
	}

	respondJSON(w, http.StatusOK, map[string]int64{"count": count})
}

// Types handles GET /api/notifications/types
func (h *NotificationHandler) Types(w http.ResponseWriter, r *http.Request) {
	types := []map[string]string{
		{"value": string(model.NotificationScanComplete), "label": "Scan Complete"},
		{"value": string(model.NotificationMediaAdded), "label": "Media Added"},
		{"value": string(model.NotificationTranscodeComplete), "label": "Transcode Complete"},
		{"value": string(model.NotificationTranscodeFailed), "label": "Transcode Failed"},
		{"value": string(model.NotificationSubtitleDownloaded), "label": "Subtitle Downloaded"},
		{"value": string(model.NotificationIdentifyComplete), "label": "Identify Complete"},
		{"value": string(model.NotificationLibraryWatcher), "label": "Library Watcher"},
	}

	respondJSON(w, http.StatusOK, types)
}
