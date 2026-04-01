package service

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/pkg/opensubs"
	"github.com/thawng/velox/pkg/subdl"
	"github.com/thawng/velox/pkg/subscene"
)

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
// Falls back to the built-in env key if none is configured.
func (s *SubtitleSearchService) buildSubdlClient(ctx context.Context) (*subdl.Client, error) {
	apiKey, _ := s.settingsRepo.Get(ctx, model.SettingSubdlAPIKey)
	if apiKey == "" {
		apiKey = s.builtinSubdlKey
	}
	if apiKey == "" {
		return nil, fmt.Errorf("no SubDL API key configured")
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
