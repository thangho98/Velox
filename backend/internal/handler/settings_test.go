package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thawng/velox/internal/service"
)

// respWrapper is a generic wrapper for API responses.
type respWrapper[T any] struct {
	Data T `json:"data"`
}

func TestGetOpenSubtitles_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/opensubtitles", nil)
	w := httptest.NewRecorder()

	h.GetOpenSubtitles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.OpenSubtitlesSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "test-api-key", wrapper.Data.APIKey)
	assert.Equal(t, "testuser", wrapper.Data.Username)
	assert.True(t, wrapper.Data.PasswordSet)
}

func TestUpdateOpenSubtitles_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{
		"api_key":  "new-key",
		"username": "newuser",
		"password": "newpass",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/opensubtitles", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateOpenSubtitles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.OpenSubtitlesSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "new-key", wrapper.Data.APIKey)
	assert.Equal(t, "newuser", wrapper.Data.Username)
	assert.True(t, wrapper.Data.PasswordSet)
}

func TestUpdateOpenSubtitles_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/opensubtitles", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateOpenSubtitles(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTMDb_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/tmdb", nil)
	w := httptest.NewRecorder()

	h.GetTMDb(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "tmdb-key", wrapper.Data.APIKey)
	assert.True(t, wrapper.Data.HasBuiltin)
}

func TestUpdateTMDb_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-tmdb-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/tmdb", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateTMDb(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "new-tmdb-key", wrapper.Data.APIKey)
}

func TestUpdateTMDb_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/tmdb", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateTMDb(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAniList_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/anilist", nil)
	w := httptest.NewRecorder()

	h.GetAniList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "anilist-key", wrapper.Data.APIKey)
}

func TestUpdateAniList_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-anilist-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/anilist", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAniList(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "new-anilist-key", wrapper.Data.APIKey)
}

func TestGetOMDb_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/omdb", nil)
	w := httptest.NewRecorder()

	h.GetOMDb(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "omdb-key", wrapper.Data.APIKey)
}

func TestUpdateOMDb_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-omdb-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/omdb", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateOMDb(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTVDB_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/tvdb", nil)
	w := httptest.NewRecorder()

	h.GetTVDB(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "tvdb-key", wrapper.Data.APIKey)
}

func TestUpdateTVDB_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-tvdb-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/tvdb", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateTVDB(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetPlayback_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/playback", nil)
	w := httptest.NewRecorder()

	h.GetPlayback(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.PlaybackSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "auto", wrapper.Data.PlaybackMode)
}

func TestUpdatePlayback_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"playback_mode": "direct_play"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/playback", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdatePlayback(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.PlaybackSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "direct_play", wrapper.Data.PlaybackMode)
}

func TestUpdatePlayback_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/playback", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdatePlayback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetFanart_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/fanart", nil)
	w := httptest.NewRecorder()

	h.GetFanart(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "fanart-key", wrapper.Data.APIKey)
}

func TestUpdateFanart_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-fanart-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/fanart", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateFanart(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetAutoSubtitles_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/auto-subtitles", nil)
	w := httptest.NewRecorder()

	h.GetAutoSubtitles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AutoSubtitleSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "en,vi", wrapper.Data.Languages)
}

func TestUpdateAutoSubtitles_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"languages": "en,vi,ja"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/auto-subtitles", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAutoSubtitles(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AutoSubtitleSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "en,vi,ja", wrapper.Data.Languages)
}

func TestGetSubdl_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/subdl", nil)
	w := httptest.NewRecorder()

	h.GetSubdl(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "subdl-key", wrapper.Data.APIKey)
}

func TestUpdateSubdl_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-subdl-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/subdl", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateSubdl(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeepL_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/deepl", nil)
	w := httptest.NewRecorder()

	h.GetDeepL(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "deepl-key", wrapper.Data.APIKey)
}

func TestUpdateDeepL_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{"api_key": "new-deepl-key"}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/deepl", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateDeepL(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.APIKeySettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "new-deepl-key", wrapper.Data.APIKey)
}

func TestUpdateDeepL_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/deepl", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateDeepL(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAITranslation_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/ai-translation", nil)
	w := httptest.NewRecorder()

	h.GetAITranslation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AITranslationSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "openai_compatible", wrapper.Data.Provider)
	assert.Equal(t, "test-key", wrapper.Data.APIKey)
	assert.NotEmpty(t, wrapper.Data.BaseURL)
	assert.Equal(t, "gpt-4", wrapper.Data.Model)
}

func TestUpdateAITranslation_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]string{
		"provider": "gemini_compatible",
		"api_key":  "new-gemini-key",
		"base_url": "https://generativelanguage.googleapis.com/v1beta",
		"model":    "gemini-pro",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/ai-translation", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAITranslation(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AITranslationSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.Equal(t, "gemini_compatible", wrapper.Data.Provider)
	assert.Equal(t, "new-gemini-key", wrapper.Data.APIKey)
	assert.Equal(t, "gemini-pro", wrapper.Data.Model)
}

func TestUpdateAITranslation_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/ai-translation", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAITranslation(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAutoTranslate_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/settings/auto-translate", nil)
	w := httptest.NewRecorder()

	h.GetAutoTranslate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AutoTranslateSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.True(t, wrapper.Data.Enabled)
	assert.Equal(t, "en,vi", wrapper.Data.Languages)
}

func TestUpdateAutoTranslate_Success(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	body := map[string]interface{}{
		"enabled":   false,
		"languages": "en",
	}
	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/auto-translate", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAutoTranslate(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var wrapper respWrapper[service.AutoTranslateSettings]
	err := json.Unmarshal(w.Body.Bytes(), &wrapper)
	require.NoError(t, err)
	assert.False(t, wrapper.Data.Enabled)
	assert.Equal(t, "en", wrapper.Data.Languages)
}

func TestUpdateAutoTranslate_InvalidJSON(t *testing.T) {
	t.Parallel()

	svc := newMockSettingsService()
	h := NewSettingsHandler(svc)

	req := httptest.NewRequest(http.MethodPut, "/api/admin/settings/auto-translate", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.UpdateAutoTranslate(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// mockSettingsService provides a test double for SettingsService.
type mockSettingsService struct {
	openSubs      *service.OpenSubtitlesSettings
	tmdb          *service.APIKeySettings
	anilist       *service.APIKeySettings
	omdb          *service.APIKeySettings
	tvdb          *service.APIKeySettings
	playback      *service.PlaybackSettings
	fanart        *service.APIKeySettings
	autoSub       *service.AutoSubtitleSettings
	subdl         *service.APIKeySettings
	deepl         *service.APIKeySettings
	aiTranslation *service.AITranslationSettings
	autoTranslate *service.AutoTranslateSettings
}

func newMockSettingsService() *mockSettingsService {
	return &mockSettingsService{
		openSubs: &service.OpenSubtitlesSettings{
			APIKey:      "test-api-key",
			Username:    "testuser",
			PasswordSet: true,
		},
		tmdb: &service.APIKeySettings{
			APIKey:     "tmdb-key",
			HasBuiltin: true,
		},
		anilist: &service.APIKeySettings{
			APIKey:     "anilist-key",
			HasBuiltin: false,
		},
		omdb: &service.APIKeySettings{
			APIKey:     "omdb-key",
			HasBuiltin: false,
		},
		tvdb: &service.APIKeySettings{
			APIKey:     "tvdb-key",
			HasBuiltin: false,
		},
		playback: &service.PlaybackSettings{
			PlaybackMode: "auto",
		},
		fanart: &service.APIKeySettings{
			APIKey:     "fanart-key",
			HasBuiltin: false,
		},
		autoSub: &service.AutoSubtitleSettings{
			Languages: "en,vi",
		},
		subdl: &service.APIKeySettings{
			APIKey:     "subdl-key",
			HasBuiltin: false,
		},
		deepl: &service.APIKeySettings{
			APIKey: "deepl-key",
		},
		aiTranslation: &service.AITranslationSettings{
			Provider: "openai_compatible",
			APIKey:   "test-key",
			BaseURL:  "https://api.openai.com/v1",
			Model:    "gpt-4",
		},
		autoTranslate: &service.AutoTranslateSettings{
			Enabled:   true,
			Languages: "en,vi",
		},
	}
}

func (m *mockSettingsService) GetOpenSubtitles(_ context.Context) (*service.OpenSubtitlesSettings, error) {
	return m.openSubs, nil
}

func (m *mockSettingsService) UpdateOpenSubtitles(_ context.Context, apiKey, username, password string) (*service.OpenSubtitlesSettings, error) {
	m.openSubs = &service.OpenSubtitlesSettings{
		APIKey:      apiKey,
		Username:    username,
		PasswordSet: password != "",
	}
	return m.openSubs, nil
}

func (m *mockSettingsService) GetTMDb(_ context.Context) (*service.APIKeySettings, error) {
	return m.tmdb, nil
}

func (m *mockSettingsService) UpdateTMDb(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.tmdb.APIKey = apiKey
	return m.tmdb, nil
}

func (m *mockSettingsService) GetAniList(_ context.Context) (*service.APIKeySettings, error) {
	return m.anilist, nil
}

func (m *mockSettingsService) UpdateAniList(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.anilist.APIKey = apiKey
	return m.anilist, nil
}

func (m *mockSettingsService) GetOMDb(_ context.Context) (*service.APIKeySettings, error) {
	return m.omdb, nil
}

func (m *mockSettingsService) UpdateOMDb(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.omdb.APIKey = apiKey
	return m.omdb, nil
}

func (m *mockSettingsService) GetTVDB(_ context.Context) (*service.APIKeySettings, error) {
	return m.tvdb, nil
}

func (m *mockSettingsService) UpdateTVDB(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.tvdb.APIKey = apiKey
	return m.tvdb, nil
}

func (m *mockSettingsService) GetPlayback(_ context.Context) (*service.PlaybackSettings, error) {
	return m.playback, nil
}

func (m *mockSettingsService) UpdatePlayback(_ context.Context, mode string) (*service.PlaybackSettings, error) {
	m.playback = &service.PlaybackSettings{PlaybackMode: mode}
	return m.playback, nil
}

func (m *mockSettingsService) GetFanart(_ context.Context) (*service.APIKeySettings, error) {
	return m.fanart, nil
}

func (m *mockSettingsService) UpdateFanart(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.fanart.APIKey = apiKey
	return m.fanart, nil
}

func (m *mockSettingsService) GetAutoSubtitles(_ context.Context) (*service.AutoSubtitleSettings, error) {
	return m.autoSub, nil
}

func (m *mockSettingsService) UpdateAutoSubtitles(_ context.Context, languages string) (*service.AutoSubtitleSettings, error) {
	m.autoSub = &service.AutoSubtitleSettings{Languages: languages}
	return m.autoSub, nil
}

func (m *mockSettingsService) GetSubdl(_ context.Context) (*service.APIKeySettings, error) {
	return m.subdl, nil
}

func (m *mockSettingsService) UpdateSubdl(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.subdl.APIKey = apiKey
	return m.subdl, nil
}

func (m *mockSettingsService) GetDeepL(_ context.Context) (*service.APIKeySettings, error) {
	return m.deepl, nil
}

func (m *mockSettingsService) UpdateDeepL(_ context.Context, apiKey string) (*service.APIKeySettings, error) {
	m.deepl.APIKey = apiKey
	return m.deepl, nil
}

func (m *mockSettingsService) GetAITranslation(_ context.Context) (*service.AITranslationSettings, error) {
	return m.aiTranslation, nil
}

func (m *mockSettingsService) UpdateAITranslation(_ context.Context, provider, apiKey, baseURL, model string) (*service.AITranslationSettings, error) {
	m.aiTranslation = &service.AITranslationSettings{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
	}
	return m.aiTranslation, nil
}

func (m *mockSettingsService) GetAutoTranslate(_ context.Context) (*service.AutoTranslateSettings, error) {
	return m.autoTranslate, nil
}

func (m *mockSettingsService) UpdateAutoTranslate(_ context.Context, enabled bool, languages string) (*service.AutoTranslateSettings, error) {
	m.autoTranslate = &service.AutoTranslateSettings{Enabled: enabled, Languages: languages}
	return m.autoTranslate, nil
}
