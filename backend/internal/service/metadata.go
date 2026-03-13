package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

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
	tmdbClient    *tmdb.Client
	omdbClient    *omdb.Client
	tvdbClient    *thetvdb.Client
	fanartClient  *fanart.Client
	tvmazeClient  *tvmaze.Client
	matcher       *metadata.Matcher
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	seriesRepo    *repository.SeriesRepo
	seasonRepo    *repository.SeasonRepo
	episodeRepo   *repository.EpisodeRepo
	genreRepo     *repository.GenreRepo
	personRepo    *repository.PersonRepo
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
func (s *MetadataService) MatchAndPersistMovie(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, force bool) error {
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
func (s *MetadataService) MatchAndPersistEpisode(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
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
func (s *MetadataService) findOrCreateSeries(ctx context.Context, result *metadata.TVMatchResult, libraryID int64) (*model.Series, error) {
	if result.SeriesID > 0 {
		tmdbID := int64(result.SeriesID)
		existing, err := s.seriesRepo.GetByTmdbID(ctx, tmdbID)
		if err == nil && existing != nil {
			return existing, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
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
		return nil, err
	}

	// Sync genres for series
	s.syncSeriesGenres(ctx, series.ID, result.Genres)
	s.syncSeriesCredits(ctx, series.ID, result.Cast, result.Crew)

	// Enrich with TheTVDB data (IMDb ID, status, etc.)
	s.enrichTVDBSeries(ctx, series)

	// Enrich with fanart.tv artwork (logo, thumb)
	s.enrichFanartShow(ctx, series)

	// Enrich with TVmaze data (network, schedule)
	s.enrichTVmazeSeries(ctx, series)

	return series, nil
}

// findOrCreateSeason looks up or creates a season for a series.
func (s *MetadataService) findOrCreateSeason(ctx context.Context, seriesID int64, seasonNumber int) (*model.Season, error) {
	if seasonNumber <= 0 {
		seasonNumber = 1
	}

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

// syncMediaGenres replaces all genres for a media item.
func (s *MetadataService) syncMediaGenres(ctx context.Context, mediaID int64, genres []metadata.GenreInfo) {
	if len(genres) == 0 {
		return
	}

	if err := s.genreRepo.ClearMediaGenres(ctx, mediaID); err != nil {
		log.Printf("Failed to clear media genres: %v", err)
		return
	}

	for _, g := range genres {
		genreID, err := s.ensureGenre(ctx, g)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", g.Name, err)
			continue
		}
		if err := s.genreRepo.LinkToMedia(ctx, mediaID, genreID); err != nil {
			log.Printf("Failed to link genre %q to media %d: %v", g.Name, mediaID, err)
		}
	}
}

// syncSeriesGenres replaces all genres for a series.
func (s *MetadataService) syncSeriesGenres(ctx context.Context, seriesID int64, genres []metadata.GenreInfo) {
	if len(genres) == 0 {
		return
	}

	if err := s.genreRepo.ClearSeriesGenres(ctx, seriesID); err != nil {
		log.Printf("Failed to clear series genres: %v", err)
		return
	}

	for _, g := range genres {
		genreID, err := s.ensureGenre(ctx, g)
		if err != nil {
			log.Printf("Failed to ensure genre %q: %v", g.Name, err)
			continue
		}
		if err := s.genreRepo.LinkToSeries(ctx, seriesID, genreID); err != nil {
			log.Printf("Failed to link genre %q to series %d: %v", g.Name, seriesID, err)
		}
	}
}

// ensureGenre gets or creates a genre, returning its ID.
func (s *MetadataService) ensureGenre(ctx context.Context, g metadata.GenreInfo) (int64, error) {
	// Try by TMDb ID first
	if g.ID > 0 {
		existing, err := s.genreRepo.GetByTmdbID(ctx, int64(g.ID))
		if err == nil {
			return existing.ID, nil
		}
	}

	// Try by name
	existing, err := s.genreRepo.GetByName(ctx, g.Name)
	if err == nil {
		return existing.ID, nil
	}

	// Create new genre
	var tmdbID *int64
	if g.ID > 0 {
		id := int64(g.ID)
		tmdbID = &id
	}
	genre := &model.Genre{
		Name:   g.Name,
		TmdbID: tmdbID,
	}
	if err := s.genreRepo.Create(ctx, genre); err != nil {
		return 0, err
	}
	return genre.ID, nil
}

// syncMediaCredits replaces all credits for a media item.
func (s *MetadataService) syncMediaCredits(ctx context.Context, mediaID int64, cast []metadata.CastInfo, crew []metadata.CrewInfo) {
	if len(cast) == 0 && len(crew) == 0 {
		return
	}

	if err := s.personRepo.ClearMediaCredits(ctx, mediaID); err != nil {
		log.Printf("Failed to clear media credits: %v", err)
		return
	}

	// Top 20 cast
	limit := 20
	if len(cast) < limit {
		limit = len(cast)
	}
	for i := 0; i < limit; i++ {
		c := cast[i]
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		credit := &model.Credit{
			MediaID:      &mediaID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         "cast",
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add cast credit: %v", err)
		}
	}

	// Key crew (director, writer)
	for _, c := range crew {
		if c.Job != "Director" && c.Job != "Writer" && c.Job != "Screenplay" {
			continue
		}
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		role := "director"
		if c.Job == "Writer" || c.Job == "Screenplay" {
			role = "writer"
		}
		credit := &model.Credit{
			MediaID:  &mediaID,
			PersonID: personID,
			Role:     role,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add crew credit: %v", err)
		}
	}
}

// syncSeriesCredits replaces all credits for a series.
func (s *MetadataService) syncSeriesCredits(ctx context.Context, seriesID int64, cast []metadata.CastInfo, crew []metadata.CrewInfo) {
	if len(cast) == 0 && len(crew) == 0 {
		return
	}

	if err := s.personRepo.ClearSeriesCredits(ctx, seriesID); err != nil {
		log.Printf("Failed to clear series credits: %v", err)
		return
	}

	limit := 20
	if len(cast) < limit {
		limit = len(cast)
	}
	for i := 0; i < limit; i++ {
		c := cast[i]
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		credit := &model.Credit{
			SeriesID:     &seriesID,
			PersonID:     personID,
			Character:    c.Character,
			Role:         "cast",
			DisplayOrder: c.Order,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add series cast credit: %v", err)
		}
	}

	for _, c := range crew {
		if c.Job != "Director" && c.Job != "Writer" && c.Job != "Screenplay" {
			continue
		}
		personID, err := s.ensurePerson(ctx, c.ID, c.Name, c.ProfilePath)
		if err != nil {
			continue
		}
		role := "director"
		if c.Job == "Writer" || c.Job == "Screenplay" {
			role = "writer"
		}
		credit := &model.Credit{
			SeriesID: &seriesID,
			PersonID: personID,
			Role:     role,
		}
		if err := s.personRepo.AddCredit(ctx, credit); err != nil {
			log.Printf("Failed to add series crew credit: %v", err)
		}
	}
}

// ensurePerson gets or creates a person, returning their local ID.
func (s *MetadataService) ensurePerson(ctx context.Context, tmdbPersonID int, name, profilePath string) (int64, error) {
	if tmdbPersonID > 0 {
		existing, err := s.personRepo.GetByTmdbID(ctx, int64(tmdbPersonID))
		if err == nil {
			return existing.ID, nil
		}
	}

	var tmdbID *int64
	if tmdbPersonID > 0 {
		id := int64(tmdbPersonID)
		tmdbID = &id
	}
	person := &model.Person{
		Name:        name,
		TmdbID:      tmdbID,
		ProfilePath: profilePath,
	}
	if err := s.personRepo.Create(ctx, person); err != nil {
		return 0, err
	}
	return person.ID, nil
}

// IdentifyByTmdbID manually identifies a media item with a specific TMDb ID.
func (s *MetadataService) IdentifyByTmdbID(ctx context.Context, media *model.Media, tmdbID int, mediaType string) error {
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

		// Step 1: Auto-match unmatched media against TMDb
		if m.TmdbID == nil {
			if err := s.AutoMatchAndRefresh(ctx, m); err != nil {
				log.Printf("Auto-match failed for media %d (%s): %v", m.ID, m.Title, err)
				continue
			}
			// Re-read after match to get updated fields
			refreshed, err := s.mediaRepo.GetByID(ctx, m.ID)
			if err != nil {
				continue
			}
			*m = *refreshed
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

// enrichOMDbRatings fetches IMDb/RT/Metacritic ratings from OMDb and updates media.
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichOMDbRatings(ctx context.Context, media *model.Media) {
	if s.omdbClient == nil || media.ImdbID == nil || *media.ImdbID == "" {
		return
	}

	result, err := s.omdbClient.GetByIMDbID(ctx, *media.ImdbID)
	if err != nil {
		log.Printf("OMDb lookup failed for %s: %v", *media.ImdbID, err)
		return
	}

	media.IMDbRating = result.IMDbRatingFloat()
	media.RTScore = result.RottenTomatoesScore()
	media.MetacriticScore = result.MetascoreInt()

	if err := s.mediaRepo.UpdateOMDbRatings(ctx, media.ID, media.IMDbRating, media.RTScore, media.MetacriticScore); err != nil {
		log.Printf("Failed to save OMDb ratings for media %d: %v", media.ID, err)
	}
}

// enrichTVDBSeries fetches series info from TheTVDB to fill in missing data
// (e.g. IMDb ID from remote IDs). Non-fatal: logs errors and continues.
func (s *MetadataService) enrichTVDBSeries(ctx context.Context, series *model.Series) {
	if s.tvdbClient == nil || series.TvdbID == nil || *series.TvdbID == 0 {
		return
	}

	tvdbSeries, err := s.tvdbClient.GetSeriesExtended(ctx, int(*series.TvdbID), true)
	if err != nil {
		log.Printf("TheTVDB lookup failed for series %d (tvdb:%d): %v", series.ID, *series.TvdbID, err)
		return
	}

	changed := false

	// Fill in IMDb ID if missing
	if (series.ImdbID == nil || *series.ImdbID == "") && len(tvdbSeries.RemoteIDs) > 0 {
		imdbID := thetvdb.IMDbID(tvdbSeries.RemoteIDs)
		if imdbID != "" {
			series.ImdbID = &imdbID
			changed = true
		}
	}

	// Fill in status if missing
	if series.Status == "" && tvdbSeries.Status != nil {
		series.Status = tvdbSeries.Status.Name
		changed = true
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to update series %d with TVDB data: %v", series.ID, err)
		}
	}
}

// enrichFanartMovie fetches artwork from fanart.tv for a movie (by TMDb ID).
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichFanartMovie(ctx context.Context, media *model.Media) {
	if s.fanartClient == nil || media.TmdbID == nil || *media.TmdbID == 0 {
		return
	}

	images, err := s.fanartClient.GetMovieImages(ctx, int(*media.TmdbID))
	if err != nil {
		if err != fanart.ErrNotFound {
			log.Printf("fanart.tv movie lookup failed for tmdb:%d: %v", *media.TmdbID, err)
		}
		return
	}

	changed := false

	if media.LogoPath == "" {
		if logo := fanart.BestImage(images.HDClearLogo); logo != "" {
			media.LogoPath = logo
			changed = true
		} else if logo := fanart.BestImage(images.MovieLogo); logo != "" {
			media.LogoPath = logo
			changed = true
		}
	}

	if media.ThumbPath == "" {
		if thumb := fanart.BestImage(images.MovieThumb); thumb != "" {
			media.ThumbPath = thumb
			changed = true
		}
	}

	if changed {
		if err := s.mediaRepo.Update(ctx, media); err != nil {
			log.Printf("Failed to save fanart.tv artwork for media %d: %v", media.ID, err)
		}
	}
}

// enrichFanartShow fetches artwork from fanart.tv for a TV show (by TVDB ID).
// Non-fatal: logs errors and continues.
func (s *MetadataService) enrichFanartShow(ctx context.Context, series *model.Series) {
	if s.fanartClient == nil || series.TvdbID == nil || *series.TvdbID == 0 {
		return
	}

	images, err := s.fanartClient.GetShowImages(ctx, int(*series.TvdbID))
	if err != nil {
		if err != fanart.ErrNotFound {
			log.Printf("fanart.tv show lookup failed for tvdb:%d: %v", *series.TvdbID, err)
		}
		return
	}

	changed := false

	if series.LogoPath == "" {
		if logo := fanart.BestImage(images.HDClearLogo); logo != "" {
			series.LogoPath = logo
			changed = true
		} else if logo := fanart.BestImage(images.ClearLogo); logo != "" {
			series.LogoPath = logo
			changed = true
		}
	}

	if series.ThumbPath == "" {
		if thumb := fanart.BestImage(images.TVThumb); thumb != "" {
			series.ThumbPath = thumb
			changed = true
		}
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to save fanart.tv artwork for series %d: %v", series.ID, err)
		}
	}
}

// enrichTVmazeSeries fetches series info from TVmaze to fill in network and missing IDs.
// Looks up by TVDB ID first, then IMDb ID. Non-fatal: logs errors and continues.
func (s *MetadataService) enrichTVmazeSeries(ctx context.Context, series *model.Series) {
	if s.tvmazeClient == nil {
		return
	}

	var show *tvmaze.Show
	var err error

	// Try TVDB ID first
	if series.TvdbID != nil && *series.TvdbID > 0 {
		show, err = s.tvmazeClient.LookupByTVDB(ctx, int(*series.TvdbID))
		if err != nil && err != tvmaze.ErrNotFound {
			log.Printf("TVmaze TVDB lookup failed for series %d (tvdb:%d): %v", series.ID, *series.TvdbID, err)
		}
	}

	// Fallback to IMDb ID
	if show == nil && series.ImdbID != nil && *series.ImdbID != "" {
		show, err = s.tvmazeClient.LookupByIMDb(ctx, *series.ImdbID)
		if err != nil && err != tvmaze.ErrNotFound {
			log.Printf("TVmaze IMDb lookup failed for series %d (%s): %v", series.ID, *series.ImdbID, err)
		}
	}

	if show == nil {
		return
	}

	changed := false

	// Fill in network if missing
	if series.Network == "" {
		network := show.NetworkName()
		if network != "" {
			series.Network = network
			changed = true
		}
	}

	// Fill in status if missing
	if series.Status == "" && show.Status != "" {
		series.Status = show.Status
		changed = true
	}

	// Fill in TVDB ID if missing
	if (series.TvdbID == nil || *series.TvdbID == 0) && show.Externals.TheTVDB != nil {
		tvdbID := int64(*show.Externals.TheTVDB)
		series.TvdbID = &tvdbID
		changed = true
	}

	// Fill in IMDb ID if missing
	if (series.ImdbID == nil || *series.ImdbID == "") && show.Externals.IMDb != "" {
		series.ImdbID = &show.Externals.IMDb
		changed = true
	}

	if changed {
		if err := s.seriesRepo.Update(ctx, series); err != nil {
			log.Printf("Failed to update series %d with TVmaze data: %v", series.ID, err)
		}
	}
}
