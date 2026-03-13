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
	"github.com/thawng/velox/pkg/subprovider"
)

// SubtitleSearchService orchestrates subtitle search across external providers.
type SubtitleSearchService struct {
	mediaRepo    *repository.MediaRepo
	mfRepo       *repository.MediaFileRepo
	subtitleRepo *repository.SubtitleRepo
	settingsRepo *repository.AppSettingsRepo
	podClient    *podnapisi.Client
	downloadDir  string // e.g. ~/.velox/subtitles/downloaded
}

// NewSubtitleSearchService creates a new subtitle search service.
func NewSubtitleSearchService(
	mediaRepo *repository.MediaRepo,
	mfRepo *repository.MediaFileRepo,
	subtitleRepo *repository.SubtitleRepo,
	settingsRepo *repository.AppSettingsRepo,
	downloadDir string,
) *SubtitleSearchService {
	return &SubtitleSearchService{
		mediaRepo:    mediaRepo,
		mfRepo:       mfRepo,
		subtitleRepo: subtitleRepo,
		settingsRepo: settingsRepo,
		podClient:    podnapisi.New(),
		downloadDir:  downloadDir,
	}
}

// Search queries both providers for subtitles matching the given media and language.
// Uses the original video filename (without extension) as the search query — providers
// match best on release names which are closer to the filename than the parsed title.
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

	// Podnapisi (always available)
	podParams := podnapisi.SearchParams{
		Keywords: query,
		Language: lang,
	}
	if year := extractYear(media.ReleaseDate); year > 0 {
		podParams.Year = year
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
func (s *SubtitleSearchService) Download(ctx context.Context, mediaID int64, provider, externalID string) (*model.Subtitle, error) {
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

	saveName := fmt.Sprintf("%s_%s%s", provider, externalID, ext)
	savePath := filepath.Join(dir, saveName)
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		return nil, fmt.Errorf("writing subtitle file: %w", err)
	}

	// Create DB record
	sub := &model.Subtitle{
		MediaFileID: mf.ID,
		Language:    "", // will be populated from search result by caller if needed
		Codec:       codec,
		Title:       fmt.Sprintf("%s (%s)", filename, provider),
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
