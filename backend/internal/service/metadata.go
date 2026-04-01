package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"sync"

	"github.com/thawng/velox/internal/metadata"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/fanart"
	"github.com/thawng/velox/pkg/nameparser"
	"github.com/thawng/velox/pkg/omdb"
	"github.com/thawng/velox/pkg/thetvdb"
	"github.com/thawng/velox/pkg/tmdb"
	"github.com/thawng/velox/pkg/tvmaze"
)

// MetadataService orchestrates metadata matching and persistence.
type MetadataService struct {
	tmdbClient      *tmdb.Client
	omdbClient      *omdb.Client
	tvdbClient      *thetvdb.Client
	fanartClient    *fanart.Client
	tvmazeClient    *tvmaze.Client
	matcher         *metadata.Matcher
	mediaRepo       *repository.MediaRepo
	mediaFileRepo   *repository.MediaFileRepo
	seriesRepo      *repository.SeriesRepo
	seasonRepo      *repository.SeasonRepo
	episodeRepo     *repository.EpisodeRepo
	genreRepo       *repository.GenreRepo
	personRepo      *repository.PersonRepo
	notificationSvc *NotificationService

	// Mutexes protecting find-or-create operations against concurrent duplicate
	// creation when the scanner processes multiple episodes in parallel.
	seriesCreateMu sync.Mutex
	seasonCreateMu sync.Mutex
}

// SetNotificationService sets the optional notification service for identify events.
func (s *MetadataService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

// SetOMDbClient sets an optional OMDb client for rating enrichment.
func (s *MetadataService) SetOMDbClient(client *omdb.Client) {
	s.omdbClient = client
}

// SetTVDBClient sets an optional TheTVDB client for additional metadata.
func (s *MetadataService) SetTVDBClient(client *thetvdb.Client) {
	s.tvdbClient = client
}

// SetFanartClient sets an optional fanart.tv client for artwork enrichment.
func (s *MetadataService) SetFanartClient(client *fanart.Client) {
	s.fanartClient = client
}

// SetTVmazeClient sets an optional TVmaze client for TV schedule/network data.
func (s *MetadataService) SetTVmazeClient(client *tvmaze.Client) {
	s.tvmazeClient = client
}

// NewMetadataService creates a new metadata service.
// Returns nil if tmdbClient is nil (no API key configured).
func NewMetadataService(
	tmdbClient *tmdb.Client,
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	seriesRepo *repository.SeriesRepo,
	seasonRepo *repository.SeasonRepo,
	episodeRepo *repository.EpisodeRepo,
	genreRepo *repository.GenreRepo,
	personRepo *repository.PersonRepo,
) *MetadataService {
	if tmdbClient == nil {
		return nil
	}
	return &MetadataService{
		tmdbClient:    tmdbClient,
		matcher:       metadata.NewMatcher(tmdbClient),
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		seriesRepo:    seriesRepo,
		seasonRepo:    seasonRepo,
		episodeRepo:   episodeRepo,
		genreRepo:     genreRepo,
		personRepo:    personRepo,
	}
}

// MatchAndPersistMovie matches a movie against TMDb and saves metadata.
// Skips if media already has a tmdb_id (unless force is true).
// Skips if metadata is locked (admin manual edit protected from rescan override).
func (s *MetadataService) MatchAndPersistMovie(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, force bool) error {
	if !force && media.MetadataLocked {
		return nil // Respect manual edits
	}
	if !force && media.TmdbID != nil {
		return nil // Already matched
	}

	result, err := s.matcher.MatchMovie(ctx, parsed, filePath)
	if err != nil {
		return err
	}
	if !result.Found {
		return nil
	}

	// Update media with TMDb metadata
	tmdbID := int64(result.TMDbID)
	media.TmdbID = &tmdbID
	if result.IMDbID != "" {
		media.ImdbID = &result.IMDbID
	}
	media.Title = result.Title
	media.SortTitle = result.Title
	media.Overview = result.Overview
	media.ReleaseDate = result.ReleaseDate
	media.Rating = result.Rating
	media.PosterPath = result.PosterPath
	media.BackdropPath = result.BackdropPath

	if err := s.mediaRepo.Update(ctx, media); err != nil {
		return err
	}

	// Sync genres and credits
	s.syncMediaGenres(ctx, media.ID, result.Genres)
	s.syncMediaCredits(ctx, media.ID, result.Cast, result.Crew)

	// Enrich with OMDb ratings (IMDb, RT, Metacritic)
	s.enrichOMDbRatings(ctx, media)

	// Enrich with fanart.tv artwork (logo, thumb)
	s.enrichFanartMovie(ctx, media)

	return nil
}

// MatchAndPersistEpisode matches a TV episode against TMDb and saves metadata.
// Skips if metadata is locked (admin manual edit protected from rescan override).
func (s *MetadataService) MatchAndPersistEpisode(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
	if !force && media.MetadataLocked {
		return nil // Respect manual edits
	}
	if !force && media.TmdbID != nil {
		return nil
	}

	result, err := s.matcher.MatchTVShow(ctx, parsed, filePath)
	if err != nil {
		return err
	}
	if !result.Found {
		return nil
	}

	// Find or create series
	series, err := s.findOrCreateSeries(ctx, result, libraryID)
	if err != nil {
		return err
	}

	// Find or create season
	season, err := s.findOrCreateSeason(ctx, series.ID, result.SeasonNumber)
	if err != nil {
		return err
	}

	// Update the media (episode) with TMDb data
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
	} else {
		media.PosterPath = result.PosterPath
	}
	media.BackdropPath = result.BackdropPath

	if err := s.mediaRepo.Update(ctx, media); err != nil {
		return err
	}

	// Link episode to series/season
	s.linkEpisode(ctx, media.ID, series.ID, season.ID, result.EpisodeNumber, media.Title, media.Overview, media.PosterPath)

	return nil
}

// findOrCreateSeries looks up a series by TMDb ID, or creates one.
// The DB check+create is protected by seriesCreateMu to prevent duplicate rows
// when multiple scanner workers process episodes from the same series in parallel.
// HTTP enrichment happens outside the lock so it doesn't stall other workers.
func (s *MetadataService) findOrCreateSeries(ctx context.Context, result *metadata.TVMatchResult, libraryID int64) (*model.Series, error) {
	series, isNew, err := s.ensureSeries(ctx, result, libraryID)
	if err != nil || !isNew {
		return series, err
	}

	// Enrichment for newly created series only — safe outside the lock because
	// series.ID is now stable and no other worker will create the same row.
	s.syncSeriesGenres(ctx, series.ID, result.Genres)
	s.syncSeriesCredits(ctx, series.ID, result.Cast, result.Crew)
	s.enrichTVDBSeries(ctx, series)
	s.enrichFanartShow(ctx, series)
	s.enrichTVmazeSeries(ctx, series)

	return series, nil
}

// ensureSeries performs the find-or-create under a mutex to prevent TOCTOU races
// when parallel workers process episodes from the same series simultaneously.
// Returns (series, isNew, err).
func (s *MetadataService) ensureSeries(ctx context.Context, result *metadata.TVMatchResult, libraryID int64) (*model.Series, bool, error) {
	s.seriesCreateMu.Lock()
	defer s.seriesCreateMu.Unlock()

	if result.SeriesID > 0 {
		tmdbID := int64(result.SeriesID)
		existing, err := s.seriesRepo.GetByTmdbID(ctx, tmdbID)
		if err == nil && existing != nil {
			return existing, false, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, false, err
		}
	}

	seriesTitle := result.SeriesTitle
	if seriesTitle == "" {
		seriesTitle = result.Title
	}
	seriesOverview := result.SeriesOverview
	if seriesOverview == "" {
		seriesOverview = result.Overview
	}
	series := &model.Series{
		LibraryID:    libraryID,
		Title:        seriesTitle,
		SortTitle:    seriesTitle,
		Overview:     seriesOverview,
		FirstAirDate: result.SeriesAirDate,
		PosterPath:   result.SeriesPoster,
		BackdropPath: result.BackdropPath,
	}
	if result.SeriesID > 0 {
		tmdbID := int64(result.SeriesID)
		series.TmdbID = &tmdbID
	}
	if result.TvdbID > 0 {
		tvdbID := int64(result.TvdbID)
		series.TvdbID = &tvdbID
	}

	if err := s.seriesRepo.Create(ctx, series); err != nil {
		return nil, false, err
	}
	return series, true, nil
}

// findOrCreateSeason looks up or creates a season for a series.
// Protected by seasonCreateMu to prevent duplicate rows when parallel workers
// process episodes from the same season simultaneously.
func (s *MetadataService) findOrCreateSeason(ctx context.Context, seriesID int64, seasonNumber int) (*model.Season, error) {
	if seasonNumber <= 0 {
		seasonNumber = 1
	}

	s.seasonCreateMu.Lock()
	defer s.seasonCreateMu.Unlock()

	existing, err := s.seasonRepo.GetBySeriesAndNumber(ctx, seriesID, seasonNumber)
	if err == nil && existing != nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	season := &model.Season{
		SeriesID:     seriesID,
		SeasonNumber: seasonNumber,
		Title:        "",
	}
	if err := s.seasonRepo.Create(ctx, season); err != nil {
		return nil, err
	}
	return season, nil
}

// linkEpisode creates or updates the episode record linking media to series/season.
func (s *MetadataService) linkEpisode(ctx context.Context, mediaID, seriesID, seasonID int64, episodeNumber int, title, overview, stillPath string) {
	if episodeNumber <= 0 {
		return
	}

	existing, err := s.episodeRepo.GetByMediaID(ctx, mediaID)
	if err == nil && existing != nil {
		// Update metadata if missing
		if (existing.Title == "" && title != "") || (existing.StillPath == "" && stillPath != "") || (existing.Overview == "" && overview != "") {
			existing.Title = title
			existing.Overview = overview
			existing.StillPath = stillPath
			if updateErr := s.episodeRepo.Update(ctx, existing); updateErr != nil {
				log.Printf("Failed to update episode metadata: %v", updateErr)
			}
		}
		return
	}

	ep := &model.Episode{
		SeriesID:      seriesID,
		SeasonID:      seasonID,
		MediaID:       mediaID,
		EpisodeNumber: episodeNumber,
		Title:         title,
		Overview:      overview,
		StillPath:     stillPath,
	}
	if err := s.episodeRepo.Create(ctx, ep); err != nil {
		log.Printf("Failed to link episode: %v", err)
	}
}

// updateEpisodeLink updates an existing episode record with the correct season/episode.
func (s *MetadataService) updateEpisodeLink(ctx context.Context, episode *model.Episode, seasonNum, episodeNum int, media *model.Media) {
	if seasonNum <= 0 || episodeNum <= 0 {
		return
	}

	// Find or create the correct season
	season, err := s.findOrCreateSeason(ctx, episode.SeriesID, seasonNum)
	if err != nil {
		log.Printf("Failed to find/create season %d for series %d: %v", seasonNum, episode.SeriesID, err)
		return
	}

	// Update the episode record with correct season and number
	episode.SeasonID = season.ID
	episode.EpisodeNumber = episodeNum
	if err := s.episodeRepo.UpdateSeasonLink(ctx, episode.ID, season.ID, episodeNum); err != nil {
		log.Printf("Failed to update episode link: %v", err)
	}
}

// GetSeries retrieves a series by ID. Returns ErrNotFound if not found.
func (s *MetadataService) GetSeries(ctx context.Context, id int64) (*model.Series, error) {
	series, err := s.seriesRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return series, nil
}
