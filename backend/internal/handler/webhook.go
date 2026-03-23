package handler

import (
	"encoding/json"
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
)

// validateEvents checks that events is a valid JSON string array and returns it normalized.
func validateEvents(raw string) (string, bool) {
	if raw == "" {
		return "[]", true
	}
	var arr []string
	if err := json.Unmarshal([]byte(raw), &arr); err != nil {
		return "", false
	}
	b, _ := json.Marshal(arr)
	return string(b), true
}

// WebhookHandler handles webhook CRUD endpoints.
type WebhookHandler struct {
	svc *service.WebhookService
}

func NewWebhookHandler(svc *service.WebhookService) *WebhookHandler {
	return &WebhookHandler{svc: svc}
}

// List returns all webhooks.
// GET /api/admin/webhooks
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	hooks, err := h.svc.List(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, hooks)
}

type createWebhookReq struct {
	URL    string `json:"url"`
	Events string `json:"events"` // JSON array string, e.g. '["scan.complete","media.added"]'
	Secret string `json:"secret"`
	Active *bool  `json:"active"`
}

// Create adds a new webhook.
// POST /api/admin/webhooks
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createWebhookReq
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.URL == "" {
		respondError(w, http.StatusBadRequest, "url is required")
		return
	}
	events, ok := validateEvents(req.Events)
	if !ok {
		respondError(w, http.StatusBadRequest, "events must be a JSON array of strings")
		return
	}

	active := true
	if req.Active != nil {
		active = *req.Active
	}

	webhook := &model.Webhook{
		URL:    req.URL,
		Events: events,
		Secret: req.Secret,
		Active: active,
	}

	if err := h.svc.Create(r.Context(), webhook); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, webhook)
}

type updateWebhookReq struct {
	URL    *string `json:"url"`
	Events *string `json:"events"`
	Secret *string `json:"secret"`
	Active *bool   `json:"active"`
}

// Update modifies an existing webhook.
// PUT /api/admin/webhooks/{id}
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	existing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusNotFound, "webhook not found")
		return
	}

	var req updateWebhookReq
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.URL != nil {
		existing.URL = *req.URL
	}
	if req.Events != nil {
		events, ok := validateEvents(*req.Events)
		if !ok {
			respondError(w, http.StatusBadRequest, "events must be a JSON array of strings")
			return
		}
		existing.Events = events
	}
	if req.Secret != nil {
		existing.Secret = *req.Secret
	}
	if req.Active != nil {
		existing.Active = *req.Active
	}

	if err := h.svc.Update(r.Context(), existing); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

// Delete removes a webhook.
// DELETE /api/admin/webhooks/{id}
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
