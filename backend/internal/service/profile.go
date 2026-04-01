package service

import (
	"context"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// ProfileService orchestrates profile, preferences, and user-media state.
type ProfileService struct {
	authSvc     *AuthService
	prefsRepo   *repository.UserPreferencesRepo
	userDataSvc *UserDataService
}

// NewProfileService creates a new profile service.
func NewProfileService(
	authSvc *AuthService,
	prefsRepo *repository.UserPreferencesRepo,
	userDataSvc *UserDataService,
) *ProfileService {
	return &ProfileService{
		authSvc:     authSvc,
		prefsRepo:   prefsRepo,
		userDataSvc: userDataSvc,
	}
}

// GetPreferences returns user preferences.
func (s *ProfileService) GetPreferences(ctx context.Context, userID int64) (*model.UserPreferences, error) {
	return s.prefsRepo.Get(ctx, userID)
}

// UpdatePreferences updates user preferences.
func (s *ProfileService) UpdatePreferences(ctx context.Context, prefs *model.UserPreferences) error {
	return s.prefsRepo.Update(ctx, prefs)
}

// GetProfile returns the current user's profile.
func (s *ProfileService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.authSvc.GetUser(ctx, userID)
}

// UpdateProfile updates the user's profile.
func (s *ProfileService) UpdateProfile(ctx context.Context, userID int64, displayName string) (*model.User, error) {
	if displayName == "" {
		return nil, ErrInvalidDisplayName
	}

	user, err := s.authSvc.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.DisplayName = displayName
	if err := s.authSvc.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetProgress returns user's watch progress for a media item.
func (s *ProfileService) GetProgress(ctx context.Context, userID, mediaID int64) (*model.UserData, error) {
	return s.userDataSvc.GetProgress(ctx, userID, mediaID)
}

// UpdateProgress updates watch progress.
func (s *ProfileService) UpdateProgress(
	ctx context.Context,
	userID, mediaID int64,
	position float64,
	completed bool,
) error {
	return s.userDataSvc.UpdateProgress(ctx, userID, mediaID, position, completed)
}

// ListFavorites returns user's favorite items.
func (s *ProfileService) ListFavorites(
	ctx context.Context,
	userID int64,
	limit, offset int,
) ([]*model.UserData, error) {
	return s.userDataSvc.ListFavorites(ctx, userID, limit, offset)
}

// ToggleFavorite toggles favorite status.
func (s *ProfileService) ToggleFavorite(ctx context.Context, userID, mediaID int64) (bool, error) {
	return s.userDataSvc.ToggleFavorite(ctx, userID, mediaID)
}

// ListRecentlyWatched returns recently watched items.
func (s *ProfileService) ListRecentlyWatched(ctx context.Context, userID int64, limit int) ([]*model.UserData, error) {
	return s.userDataSvc.ListRecentlyWatched(ctx, userID, limit)
}

// ContinueWatching returns in-progress items.
func (s *ProfileService) ContinueWatching(
	ctx context.Context,
	userID int64,
	limit int,
) ([]*model.ContinueWatchingItem, error) {
	return s.userDataSvc.ContinueWatching(ctx, userID, limit)
}

// NextUp returns the next unwatched episode for each series.
func (s *ProfileService) NextUp(ctx context.Context, userID int64, limit int) ([]*model.NextUpItem, error) {
	return s.userDataSvc.NextUp(ctx, userID, limit)
}

// DismissProgress resets progress for a media item.
func (s *ProfileService) DismissProgress(ctx context.Context, userID, mediaID int64) error {
	return s.userDataSvc.DismissProgress(ctx, userID, mediaID)
}
