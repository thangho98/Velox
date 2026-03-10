package service

import (
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type ProgressService struct {
	repo *repository.ProgressRepo
}

func NewProgressService(repo *repository.ProgressRepo) *ProgressService {
	return &ProgressService{repo: repo}
}

func (s *ProgressService) Get(mediaID int64) (*model.Progress, error) {
	return s.repo.Get(mediaID)
}

func (s *ProgressService) Update(p *model.Progress) error {
	return s.repo.Upsert(p)
}
