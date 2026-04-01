package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"path/filepath"

	"github.com/thawng/velox/internal/metadata"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/nameparser"
	"github.com/thawng/velox/pkg/nfo"
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
		media.PosterPath = details.PosterPath
		media.BackdropPath = details.BackdropPath
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
		media.PosterPath = details.PosterPath
		media.BackdropPath = details.BackdropPath

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
		media.PosterPath = result.StillPath
	} else if result.PosterPath != "" {
		media.PosterPath = result.PosterPath
	}
	media.BackdropPath = result.BackdropPath

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
		if errors.Is(err, sql.ErrNoRows) {
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
