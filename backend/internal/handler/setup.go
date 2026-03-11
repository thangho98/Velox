package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/service"
)

type SetupHandler struct {
	authSvc *service.AuthService
}

func NewSetupHandler(authSvc *service.AuthService) *SetupHandler {
	return &SetupHandler{authSvc: authSvc}
}

// Status returns whether the system is configured
func (h *SetupHandler) Status(w http.ResponseWriter, r *http.Request) {
	configured, err := h.authSvc.IsConfigured(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]bool{"configured": configured})
}

type setupReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

// Setup creates the first admin user (only works when not configured)
func (h *SetupHandler) Setup(w http.ResponseWriter, r *http.Request) {
	// Check if already configured
	configured, err := h.authSvc.IsConfigured(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if configured {
		respondError(w, http.StatusForbidden, "setup already completed")
		return
	}

	var req setupReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" || req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "username, password and display_name are required")
		return
	}

	user, err := h.authSvc.CreateFirstAdmin(r.Context(), req.Username, req.Password, req.DisplayName)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUsername):
			respondError(w, http.StatusBadRequest, "username must be 3-32 alphanumeric characters")
		case errors.Is(err, service.ErrInvalidPassword):
			respondError(w, http.StatusBadRequest, "password must be at least 8 characters")
		case errors.Is(err, service.ErrUserExists):
			respondError(w, http.StatusConflict, "username already exists")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	// Don't return password hash
	user.PasswordHash = ""
	respondJSON(w, http.StatusCreated, user)
}
