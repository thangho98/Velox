package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/bsplayer"
	"github.com/thawng/velox/pkg/opensubs"
	"github.com/thawng/velox/pkg/opensubslegacy"
	"github.com/thawng/velox/pkg/podnapisi"
	"github.com/thawng/velox/pkg/subprovider"
	"github.com/thawng/velox/pkg/subscene"
)

// SubtitleSearchService orchestrates subtitle search across external providers.
type SubtitleSearchService struct {
	mediaRepo       *repository.MediaRepo
	mfRepo          *repository.MediaFileRepo
	subtitleRepo    *repository.SubtitleRepo
	settingsRepo    *repository.AppSettingsRepo
	prefsRepo       *repository.UserPreferencesRepo
	episodeRepo     *repository.EpisodeRepo
	seasonRepo      *repository.SeasonRepo
	seriesRepo      *repository.SeriesRepo
	downloadDir     string // e.g. ~/.velox/subtitles/downloaded
	builtinSubdlKey string // VELOX_SUBDL_API_KEY from env (optional)
	notificationSvc *NotificationService
}

// NewSubtitleSearchService creates a new subtitle search service.
func NewSubtitleSearchService(
	mediaRepo *repository.MediaRepo,
	mfRepo *repository.MediaFileRepo,
	subtitleRepo *repository.SubtitleRepo,
	settingsRepo *repository.AppSettingsRepo,
	prefsRepo *repository.UserPreferencesRepo,
	episodeRepo *repository.EpisodeRepo,
	seasonRepo *repository.SeasonRepo,
	seriesRepo *repository.SeriesRepo,
	downloadDir string,
) *SubtitleSearchService {
	return &SubtitleSearchService{
		mediaRepo:    mediaRepo,
		mfRepo:       mfRepo,
		subtitleRepo: subtitleRepo,
		settingsRepo: settingsRepo,
		prefsRepo:    prefsRepo,
		episodeRepo:  episodeRepo,
		seasonRepo:   seasonRepo,
		seriesRepo:   seriesRepo,
		downloadDir:  downloadDir,
	}
}

// SetBuiltinSubdlKey sets the built-in Subdl API key from env (optional).
// This allows open-source distributions to provide a default key.
func (s *SubtitleSearchService) SetBuiltinSubdlKey(key string) {
	s.builtinSubdlKey = key
}

// SetNotificationService sets the optional notification service for subtitle events.
func (s *SubtitleSearchService) SetNotificationService(svc *NotificationService) {
	s.notificationSvc = svc
}

// episodeInfo holds series-level metadata resolved from an episode media item.
type episodeInfo struct {
	seriesTitle   string
	seriesTmdbID  int
	seriesImdbID  string
	seasonNumber  int
	episodeNumber int
}

// resolveEpisodeInfo looks up series-level metadata for an episode media item.
// Returns nil if the media is not an episode or lookup fails.
func (s *SubtitleSearchService) resolveEpisodeInfo(ctx context.Context, mediaID int64) *episodeInfo {
	ep, err := s.episodeRepo.GetByMediaID(ctx, mediaID)
	if err != nil {
		return nil
	}

	season, err := s.seasonRepo.GetByID(ctx, ep.SeasonID)
	if err != nil {
		return nil
	}

	series, err := s.seriesRepo.GetByID(ctx, ep.SeriesID)
	if err != nil {
		return nil
	}

	info := &episodeInfo{
		seriesTitle:   series.Title,
		seasonNumber:  season.SeasonNumber,
		episodeNumber: ep.EpisodeNumber,
	}
	if series.TmdbID != nil && *series.TmdbID > 0 {
		info.seriesTmdbID = int(*series.TmdbID)
	}
	if series.ImdbID != nil && *series.ImdbID != "" {
		info.seriesImdbID = *series.ImdbID
	}
	return info
}

// Search queries all providers for subtitles matching the given media and language.
// Uses the original video filename (without extension) as the search query — providers
// match best on release names which are closer to the filename than the parsed title.
// For episodes, resolves series-level tmdb_id and passes season/episode numbers.
func (s *SubtitleSearchService) Search(ctx context.Context, mediaID int64, lang string) ([]subprovider.Result, error) {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("loading media %d: %w", mediaID, err)
	}

	// Prefer original video filename over parsed title for search accuracy
	query := media.Title
	mf, err := s.mfRepo.GetPrimaryByMediaID(ctx, mediaID)
	if err == nil && mf != nil {
		base := filepath.Base(mf.FilePath)
		query = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// For episodes, resolve series-level metadata (series tmdb_id, season/episode numbers)
	var epInfo *episodeInfo
	if media.MediaType == "episode" {
		epInfo = s.resolveEpisodeInfo(ctx, mediaID)
	}

	var results []subprovider.Result

	// OpenSubtitles — use REST v2 if user configured API key, otherwise use legacy API (no key needed)
	osClient, osErr := s.buildOpenSubsClient(ctx)
	if osErr == nil && osClient != nil {
		// REST v2 (user has configured API key + credentials)
		osParams := opensubs.SearchParams{Query: query, Language: lang}
		if media.ImdbID != nil && *media.ImdbID != "" {
			osParams.ImdbID = *media.ImdbID
		}
		if media.TmdbID != nil && *media.TmdbID > 0 {
			osParams.TmdbID = int(*media.TmdbID)
		}
		if year := extractYear(media.ReleaseDate); year > 0 {
			osParams.Year = year
		}
		osResults, err := osClient.Search(ctx, osParams)
		if err != nil {
			log.Printf("opensubtitles v2 search error: %v", err)
		} else {
			results = append(results, osResults...)
		}
	} else {
		// Legacy API (no key needed, search by IMDB ID)
		imdbID := ""
		if media.ImdbID != nil {
			imdbID = *media.ImdbID
		}
		if epInfo != nil && epInfo.seriesImdbID != "" {
			imdbID = epInfo.seriesImdbID
		}
		if imdbID != "" {
			legacyParams := opensubslegacy.SearchParams{
				ImdbID:   imdbID,
				Language: lang,
			}
			if epInfo != nil {
				legacyParams.SeasonNumber = epInfo.seasonNumber
				legacyParams.EpisodeNumber = epInfo.episodeNumber
			}
			legacyResults, err := opensubslegacy.New().Search(ctx, legacyParams)
			if err != nil {
				log.Printf("opensubtitles legacy search error: %v", err)
			} else {
				results = append(results, legacyResults...)
			}
		}
	}

	// Subdl (if configured)
	subdlClient, err := s.buildSubdlClient(ctx)
	if err != nil {
		log.Printf("subdl not configured: %v", err)
	} else {
		log.Printf("subdl client ready (key present)")
	}
	if subdlClient != nil {
		sdParams := buildSubdlSearchParams(media, epInfo, lang, "")
		sdResults, err := subdlClient.Search(ctx, sdParams)
		if err != nil {
			log.Printf("subdl search error: %v", err)
		}
		if len(sdResults) == 0 && shouldFallbackToSubdlFileNameSearch(media, epInfo) {
			fallbackParams := buildSubdlSearchParams(media, epInfo, lang, query)
			fallbackResults, fallbackErr := subdlClient.Search(ctx, fallbackParams)
			if fallbackErr != nil {
				log.Printf("subdl fallback search error: %v", fallbackErr)
			} else {
				sdResults = fallbackResults
			}
		}
		if len(sdResults) > 0 {
			results = append(results, sdResults...)
		}
	}

	// Podnapisi (no API key needed)
	podClient := podnapisi.New()
	podParams := podnapisi.SearchParams{
		Keywords: query,
		Language: lang,
	}
	if year := extractYear(media.ReleaseDate); year > 0 {
		podParams.Year = year
	}
	if epInfo != nil {
		podParams.Season = epInfo.seasonNumber
		podParams.Episode = epInfo.episodeNumber
		podParams.Keywords = epInfo.seriesTitle
	}
	podResults, err := podClient.Search(ctx, podParams)
	if err != nil {
		log.Printf("podnapisi search error: %v", err)
	} else {
		results = append(results, podResults...)
	}

	// BSPlayer (no API key needed, searches by IMDB ID)
	if media.ImdbID != nil && *media.ImdbID != "" {
		bsClient := bsplayer.New()
		bsParams := bsplayer.SearchParams{
			ImdbID:   *media.ImdbID,
			Language: lang,
		}
		if epInfo != nil && epInfo.seriesImdbID != "" {
			bsParams.ImdbID = epInfo.seriesImdbID
		}
		if mf != nil {
			bsParams.FileSize = mf.FileSize
		}
		bsResults, err := bsClient.Search(ctx, bsParams)
		if err != nil {
			log.Printf("bsplayer search error: %v", err)
		} else {
			results = append(results, bsResults...)
		}
	}

	if results == nil {
		results = []subprovider.Result{}
	}
	results = filterAndRankSubtitleSearchResults(results, epInfo, query)
	return results, nil
}

// Download fetches a subtitle from the given provider and saves it to disk + DB.
// If language is non-empty, it is stored on the subtitle record.
func (s *SubtitleSearchService) Download(ctx context.Context, mediaID int64, provider, externalID, language string) (*model.Subtitle, error) {
	// Get primary file for this media
	mf, err := s.mfRepo.GetPrimaryByMediaID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("getting primary file for media %d: %w", mediaID, err)
	}

	var data []byte
	var filename string

	switch provider {
	case "opensubtitles":
		// Try REST v2 first (if configured), fall back to legacy API
		osClient, osErr := s.buildOpenSubsClient(ctx)
		if osErr == nil && osClient != nil {
			data, filename, err = osClient.Download(ctx, externalID)
		} else {
			data, filename, err = opensubslegacy.New().Download(ctx, externalID)
		}
		if err != nil {
			return nil, fmt.Errorf("downloading from opensubtitles: %w", err)
		}

	case "subdl":
		sdClient, err := s.buildSubdlClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("subdl not configured: %w", err)
		}
		data, filename, err = sdClient.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from subdl: %w", err)
		}

	case "podnapisi":
		podClient := podnapisi.New()
		data, filename, err = podClient.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from podnapisi: %w", err)
		}

	case "bsplayer":
		bsClient := bsplayer.New()
		data, filename, err = bsClient.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from bsplayer: %w", err)
		}

	case "subscene":
		scScraper := subscene.New()
		data, filename, err = scScraper.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from subscene: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}

	// Determine format from filename
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ".srt"
	}
	codec := strings.TrimPrefix(ext, ".")

	// Save file to disk
	dir := filepath.Join(s.downloadDir, strconv.FormatInt(mediaID, 10))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating download dir: %w", err)
	}

	// Sanitize externalID for use as filename (may contain slashes or .zip suffix)
	safeID := strings.ReplaceAll(externalID, "/", "_")
	safeID = strings.TrimSuffix(safeID, ".zip")
	saveName := fmt.Sprintf("%s_%s%s", provider, safeID, ext)
	savePath := filepath.Join(dir, saveName)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return nil, fmt.Errorf("writing subtitle file: %w", err)
	}

	// Create DB record
	sub := &model.Subtitle{
		MediaFileID: mf.ID,
		Language:    language,
		Codec:       codec,
		Title:       subtitleTitle(language, provider),
		IsEmbedded:  false,
		StreamIndex: -1,
		FilePath:    savePath,
		IsForced:    false,
		IsDefault:   false,
		IsSDH:       false,
	}

	if err := s.subtitleRepo.Create(ctx, sub); err != nil {
		return nil, fmt.Errorf("creating subtitle record: %w", err)
	}

	return sub, nil
}

// AutoDownloadBestMatch searches for subtitles matching the given language and
// downloads the highest-ranked result for an explicit user request.
func (s *SubtitleSearchService) AutoDownloadBestMatch(ctx context.Context, mediaID int64, language string) (*model.Subtitle, error) {
	// First search for available subtitles
	results, err := s.Search(ctx, mediaID, language)
	if err != nil {
		return nil, fmt.Errorf("searching subtitles: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no %s subtitles found for media %d", language, mediaID)
	}

	// Results are already ranked, so taking the first one yields the best match
	bestMatch := results[0]

	// Download the best match
	sub, err := s.Download(ctx, mediaID, bestMatch.Provider, bestMatch.ExternalID, language)
	if err != nil {
		return nil, fmt.Errorf("downloading best subtitle (provider: %s, external_id: %s): %w", bestMatch.Provider, bestMatch.ExternalID, err)
	}

	return sub, nil
}

// AutoDownloadForLanguages searches and downloads the best match for each requested language.
func (s *SubtitleSearchService) AutoDownloadForLanguages(ctx context.Context, mediaID int64, languages []string) ([]*model.Subtitle, error) {
	var downloaded []*model.Subtitle
	var firstErr error

	for _, lang := range languages {
		lang = strings.TrimSpace(lang)
		if lang == "" {
			continue
		}
		sub, err := s.AutoDownloadBestMatch(ctx, mediaID, lang)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			log.Printf("failed to auto download %s for media %d: %v", lang, mediaID, err)
		} else {
			downloaded = append(downloaded, sub)
		}
	}

	if len(downloaded) == 0 && firstErr != nil {
		return nil, firstErr
	}

	return downloaded, nil
}

// AutoDownloadForUser uses the user's configured subtitle preferences to download subtitles.
func (s *SubtitleSearchService) AutoDownloadForUser(ctx context.Context, mediaID int64, userID int64) ([]*model.Subtitle, error) {
	prefs, err := s.prefsRepo.Get(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("getting preferences: %w", err)
	}

	var langs []string
	if prefs.SubtitleLanguage != "" {
		langs = strings.Split(prefs.SubtitleLanguage, ",")
	}
	if len(langs) == 0 {
		langs = []string{"vi", "en"} // Default to common if nothing configured
	}

	return s.AutoDownloadForLanguages(ctx, mediaID, langs)
}
