package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/service"
)

type AuthHandler struct {
	authSvc     *service.AuthService
	activitySvc *service.ActivityService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) SetActivityService(svc *service.ActivityService) {
	h.activitySvc = svc
}

type loginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResp struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"` // seconds
	User         userInfo `json:"user"`
}

type userInfo struct {
	ID          int64  `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

// Login validates credentials and returns tokens
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	// Get client info
	deviceName := r.Header.Get("X-Device-Name")
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	user, tokens, err := h.authSvc.Login(r.Context(), req.Username, req.Password, deviceName, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			respondError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Log successful login
	if h.activitySvc != nil {
		h.activitySvc.Log(&user.ID, "login", ipAddress, nil, "")
	}

	respondJSON(w, http.StatusOK, loginResp{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresIn:    900, // 15 minutes
		User: userInfo{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			IsAdmin:     user.IsAdmin,
		},
	})
}

type refreshReq struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh rotates refresh token and returns new tokens
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	deviceName := r.Header.Get("X-Device-Name")
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	tokens, err := h.authSvc.Refresh(r.Context(), req.RefreshToken, deviceName, ipAddress, userAgent)
	if err != nil {
		if errors.Is(err, service.ErrInvalidToken) {
			respondError(w, http.StatusUnauthorized, "invalid or expired refresh token")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token":  tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    900,
	})
}

type changePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword allows user to change their password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req changePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		respondError(w, http.StatusBadRequest, "old_password and new_password are required")
		return
	}

	if err := h.authSvc.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			respondError(w, http.StatusForbidden, "current password is incorrect")
		case errors.Is(err, service.ErrInvalidPassword):
			respondError(w, http.StatusBadRequest, "new password must be at least 8 characters")
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "user not found")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "password changed - please login again"})
}

// Me returns current user info
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, isAdmin, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.authSvc.GetUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, userInfo{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		IsAdmin:     isAdmin,
	})
}

// Logout invalidates the refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	if err := h.authSvc.Logout(r.Context(), req.RefreshToken); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// ListSessions returns all active sessions for the current user
func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessions, err := h.authSvc.ListSessions(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, sessions)
}

// RevokeSession revokes a specific session
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	sessionID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid session id")
		return
	}

	if err := h.authSvc.RevokeSession(r.Context(), sessionID, userID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "session not found")
		case errors.Is(err, service.ErrNotOwner):
			respondError(w, http.StatusForbidden, "session does not belong to you")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "session revoked"})
}
