package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

// UserDataService handles user-specific data (progress, favorites, ratings)
type UserDataService struct {
	userDataRepo *repository.UserDataRepo
}

// NewUserDataService creates a new user data service
func NewUserDataService(userDataRepo *repository.UserDataRepo) *UserDataService {
	return &UserDataService{userDataRepo: userDataRepo}
}

// GetProgress retrieves watch progress for a media item
func (s *UserDataService) GetProgress(ctx context.Context, userID, mediaID int64) (*model.UserData, error) {
	progress, err := s.userDataRepo.GetProgress(ctx, userID, mediaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("getting progress: %w", err)
	}
	return progress, nil
}

// UpdateProgress creates or updates watch progress
func (s *UserDataService) UpdateProgress(ctx context.Context, userID, mediaID int64, position float64, completed bool) error {
	if position < 0 {
		return fmt.Errorf("position cannot be negative")
	}
	if err := s.userDataRepo.UpsertProgress(ctx, userID, mediaID, position, completed); err != nil {
		return fmt.Errorf("updating progress: %w", err)
	}
	return nil
}

// ListFavorites returns user's favorite items
func (s *UserDataService) ListFavorites(ctx context.Context, userID int64, limit, offset int) ([]*model.UserData, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	favorites, err := s.userDataRepo.ListFavorites(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("listing favorites: %w", err)
	}
	return favorites, nil
}

// ToggleFavorite toggles favorite status for a media item
func (s *UserDataService) ToggleFavorite(ctx context.Context, userID, mediaID int64) (bool, error) {
	isFavorite, err := s.userDataRepo.ToggleFavorite(ctx, userID, mediaID)
	if err != nil {
		return false, fmt.Errorf("toggling favorite: %w", err)
	}
	return isFavorite, nil
}

// ListRecentlyWatched returns recently watched items
func (s *UserDataService) ListRecentlyWatched(ctx context.Context, userID int64, limit int) ([]*model.UserData, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	items, err := s.userDataRepo.ListRecentlyWatched(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing recently watched: %w", err)
	}
	return items, nil
}

// IsFavorite checks if a media item is in user's favorites
func (s *UserDataService) IsFavorite(ctx context.Context, userID, mediaID int64) bool {
	data, err := s.userDataRepo.GetProgress(ctx, userID, mediaID)
	if err != nil {
		return false
	}
	return data.IsFavorite
}

// SetRating sets user rating for a media item
func (s *UserDataService) SetRating(ctx context.Context, userID, mediaID int64, rating *float64) error {
	if rating != nil && (*rating < 1.0 || *rating > 10.0) {
		return fmt.Errorf("rating must be between 1.0 and 10.0")
	}
	if err := s.userDataRepo.SetRating(ctx, userID, mediaID, rating); err != nil {
		return fmt.Errorf("setting rating: %w", err)
	}
	return nil
}

// DismissProgress resets progress while preserving favorite/rating/play_count
func (s *UserDataService) DismissProgress(ctx context.Context, userID, mediaID int64) error {
	if err := s.userDataRepo.DismissProgress(ctx, userID, mediaID); err != nil {
		return fmt.Errorf("dismissing progress: %w", err)
	}
	return nil
}

// ContinueWatching returns in-progress items (position > 0, not completed)
func (s *UserDataService) ContinueWatching(ctx context.Context, userID int64, limit int) ([]*model.ContinueWatchingItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	items, err := s.userDataRepo.ListContinueWatching(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing continue watching: %w", err)
	}
	return items, nil
}

// NextUp returns the next unwatched episode for each series being followed
func (s *UserDataService) NextUp(ctx context.Context, userID int64, limit int) ([]*model.NextUpItem, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	items, err := s.userDataRepo.ListNextUp(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("listing next up: %w", err)
	}
	return items, nil
}
