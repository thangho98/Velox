package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/nfo"
)

// EditMediaMetadata performs a partial metadata update for a media item.
// Syncs genres and credits if provided. Auto-locks metadata unless explicitly set to false.
func (s *MetadataService) EditMediaMetadata(ctx context.Context, mediaID int64, req model.MetadataEditRequest) error {
	// Auto-set metadata_locked = true if not explicitly set
	if req.MetadataLocked == nil {
		locked := true
		req.MetadataLocked = &locked
	}

	// Update scalar fields
	if err := s.mediaRepo.UpdateMetadata(ctx, mediaID, req); err != nil {
		return fmt.Errorf("updating media metadata: %w", err)
	}

	// Sync genres if provided (nil = don't change, non-nil = replace)
	if req.Genres != nil {
		if err := s.syncMediaGenresByName(ctx, mediaID, req.Genres); err != nil {
			return fmt.Errorf("syncing genres: %w", err)
		}
	}

	// Sync credits if provided
	if req.Credits != nil {
		if err := s.syncMediaCreditsByInput(ctx, mediaID, req.Credits); err != nil {
			return fmt.Errorf("syncing credits: %w", err)
		}
	}

	// Write NFO if requested
	if req.SaveNFO {
		if err := s.WriteMediaNFO(ctx, mediaID); err != nil {
			return fmt.Errorf("metadata saved but NFO write failed: %w", err)
		}
	}

	return nil
}

// EditSeriesMetadata performs a partial metadata update for a series.
func (s *MetadataService) EditSeriesMetadata(ctx context.Context, seriesID int64, req model.SeriesMetadataEditRequest) error {
	if req.MetadataLocked == nil {
		locked := true
		req.MetadataLocked = &locked
	}

	if err := s.seriesRepo.UpdateMetadata(ctx, seriesID, req); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("updating series metadata: %w", err)
	}

	if req.Genres != nil {
		if err := s.syncSeriesGenresByName(ctx, seriesID, req.Genres); err != nil {
			return fmt.Errorf("syncing genres: %w", err)
		}
	}

	if req.Credits != nil {
		if err := s.syncSeriesCreditsByInput(ctx, seriesID, req.Credits); err != nil {
			return fmt.Errorf("syncing credits: %w", err)
		}
	}

	if req.SaveNFO {
		if err := s.WriteSeriesNFO(ctx, seriesID); err != nil {
			return fmt.Errorf("metadata saved but NFO write failed: %w", err)
		}
	}

	return nil
}

// EditEpisodeMetadata updates episode-level metadata (title, overview, air_date)
// and syncs the linked media record. Locks the media from rescan override.
func (s *MetadataService) EditEpisodeMetadata(ctx context.Context, episodeID int64, req model.EpisodeMetadataEditRequest) error {
	episode, err := s.episodeRepo.GetByID(ctx, episodeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("getting episode: %w", err)
	}

	// Update episode fields
	if req.Title != nil {
		episode.Title = *req.Title
	}
	if req.Overview != nil {
		episode.Overview = *req.Overview
	}
	if req.AirDate != nil {
		episode.AirDate = *req.AirDate
	}
	if req.EpisodeNumber != nil {
		episode.EpisodeNumber = *req.EpisodeNumber
	}

	if err := s.episodeRepo.Update(ctx, episode); err != nil {
		return fmt.Errorf("updating episode: %w", err)
	}

	// Sync title + overview to the linked media record
	mediaReq := model.MetadataEditRequest{}
	if req.Title != nil {
		mediaReq.Title = req.Title
		mediaReq.SortTitle = req.Title
	}
	if req.Overview != nil {
		mediaReq.Overview = req.Overview
	}
	if req.AirDate != nil {
		mediaReq.ReleaseDate = req.AirDate
	}

	// Auto-lock unless explicitly false
	if req.MetadataLocked == nil {
		locked := true
		mediaReq.MetadataLocked = &locked
	} else {
		mediaReq.MetadataLocked = req.MetadataLocked
	}

	if err := s.mediaRepo.UpdateMetadata(ctx, episode.MediaID, mediaReq); err != nil {
		if !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("syncing media metadata: %w", err)
		}
	}

	return nil
}

// UnlockMediaMetadata sets metadata_locked = false for a media item.
func (s *MetadataService) UnlockMediaMetadata(ctx context.Context, mediaID int64) error {
	return s.mediaRepo.SetMetadataLocked(ctx, mediaID, false)
}

// UnlockSeriesMetadata sets metadata_locked = false for a series.
func (s *MetadataService) UnlockSeriesMetadata(ctx context.Context, seriesID int64) error {
	return s.seriesRepo.SetMetadataLocked(ctx, seriesID, false)
}

// UpdateMediaImagePath updates poster_path or backdrop_path for a media item and auto-locks metadata.
func (s *MetadataService) UpdateMediaImagePath(ctx context.Context, mediaID int64, imageType, path string) error {
	if err := s.mediaRepo.UpdateImagePath(ctx, mediaID, imageType, path); err != nil {
		return err
	}
	return s.mediaRepo.SetMetadataLocked(ctx, mediaID, true)
}

// UpdateSeriesImagePath updates poster_path or backdrop_path for a series and auto-locks metadata.
func (s *MetadataService) UpdateSeriesImagePath(ctx context.Context, seriesID int64, imageType, path string) error {
	if err := s.seriesRepo.UpdateImagePath(ctx, seriesID, imageType, path); err != nil {
		return err
	}
	return s.seriesRepo.SetMetadataLocked(ctx, seriesID, true)
}

// WriteMediaNFO generates and writes an NFO file for a media item.
// For movies, writes a movie.nfo; for episodes, writes an episodedetails .nfo.
func (s *MetadataService) WriteMediaNFO(ctx context.Context, mediaID int64) error {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("getting media: %w", err)
	}

	// Get primary file path for NFO location
	primaryFile, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("getting primary file: %w", err)
	}

	nfoPath := nfo.MovieNFOPath(primaryFile.FilePath)

	if media.MediaType == "episode" {
		return s.writeEpisodeNFO(ctx, media, nfoPath)
	}

	data := s.buildMediaNFOData(ctx, media)
	nfoMovie := nfo.MovieFromData(data)
	return nfo.WriteMovie(nfoMovie, nfoPath)
}

func (s *MetadataService) writeEpisodeNFO(ctx context.Context, media *model.Media, nfoPath string) error {
	episode, err := s.episodeRepo.GetByMediaID(ctx, media.ID)
	if err != nil {
		return fmt.Errorf("getting episode: %w", err)
	}

	series, err := s.seriesRepo.GetByID(ctx, episode.SeriesID)
	if err != nil {
		return fmt.Errorf("getting series: %w", err)
	}

	season, err := s.seasonRepo.GetByID(ctx, episode.SeasonID)
	if err != nil {
		return fmt.Errorf("getting season: %w", err)
	}

	epData := nfo.EpisodeData{
		Title:         media.Title,
		ShowTitle:     series.Title,
		SeasonNumber:  season.SeasonNumber,
		EpisodeNumber: episode.EpisodeNumber,
		Overview:      media.Overview,
		ReleaseDate:   media.ReleaseDate,
		Rating:        media.Rating,
		TmdbID:        media.TmdbID,
		ImdbID:        media.ImdbID,
	}

	epNFO := nfo.EpisodeFromData(epData)
	return nfo.WriteEpisode(epNFO, nfoPath)
}

// WriteSeriesNFO generates and writes a tvshow.nfo for a series.
func (s *MetadataService) WriteSeriesNFO(ctx context.Context, seriesID int64) error {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil {
		return fmt.Errorf("getting series: %w", err)
	}

	// Find series directory from first episode's file path
	episodes, err := s.episodeRepo.ListBySeriesID(ctx, seriesID)
	if err != nil || len(episodes) == 0 {
		return fmt.Errorf("no episodes found for series %d", seriesID)
	}
	file, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, episodes[0].MediaID)
	if err != nil {
		return fmt.Errorf("getting episode file: %w", err)
	}

	// Series directory is typically 2 levels up from episode file (series/season/episode.mkv)
	seriesDir := filepath.Dir(filepath.Dir(file.FilePath))

	data := s.buildSeriesNFOData(ctx, series)
	nfoShow := nfo.TVShowFromData(data)
	nfoPath := nfo.TVShowNFOPath(seriesDir)

	return nfo.WriteTVShow(nfoShow, nfoPath)
}

func (s *MetadataService) buildMediaNFOData(ctx context.Context, media *model.Media) nfo.MediaData {
	data := nfo.MediaData{
		Title:        media.Title,
		SortTitle:    media.SortTitle,
		Overview:     media.Overview,
		Tagline:      media.Tagline,
		ReleaseDate:  media.ReleaseDate,
		Rating:       media.Rating,
		TmdbID:       media.TmdbID,
		ImdbID:       media.ImdbID,
		TvdbID:       media.TvdbID,
		PosterPath:   media.PosterPath,
		BackdropPath: media.BackdropPath,
	}

	// Genres
	genres, err := s.genreRepo.ListByMediaID(ctx, media.ID)
	if err == nil {
		for _, g := range genres {
			data.Genres = append(data.Genres, g.Name)
		}
	}

	// Credits
	credits, err := s.personRepo.ListCreditsByMedia(ctx, media.ID)
	if err == nil {
		for _, c := range credits {
			switch c.Credit.Role {
			case "cast":
				data.Cast = append(data.Cast, nfo.CastData{
					Name:        c.Person.Name,
					Character:   c.Credit.Character,
					Order:       c.Credit.DisplayOrder,
					ProfilePath: c.Person.ProfilePath,
				})
			case "director":
				data.Directors = append(data.Directors, c.Person.Name)
			case "writer":
				data.Writers = append(data.Writers, c.Person.Name)
			}
		}
	}

	return data
}

func (s *MetadataService) buildSeriesNFOData(ctx context.Context, series *model.Series) nfo.SeriesData {
	data := nfo.SeriesData{
		Title:        series.Title,
		SortTitle:    series.SortTitle,
		Overview:     series.Overview,
		Status:       series.Status,
		Network:      series.Network,
		FirstAirDate: series.FirstAirDate,
		TmdbID:       series.TmdbID,
		ImdbID:       series.ImdbID,
		TvdbID:       series.TvdbID,
		PosterPath:   series.PosterPath,
		BackdropPath: series.BackdropPath,
	}

	genres, err := s.genreRepo.ListBySeriesID(ctx, series.ID)
	if err == nil {
		for _, g := range genres {
			data.Genres = append(data.Genres, g.Name)
		}
	}

	credits, err := s.personRepo.ListCreditsBySeries(ctx, series.ID)
	if err == nil {
		for _, c := range credits {
			switch c.Credit.Role {
			case "cast":
				data.Cast = append(data.Cast, nfo.CastData{
					Name:        c.Person.Name,
					Character:   c.Credit.Character,
					Order:       c.Credit.DisplayOrder,
					ProfilePath: c.Person.ProfilePath,
				})
			case "director":
				data.Directors = append(data.Directors, c.Person.Name)
			case "writer":
				data.Writers = append(data.Writers, c.Person.Name)
			}
		}
	}

	return data
}
