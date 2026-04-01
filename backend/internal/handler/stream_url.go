package handler

import (
	"fmt"
	"net/http"

	"github.com/thawng/velox/internal/auth"
)

type StreamURLHandler struct {
	apiKeyStore *auth.APIKeyStore
}

func NewStreamURLHandler(apiKeyStore *auth.APIKeyStore) *StreamURLHandler {
	return &StreamURLHandler{apiKeyStore: apiKeyStore}
}

func (h *StreamURLHandler) Create(w http.ResponseWriter, r *http.Request) {
	mediaID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	userID, isAdmin, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	apiKey := h.apiKeyStore.Generate(userID, isAdmin)
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"direct_url": fmt.Sprintf("%s://%s/api/stream/%d?api_key=%s", scheme, r.Host, mediaID, apiKey),
		"hls_url":    fmt.Sprintf("%s://%s/api/stream/%d/hls/master.m3u8?api_key=%s", scheme, r.Host, mediaID, apiKey),
		"api_key":    apiKey,
		"expires_in": int(auth.StreamTokenExpiry.Seconds()),
	})
}
