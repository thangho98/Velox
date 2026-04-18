package service

import (
	"context"
	"errors"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
)

type MediaService struct {
	repo          *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	episodeRepo   *repository.EpisodeRepo
	seasonRepo    *repository.SeasonRepo
	imageMetaRepo *repository.ImageMetadataRepo
}

func NewMediaService(repo *repository.MediaRepo, mediaFileRepo *repository.MediaFileRepo) *MediaService {
	return &MediaService{
		repo:          repo,
		mediaFileRepo: mediaFileRepo,
	}
}

// SetEpisodeRepo sets the episode repo for episode info enrichment.
func (s *MediaService) SetEpisodeRepo(r *repository.EpisodeRepo) { s.episodeRepo = r }

// SetSeasonRepo sets the season repo for episode info enrichment.
func (s *MediaService) SetSeasonRepo(r *repository.SeasonRepo) { s.seasonRepo = r }

// SetImageMetadataRepo sets the image metadata repo for attaching rich image resources.
func (s *MediaService) SetImageMetadataRepo(r *repository.ImageMetadataRepo) { s.imageMetaRepo = r }

func (s *MediaService) List(ctx context.Context, libraryID int64, mediaType string, limit, offset int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.List(ctx, libraryID, mediaType, limit, offset)
}

// ListFiltered returns media items with advanced filtering, sorting, and pagination.
// Supports filtering by library, media type, search query, genre, and year.
func (s *MediaService) ListFiltered(ctx context.Context, filter model.MediaListFilter) ([]model.MediaListItem, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	return s.repo.ListFiltered(ctx, filter)
}

func (s *MediaService) GetAlphabet(ctx context.Context, filter model.MediaListFilter) ([]model.AlphabetCount, error) {
	return s.repo.GetAlphabet(ctx, filter)
}

func (s *MediaService) Get(ctx context.Context, id int64) (*model.Media, error) {
	media, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return media, err
}

func (s *MediaService) GetWithFiles(ctx context.Context, id int64) (*model.MediaWithFiles, error) {
	media, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	files, err := s.mediaFileRepo.ListByMediaID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &model.MediaWithFiles{
		Media: *media,
		Files: files,
	}

	// Enrich with episode/season info when applicable
	if media.MediaType == "episode" && s.episodeRepo != nil && s.seasonRepo != nil {
		if ep, err := s.episodeRepo.GetByMediaID(ctx, id); err == nil {
			result.SeriesID = ep.SeriesID
			result.SeasonID = ep.SeasonID
			result.EpisodeNumber = ep.EpisodeNumber
			if season, err := s.seasonRepo.GetByID(ctx, ep.SeasonID); err == nil {
				result.SeasonNumber = season.SeasonNumber
			}
		}
	}

	return result, nil
}

func (s *MediaService) Search(ctx context.Context, query string, limit int) ([]model.Media, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.Search(ctx, query, limit)
}

// ListVersions returns all physical files for the given media, ordered by is_primary DESC.
// Returns ErrNotFound if the media ID does not exist.
func (s *MediaService) ListVersions(ctx context.Context, mediaID int64) ([]model.MediaFile, error) {
	if _, err := s.repo.GetByID(ctx, mediaID); errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}
	return s.mediaFileRepo.ListByMediaID(ctx, mediaID)
}

func (s *MediaService) AttachImageResources(ctx context.Context, m *model.Media) error {
	if s.imageMetaRepo == nil {
		return nil
	}
	paths := []string{string(m.PosterPath), string(m.BackdropPath), string(m.LogoPath), string(m.ThumbPath)}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	m.Poster = model.BuildImageResource(string(m.PosterPath), "poster", meta[string(m.PosterPath)])
	m.Backdrop = model.BuildImageResource(string(m.BackdropPath), "backdrop", meta[string(m.BackdropPath)])
	m.Logo = model.BuildImageResource(string(m.LogoPath), "logo", meta[string(m.LogoPath)])
	m.Thumb = model.BuildImageResource(string(m.ThumbPath), "backdrop", meta[string(m.ThumbPath)])
	return nil
}

func (s *MediaService) AttachImageResourcesForList(ctx context.Context, items []model.MediaListItem) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.PosterPath != "" {
			paths = append(paths, string(item.PosterPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].PosterPath)
		items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForUserData(ctx context.Context, items []*model.UserData) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.MediaPoster != "" {
			paths = append(paths, string(item.MediaPoster))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].MediaPoster)
		if p != "" {
			items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
		}
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForContinueWatching(ctx context.Context, items []*model.ContinueWatchingItem) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.PosterPath != "" {
			paths = append(paths, string(item.PosterPath))
		}
		if item.BackdropPath != "" {
			paths = append(paths, string(item.BackdropPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].PosterPath != "" {
			p := string(items[i].PosterPath)
			items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
		}
		if items[i].BackdropPath != "" {
			p := string(items[i].BackdropPath)
			items[i].Backdrop = model.BuildImageResource(p, "backdrop", meta[p])
		}
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForNextUp(ctx context.Context, items []*model.NextUpItem) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.SeriesPoster != "" {
			paths = append(paths, string(item.SeriesPoster))
		}
		if item.BackdropPath != "" {
			paths = append(paths, string(item.BackdropPath))
		}
		if item.StillPath != "" {
			paths = append(paths, string(item.StillPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		if items[i].SeriesPoster != "" {
			p := string(items[i].SeriesPoster)
			items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
		}
		if items[i].BackdropPath != "" {
			p := string(items[i].BackdropPath)
			items[i].Backdrop = model.BuildImageResource(p, "backdrop", meta[p])
		}
		if items[i].StillPath != "" {
			p := string(items[i].StillPath)
			items[i].Still = model.BuildImageResource(p, "still", meta[p])
		}
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForSeriesList(ctx context.Context, items []model.SeriesListItem) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.PosterPath != "" {
			paths = append(paths, string(item.PosterPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].PosterPath)
		items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForSeries(ctx context.Context, m *model.Series) error {
	if s.imageMetaRepo == nil || m == nil {
		return nil
	}
	paths := []string{string(m.PosterPath), string(m.BackdropPath), string(m.LogoPath), string(m.ThumbPath)}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	m.Poster = model.BuildImageResource(string(m.PosterPath), "poster", meta[string(m.PosterPath)])
	m.Backdrop = model.BuildImageResource(string(m.BackdropPath), "backdrop", meta[string(m.BackdropPath)])
	m.Logo = model.BuildImageResource(string(m.LogoPath), "logo", meta[string(m.LogoPath)])
	m.Thumb = model.BuildImageResource(string(m.ThumbPath), "backdrop", meta[string(m.ThumbPath)])
	return nil
}

// AttachImageResourcesForSeriesBatch enriches a slice of full Series with
// ImageResource metadata using a single GetBatch round-trip (vs N calls for
// per-row AttachImageResourcesForSeries in search/list loops).
func (s *MediaService) AttachImageResourcesForSeriesBatch(ctx context.Context, items []model.Series) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for i := range items {
		if items[i].PosterPath != "" {
			paths = append(paths, string(items[i].PosterPath))
		}
		if items[i].BackdropPath != "" {
			paths = append(paths, string(items[i].BackdropPath))
		}
		if items[i].LogoPath != "" {
			paths = append(paths, string(items[i].LogoPath))
		}
		if items[i].ThumbPath != "" {
			paths = append(paths, string(items[i].ThumbPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		items[i].Poster = model.BuildImageResource(string(items[i].PosterPath), "poster", meta[string(items[i].PosterPath)])
		items[i].Backdrop = model.BuildImageResource(string(items[i].BackdropPath), "backdrop", meta[string(items[i].BackdropPath)])
		items[i].Logo = model.BuildImageResource(string(items[i].LogoPath), "logo", meta[string(items[i].LogoPath)])
		items[i].Thumb = model.BuildImageResource(string(items[i].ThumbPath), "backdrop", meta[string(items[i].ThumbPath)])
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForSeasons(ctx context.Context, items []model.Season) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.PosterPath != "" {
			paths = append(paths, string(item.PosterPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].PosterPath)
		items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForEpisodes(ctx context.Context, items []model.Episode) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.StillPath != "" {
			paths = append(paths, string(item.StillPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].StillPath)
		items[i].Still = model.BuildImageResource(p, "still", meta[p])
	}
	return nil
}

func (s *MediaService) AttachImageResourcesForMostWatched(ctx context.Context, items []model.MostWatchedItem) error {
	if s.imageMetaRepo == nil || len(items) == 0 {
		return nil
	}
	var paths []string
	for _, item := range items {
		if item.PosterPath != "" {
			paths = append(paths, string(item.PosterPath))
		}
	}
	meta, err := s.imageMetaRepo.GetBatch(ctx, paths)
	if err != nil {
		return err
	}
	for i := range items {
		p := string(items[i].PosterPath)
		items[i].Poster = model.BuildImageResource(p, "poster", meta[p])
	}
	return nil
}
