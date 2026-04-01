package handler

import (
	"context"
	"io"
	"net/http"
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

type CinemaHandler struct {
	appSettingsRepo *repository.AppSettingsRepo
	mediaRepo       *repository.MediaRepo
	episodeRepo     *repository.EpisodeRepo
	seriesRepo      *repository.SeriesRepo
	tmdbClient      *tmdb.Client
	dataDir         string
}

type cinemaItem struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	URL       string `json:"url"`
	Duration  int    `json:"duration"`
	Skippable bool   `json:"skippable"`
}

func NewCinemaHandler(
	appSettingsRepo *repository.AppSettingsRepo,
	mediaRepo *repository.MediaRepo,
	episodeRepo *repository.EpisodeRepo,
	seriesRepo *repository.SeriesRepo,
	tmdbClient *tmdb.Client,
	dataDir string,
) *CinemaHandler {
	return &CinemaHandler{
		appSettingsRepo: appSettingsRepo,
		mediaRepo:       mediaRepo,
		episodeRepo:     episodeRepo,
		seriesRepo:      seriesRepo,
		tmdbClient:      tmdbClient,
		dataDir:         dataDir,
	}
}

func (h *CinemaHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	enabled, err := h.appSettingsRepo.Get(r.Context(), settingCinemaModeEnabled)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load cinema setting")
		return
	}

	maxTrailers, err := h.appSettingsRepo.Get(r.Context(), settingCinemaMaxTrailers)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load cinema setting")
		return
	}

	introPath, err := h.appSettingsRepo.Get(r.Context(), settingCinemaIntroPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to load cinema setting")
		return
	}

	if maxTrailers == "" {
		maxTrailers = strconv.Itoa(defaultCinemaMaxTrailers)
	}

	respondJSON(w, http.StatusOK, map[string]any{
		"enabled":      enabled == "true",
		"max_trailers": maxTrailers,
		"has_intro":    introPath != "",
	})
}

func (h *CinemaHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled     *bool   `json:"enabled"`
		MaxTrailers *string `json:"max_trailers"`
	}

	if err := parseJSON(r, &body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}

	if body.Enabled != nil {
		value := "false"
		if *body.Enabled {
			value = "true"
		}
		if err := h.appSettingsRepo.Set(r.Context(), settingCinemaModeEnabled, value); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save cinema setting")
			return
		}
	}

	if body.MaxTrailers != nil {
		if err := h.appSettingsRepo.Set(r.Context(), settingCinemaMaxTrailers, *body.MaxTrailers); err != nil {
			respondError(w, http.StatusInternalServerError, "failed to save cinema setting")
			return
		}
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *CinemaHandler) MediaCinema(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	items := make([]cinemaItem, 0, defaultCinemaMaxTrailers+1)

	cinemaEnabled, _ := h.appSettingsRepo.Get(r.Context(), settingCinemaModeEnabled)
	introPath, _ := h.appSettingsRepo.Get(r.Context(), settingCinemaIntroPath)
	if introPath != "" && cinemaEnabled == "true" {
		items = append(items, cinemaItem{
			Type:      "intro",
			Title:     "Cinema Intro",
			URL:       "/api/cinema/intro",
			Skippable: true,
		})
	}

	maxTrailers := h.trailerLimit(r.Context())
	if h.tmdbClient != nil && maxTrailers > 0 {
		items = append(items, trailerItems(h.mediaVideos(r.Context(), id), maxTrailers)...)
	}

	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CinemaHandler) SeriesCinema(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid id")
		return
	}

	items := []cinemaItem{}
	maxTrailers := h.trailerLimit(r.Context())
	if h.tmdbClient != nil && maxTrailers > 0 {
		items = append(items, trailerItems(h.seriesVideos(r.Context(), id), maxTrailers)...)
	}

	respondJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *CinemaHandler) ServeIntro(w http.ResponseWriter, r *http.Request) {
	introPath, _ := h.appSettingsRepo.Get(r.Context(), settingCinemaIntroPath)
	if introPath == "" {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, introPath)
}

func (h *CinemaHandler) UploadIntro(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)
	if err := r.ParseMultipartForm(100 << 20); err != nil {
		respondError(w, http.StatusBadRequest, "file too large (max 100MB)")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		respondError(w, http.StatusBadRequest, "missing file")
		return
	}
	defer file.Close()

	cinemaDir := filepath.Join(h.dataDir, "cinema")
	if err := os.MkdirAll(cinemaDir, 0755); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create cinema directory")
		return
	}

	introPath := filepath.Join(cinemaDir, "intro.mp4")
	dst, err := os.Create(introPath)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	if err := h.appSettingsRepo.Set(r.Context(), settingCinemaIntroPath, introPath); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save cinema setting")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"path": introPath})
}

func (h *CinemaHandler) trailerLimit(ctx context.Context) int {
	maxTrailers := defaultCinemaMaxTrailers
	if value, _ := h.appSettingsRepo.Get(ctx, settingCinemaMaxTrailers); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 0 {
			maxTrailers = parsed
		}
	}
	return maxTrailers
}

func (h *CinemaHandler) mediaVideos(ctx context.Context, mediaID int64) *tmdb.VideoList {
	media, err := h.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil
	}

	if media.MediaType == "episode" {
		episode, err := h.episodeRepo.GetByMediaID(ctx, media.ID)
		if err != nil {
			return nil
		}

		series, err := h.seriesRepo.GetByID(ctx, episode.SeriesID)
		if err != nil || series.TmdbID == nil {
			return nil
		}

		details, err := h.tmdbClient.GetTVDetails(ctx, int(*series.TmdbID))
		if err != nil {
			return nil
		}

		return details.Videos
	}

	if media.TmdbID == nil {
		return nil
	}

	details, err := h.tmdbClient.GetMovieDetails(ctx, int(*media.TmdbID))
	if err != nil {
		return nil
	}

	return details.Videos
}

func (h *CinemaHandler) seriesVideos(ctx context.Context, seriesID int64) *tmdb.VideoList {
	series, err := h.seriesRepo.GetByID(ctx, seriesID)
	if err != nil || series.TmdbID == nil {
		return nil
	}

	details, err := h.tmdbClient.GetTVDetails(ctx, int(*series.TmdbID))
	if err != nil {
		return nil
	}

	return details.Videos
}

func trailerItems(videos *tmdb.VideoList, maxTrailers int) []cinemaItem {
	if videos == nil || maxTrailers <= 0 {
		return nil
	}

	items := make([]cinemaItem, 0, maxTrailers)
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

		items = append(items, cinemaItem{
			Type:      "trailer",
			Title:     video.Name,
			URL:       "https://www.youtube.com/embed/" + video.Key + "?autoplay=1&controls=0",
			Skippable: true,
		})
	}

	return items
}
