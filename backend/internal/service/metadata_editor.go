package service

import (
	"context"
	"fmt"
	"log"

	"github.com/thawng/velox/internal/metadata"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/pkg/nameparser"
)

// IdentifyByTmdbID manually identifies a media item with a specific TMDb ID.
// Auto-unlocks metadata_locked since the admin explicitly chose to re-match.
func (s *MetadataService) IdentifyByTmdbID(ctx context.Context, media *model.Media, tmdbID int, mediaType string) error {
	media.MetadataLocked = false
	if mediaType == "tv" || media.MediaType == "episode" {
		// Fetch TV details and update
		details, err := s.tmdbClient.GetTVDetails(ctx, tmdbID)
		if err != nil {
			return err
		}
		id := int64(details.ID)
		media.TmdbID = &id
		if details.ExternalIDs != nil {
			if details.ExternalIDs.IMDbID != "" {
				media.ImdbID = &details.ExternalIDs.IMDbID
			}
			if details.ExternalIDs.TVDBID > 0 {
				tvdbID := int64(details.ExternalIDs.TVDBID)
				media.TvdbID = &tvdbID
			}
		}
		media.Title = details.Name
		media.SortTitle = details.Name
		media.Overview = details.Overview
		media.ReleaseDate = details.FirstAirDate
		media.Rating = details.VoteAverage
		media.PosterPath = model.PosterPath(details.PosterPath)
		media.BackdropPath = model.BackdropPath(details.BackdropPath)
	} else {
		details, err := s.tmdbClient.GetMovieDetails(ctx, tmdbID)
		if err != nil {
			return err
		}
		id := int64(details.ID)
		media.TmdbID = &id
		if details.IMDbID != "" {
			media.ImdbID = &details.IMDbID
		}
		media.Title = details.Title
		media.SortTitle = details.Title
		media.Overview = details.Overview
		media.ReleaseDate = details.ReleaseDate
		media.Rating = details.VoteAverage
		media.PosterPath = model.PosterPath(details.PosterPath)
		media.BackdropPath = model.BackdropPath(details.BackdropPath)

		if err := s.mediaRepo.Update(ctx, media); err != nil {
			return err
		}

		// Sync genres and credits
		var genres []metadata.GenreInfo
		for _, g := range details.Genres {
			genres = append(genres, metadata.GenreInfo{ID: g.ID, Name: g.Name})
		}
		s.syncMediaGenres(ctx, media.ID, genres)

		if details.Credits != nil {
			var cast []metadata.CastInfo
			for _, c := range details.Credits.Cast {
				cast = append(cast, metadata.CastInfo{ID: c.ID, Name: c.Name, Character: c.Character, ProfilePath: c.ProfilePath, Order: c.Order})
			}
			var crew []metadata.CrewInfo
			for _, c := range details.Credits.Crew {
				crew = append(crew, metadata.CrewInfo{ID: c.ID, Name: c.Name, Job: c.Job, Department: c.Department, ProfilePath: c.ProfilePath})
			}
			s.syncMediaCredits(ctx, media.ID, cast, crew)
		}
		s.enrichOMDbRatings(ctx, media)
		s.enrichFanartMovie(ctx, media)
		return nil
	}

	return s.mediaRepo.Update(ctx, media)
}

// RefreshMetadata re-fetches metadata from TMDb for a media item that already has a tmdb_id.
func (s *MetadataService) RefreshMetadata(ctx context.Context, media *model.Media) error {
	if media.TmdbID == nil {
		return nil
	}

	// For episodes, media.tmdb_id stores the EPISODE TMDb ID, not the SERIES ID.
	// We must look up the linked series and use its tmdb_id to re-fetch correctly.
	if media.MediaType == "episode" {
		return s.refreshEpisodeMetadata(ctx, media)
	}

	if err := s.IdentifyByTmdbID(ctx, media, int(*media.TmdbID), media.MediaType); err != nil {
		return err
	}
	s.enrichOMDbRatings(ctx, media)
	return nil
}

// refreshEpisodeMetadata re-fetches episode metadata via its linked series.
func (s *MetadataService) refreshEpisodeMetadata(ctx context.Context, media *model.Media) error {
	// Look up the episode record to find series + season/episode numbers
	episode, err := s.episodeRepo.GetByMediaID(ctx, media.ID)
	if err != nil {
		// No episode link — fall back to auto-match from filename
		return s.AutoMatchAndRefresh(ctx, media)
	}

	// Get the series to find its TMDb ID
	series, err := s.seriesRepo.GetByID(ctx, episode.SeriesID)
	if err != nil || series.TmdbID == nil {
		return s.AutoMatchAndRefresh(ctx, media)
	}

	// Get primary file to re-parse season/episode numbers
	primaryFile, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, media.ID)
	if err != nil {
		return fmt.Errorf("no media file found: %w", err)
	}
	parsed := nameparser.Parse(primaryFile.FilePath)

	// Re-match using the correct series TMDb ID
	result, err := s.matcher.MatchTVEpisodeBySeriesID(ctx, int(*series.TmdbID), parsed, primaryFile.FilePath)
	if err != nil {
		return err
	}
	if !result.Found {
		return nil
	}

	// Update the media with fresh episode data
	if result.TMDbID > 0 {
		tmdbID := int64(result.TMDbID)
		media.TmdbID = &tmdbID
	}
	if result.TvdbID > 0 {
		tvdbID := int64(result.TvdbID)
		media.TvdbID = &tvdbID
	}
	media.Title = result.EpisodeTitle
	if media.Title == "" {
		media.Title = result.Title
	}
	media.SortTitle = media.Title
	media.Overview = result.Overview
	media.ReleaseDate = result.ReleaseDate
	media.Rating = result.Rating
	if result.StillPath != "" {
		media.PosterPath = model.PosterPath(result.StillPath)
	} else if result.PosterPath != "" {
		media.PosterPath = model.PosterPath(result.PosterPath)
	}
	media.BackdropPath = model.BackdropPath(result.BackdropPath)

	if err := s.mediaRepo.Update(ctx, media); err != nil {
		return err
	}

	// Update episode link with correct season/episode
	s.updateEpisodeLink(ctx, episode, result.SeasonNumber, result.EpisodeNumber, media)

	return nil
}

// AutoMatchAndRefresh tries to auto-match a media item against TMDb using its file path,
// then refreshes metadata. Works even if media has no tmdb_id yet.
func (s *MetadataService) AutoMatchAndRefresh(ctx context.Context, media *model.Media) error {
	// Get primary file to extract path for name parsing
	primaryFile, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, media.ID)
	if err != nil {
		return fmt.Errorf("no media file found: %w", err)
	}

	parsed := nameparser.Parse(primaryFile.FilePath)

	if media.MediaType == "episode" {
		return s.MatchAndPersistEpisode(ctx, media, parsed, primaryFile.FilePath, media.LibraryID, true)
	}
	return s.MatchAndPersistMovie(ctx, media, parsed, primaryFile.FilePath, true)
}

// BulkRefreshAllMetadata auto-matches all unmatched media against TMDb,
// then fetches OMDb ratings for everything with an IMDb ID.
// Returns the number of items updated.
func (s *MetadataService) BulkRefreshAllMetadata(ctx context.Context) (int, error) {
	items, err := s.mediaRepo.List(ctx, 0, "", 0, 0)
	if err != nil {
		return 0, fmt.Errorf("listing all media: %w", err)
	}

	updated := 0
	for i := range items {
		m := &items[i]

		// Step 1: Auto-match unmatched media against TMDb.
		// MatchAndPersistMovie mutates *m in-place, so no re-read needed.
		if m.TmdbID == nil {
			if err := s.AutoMatchAndRefresh(ctx, m); err != nil {
				log.Printf("Auto-match failed for media %d (%s): %v", m.ID, m.Title, err)
				continue
			}
			if m.TmdbID != nil {
				updated++
			}
		}

		// Step 2: Enrich with OMDb ratings if we have an IMDb ID
		if s.omdbClient != nil && m.ImdbID != nil && *m.ImdbID != "" {
			prevIMDb := m.IMDbRating
			prevRT := m.RTScore
			prevMeta := m.MetacriticScore
			s.enrichOMDbRatings(ctx, m)
			if m.IMDbRating != prevIMDb || m.RTScore != prevRT || m.MetacriticScore != prevMeta {
				updated++
			}
		}
	}
	return updated, nil
}
