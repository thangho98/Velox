package handler

import (
	"net/http"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SettingsHandler handles admin settings endpoints.
type SettingsHandler struct {
	repo             *repository.AppSettingsRepo
	hasBuiltinTMDb   bool
	hasBuiltinOMDb   bool
	hasBuiltinTVDB   bool
	hasBuiltinFanart bool
	hasBuiltinSubdl  bool
}

// NewSettingsHandler creates a new settings handler.
// builtinKeys indicates which providers have env-based default keys configured.
func NewSettingsHandler(repo *repository.AppSettingsRepo, builtinKeys map[string]bool) *SettingsHandler {
	return &SettingsHandler{
		repo:             repo,
		hasBuiltinTMDb:   builtinKeys["tmdb"],
		hasBuiltinOMDb:   builtinKeys["omdb"],
		hasBuiltinTVDB:   builtinKeys["tvdb"],
		hasBuiltinFanart: builtinKeys["fanart"],
		hasBuiltinSubdl:  builtinKeys["subdl"],
	}
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
	ctx := r.Context()

	vals, err := h.repo.GetMulti(ctx,
		model.SettingOpenSubsAPIKey,
		model.SettingOpenSubsUsername,
		model.SettingOpenSubsPassword,
	)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}

	respondJSON(w, http.StatusOK, openSubsResponse{
		APIKey:      vals[model.SettingOpenSubsAPIKey],
		Username:    vals[model.SettingOpenSubsUsername],
		PasswordSet: vals[model.SettingOpenSubsPassword] != "",
	})
}

// UpdateOpenSubtitles saves OpenSubtitles credentials.
// PUT /api/admin/settings/opensubtitles
func (h *SettingsHandler) UpdateOpenSubtitles(w http.ResponseWriter, r *http.Request) {
	var req openSubsRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	ctx := r.Context()

	if err := h.repo.Set(ctx, model.SettingOpenSubsAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	if err := h.repo.Set(ctx, model.SettingOpenSubsUsername, req.Username); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save username")
		return
	}
	if err := h.repo.Set(ctx, model.SettingOpenSubsPassword, req.Password); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save password")
		return
	}

	respondJSON(w, http.StatusOK, openSubsResponse{
		APIKey:      req.APIKey,
		Username:    req.Username,
		PasswordSet: req.Password != "",
	})
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
	val, err := h.repo.Get(r.Context(), model.SettingTMDbAPIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, tmdbResponse{APIKey: val, HasBuiltin: h.hasBuiltinTMDb})
}

// UpdateTMDb saves the TMDb API key.
// PUT /api/admin/settings/tmdb
func (h *SettingsHandler) UpdateTMDb(w http.ResponseWriter, r *http.Request) {
	var req tmdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingTMDbAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}

	respondJSON(w, http.StatusOK, tmdbResponse{APIKey: req.APIKey})
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
	val, err := h.repo.Get(r.Context(), model.SettingOMDbAPIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, omdbResponse{APIKey: val, HasBuiltin: h.hasBuiltinOMDb})
}

// UpdateOMDb saves the OMDb API key.
// PUT /api/admin/settings/omdb
func (h *SettingsHandler) UpdateOMDb(w http.ResponseWriter, r *http.Request) {
	var req omdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingOMDbAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}

	respondJSON(w, http.StatusOK, omdbResponse{APIKey: req.APIKey})
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
	val, err := h.repo.Get(r.Context(), model.SettingTVDBAPIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, tvdbResponse{APIKey: val, HasBuiltin: h.hasBuiltinTVDB})
}

// UpdateTVDB saves the TheTVDB API key.
// PUT /api/admin/settings/tvdb
func (h *SettingsHandler) UpdateTVDB(w http.ResponseWriter, r *http.Request) {
	var req tvdbRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingTVDBAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}

	respondJSON(w, http.StatusOK, tvdbResponse{APIKey: req.APIKey})
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
	val, err := h.repo.Get(r.Context(), model.SettingPlaybackMode)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	if val == "" {
		val = "auto"
	}
	respondJSON(w, http.StatusOK, playbackResponse{PlaybackMode: val})
}

// UpdatePlayback saves the playback policy.
// PUT /api/admin/settings/playback
func (h *SettingsHandler) UpdatePlayback(w http.ResponseWriter, r *http.Request) {
	var req playbackRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if req.PlaybackMode != "auto" && req.PlaybackMode != "direct_play" {
		respondError(w, http.StatusBadRequest, "playback_mode must be 'auto' or 'direct_play'")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingPlaybackMode, req.PlaybackMode); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save setting")
		return
	}

	respondJSON(w, http.StatusOK, playbackResponse{PlaybackMode: req.PlaybackMode})
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
	val, err := h.repo.Get(r.Context(), model.SettingFanartAPIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, fanartResponse{APIKey: val, HasBuiltin: h.hasBuiltinFanart})
}

// UpdateFanart saves the fanart.tv API key.
// PUT /api/admin/settings/fanart
func (h *SettingsHandler) UpdateFanart(w http.ResponseWriter, r *http.Request) {
	var req fanartRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingFanartAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}

	respondJSON(w, http.StatusOK, fanartResponse{APIKey: req.APIKey})
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
	val, err := h.repo.Get(r.Context(), model.SettingAutoSubLanguages)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, autoSubResponse{Languages: val})
}

// UpdateAutoSubtitles saves the auto-download subtitle configuration.
// PUT /api/admin/settings/auto-subtitles
func (h *SettingsHandler) UpdateAutoSubtitles(w http.ResponseWriter, r *http.Request) {
	var req autoSubRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingAutoSubLanguages, req.Languages); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save setting")
		return
	}

	respondJSON(w, http.StatusOK, autoSubResponse{Languages: req.Languages})
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
	val, err := h.repo.Get(r.Context(), model.SettingSubdlAPIKey)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	respondJSON(w, http.StatusOK, subdlResponse{APIKey: val, HasBuiltin: h.hasBuiltinSubdl})
}

// UpdateSubdl saves the Subdl API key.
// PUT /api/admin/settings/subdl
func (h *SettingsHandler) UpdateSubdl(w http.ResponseWriter, r *http.Request) {
	var req subdlRequest
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := h.repo.Set(r.Context(), model.SettingSubdlAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}

	respondJSON(w, http.StatusOK, subdlResponse{APIKey: req.APIKey})
}

// deeplResponse is the JSON shape for GET /api/admin/settings/deepl.
type deeplResponse struct {
	APIKey string `json:"api_key"`
}

// GetDeepL returns the current DeepL configuration.
// GET /api/admin/settings/deepl
func (h *SettingsHandler) GetDeepL(w http.ResponseWriter, r *http.Request) {
	val, err := h.repo.Get(r.Context(), model.SettingDeepLAPIKey)
	if err != nil {
		val = ""
	}
	respondJSON(w, http.StatusOK, deeplResponse{APIKey: val})
}

// UpdateDeepL saves the DeepL API key.
// PUT /api/admin/settings/deepl
func (h *SettingsHandler) UpdateDeepL(w http.ResponseWriter, r *http.Request) {
	var req deeplResponse
	if err := parseJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := h.repo.Set(r.Context(), model.SettingDeepLAPIKey, req.APIKey); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save api_key")
		return
	}
	respondJSON(w, http.StatusOK, deeplResponse{APIKey: req.APIKey})
}
