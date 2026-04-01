package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/service"
)

// SettingsHandler handles admin settings endpoints.
type SettingsHandler struct {
	settingsSvc *service.SettingsService
}

// NewSettingsHandler creates a new settings handler.
func NewSettingsHandler(settingsSvc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsSvc: settingsSvc}
}

// openSubsResponse is the JSON shape for GET /api/admin/settings/opensubtitles.
type openSubsResponse struct {
	APIKey      string `json:"api_key"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"password_set"`
}

// openSubsRequest is the JSON shape for PUT /api/admin/settings/opensubtitles.
type openSubsRequest struct {
	APIKey   string `json:"api_key"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// GetOpenSubtitles returns the current OpenSubtitles configuration.
// GET /api/admin/settings/opensubtitles
func (h *SettingsHandler) GetOpenSubtitles(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetOpenSubtitles(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateOpenSubtitles saves OpenSubtitles credentials.
// PUT /api/admin/settings/opensubtitles
func (h *SettingsHandler) UpdateOpenSubtitles(w http.ResponseWriter, r *http.Request) {
	var req openSubsRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateOpenSubtitles(r.Context(), req.APIKey, req.Username, req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save OpenSubtitles settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// tmdbResponse is the JSON shape for GET /api/admin/settings/tmdb.
type tmdbResponse struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin"` // true if VELOX_TMDB_API_KEY env var is set
}

// tmdbRequest is the JSON shape for PUT /api/admin/settings/tmdb.
type tmdbRequest struct {
	APIKey string `json:"api_key"`
}

// GetTMDb returns the current TMDb configuration.
// GET /api/admin/settings/tmdb
func (h *SettingsHandler) GetTMDb(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetTMDb(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateTMDb saves the TMDb API key.
// PUT /api/admin/settings/tmdb
func (h *SettingsHandler) UpdateTMDb(w http.ResponseWriter, r *http.Request) {
	var req tmdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateTMDb(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// omdbResponse is the JSON shape for GET /api/admin/settings/omdb.
type omdbResponse struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin"` // true if VELOX_OMDB_API_KEY env var is set
}

// omdbRequest is the JSON shape for PUT /api/admin/settings/omdb.
type omdbRequest struct {
	APIKey string `json:"api_key"`
}

// GetOMDb returns the current OMDb configuration.
// GET /api/admin/settings/omdb
func (h *SettingsHandler) GetOMDb(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetOMDb(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateOMDb saves the OMDb API key.
// PUT /api/admin/settings/omdb
func (h *SettingsHandler) UpdateOMDb(w http.ResponseWriter, r *http.Request) {
	var req omdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateOMDb(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// tvdbResponse is the JSON shape for GET /api/admin/settings/tvdb.
type tvdbResponse struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin"` // true if VELOX_TVDB_API_KEY env var is set
}

// tvdbRequest is the JSON shape for PUT /api/admin/settings/tvdb.
type tvdbRequest struct {
	APIKey string `json:"api_key"`
}

// GetTVDB returns the current TheTVDB configuration.
// GET /api/admin/settings/tvdb
func (h *SettingsHandler) GetTVDB(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetTVDB(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateTVDB saves the TheTVDB API key.
// PUT /api/admin/settings/tvdb
func (h *SettingsHandler) UpdateTVDB(w http.ResponseWriter, r *http.Request) {
	var req tvdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateTVDB(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// playbackResponse is the JSON shape for GET /api/admin/settings/playback.
type playbackResponse struct {
	PlaybackMode string `json:"playback_mode"` // "auto" or "direct_play"
}

// playbackRequest is the JSON shape for PUT /api/admin/settings/playback.
type playbackRequest struct {
	PlaybackMode string `json:"playback_mode"`
}

// GetPlayback returns the current playback policy.
// GET /api/admin/settings/playback
func (h *SettingsHandler) GetPlayback(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetPlayback(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdatePlayback saves the playback policy.
// PUT /api/admin/settings/playback
func (h *SettingsHandler) UpdatePlayback(w http.ResponseWriter, r *http.Request) {
	var req playbackRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdatePlayback(r.Context(), req.PlaybackMode)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// fanartResponse is the JSON shape for GET /api/admin/settings/fanart.
type fanartResponse struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin"` // true if VELOX_FANART_API_KEY env var is set
}

// fanartRequest is the JSON shape for PUT /api/admin/settings/fanart.
type fanartRequest struct {
	APIKey string `json:"api_key"`
}

// GetFanart returns the current fanart.tv configuration.
// GET /api/admin/settings/fanart
func (h *SettingsHandler) GetFanart(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetFanart(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateFanart saves the fanart.tv API key.
// PUT /api/admin/settings/fanart
func (h *SettingsHandler) UpdateFanart(w http.ResponseWriter, r *http.Request) {
	var req fanartRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateFanart(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// autoSubResponse is the JSON shape for GET /api/admin/settings/auto-subtitles.
type autoSubResponse struct {
	Languages string `json:"languages"` // comma-separated: "en,vi"
}

// autoSubRequest is the JSON shape for PUT /api/admin/settings/auto-subtitles.
type autoSubRequest struct {
	Languages string `json:"languages"`
}

// GetAutoSubtitles returns the auto-download subtitle configuration.
// GET /api/admin/settings/auto-subtitles
func (h *SettingsHandler) GetAutoSubtitles(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetAutoSubtitles(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateAutoSubtitles saves the auto-download subtitle configuration.
// PUT /api/admin/settings/auto-subtitles
func (h *SettingsHandler) UpdateAutoSubtitles(w http.ResponseWriter, r *http.Request) {
	var req autoSubRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateAutoSubtitles(r.Context(), req.Languages)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save setting")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// subdlResponse is the JSON shape for GET /api/admin/settings/subdl.
type subdlResponse struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin"` // true if VELOX_SUBDL_API_KEY env var is set
}

// subdlRequest is the JSON shape for PUT /api/admin/settings/subdl.
type subdlRequest struct {
	APIKey string `json:"api_key"`
}

// GetSubdl returns the current Subdl configuration.
// GET /api/admin/settings/subdl
func (h *SettingsHandler) GetSubdl(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetSubdl(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateSubdl saves the Subdl API key.
// PUT /api/admin/settings/subdl
func (h *SettingsHandler) UpdateSubdl(w http.ResponseWriter, r *http.Request) {
	var req subdlRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	settings, err := h.settingsSvc.UpdateSubdl(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// deeplResponse is the JSON shape for GET /api/admin/settings/deepl.
type deeplResponse struct {
	APIKey string `json:"api_key"`
}

// GetDeepL returns the current DeepL configuration.
// GET /api/admin/settings/deepl
func (h *SettingsHandler) GetDeepL(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsSvc.GetDeepL(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}

// UpdateDeepL saves the DeepL API key.
// PUT /api/admin/settings/deepl
func (h *SettingsHandler) UpdateDeepL(w http.ResponseWriter, r *http.Request) {
	var req deeplResponse
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	settings, err := h.settingsSvc.UpdateDeepL(r.Context(), req.APIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, settings)
}
