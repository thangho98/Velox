package service

import (
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type MediaService struct {
	repo *repository.MediaRepo
}

func NewMediaService(repo *repository.MediaRepo) *MediaService {
	return &MediaService{repo: repo}
}

func (s *MediaService) List(libraryID int64, search string, limit, offset int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(libraryID, search, limit, offset)
}

func (s *MediaService) Get(id int64) (*model.Media, error) {
	return s.repo.GetByID(id)
}
