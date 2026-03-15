package service

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/bsplayer"
	"github.com/thawng/velox/pkg/opensubs"
	"github.com/thawng/velox/pkg/podnapisi"
	"github.com/thawng/velox/pkg/subdl"
	"github.com/thawng/velox/pkg/subprovider"
	"github.com/thawng/velox/pkg/subscene"
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

	// OpenSubtitles — disabled: requires VIP subscription for media server apps.
	// Re-enable by uncommenting the block below when a VIP API key is available.
	// osClient, err := s.buildOpenSubsClient(ctx)
	// if err == nil && osClient != nil {
	// 	osParams := opensubs.SearchParams{Query: query, Language: lang}
	// 	if media.ImdbID != nil && *media.ImdbID != "" { osParams.ImdbID = *media.ImdbID }
	// 	if media.TmdbID != nil && *media.TmdbID > 0 { osParams.TmdbID = int(*media.TmdbID) }
	// 	if year := extractYear(media.ReleaseDate); year > 0 { osParams.Year = year }
	// 	osResults, osErr := osClient.Search(ctx, osParams)
	// 	if osErr == nil { results = append(results, osResults...) }
	// }

	// Subdl (if configured)
	subdlClient, err := s.buildSubdlClient(ctx)
	if err != nil {
		log.Printf("subdl not configured: %v", err)
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
		osClient, err := s.buildOpenSubsClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("opensubtitles not configured: %w", err)
		}
		data, filename, err = osClient.Download(ctx, externalID)
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

// buildOpenSubsClient loads credentials from DB and creates a client.
// Note: OpenSubtitles search is currently disabled (requires VIP subscription),
// but download still works if results were previously obtained.
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

	// Check which languages already have text-based subtitles (srt/ass/vtt).
	// Image-based subs (PGS, VOBSUB) can't be used with Direct Play,
	// so they don't count as "having" a subtitle for auto-download purposes.
	existing, err := s.subtitleRepo.ListByMediaFileID(ctx, mediaFileID)
	if err != nil {
		return fmt.Errorf("listing existing subtitles: %w", err)
	}
	haveLang := make(map[string]bool)
	for _, sub := range existing {
		if sub.Language != "" && isTextBasedSubtitle(sub.Codec) {
			haveLang[strings.ToLower(sub.Language)] = true
		}
	}

	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("auto-sub: loading media %d: %w", mediaID, err)
	}

	for _, lang := range targetLangs {
		if haveLang[lang] {
			log.Printf("auto-sub: media %d already has %s subtitle, skipping", mediaID, lang)
			continue
		}

		// Phase 1: Fast API providers (Subdl, Podnapisi, BSPlayer)
		results, err := s.Search(ctx, mediaID, lang)
		if err != nil {
			log.Printf("auto-sub: search failed for media %d lang %s: %v", mediaID, lang, err)
		}

		if len(results) > 0 {
			best := results[0]
			_, err = s.Download(ctx, mediaID, best.Provider, best.ExternalID, lang)
			if err != nil {
				log.Printf("auto-sub: download failed for media %d lang %s: %v", mediaID, lang, err)
			} else {
				log.Printf("auto-sub: downloaded %s subtitle for media %d from %s", lang, mediaID, best.Provider)
				continue
			}
		}

		// Phase 2: Subscene scraper (slow, DrissionPage — background only)
		scQuery := media.Title
		scSeason := 0
		if epInfo := s.resolveEpisodeInfo(ctx, mediaID); epInfo != nil {
			scQuery = epInfo.seriesTitle
			scSeason = epInfo.seasonNumber
		}
		log.Printf("auto-sub: trying subscene for media %d lang %s query %q season %d", mediaID, lang, scQuery, scSeason)
		scScraper := subscene.New()
		scResults, scErr := scScraper.Search(ctx, subscene.SearchParams{
			Query:    scQuery,
			Language: lang,
			Season:   scSeason,
		})
		if scErr != nil {
			log.Printf("auto-sub: subscene search error for media %d: %v", mediaID, scErr)
			continue
		}
		if len(scResults) == 0 {
			log.Printf("auto-sub: no %s subtitles found on subscene for media %d", lang, mediaID)
			continue
		}

		best := scResults[0]
		_, err = s.Download(ctx, mediaID, best.Provider, best.ExternalID, lang)
		if err != nil {
			log.Printf("auto-sub: subscene download failed for media %d lang %s: %v", mediaID, lang, err)
			continue
		}
		log.Printf("auto-sub: downloaded %s subtitle for media %d from subscene", lang, mediaID)
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

func buildSubdlSearchParams(
	media *model.Media,
	epInfo *episodeInfo,
	lang string,
	fileName string,
) subdl.SearchParams {
	params := subdl.SearchParams{
		FilmName: media.Title,
		Language: lang,
		FileName: fileName,
	}

	if epInfo != nil {
		params.FilmName = epInfo.seriesTitle
		params.SeasonNumber = epInfo.seasonNumber
		params.EpisodeNumber = epInfo.episodeNumber
		params.Type = "tv"
		params.ImdbID = epInfo.seriesImdbID
		params.TmdbID = epInfo.seriesTmdbID
	} else {
		if media.ImdbID != nil && *media.ImdbID != "" {
			params.ImdbID = *media.ImdbID
		}
		if media.TmdbID != nil && *media.TmdbID > 0 {
			params.TmdbID = int(*media.TmdbID)
		}
	}

	if year := extractYear(media.ReleaseDate); year > 0 {
		params.Year = year
	}

	// When no canonical IDs exist, fall back to filename-driven search immediately.
	if !shouldFallbackToSubdlFileNameSearch(media, epInfo) && params.FileName == "" {
		params.FileName = fileName
	}
	if !hasCanonicalSubdlIDs(media, epInfo) && params.FileName == "" {
		params.FileName = media.Title
	}

	return params
}

func shouldFallbackToSubdlFileNameSearch(media *model.Media, epInfo *episodeInfo) bool {
	return hasCanonicalSubdlIDs(media, epInfo)
}

func hasCanonicalSubdlIDs(media *model.Media, epInfo *episodeInfo) bool {
	if epInfo != nil {
		return epInfo.seriesImdbID != "" || epInfo.seriesTmdbID > 0
	}
	return (media.ImdbID != nil && *media.ImdbID != "") ||
		(media.TmdbID != nil && *media.TmdbID > 0)
}

var subtitleSearchTokenSplitter = regexp.MustCompile(`[^a-z0-9]+`)

func filterAndRankSubtitleSearchResults(
	results []subprovider.Result,
	epInfo *episodeInfo,
	query string,
) []subprovider.Result {
	if len(results) == 0 || epInfo == nil {
		return results
	}

	filtered := make([]subprovider.Result, 0, len(results))
	for _, result := range results {
		if !isExactEpisodeSubtitleMatch(result.Title, epInfo.seasonNumber, epInfo.episodeNumber) {
			continue
		}
		filtered = append(filtered, result)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		left := subtitleSearchScore(filtered[i], query)
		right := subtitleSearchScore(filtered[j], query)
		if left != right {
			return left > right
		}
		return filtered[i].Title < filtered[j].Title
	})

	return filtered
}

func isExactEpisodeSubtitleMatch(title string, seasonNumber, episodeNumber int) bool {
	normalized := normalizeSubtitleSearchText(title)
	if normalized == "" {
		return false
	}

	season := strconv.Itoa(seasonNumber)
	episode := strconv.Itoa(episodeNumber)
	seasonPadded := fmt.Sprintf("%02d", seasonNumber)
	episodePadded := fmt.Sprintf("%02d", episodeNumber)

	exactTokens := []string{
		"s" + season + "e" + episode,
		"s" + seasonPadded + "e" + episodePadded,
		season + "x" + episodePadded,
		seasonPadded + "x" + episodePadded,
		"season " + season + " episode " + episode,
		"season " + seasonPadded + " episode " + episodePadded,
	}
	for _, token := range exactTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func subtitleSearchScore(result subprovider.Result, query string) int {
	score := subtitleSearchOverlapScore(query, result.Title) * 100
	score += result.Downloads
	if !result.HearingImpaired {
		score += 25
	}
	switch result.Provider {
	case "subdl":
		score += 10
	case "podnapisi":
		score += 5
	case "bsplayer":
		score += 3
	}
	return score
}

func subtitleSearchOverlapScore(query, title string) int {
	queryTokens := subtitleSearchTokens(query)
	if len(queryTokens) == 0 {
		return 0
	}
	titleTokens := subtitleSearchTokens(title)
	overlap := 0
	for token := range titleTokens {
		if _, ok := queryTokens[token]; ok {
			overlap++
		}
	}
	return overlap
}

func subtitleSearchTokens(value string) map[string]struct{} {
	normalized := normalizeSubtitleSearchText(value)
	parts := subtitleSearchTokenSplitter.Split(normalized, -1)
	tokens := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		tokens[part] = struct{}{}
	}
	return tokens
}

func normalizeSubtitleSearchText(value string) string {
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}

// isTextBasedSubtitle returns true for text-based subtitle codecs (srt, ass, vtt).
// Image-based codecs (PGS, VOBSUB) return false — they can't be used with Direct Play.
func isTextBasedSubtitle(codec string) bool {
	switch strings.ToLower(codec) {
	case "subrip", "srt", "ass", "ssa", "webvtt", "vtt", "mov_text", "text":
		return true
	}
	return false
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
