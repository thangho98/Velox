package service

import (
	"context"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// PlaybackConfigService provides preference and playback-policy config for playback decisions.
type PlaybackConfigService struct {
	prefsRepo       *repository.UserPreferencesRepo
	appSettingsRepo *repository.AppSettingsRepo
}

// NewPlaybackConfigService creates a new playback config service.
func NewPlaybackConfigService(
	prefsRepo *repository.UserPreferencesRepo,
	appSettingsRepo *repository.AppSettingsRepo,
) *PlaybackConfigService {
	return &PlaybackConfigService{
		prefsRepo:       prefsRepo,
		appSettingsRepo: appSettingsRepo,
	}
}

// GetUserPreferences returns playback-related user preferences.
func (s *PlaybackConfigService) GetUserPreferences(ctx context.Context, userID int64) (*model.UserPreferences, error) {
	if s == nil || s.prefsRepo == nil {
		return &model.UserPreferences{
			UserID:              userID,
			MaxStreamingQuality: "auto",
		}, nil
	}

	return s.prefsRepo.Get(ctx, userID)
}

// GetPlaybackMode returns the admin-configured playback mode.
func (s *PlaybackConfigService) GetPlaybackMode(ctx context.Context) (string, error) {
	if s == nil || s.appSettingsRepo == nil {
		return "auto", nil
	}

	val, err := s.appSettingsRepo.Get(ctx, model.SettingPlaybackMode)
	if err != nil {
		return "", err
	}
	if val == "" {
		val = "auto"
	}
	return val, nil
}
