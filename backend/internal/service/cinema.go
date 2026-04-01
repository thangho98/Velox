package service

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/tmdb"
)

const (
	defaultCinemaMaxTrailers = 2
	settingCinemaModeEnabled = "cinema_mode_enabled"
	settingCinemaMaxTrailers = "cinema_max_trailers"
	settingCinemaIntroPath   = "cinema_intro_path"
)

// CinemaService orchestrates cinema-mode settings and trailer intro playback.
type CinemaService struct {
	appSettingsRepo *repository.AppSettingsRepo
	mediaRepo       *repository.MediaRepo
	episodeRepo     *repository.EpisodeRepo
	seriesRepo      *repository.SeriesRepo
	tmdbClient      *tmdb.Client
	dataDir         string
}

// CinemaSettings is the response shape for cinema-mode settings.
type CinemaSettings struct {
	Enabled     bool   `json:"enabled"`
	MaxTrailers string `json:"max_trailers"`
	HasIntro    bool   `json:"has_intro"`
}

// CinemaSettingsUpdate is the writable cinema-mode settings payload.
type CinemaSettingsUpdate struct {
	Enabled     *bool
	MaxTrailers *string
}

// CinemaItem represents an intro or trailer playable before media starts.
type CinemaItem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
	Skippable bool   `json:"skippable"`
}

// NewCinemaService creates a new cinema service.
func NewCinemaService(
	appSettingsRepo *repository.AppSettingsRepo,
	mediaRepo *repository.MediaRepo,
	episodeRepo *repository.EpisodeRepo,
	seriesRepo *repository.SeriesRepo,
	tmdbClient *tmdb.Client,
	dataDir string,
) *CinemaService {
	return &CinemaService{
		appSettingsRepo: appSettingsRepo,
		mediaRepo:       mediaRepo,
		episodeRepo:     episodeRepo,
		seriesRepo:      seriesRepo,
		tmdbClient:      tmdbClient,
		dataDir:         dataDir,
	}
}

func (s *CinemaService) GetSettings(ctx context.Context) (*CinemaSettings, error) {
	enabled, err := s.appSettingsRepo.Get(ctx, settingCinemaModeEnabled)
	if err != nil {
		return nil, err
	}

	maxTrailers, err := s.appSettingsRepo.Get(ctx, settingCinemaMaxTrailers)
	if err != nil {
		return nil, err
	}

	introPath, err := s.appSettingsRepo.Get(ctx, settingCinemaIntroPath)
	if err != nil {
		return nil, err
	}

	if maxTrailers == "" {
		maxTrailers = strconv.Itoa(defaultCinemaMaxTrailers)
	}

	return &CinemaSettings{
		Enabled:     enabled == "true",
		MaxTrailers: maxTrailers,
		HasIntro:    introPath != "",
	}, nil
}

func (s *CinemaService) UpdateSettings(ctx context.Context, update CinemaSettingsUpdate) error {
	if update.Enabled != nil {
		value := "false"
		if *update.Enabled {
			value = "true"
		}
		if err := s.appSettingsRepo.Set(ctx, settingCinemaModeEnabled, value); err != nil {
			return err
		}
	}

	if update.MaxTrailers != nil {
		if err := s.appSettingsRepo.Set(ctx, settingCinemaMaxTrailers, *update.MaxTrailers); err != nil {
			return err
		}
	}

	return nil
}

func (s *CinemaService) MediaItems(ctx context.Context, mediaID int64) ([]CinemaItem, error) {
	items := make([]CinemaItem, 0, defaultCinemaMaxTrailers+1)

	cinemaEnabled, _ := s.appSettingsRepo.Get(ctx, settingCinemaModeEnabled)
	introPath, _ := s.appSettingsRepo.Get(ctx, settingCinemaIntroPath)
	if introPath != "" && cinemaEnabled == "true" {
		items = append(items, CinemaItem{
			Type:      "intro",
			Title:     "Cinema Intro",
			URL:       "/api/cinema/intro",
			Skippable: true,
		})
	}

	maxTrailers := s.trailerLimit(ctx)
	if s.tmdbClient != nil && maxTrailers > 0 {
		items = append(items, trailerItems(s.mediaVideos(ctx, mediaID), maxTrailers)...)
	}

	return items, nil
}

func (s *CinemaService) SeriesItems(ctx context.Context, seriesID int64) ([]CinemaItem, error) {
	maxTrailers := s.trailerLimit(ctx)
	if s.tmdbClient == nil || maxTrailers <= 0 {
		return []CinemaItem{}, nil
	}

	return trailerItems(s.seriesVideos(ctx, seriesID), maxTrailers), nil
}

func (s *CinemaService) GetIntroPath(ctx context.Context) (string, error) {
	introPath, err := s.appSettingsRepo.Get(ctx, settingCinemaIntroPath)
	if err != nil {
		return "", err
	}
	if introPath == "" {
		return "", ErrNotFound
	}
	return introPath, nil
}

func (s *CinemaService) SaveIntro(ctx context.Context, src io.Reader) (string, error) {
	cinemaDir := filepath.Join(s.dataDir, "cinema")
	if err := os.MkdirAll(cinemaDir, 0755); err != nil {
		return "", err
	}

	introPath := filepath.Join(cinemaDir, "intro.mp4")
	dst, err := os.Create(introPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	if err := s.appSettingsRepo.Set(ctx, settingCinemaIntroPath, introPath); err != nil {
		return "", err
	}

	return introPath, nil
}

func (s *CinemaService) trailerLimit(ctx context.Context) int {
	maxTrailers := defaultCinemaMaxTrailers
	if value, _ := s.appSettingsRepo.Get(ctx, settingCinemaMaxTrailers); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			maxTrailers = parsed
		}
	}
	return maxTrailers
}

func (s *CinemaService) mediaVideos(ctx context.Context, mediaID int64) *tmdb.VideoList {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil
	}

	if media.MediaType == "episode" {
		episode, err := s.episodeRepo.GetByMediaID(ctx, media.ID)
		if err != nil {
			return nil
		}

		series, err := s.seriesRepo.GetByID(ctx, episode.SeriesID)
		if err != nil || series.TmdbID == nil {
			return nil
		}

		details, err := s.tmdbClient.GetTVDetails(ctx, int(*series.TmdbID))
		if err != nil {
			return nil
		}

		return details.Videos
	}

	if media.TmdbID == nil {
		return nil
	}

	details, err := s.tmdbClient.GetMovieDetails(ctx, int(*media.TmdbID))
	if err != nil {
		return nil
	}

	return details.Videos
}

func (s *CinemaService) seriesVideos(ctx context.Context, seriesID int64) *tmdb.VideoList {
	series, err := s.seriesRepo.GetByID(ctx, seriesID)
	if err != nil || series.TmdbID == nil {
		return nil
	}

	details, err := s.tmdbClient.GetTVDetails(ctx, int(*series.TmdbID))
	if err != nil {
		return nil
	}

	return details.Videos
}

func trailerItems(videos *tmdb.VideoList, maxTrailers int) []CinemaItem {
	if videos == nil || maxTrailers <= 0 {
		return nil
	}

	items := make([]CinemaItem, 0, maxTrailers)
	for _, video := range videos.Results {
		if len(items) >= maxTrailers {
			break
		}
		if video.Site != "YouTube" {
			continue
		}
		if video.Type != "Trailer" && video.Type != "Teaser" {
			continue
		}

		items = append(items, CinemaItem{
			Type:      "trailer",
			Title:     video.Name,
			URL:       "https://www.youtube.com/embed/" + video.Key + "?autoplay=1&controls=0",
			Skippable: true,
		})
	}

	return items
}
