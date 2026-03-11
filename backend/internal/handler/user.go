package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/service"
)

type UserHandler struct {
	authSvc *service.AuthService
}

func NewUserHandler(authSvc *service.AuthService) *UserHandler {
	return &UserHandler{authSvc: authSvc}
}

// List returns all users (admin only)
func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.authSvc.ListUsers(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// json:"-" tag on PasswordHash already handles hiding it
	respondJSON(w, http.StatusOK, users)
}

type createUserReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	IsAdmin     bool   `json:"is_admin"`
}

// Create creates a new user (admin only)
func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" || req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "username, password and display_name are required")
		return
	}

	user, err := h.authSvc.CreateUser(r.Context(), req.Username, req.Password, req.DisplayName, req.IsAdmin)
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

	respondJSON(w, http.StatusCreated, user)
}

type updateUserReq struct {
	DisplayName string `json:"display_name"`
	IsAdmin     *bool  `json:"is_admin,omitempty"`
}

// Update updates a user (admin only)
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := h.authSvc.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	var req updateUserReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.IsAdmin != nil {
		user.IsAdmin = *req.IsAdmin
	}

	if err := h.authSvc.UpdateUser(r.Context(), user); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrLastAdmin):
			respondError(w, http.StatusBadRequest, "cannot remove the last admin")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// Delete deletes a user (admin only)
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	// Get current user from context
	currentUserID, _, _ := auth.UserFromContext(r.Context())

	if err := h.authSvc.DeleteUser(r.Context(), id, currentUserID); err != nil {
		switch {
		case errors.Is(err, service.ErrNotFound):
			respondError(w, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrLastAdmin):
			respondError(w, http.StatusBadRequest, "cannot remove the last admin")
		case errors.Is(err, service.ErrDeleteSelf):
			respondError(w, http.StatusBadRequest, "cannot delete your own account")
		default:
			respondError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type setLibraryAccessReq struct {
	LibraryIDs []int64 `json:"library_ids"`
}

// SetLibraryAccess sets which libraries a user can access (admin only)
func (h *UserHandler) SetLibraryAccess(w http.ResponseWriter, r *http.Request) {
	userID, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req setLibraryAccessReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.authSvc.SetLibraryAccess(r.Context(), userID, req.LibraryIDs); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondError(w, http.StatusNotFound, "user not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "library access updated"})
}
