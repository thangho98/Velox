package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"log"
	"regexp"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/anilist"
	"github.com/thawng/velox/pkg/nameparser"
)

var animeDescriptionTagPattern = regexp.MustCompile(`<[^>]+>`)

// MatchAndPersistAnime matches anime content against AniList and persists
// show-level metadata into series plus episode/movie metadata into media.
// Falls back to TMDb when AniList cannot produce a confident match.
func (s *MetadataService) MatchAndPersistAnime(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
	if !force && media.MetadataLocked {
		return nil
	}
	if !force && media.AnilistID != nil {
		return nil
	}

	anime, err := s.matchAnime(ctx, parsed)
	if err != nil {
		return err
	}
	if anime == nil {
		return s.fallbackAnimeToTMDb(ctx, media, parsed, filePath, libraryID, force)
	}

	switch parsed.MediaType {
	case "episode":
		return s.persistAnimeEpisode(ctx, media, parsed, anime, libraryID)
	default:
		return s.persistAnimeMovie(ctx, media, parsed, anime)
	}
}

func (s *MetadataService) isAnimeLibrary(ctx context.Context, libraryID int64) bool {
	if s.libraryRepo == nil || libraryID <= 0 {
		return false
	}

	lib, err := s.libraryRepo.GetByID(ctx, libraryID)
	if err != nil {
		return false
	}

	return lib.Type == model.LibraryTypeAnime
}

func (s *MetadataService) fallbackAnimeToTMDb(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error {
	if parsed.MediaType == "episode" {
		return s.MatchAndPersistEpisode(ctx, media, parsed, filePath, libraryID, force)
	}
	return s.MatchAndPersistMovie(ctx, media, parsed, filePath, force)
}

func (s *MetadataService) matchAnime(ctx context.Context, parsed nameparser.ParsedMedia) (*anilist.Media, error) {
	if s.anilistClient == nil {
		return nil, nil
	}

	queries := uniqueNonEmptyStrings(parsed.Title, parsed.UnstrippedTitle)
	var (
		best      *anilist.Media
		bestScore float64
	)

	for _, query := range queries {
		page, err := s.anilistClient.SearchAnime(ctx, query, 1, 10)
		if err != nil {
			return nil, err
		}

		for i := range page.Media {
			candidate := &page.Media[i]
			score := scoreAnimeCandidate(parsed, candidate)
			if score > bestScore {
				best = candidate
				bestScore = score
			}
		}
	}

	if bestScore < 0.55 {
		return nil, nil
	}

	return best, nil
}

func (s *MetadataService) persistAnimeMovie(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, anime *anilist.Media) error {
	anilistID := int64(anime.ID)
	media.AnilistID = &anilistID
	media.Title = firstNonEmpty(anime.PreferredTitle(), parsed.Title, media.Title)
	media.SortTitle = media.Title
	media.RomajiTitle = anime.Title.Romaji
	media.Studio = anime.PrimaryStudio()
	media.Overview = cleanAnimeDescription(anime.Description)
	media.ReleaseDate = anime.StartDate.String()
	media.PosterPath = model.PosterPath(bestAnimePoster(anime))
	media.BackdropPath = model.BackdropPath(anime.BannerImage)

	s.enqueueImageForCompute(string(media.PosterPath))
	s.enqueueImageForCompute(string(media.BackdropPath))

	return s.mediaRepo.Update(ctx, media)
}

func (s *MetadataService) persistAnimeEpisode(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, anime *anilist.Media, libraryID int64) error {
	series, err := s.findOrCreateAnimeSeries(ctx, anime, libraryID)
	if err != nil {
		return err
	}

	seasonNumber := parsed.Season
	if seasonNumber <= 0 {
		seasonNumber = 1
	}

	season, err := s.findOrCreateSeason(ctx, series.ID, seasonNumber)
	if err != nil {
		return err
	}

	seasonChanged := false
	if season.Title == "" {
		season.Title = fmt.Sprintf("Season %d", seasonNumber)
		seasonChanged = true
	}
	if season.Overview == "" {
		season.Overview = cleanAnimeDescription(anime.Description)
		seasonChanged = true
	}
	if season.PosterPath == "" {
		season.PosterPath = model.PosterPath(bestAnimePoster(anime))
		seasonChanged = true
	}
	if seasonChanged {
		if err := s.seasonRepo.Update(ctx, season); err != nil {
			log.Printf("update anime season %d: %v", season.ID, err)
		}
	}

	anilistID := int64(anime.ID)
	media.AnilistID = &anilistID
	media.Title = resolveAnimeEpisodeTitle(parsed, anime)
	media.SortTitle = media.Title
	media.RomajiTitle = anime.Title.Romaji
	media.Studio = anime.PrimaryStudio()
	media.Overview = cleanAnimeDescription(anime.Description)
	media.ReleaseDate = anime.StartDate.String()
	media.PosterPath = model.PosterPath(resolveAnimeEpisodeStill(parsed, anime))
	media.BackdropPath = model.BackdropPath(anime.BannerImage)

	s.enqueueImageForCompute(string(media.PosterPath))
	s.enqueueImageForCompute(string(media.BackdropPath))

	if err := s.mediaRepo.Update(ctx, media); err != nil {
		return err
	}

	s.linkEpisode(ctx, media.ID, series.ID, season.ID, parsed.Episode, media.Title, media.Overview, string(media.PosterPath))
	if episode, err := s.episodeRepo.GetByMediaID(ctx, media.ID); err == nil && episode != nil {
		episode.Title = media.Title
		episode.Overview = media.Overview
		episode.StillPath = model.StillPath(media.PosterPath)
		episode.AirDate = media.ReleaseDate
		if updateErr := s.episodeRepo.Update(ctx, episode); updateErr != nil {
			log.Printf("update anime episode %d: %v", episode.ID, updateErr)
		}
	}

	return nil
}

func (s *MetadataService) findOrCreateAnimeSeries(ctx context.Context, anime *anilist.Media, libraryID int64) (*model.Series, error) {
	s.seriesCreateMu.Lock()
	defer s.seriesCreateMu.Unlock()

	anilistID := int64(anime.ID)
	existing, err := s.seriesRepo.GetByAnilistID(ctx, libraryID, anilistID)
	if err == nil && existing != nil {
		if !existing.MetadataLocked {
			s.applyAnimeSeriesMetadata(existing, anime)
			if updateErr := s.seriesRepo.Update(ctx, existing); updateErr != nil {
				log.Printf("update anime series %d: %v", existing.ID, updateErr)
			}
		}
		return existing, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) && !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	series := &model.Series{
		LibraryID: libraryID,
	}
	s.applyAnimeSeriesMetadata(series, anime)
	series.AnilistID = &anilistID

	if err := s.seriesRepo.Create(ctx, series); err != nil {
		return nil, err
	}

	s.enqueueImageForCompute(string(series.PosterPath))
	s.enqueueImageForCompute(string(series.BackdropPath))

	return series, nil
}

func (s *MetadataService) applyAnimeSeriesMetadata(series *model.Series, anime *anilist.Media) {
	series.Title = anime.PreferredTitle()
	series.SortTitle = series.Title
	series.AnilistID = int64Ptr(int64(anime.ID))
	series.RomajiTitle = anime.Title.Romaji
	series.Studio = anime.PrimaryStudio()
	series.Overview = cleanAnimeDescription(anime.Description)
	series.Status = mapAnimeStatus(anime.Status)
	series.FirstAirDate = anime.StartDate.String()
	series.PosterPath = model.PosterPath(bestAnimePoster(anime))
	series.BackdropPath = model.BackdropPath(anime.BannerImage)
}

func scoreAnimeCandidate(parsed nameparser.ParsedMedia, candidate *anilist.Media) float64 {
	score := bestAnimeTitleScore(parsed, candidate)

	if parsed.Year > 0 {
		candidateYear := animeYear(candidate)
		if candidateYear == parsed.Year {
			score += 0.2
		} else if candidateYear > 0 {
			diff := absInt(candidateYear - parsed.Year)
			if diff == 1 {
				score += 0.1
			}
		}
	}

	switch parsed.MediaType {
	case "episode":
		if isEpisodeAnimeFormat(candidate.Format) {
			score += 0.15
		} else if candidate.Format == "MOVIE" {
			score -= 0.2
		}
	case "movie":
		if candidate.Format == "MOVIE" {
			score += 0.2
		} else if isEpisodeAnimeFormat(candidate.Format) {
			score -= 0.1
		}
	}

	return score
}

func bestAnimeTitleScore(parsed nameparser.ParsedMedia, candidate *anilist.Media) float64 {
	queryTitles := uniqueNonEmptyStrings(parsed.Title, parsed.UnstrippedTitle)
	candidateTitles := uniqueNonEmptyStrings(
		candidate.Title.UserPreferred,
		candidate.Title.English,
		candidate.Title.Romaji,
		candidate.Title.Native,
	)

	best := 0.0
	for _, queryTitle := range queryTitles {
		for _, candidateTitle := range candidateTitles {
			score := titleSimilarityScore(queryTitle, candidateTitle)
			if score > best {
				best = score
			}
		}
	}

	return best
}

func titleSimilarityScore(a, b string) float64 {
	na := normalizeAnimeTitle(a)
	nb := normalizeAnimeTitle(b)
	if na == "" || nb == "" {
		return 0
	}
	if na == nb {
		return 0.9
	}
	if strings.Contains(na, nb) || strings.Contains(nb, na) {
		return 0.72
	}

	aTokens := tokenSet(na)
	bTokens := tokenSet(nb)
	if len(aTokens) == 0 || len(bTokens) == 0 {
		return 0
	}

	shared := 0
	for token := range aTokens {
		if bTokens[token] {
			shared++
		}
	}

	union := len(aTokens) + len(bTokens) - shared
	if union == 0 {
		return 0
	}

	return 0.6 * float64(shared) / float64(union)
}

func resolveAnimeEpisodeTitle(parsed nameparser.ParsedMedia, anime *anilist.Media) string {
	if parsed.Episode > 0 && parsed.Episode <= len(anime.StreamingEpisodes) {
		if title := strings.TrimSpace(anime.StreamingEpisodes[parsed.Episode-1].Title); title != "" {
			return title
		}
	}
	if parsed.EpisodeTitle != "" {
		return parsed.EpisodeTitle
	}
	if parsed.Episode > 0 {
		return fmt.Sprintf("Episode %d", parsed.Episode)
	}
	return anime.PreferredTitle()
}

func resolveAnimeEpisodeStill(parsed nameparser.ParsedMedia, anime *anilist.Media) string {
	if parsed.Episode > 0 && parsed.Episode <= len(anime.StreamingEpisodes) {
		if still := strings.TrimSpace(anime.StreamingEpisodes[parsed.Episode-1].Thumbnail); still != "" {
			return still
		}
	}
	return bestAnimePoster(anime)
}

func bestAnimePoster(anime *anilist.Media) string {
	for _, candidate := range []string{
		anime.CoverImage.ExtraLarge,
		anime.CoverImage.Large,
		anime.CoverImage.Medium,
	} {
		if strings.TrimSpace(candidate) != "" {
			return candidate
		}
	}
	return ""
}

func cleanAnimeDescription(description string) string {
	description = html.UnescapeString(description)
	description = animeDescriptionTagPattern.ReplaceAllString(description, " ")
	description = strings.ReplaceAll(description, "\n", " ")
	description = strings.TrimSpace(description)
	return strings.Join(strings.Fields(description), " ")
}

func mapAnimeStatus(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "RELEASING":
		return "Returning Series"
	case "FINISHED":
		return "Ended"
	case "CANCELLED":
		return "Canceled"
	case "HIATUS":
		return "Hiatus"
	case "NOT_YET_RELEASED":
		return "Planned"
	default:
		return status
	}
}

func animeYear(anime *anilist.Media) int {
	if anime.SeasonYear != nil && *anime.SeasonYear > 0 {
		return *anime.SeasonYear
	}
	if anime.StartDate.Year > 0 {
		return anime.StartDate.Year
	}
	return 0
}

func isEpisodeAnimeFormat(format string) bool {
	switch strings.ToUpper(strings.TrimSpace(format)) {
	case "TV", "TV_SHORT", "ONA", "OVA", "SPECIAL":
		return true
	default:
		return false
	}
}

func normalizeAnimeTitle(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	replacer := strings.NewReplacer(
		"&", " and ",
		":", " ",
		"-", " ",
		"_", " ",
		".", " ",
		"/", " ",
	)
	value = replacer.Replace(value)
	return strings.Join(strings.Fields(value), " ")
}

func tokenSet(value string) map[string]bool {
	tokens := strings.Fields(value)
	set := make(map[string]bool, len(tokens))
	for _, token := range tokens {
		set[token] = true
	}
	return set
}

func uniqueNonEmptyStrings(values ...string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, trimmed)
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func int64Ptr(value int64) *int64 {
	return &value
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
