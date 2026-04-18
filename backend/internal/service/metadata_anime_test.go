package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/anilist"
	"github.com/thawng/velox/pkg/nameparser"
	"github.com/thawng/velox/pkg/tmdb"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type metadataTestHarness struct {
	db            *sql.DB
	libraryRepo   *repository.LibraryRepo
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	seriesRepo    *repository.SeriesRepo
	seasonRepo    *repository.SeasonRepo
	episodeRepo   *repository.EpisodeRepo
	svc           *MetadataService
}

func openMetadataTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "metadata-test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}

	return db
}

func newMetadataTestHarness(t *testing.T, aniRT http.RoundTripper, tmdbRT http.RoundTripper) *metadataTestHarness {
	t.Helper()

	db := openMetadataTestDB(t)

	libraryRepo := repository.NewLibraryRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	mediaFileRepo := repository.NewMediaFileRepo(db)
	seriesRepo := repository.NewSeriesRepo(db)
	seasonRepo := repository.NewSeasonRepo(db)
	episodeRepo := repository.NewEpisodeRepo(db)
	genreRepo := repository.NewGenreRepo(db)
	personRepo := repository.NewPersonRepo(db)

	var aniClient *anilist.Client
	if aniRT != nil {
		aniClient = anilist.NewWithHTTPClient("", &http.Client{Transport: aniRT})
	}

	var tmdbClient *tmdb.Client
	if tmdbRT != nil {
		tmdbClient = tmdb.NewWithHTTPClient("test-token", &http.Client{Transport: tmdbRT})
	}

	svc := NewMetadataService(
		tmdbClient,
		aniClient,
		libraryRepo,
		mediaRepo,
		mediaFileRepo,
		seriesRepo,
		seasonRepo,
		episodeRepo,
		genreRepo,
		personRepo,
	)

	return &metadataTestHarness{
		db:            db,
		libraryRepo:   libraryRepo,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		seriesRepo:    seriesRepo,
		seasonRepo:    seasonRepo,
		episodeRepo:   episodeRepo,
		svc:           svc,
	}
}

func mustJSONResponse(t *testing.T, status int, payload any) *http.Response {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal response: %v", err)
	}

	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(string(body))),
	}
}

func TestMatchAndPersistAnimeEpisodePersistsAniListMetadata(t *testing.T) {
	t.Parallel()

	aniRT := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost || req.URL.Host != "graphql.anilist.co" {
			t.Fatalf("unexpected AniList request: %s %s", req.Method, req.URL.String())
		}

		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"Page": map[string]any{
					"pageInfo": map[string]any{
						"total":       1,
						"perPage":     10,
						"currentPage": 1,
						"lastPage":    1,
						"hasNextPage": false,
					},
					"media": []map[string]any{
						{
							"id":          52991,
							"type":        "ANIME",
							"format":      "TV",
							"status":      "RELEASING",
							"description": "<b>Elf mage</b> on a quiet journey.",
							"title": map[string]any{
								"romaji":        "Sousou no Frieren",
								"english":       "Frieren: Beyond Journey's End",
								"native":        "葬送のフリーレン",
								"userPreferred": "Sousou no Frieren",
							},
							"coverImage": map[string]any{
								"extraLarge": "https://img.anilist.co/frieren-xl.jpg",
								"large":      "https://img.anilist.co/frieren-lg.jpg",
								"medium":     "https://img.anilist.co/frieren-md.jpg",
							},
							"bannerImage": "https://img.anilist.co/frieren-banner.jpg",
							"startDate": map[string]any{
								"year":  2023,
								"month": 9,
								"day":   29,
							},
							"studios": map[string]any{
								"nodes": []map[string]any{
									{
										"id":                11,
										"name":              "Madhouse",
										"isAnimationStudio": true,
									},
								},
							},
							"streamingEpisodes": []map[string]any{
								{"title": "Episode 1", "thumbnail": "https://cdn.example/ep1.jpg"},
								{"title": "Episode 2", "thumbnail": "https://cdn.example/ep2.jpg"},
								{"title": "Episode 3", "thumbnail": "https://cdn.example/ep3.jpg"},
								{"title": "Episode 4", "thumbnail": "https://cdn.example/ep4.jpg"},
								{"title": "Episode 5", "thumbnail": "https://cdn.example/ep5.jpg"},
								{"title": "Episode 6", "thumbnail": "https://cdn.example/ep6.jpg"},
								{"title": "A Fairytale Hero", "thumbnail": "https://cdn.example/ep7.jpg"},
							},
						},
					},
				},
			},
		}), nil
	})

	h := newMetadataTestHarness(t, aniRT, nil)
	ctx := context.Background()

	lib, err := h.libraryRepo.Create(ctx, "Anime", model.LibraryTypeAnime, []string{t.TempDir()})
	if err != nil {
		t.Fatalf("create library: %v", err)
	}

	media := &model.Media{
		LibraryID: lib.ID,
		MediaType: "episode",
		Title:     "placeholder",
		SortTitle: "placeholder",
	}
	if err := h.mediaRepo.Create(ctx, media); err != nil {
		t.Fatalf("create media: %v", err)
	}

	parsed := nameparser.ParsedMedia{
		Title:        "Sousou no Frieren",
		MediaType:    "episode",
		Season:       1,
		Episode:      7,
		EpisodeTitle: "",
	}

	if err := h.svc.MatchAndPersistAnime(ctx, media, parsed, "/library/Sousou no Frieren - 07.mkv", lib.ID, false); err != nil {
		t.Fatalf("match anime: %v", err)
	}

	gotMedia, err := h.mediaRepo.GetByID(ctx, media.ID)
	if err != nil {
		t.Fatalf("reload media: %v", err)
	}
	if gotMedia.AnilistID == nil || *gotMedia.AnilistID != 52991 {
		t.Fatalf("expected AniList ID 52991, got %+v", gotMedia.AnilistID)
	}
	if gotMedia.Title != "A Fairytale Hero" {
		t.Fatalf("expected AniList episode title, got %q", gotMedia.Title)
	}
	if gotMedia.RomajiTitle != "Sousou no Frieren" {
		t.Fatalf("expected romaji title to persist, got %q", gotMedia.RomajiTitle)
	}
	if gotMedia.Studio != "Madhouse" {
		t.Fatalf("expected studio Madhouse, got %q", gotMedia.Studio)
	}
	if string(gotMedia.PosterPath) != "https://cdn.example/ep7.jpg" {
		t.Fatalf("expected episode thumbnail poster, got %q", gotMedia.PosterPath)
	}
	if string(gotMedia.BackdropPath) != "https://img.anilist.co/frieren-banner.jpg" {
		t.Fatalf("expected banner backdrop, got %q", gotMedia.BackdropPath)
	}
	if gotMedia.ReleaseDate != "2023-09-29" {
		t.Fatalf("expected AniList start date, got %q", gotMedia.ReleaseDate)
	}

	series, err := h.seriesRepo.GetByAnilistID(ctx, 1, 52991)
	if err != nil {
		t.Fatalf("get series by AniList ID: %v", err)
	}
	if series.Title != "Sousou no Frieren" {
		t.Fatalf("expected series title from AniList, got %q", series.Title)
	}
	if series.RomajiTitle != "Sousou no Frieren" {
		t.Fatalf("expected series romaji title, got %q", series.RomajiTitle)
	}
	if series.Status != "Returning Series" {
		t.Fatalf("expected mapped status, got %q", series.Status)
	}
	if string(series.PosterPath) != "https://img.anilist.co/frieren-xl.jpg" {
		t.Fatalf("expected AniList poster, got %q", series.PosterPath)
	}

	season, err := h.seasonRepo.GetBySeriesAndNumber(ctx, series.ID, 1)
	if err != nil {
		t.Fatalf("get season: %v", err)
	}
	if season.Title != "Season 1" {
		t.Fatalf("expected season title, got %q", season.Title)
	}

	episode, err := h.episodeRepo.GetByMediaID(ctx, media.ID)
	if err != nil {
		t.Fatalf("get episode: %v", err)
	}
	if episode.SeriesID != series.ID || episode.SeasonID != season.ID {
		t.Fatalf("expected linked episode to series %d season %d, got series=%d season=%d", series.ID, season.ID, episode.SeriesID, episode.SeasonID)
	}
	if episode.Title != "A Fairytale Hero" {
		t.Fatalf("expected linked episode title, got %q", episode.Title)
	}
	if string(episode.StillPath) != "https://cdn.example/ep7.jpg" {
		t.Fatalf("expected linked still path, got %q", episode.StillPath)
	}
}

func TestMatchAndPersistAnimeFallsBackToTMDbWhenAniListHasNoMatch(t *testing.T) {
	t.Parallel()

	aniRT := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return mustJSONResponse(t, http.StatusOK, map[string]any{
			"data": map[string]any{
				"Page": map[string]any{
					"pageInfo": map[string]any{
						"total":       0,
						"perPage":     10,
						"currentPage": 1,
						"lastPage":    1,
						"hasNextPage": false,
					},
					"media": []any{},
				},
			},
		}), nil
	})

	tmdbRT := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/3/search/tv":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"page":          1,
				"total_results": 1,
				"total_pages":   1,
				"results": []map[string]any{
					{
						"id":             2001,
						"name":           "Gokusen",
						"original_name":  "Gokusen",
						"overview":       "A rookie teacher enters a difficult classroom.",
						"first_air_date": "2002-04-17",
						"vote_average":   7.8,
					},
				},
			}), nil
		case "/3/tv/2001":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"id":             2001,
				"name":           "Gokusen",
				"original_name":  "Gokusen",
				"overview":       "A rookie teacher enters a difficult classroom.",
				"first_air_date": "2002-04-17",
				"poster_path":    "/gokusen-poster.jpg",
				"backdrop_path":  "/gokusen-backdrop.jpg",
				"vote_average":   7.8,
				"genres":         []map[string]any{},
				"external_ids": map[string]any{
					"tvdb_id": 4444,
				},
			}), nil
		case "/3/tv/2001/season/1":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"_id":           "season-1",
				"season_number": 1,
				"name":          "Season 1",
				"overview":      "Season overview",
				"poster_path":   "/season-1.jpg",
				"episodes": []map[string]any{
					{
						"id":             3001,
						"name":           "The Homeroom Teacher Arrives",
						"overview":       "Kumiko arrives at the school.",
						"air_date":       "2002-04-17",
						"episode_number": 1,
						"season_number":  1,
						"still_path":     "/episode-1-still.jpg",
						"vote_average":   8.2,
					},
				},
			}), nil
		default:
			t.Fatalf("unexpected TMDb path: %s", req.URL.String())
			return nil, nil
		}
	})

	h := newMetadataTestHarness(t, aniRT, tmdbRT)
	ctx := context.Background()

	lib, err := h.libraryRepo.Create(ctx, "Anime", model.LibraryTypeAnime, []string{t.TempDir()})
	if err != nil {
		t.Fatalf("create library: %v", err)
	}

	media := &model.Media{
		LibraryID: lib.ID,
		MediaType: "episode",
		Title:     "placeholder",
		SortTitle: "placeholder",
	}
	if err := h.mediaRepo.Create(ctx, media); err != nil {
		t.Fatalf("create media: %v", err)
	}

	parsed := nameparser.ParsedMedia{
		Title:     "Gokusen",
		MediaType: "episode",
		Season:    1,
		Episode:   1,
	}

	if err := h.svc.MatchAndPersistAnime(ctx, media, parsed, "/library/Gokusen - 01.mkv", lib.ID, false); err != nil {
		t.Fatalf("match anime with TMDb fallback: %v", err)
	}

	gotMedia, err := h.mediaRepo.GetByID(ctx, media.ID)
	if err != nil {
		t.Fatalf("reload media: %v", err)
	}
	if gotMedia.AnilistID != nil {
		t.Fatalf("expected AniList ID to stay nil on fallback, got %+v", gotMedia.AnilistID)
	}
	if gotMedia.TmdbID == nil || *gotMedia.TmdbID != 3001 {
		t.Fatalf("expected TMDb episode ID 3001, got %+v", gotMedia.TmdbID)
	}
	if gotMedia.TvdbID == nil || *gotMedia.TvdbID != 4444 {
		t.Fatalf("expected TVDB ID 4444, got %+v", gotMedia.TvdbID)
	}
	if gotMedia.Title != "The Homeroom Teacher Arrives" {
		t.Fatalf("expected TMDb fallback episode title, got %q", gotMedia.Title)
	}

	series, err := h.seriesRepo.GetByTmdbID(ctx, 1, 2001)
	if err != nil {
		t.Fatalf("get series by TMDb ID: %v", err)
	}
	if series.TvdbID == nil || *series.TvdbID != 4444 {
		t.Fatalf("expected fallback series TVDB ID 4444, got %+v", series.TvdbID)
	}

	episode, err := h.episodeRepo.GetByMediaID(ctx, media.ID)
	if err != nil {
		t.Fatalf("get episode: %v", err)
	}
	if episode.Title != "The Homeroom Teacher Arrives" {
		t.Fatalf("expected linked fallback episode title, got %q", episode.Title)
	}
}

func TestBulkRefreshAllMetadataKeepsMovieTMDbFlowWorking(t *testing.T) {
	t.Parallel()

	tmdbRT := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/3/search/movie":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"page":          1,
				"total_results": 1,
				"total_pages":   1,
				"results": []map[string]any{
					{
						"id":             5001,
						"title":          "Paprika",
						"original_title": "Paprika",
						"overview":       "Dream and reality begin to blur.",
						"release_date":   "2006-11-25",
						"poster_path":    "/paprika-poster.jpg",
						"backdrop_path":  "/paprika-backdrop.jpg",
						"vote_average":   7.9,
					},
				},
			}), nil
		case "/3/movie/5001":
			return mustJSONResponse(t, http.StatusOK, map[string]any{
				"id":             5001,
				"imdb_id":        "tt0851578",
				"title":          "Paprika",
				"original_title": "Paprika",
				"overview":       "Dream and reality begin to blur.",
				"release_date":   "2006-11-25",
				"poster_path":    "/paprika-poster.jpg",
				"backdrop_path":  "/paprika-backdrop.jpg",
				"vote_average":   7.9,
				"genres":         []map[string]any{},
			}), nil
		default:
			t.Fatalf("unexpected TMDb path: %s", req.URL.String())
			return nil, nil
		}
	})

	h := newMetadataTestHarness(t, nil, tmdbRT)
	ctx := context.Background()

	lib, err := h.libraryRepo.Create(ctx, "Movies", model.LibraryTypeMovies, []string{t.TempDir()})
	if err != nil {
		t.Fatalf("create library: %v", err)
	}

	media := &model.Media{
		LibraryID: lib.ID,
		MediaType: "movie",
		Title:     "Paprika",
		SortTitle: "Paprika",
	}
	if err := h.mediaRepo.Create(ctx, media); err != nil {
		t.Fatalf("create media: %v", err)
	}

	if err := h.mediaFileRepo.Create(ctx, &model.MediaFile{
		MediaID:     media.ID,
		FilePath:    filepath.Join(t.TempDir(), "Paprika (2006).mkv"),
		FileSize:    1234,
		Duration:    5400,
		VideoCodec:  "h264",
		AudioCodec:  "aac",
		Container:   "matroska",
		Fingerprint: "1234:paprika",
		IsPrimary:   true,
	}); err != nil {
		t.Fatalf("create media file: %v", err)
	}

	updated, err := h.svc.BulkRefreshAllMetadata(ctx)
	if err != nil {
		t.Fatalf("bulk refresh metadata: %v", err)
	}
	if updated != 1 {
		t.Fatalf("expected one updated movie, got %d", updated)
	}

	gotMedia, err := h.mediaRepo.GetByID(ctx, media.ID)
	if err != nil {
		t.Fatalf("reload media: %v", err)
	}
	if gotMedia.TmdbID == nil || *gotMedia.TmdbID != 5001 {
		t.Fatalf("expected TMDb ID 5001 after bulk refresh, got %+v", gotMedia.TmdbID)
	}
	if gotMedia.Title != "Paprika" {
		t.Fatalf("expected movie title Paprika, got %q", gotMedia.Title)
	}
	if gotMedia.ImdbID == nil || *gotMedia.ImdbID != "tt0851578" {
		t.Fatalf("expected IMDb ID tt0851578, got %+v", gotMedia.ImdbID)
	}
	if string(gotMedia.PosterPath) != "/paprika-poster.jpg" {
		t.Fatalf("expected TMDb poster path, got %q", gotMedia.PosterPath)
	}
}
