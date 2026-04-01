package service

import (
	"context"
	"fmt"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// SettingsService orchestrates admin-facing application settings.
type SettingsService struct {
	repo             *repository.AppSettingsRepo
	hasBuiltinTMDb   bool
	hasBuiltinOMDb   bool
	hasBuiltinTVDB   bool
	hasBuiltinFanart bool
	hasBuiltinSubdl  bool
}

// OpenSubtitlesSettings is the admin-facing OpenSubtitles config payload.
type OpenSubtitlesSettings struct {
	APIKey      string `json:"api_key"`
	Username    string `json:"username"`
	PasswordSet bool   `json:"password_set"`
}

// APIKeySettings is the shared payload for API-key-backed providers.
type APIKeySettings struct {
	APIKey     string `json:"api_key"`
	HasBuiltin bool   `json:"has_builtin,omitempty"`
}

// AutoSubtitleSettings is the auto subtitle download payload.
type AutoSubtitleSettings struct {
	Languages string `json:"languages"`
}

// PlaybackSettings is the admin playback-policy payload.
type PlaybackSettings struct {
	PlaybackMode string `json:"playback_mode"`
}

// NewSettingsService creates a new settings service.
func NewSettingsService(repo *repository.AppSettingsRepo, builtinKeys map[string]bool) *SettingsService {
	return &SettingsService{
		repo:             repo,
		hasBuiltinTMDb:   builtinKeys["tmdb"],
		hasBuiltinOMDb:   builtinKeys["omdb"],
		hasBuiltinTVDB:   builtinKeys["tvdb"],
		hasBuiltinFanart: builtinKeys["fanart"],
		hasBuiltinSubdl:  builtinKeys["subdl"],
	}
}

func (s *SettingsService) getProviderAPIKey(
	ctx context.Context,
	settingKey string,
	hasBuiltin bool,
) (*APIKeySettings, error) {
	val, err := s.repo.Get(ctx, settingKey)
	if err != nil {
		return nil, err
	}

	return &APIKeySettings{
		APIKey:     val,
		HasBuiltin: hasBuiltin,
	}, nil
}

func (s *SettingsService) updateProviderAPIKey(
	ctx context.Context,
	settingKey string,
	apiKey string,
	hasBuiltin bool,
) (*APIKeySettings, error) {
	if err := s.repo.Set(ctx, settingKey, apiKey); err != nil {
		return nil, err
	}

	return &APIKeySettings{
		APIKey:     apiKey,
		HasBuiltin: hasBuiltin,
	}, nil
}

// GetOpenSubtitles returns the current OpenSubtitles configuration.
func (s *SettingsService) GetOpenSubtitles(ctx context.Context) (*OpenSubtitlesSettings, error) {
	vals, err := s.repo.GetMulti(
		ctx,
		model.SettingOpenSubsAPIKey,
		model.SettingOpenSubsUsername,
		model.SettingOpenSubsPassword,
	)
	if err != nil {
		return nil, err
	}

	return &OpenSubtitlesSettings{
		APIKey:      vals[model.SettingOpenSubsAPIKey],
		Username:    vals[model.SettingOpenSubsUsername],
		PasswordSet: vals[model.SettingOpenSubsPassword] != "",
	}, nil
}

// UpdateOpenSubtitles saves OpenSubtitles credentials.
func (s *SettingsService) UpdateOpenSubtitles(
	ctx context.Context,
	apiKey, username, password string,
) (*OpenSubtitlesSettings, error) {
	if err := s.repo.Set(ctx, model.SettingOpenSubsAPIKey, apiKey); err != nil {
		return nil, err
	}
	if err := s.repo.Set(ctx, model.SettingOpenSubsUsername, username); err != nil {
		return nil, err
	}
	if err := s.repo.Set(ctx, model.SettingOpenSubsPassword, password); err != nil {
		return nil, err
	}

	return &OpenSubtitlesSettings{
		APIKey:      apiKey,
		Username:    username,
		PasswordSet: password != "",
	}, nil
}

// GetTMDb returns the current TMDb configuration.
func (s *SettingsService) GetTMDb(ctx context.Context) (*APIKeySettings, error) {
	return s.getProviderAPIKey(ctx, model.SettingTMDbAPIKey, s.hasBuiltinTMDb)
}

// UpdateTMDb saves the TMDb API key.
func (s *SettingsService) UpdateTMDb(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	return s.updateProviderAPIKey(ctx, model.SettingTMDbAPIKey, apiKey, s.hasBuiltinTMDb)
}

// GetOMDb returns the current OMDb configuration.
func (s *SettingsService) GetOMDb(ctx context.Context) (*APIKeySettings, error) {
	return s.getProviderAPIKey(ctx, model.SettingOMDbAPIKey, s.hasBuiltinOMDb)
}

// UpdateOMDb saves the OMDb API key.
func (s *SettingsService) UpdateOMDb(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	return s.updateProviderAPIKey(ctx, model.SettingOMDbAPIKey, apiKey, s.hasBuiltinOMDb)
}

// GetTVDB returns the current TheTVDB configuration.
func (s *SettingsService) GetTVDB(ctx context.Context) (*APIKeySettings, error) {
	return s.getProviderAPIKey(ctx, model.SettingTVDBAPIKey, s.hasBuiltinTVDB)
}

// UpdateTVDB saves the TheTVDB API key.
func (s *SettingsService) UpdateTVDB(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	return s.updateProviderAPIKey(ctx, model.SettingTVDBAPIKey, apiKey, s.hasBuiltinTVDB)
}

// GetPlayback returns the current playback policy.
func (s *SettingsService) GetPlayback(ctx context.Context) (*PlaybackSettings, error) {
	val, err := s.repo.Get(ctx, model.SettingPlaybackMode)
	if err != nil {
		return nil, err
	}
	if val == "" {
		val = "auto"
	}

	return &PlaybackSettings{PlaybackMode: val}, nil
}

// UpdatePlayback saves the playback policy.
func (s *SettingsService) UpdatePlayback(ctx context.Context, playbackMode string) (*PlaybackSettings, error) {
	if playbackMode != "auto" && playbackMode != "direct_play" {
		return nil, fmt.Errorf("playback_mode must be 'auto' or 'direct_play'")
	}

	if err := s.repo.Set(ctx, model.SettingPlaybackMode, playbackMode); err != nil {
		return nil, err
	}

	return &PlaybackSettings{PlaybackMode: playbackMode}, nil
}

// GetFanart returns the current fanart.tv configuration.
func (s *SettingsService) GetFanart(ctx context.Context) (*APIKeySettings, error) {
	return s.getProviderAPIKey(ctx, model.SettingFanartAPIKey, s.hasBuiltinFanart)
}

// UpdateFanart saves the fanart.tv API key.
func (s *SettingsService) UpdateFanart(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	return s.updateProviderAPIKey(ctx, model.SettingFanartAPIKey, apiKey, s.hasBuiltinFanart)
}

// GetAutoSubtitles returns the auto-download subtitle configuration.
func (s *SettingsService) GetAutoSubtitles(ctx context.Context) (*AutoSubtitleSettings, error) {
	val, err := s.repo.Get(ctx, model.SettingAutoSubLanguages)
	if err != nil {
		return nil, err
	}

	return &AutoSubtitleSettings{Languages: val}, nil
}

// UpdateAutoSubtitles saves the auto-download subtitle configuration.
func (s *SettingsService) UpdateAutoSubtitles(
	ctx context.Context,
	languages string,
) (*AutoSubtitleSettings, error) {
	if err := s.repo.Set(ctx, model.SettingAutoSubLanguages, languages); err != nil {
		return nil, err
	}

	return &AutoSubtitleSettings{Languages: languages}, nil
}

// GetSubdl returns the current Subdl configuration.
func (s *SettingsService) GetSubdl(ctx context.Context) (*APIKeySettings, error) {
	return s.getProviderAPIKey(ctx, model.SettingSubdlAPIKey, s.hasBuiltinSubdl)
}

// UpdateSubdl saves the Subdl API key.
func (s *SettingsService) UpdateSubdl(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	return s.updateProviderAPIKey(ctx, model.SettingSubdlAPIKey, apiKey, s.hasBuiltinSubdl)
}

// GetDeepL returns the current DeepL configuration.
func (s *SettingsService) GetDeepL(ctx context.Context) (*APIKeySettings, error) {
	val, err := s.repo.Get(ctx, model.SettingDeepLAPIKey)
	if err != nil {
		return nil, fmt.Errorf("getting DeepL API key: %w", err)
	}

	return &APIKeySettings{APIKey: val}, nil
}

// UpdateDeepL saves the DeepL API key.
func (s *SettingsService) UpdateDeepL(ctx context.Context, apiKey string) (*APIKeySettings, error) {
	if err := s.repo.Set(ctx, model.SettingDeepLAPIKey, apiKey); err != nil {
		return nil, err
	}

	return &APIKeySettings{APIKey: apiKey}, nil
}
