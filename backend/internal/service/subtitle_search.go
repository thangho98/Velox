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
	"github.com/thawng/velox/pkg/opensubs"
	"github.com/thawng/velox/pkg/podnapisi"
	"github.com/thawng/velox/pkg/subdl"
	"github.com/thawng/velox/pkg/subprovider"
)

// SubtitleSearchService orchestrates subtitle search across external providers.
type SubtitleSearchService struct {
	mediaRepo    *repository.MediaRepo
	mfRepo       *repository.MediaFileRepo
	subtitleRepo *repository.SubtitleRepo
	settingsRepo *repository.AppSettingsRepo
	episodeRepo  *repository.EpisodeRepo
	seasonRepo   *repository.SeasonRepo
	seriesRepo   *repository.SeriesRepo
	podClient    *podnapisi.Client
	downloadDir  string // e.g. ~/.velox/subtitles/downloaded
}

// NewSubtitleSearchService creates a new subtitle search service.
func NewSubtitleSearchService(
	mediaRepo *repository.MediaRepo,
	mfRepo *repository.MediaFileRepo,
	subtitleRepo *repository.SubtitleRepo,
	settingsRepo *repository.AppSettingsRepo,
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
		episodeRepo:  episodeRepo,
		seasonRepo:   seasonRepo,
		seriesRepo:   seriesRepo,
		podClient:    podnapisi.New(),
		downloadDir:  downloadDir,
	}
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

	// OpenSubtitles (if configured)
	osClient, err := s.buildOpenSubsClient(ctx)
	if err != nil {
		log.Printf("opensubtitles not configured: %v", err)
	}
	if osClient != nil {
		osParams := opensubs.SearchParams{
			Query:    query,
			Language: lang,
		}
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
			log.Printf("opensubtitles search error: %v", err)
		} else {
			results = append(results, osResults...)
		}
	}

	// Subdl (if configured)
	subdlClient, err := s.buildSubdlClient(ctx)
	if err != nil {
		log.Printf("subdl not configured: %v", err)
	}
	if subdlClient != nil {
		sdParams := subdl.SearchParams{
			FilmName: media.Title,
			FileName: query,
			Language: lang,
		}
		// For episodes, use series-level IDs and pass season/episode numbers
		if epInfo != nil {
			sdParams.FilmName = epInfo.seriesTitle
			if epInfo.seriesTmdbID > 0 {
				sdParams.TmdbID = epInfo.seriesTmdbID
			}
			if epInfo.seriesImdbID != "" {
				sdParams.ImdbID = epInfo.seriesImdbID
			}
			sdParams.SeasonNumber = epInfo.seasonNumber
			sdParams.EpisodeNumber = epInfo.episodeNumber
			sdParams.Type = "tv"
		} else {
			if media.ImdbID != nil && *media.ImdbID != "" {
				sdParams.ImdbID = *media.ImdbID
			}
			if media.TmdbID != nil && *media.TmdbID > 0 {
				sdParams.TmdbID = int(*media.TmdbID)
			}
		}
		if year := extractYear(media.ReleaseDate); year > 0 {
			sdParams.Year = year
		}

		sdResults, err := subdlClient.Search(ctx, sdParams)
		if err != nil {
			log.Printf("subdl search error: %v", err)
		} else {
			results = append(results, sdResults...)
		}
	}

	// Podnapisi (always available)
	podParams := podnapisi.SearchParams{
		Keywords: query,
		Language: lang,
	}
	if year := extractYear(media.ReleaseDate); year > 0 {
		podParams.Year = year
	}
	// For episodes, pass season/episode to Podnapisi too
	if epInfo != nil {
		podParams.Season = epInfo.seasonNumber
		podParams.Episode = epInfo.episodeNumber
	}

	podResults, err := s.podClient.SearchJSON(ctx, podParams)
	if err != nil {
		log.Printf("podnapisi search error: %v", err)
	} else {
		results = append(results, podResults...)
	}

	if results == nil {
		results = []subprovider.Result{}
	}
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
		osClient, err := s.buildOpenSubsClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("opensubtitles not configured: %w", err)
		}
		data, filename, err = osClient.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from opensubtitles: %w", err)
		}

	case "podnapisi":
		data, filename, err = s.podClient.Download(ctx, externalID)
		if err != nil {
			return nil, fmt.Errorf("downloading from podnapisi: %w", err)
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

// buildOpenSubsClient loads credentials from DB and creates a client.
func (s *SubtitleSearchService) buildOpenSubsClient(ctx context.Context) (*opensubs.Client, error) {
	vals, err := s.settingsRepo.GetMulti(ctx,
		model.SettingOpenSubsAPIKey,
		model.SettingOpenSubsUsername,
		model.SettingOpenSubsPassword,
	)
	if err != nil {
		return nil, fmt.Errorf("loading opensubs settings: %w", err)
	}

	apiKey := vals[model.SettingOpenSubsAPIKey]
	username := vals[model.SettingOpenSubsUsername]
	password := vals[model.SettingOpenSubsPassword]

	if apiKey == "" || username == "" || password == "" {
		return nil, fmt.Errorf("incomplete credentials")
	}

	return opensubs.New(apiKey, username, password), nil
}

// buildSubdlClient loads the API key from DB and creates a client.
// Falls back to the built-in key if none is configured.
func (s *SubtitleSearchService) buildSubdlClient(ctx context.Context) (*subdl.Client, error) {
	apiKey, _ := s.settingsRepo.Get(ctx, model.SettingSubdlAPIKey)
	if apiKey == "" {
		apiKey = subdl.DefaultAPIKey
	}
	return subdl.New(apiKey), nil
}

// AutoDownload fetches subtitles for configured languages if the media file
// doesn't already have them (embedded or external). Designed to be called
// from the scan pipeline — non-critical, errors are logged but not fatal.
func (s *SubtitleSearchService) AutoDownload(ctx context.Context, mediaID, mediaFileID int64) error {
	// Check configured languages
	langsStr, err := s.settingsRepo.Get(ctx, model.SettingAutoSubLanguages)
	if err != nil || langsStr == "" {
		return nil // not configured or disabled
	}

	var targetLangs []string
	for _, l := range strings.Split(langsStr, ",") {
		l = strings.TrimSpace(strings.ToLower(l))
		if l != "" {
			targetLangs = append(targetLangs, l)
		}
	}
	if len(targetLangs) == 0 {
		return nil
	}

	// Check which languages already exist (embedded + sidecar)
	existing, err := s.subtitleRepo.ListByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return fmt.Errorf("listing existing subtitles: %w", err)
	}
	haveLang := make(map[string]bool)
	for _, sub := range existing {
		if sub.Language != "" {
			haveLang[strings.ToLower(sub.Language)] = true
		}
	}

	for _, lang := range targetLangs {
		if haveLang[lang] {
			log.Printf("auto-sub: media %d already has %s subtitle, skipping", mediaID, lang)
			continue
		}

		results, err := s.Search(ctx, mediaID, lang)
		if err != nil {
			log.Printf("auto-sub: search failed for media %d lang %s: %v", mediaID, lang, err)
			continue
		}
		if len(results) == 0 {
			log.Printf("auto-sub: no %s subtitles found for media %d", lang, mediaID)
			continue
		}

		// Pick the first (best) result
		best := results[0]
		_, err = s.Download(ctx, mediaID, best.Provider, best.ExternalID, lang)
		if err != nil {
			log.Printf("auto-sub: download failed for media %d lang %s: %v", mediaID, lang, err)
			continue
		}

		log.Printf("auto-sub: downloaded %s subtitle for media %d from %s", lang, mediaID, best.Provider)
	}

	return nil
}

// subtitleTitle builds a human-readable title for a downloaded subtitle.
func subtitleTitle(langCode, provider string) string {
	names := map[string]string{
		"en": "English", "vi": "Vietnamese", "fr": "French", "de": "German",
		"es": "Spanish", "pt": "Portuguese", "it": "Italian", "ja": "Japanese",
		"ko": "Korean", "zh": "Chinese", "nl": "Dutch", "pl": "Polish",
		"ru": "Russian", "ar": "Arabic", "tr": "Turkish", "sv": "Swedish",
		"th": "Thai", "id": "Indonesian",
	}
	name := names[strings.ToLower(langCode)]
	if name == "" {
		name = langCode
	}
	return fmt.Sprintf("%s (%s)", name, provider)
}

// extractYear parses "2023-01-15" → 2023
func extractYear(releaseDate string) int {
	if len(releaseDate) >= 4 {
		year, err := strconv.Atoi(releaseDate[:4])
		if err == nil {
			return year
		}
	}
	return 0
}
