package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

type ProfileHandler struct {
	authSvc     *service.AuthService
	prefsRepo   *repository.UserPreferencesRepo
	userDataSvc *service.UserDataService
}

func NewProfileHandler(authSvc *service.AuthService, prefsRepo *repository.UserPreferencesRepo, userDataSvc *service.UserDataService) *ProfileHandler {
	return &ProfileHandler{
		authSvc:     authSvc,
		prefsRepo:   prefsRepo,
		userDataSvc: userDataSvc,
	}
}

// GetPreferences returns user preferences
func (h *ProfileHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	prefs, err := h.prefsRepo.Get(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, prefs)
}

// UpdatePreferences updates user preferences
func (h *ProfileHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var prefs model.UserPreferences
	if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	prefs.UserID = userID

	if err := h.prefsRepo.Update(r.Context(), &prefs); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, prefs)
}

// UpdateProfile updates user profile (display_name)
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.DisplayName == "" {
		respondError(w, http.StatusBadRequest, "display_name is required")
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

	user.DisplayName = req.DisplayName
	if err := h.authSvc.UpdateUser(r.Context(), user); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, user)
}

// GetProfile returns the current user's profile
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
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

	respondJSON(w, http.StatusOK, user)
}

// GetProgress returns user's watch progress for a media item
func (h *ProfileHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseID(r, "mediaId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	progress, err := h.userDataSvc.GetProgress(r.Context(), userID, mediaID)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			respondJSON(w, http.StatusOK, nil) // No progress yet
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, progress)
}

// UpdateProgress updates watch progress
func (h *ProfileHandler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseID(r, "mediaId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	var req struct {
		Position  float64 `json:"position"`
		Completed bool    `json:"completed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.userDataSvc.UpdateProgress(r.Context(), userID, mediaID, req.Position, req.Completed); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "progress updated"})
}

// ListFavorites returns user's favorite items
func (h *ProfileHandler) ListFavorites(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)

	favorites, err := h.userDataSvc.ListFavorites(r.Context(), userID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, favorites)
}

// ToggleFavorite toggles favorite status
func (h *ProfileHandler) ToggleFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseID(r, "mediaId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	isFavorite, err := h.userDataSvc.ToggleFavorite(r.Context(), userID, mediaID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]bool{"is_favorite": isFavorite})
}

// ListRecentlyWatched returns recently watched items
func (h *ProfileHandler) ListRecentlyWatched(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQuery(r, "limit", 20)

	items, err := h.userDataSvc.ListRecentlyWatched(r.Context(), userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// ContinueWatching returns in-progress items
func (h *ProfileHandler) ContinueWatching(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQuery(r, "limit", 20)

	items, err := h.userDataSvc.ContinueWatching(r.Context(), userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// NextUp returns the next unwatched episode for each series
func (h *ProfileHandler) NextUp(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := parseIntQuery(r, "limit", 20)

	items, err := h.userDataSvc.NextUp(r.Context(), userID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, items)
}

// DismissProgress resets progress for a media item (preserves favorite/rating)
func (h *ProfileHandler) DismissProgress(w http.ResponseWriter, r *http.Request) {
	userID, _, ok := auth.UserFromContext(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	mediaID, err := parseID(r, "mediaId")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid media id")
		return
	}

	if err := h.userDataSvc.DismissProgress(r.Context(), userID, mediaID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "progress dismissed"})
}
