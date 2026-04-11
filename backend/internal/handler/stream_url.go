package handler

import (
	"fmt"
	"net/http"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/service"
)

type StreamURLHandler struct {
	apiKeyStore *auth.APIKeyStore
	streamSvc   *service.StreamService
}

func NewStreamURLHandler(apiKeyStore *auth.APIKeyStore, streamSvc *service.StreamService) *StreamURLHandler {
	return &StreamURLHandler{
		apiKeyStore: apiKeyStore,
		streamSvc:   streamSvc,
	}
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

	streamTTL := auth.MinStreamAPIKeyTTL
	if h.streamSvc != nil {
		if primaryFile, err := h.streamSvc.GetPrimaryFile(r.Context(), mediaID, 0); err == nil && primaryFile != nil {
			streamTTL = auth.StreamAPIKeyTTL(primaryFile.Duration)
		}
	}

	apiKey := h.apiKeyStore.Generate(userID, isAdmin, streamTTL)
	streamSessionID := newStreamSessionID()
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"direct_url":        fmt.Sprintf("%s://%s/api/stream/%d?api_key=%s&%s=%s", scheme, r.Host, mediaID, apiKey, streamSessionQueryKey, streamSessionID),
		"hls_url":           fmt.Sprintf("%s://%s/api/stream/%d/hls/master.m3u8?api_key=%s&%s=%s", scheme, r.Host, mediaID, apiKey, streamSessionQueryKey, streamSessionID),
		"stream_session_id": streamSessionID,
		"api_key":           apiKey,
		"expires_in":        int(streamTTL.Seconds()),
	})
}
