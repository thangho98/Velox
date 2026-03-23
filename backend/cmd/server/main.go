package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/config"
	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/handler"
	"github.com/thawng/velox/internal/middleware"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/storage"
	"github.com/thawng/velox/internal/transcoder"
	"github.com/thawng/velox/internal/trickplay"
	"github.com/thawng/velox/internal/watcher"
	"github.com/thawng/velox/internal/websocket"
	"github.com/thawng/velox/pkg/fanart"
	"github.com/thawng/velox/pkg/omdb"
	"github.com/thawng/velox/pkg/thetvdb"
	"github.com/thawng/velox/pkg/tmdb"
	"github.com/thawng/velox/pkg/tvmaze"
)

func main() {
	if err := config.LoadDotEnv(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			runMigrate()
			return
		case "version":
			fmt.Println("velox v0.1.0")
			return
		}
	}

	runServer()
}

func runMigrate() {
	cfg := config.Load()
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	subcmd := "up"
	if len(os.Args) > 2 {
		subcmd = os.Args[2]
	}

	switch subcmd {
	case "up":
		if err := database.Migrate(db); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		log.Println("migrations applied successfully")

	case "rollback":
		if err := database.MigrateRollback(db); err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		log.Println("rollback completed")

	case "status":
		statuses, err := database.MigrateStatus(db)
		if err != nil {
			log.Fatalf("failed to get status: %v", err)
		}
		fmt.Printf("%-8s %-30s %-10s %s\n", "VERSION", "NAME", "STATUS", "APPLIED AT")
		fmt.Println("-------- ------------------------------ ---------- -------------------")
		for _, s := range statuses {
			status := "pending"
			appliedAt := ""
			if s.Applied {
				status = "applied"
				appliedAt = s.AppliedAt.Format("2006-01-02 15:04:05")
			}
			fmt.Printf("%03d      %-30s %-10s %s\n", s.Version, s.Name, status, appliedAt)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown migrate command: %s\n", subcmd)
		fmt.Fprintln(os.Stderr, "usage: velox migrate [up|rollback|status]")
		os.Exit(1)
	}
}

func runServer() {
	cfg := config.Load()

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	// JWT Manager
	jwtSecret, err := auth.LoadOrCreateSecret(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to initialize JWT: %v", err)
	}
	jwtManager := auth.NewJWTManager(jwtSecret)
	apiKeyStore := auth.NewAPIKeyStore()

	// Repositories
	libraryRepo := repository.NewLibraryRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	mediaFileRepo := repository.NewMediaFileRepo(db)
	seriesRepo := repository.NewSeriesRepo(db)
	seasonRepo := repository.NewSeasonRepo(db)
	episodeRepo := repository.NewEpisodeRepo(db)
	scanJobRepo := repository.NewScanJobRepo(db)
	subtitleRepo := repository.NewSubtitleRepo(db)
	audioTrackRepo := repository.NewAudioTrackRepo(db)
	userRepo := repository.NewUserRepo(db)
	refreshTokenRepo := repository.NewRefreshTokenRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	prefsRepo := repository.NewUserPreferencesRepo(db)
	userDataRepo := repository.NewUserDataRepo(db)
	genreRepo := repository.NewGenreRepo(db)
	personRepo := repository.NewPersonRepo(db)
	activityRepo := repository.NewActivityRepo(db)
	webhookRepo := repository.NewWebhookRepo(db)
	markerRepo := repository.NewMediaMarkerRepo(db)
	fpRepo := repository.NewAudioFingerprintRepo(db)
	notificationRepo := repository.NewNotificationRepo(db)

	// WebSocket Hub
	wsHub := websocket.NewHub(slog.Default())
	go wsHub.Run()

	// Services
	pipeline := scanner.NewPipeline(
		db, libraryRepo, mediaRepo, mediaFileRepo,
		seriesRepo, seasonRepo, episodeRepo,
		scanJobRepo, subtitleRepo, audioTrackRepo,
		markerRepo, // NEW: marker repo
	)
	// TMDb metadata enrichment (Phase 04)
	appSettingsRepo := repository.NewAppSettingsRepo(db)

	var tmdbClient *tmdb.Client
	var metadataSvc *service.MetadataService

	tmdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingTMDbAPIKey)
	if tmdbAPIKey == "" {
		tmdbAPIKey = cfg.TMDbAPIKey
	}
	if tmdbAPIKey != "" {
		if cfg.TMDbAPIKey != "" && tmdbAPIKey == cfg.TMDbAPIKey {
			log.Println("TMDb using built-in API key from env (override in Settings → Metadata)")
		} else {
			log.Println("TMDb using custom API key from settings")
		}
		tmdbClient = tmdb.New(tmdbAPIKey)
		metadataSvc = service.NewMetadataService(tmdbClient, mediaRepo, mediaFileRepo, seriesRepo, seasonRepo, episodeRepo, genreRepo, personRepo)
		if metadataSvc != nil {
			pipeline.SetMetadataMatcher(metadataSvc)
			log.Println("TMDb metadata enrichment enabled")
		}
		// OMDb ratings enrichment
		omdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingOMDbAPIKey)
		if omdbAPIKey == "" {
			omdbAPIKey = cfg.OMDbAPIKey
		}
		if omdbAPIKey != "" {
			if cfg.OMDbAPIKey != "" && omdbAPIKey == cfg.OMDbAPIKey {
				log.Println("OMDb using built-in API key from env (override in Settings → Metadata)")
			} else {
				log.Println("OMDb using custom API key from settings")
			}
			omdbClient := omdb.New(omdbAPIKey)
			metadataSvc.SetOMDbClient(omdbClient)
			log.Println("OMDb ratings enrichment enabled (IMDb, Rotten Tomatoes, Metacritic)")
		}

		// TheTVDB metadata enrichment
		tvdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingTVDBAPIKey)
		if tvdbAPIKey == "" {
			tvdbAPIKey = cfg.TVDBAPIKey
		}
		if tvdbAPIKey != "" {
			if cfg.TVDBAPIKey != "" && tvdbAPIKey == cfg.TVDBAPIKey {
				log.Println("TheTVDB using built-in API key from env (override in Settings → Metadata)")
			} else {
				log.Println("TheTVDB using custom API key from settings")
			}
			tvdbClient := thetvdb.New(tvdbAPIKey)
			metadataSvc.SetTVDBClient(tvdbClient)
			log.Println("TheTVDB metadata enrichment enabled")
		}

		// Fanart.tv artwork enrichment
		fanartAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingFanartAPIKey)
		if fanartAPIKey == "" {
			fanartAPIKey = cfg.FanartAPIKey
		}
		if fanartAPIKey != "" {
			if cfg.FanartAPIKey != "" && fanartAPIKey == cfg.FanartAPIKey {
				log.Println("fanart.tv using built-in API key from env (override in Settings → Metadata)")
			} else {
				log.Println("fanart.tv using custom API key from settings")
			}
			fanartClient := fanart.New(fanartAPIKey)
			metadataSvc.SetFanartClient(fanartClient)
			log.Println("fanart.tv artwork enrichment enabled (logos, thumbs)")
		}

		// TVmaze TV enrichment (free, no API key)
		tvmazeClient := tvmaze.New()
		metadataSvc.SetTVmazeClient(tvmazeClient)
		log.Println("TVmaze enrichment enabled (network, schedule, ID cross-reference)")
	} else {
		log.Println("TMDb API key not configured. Metadata enrichment disabled.")
		log.Println("Set VELOX_TMDB_API_KEY env var or configure in Settings → Metadata")
	}

	// Resolve hardware accelerator
	hwAccel := cfg.HWAccel
	switch hwAccel {
	case "auto":
		hwAccel = playback.DetectHWAccel()
		if hwAccel != "" {
			log.Printf("hardware acceleration: %s", hwAccel)
		} else {
			log.Printf("hardware acceleration: none detected, using software encoder")
		}
	case "none":
		hwAccel = ""
	default:
		log.Printf("hardware acceleration: %s (configured)", hwAccel)
	}

	startTime := time.Now()
	transcoderSvc := transcoder.New(cfg.TranscodePath, hwAccel, cfg.MaxTranscodes)
	librarySvc := service.NewLibraryService(libraryRepo, scanJobRepo, pipeline)
	mediaSvc := service.NewMediaService(mediaRepo, mediaFileRepo)
	mediaSvc.SetEpisodeRepo(episodeRepo)
	mediaSvc.SetSeasonRepo(seasonRepo)
	streamSvc := service.NewStreamService(mediaFileRepo, audioTrackRepo, transcoderSvc)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, sessionRepo, jwtManager, db)
	userDataSvc := service.NewUserDataService(userDataRepo)
	subtitleSvc := service.NewSubtitleService(subtitleRepo, mediaFileRepo)
	audioTrackSvc := service.NewAudioTrackService(audioTrackRepo)
	markerSvc := service.NewMarkerService(markerRepo, mediaFileRepo, fpRepo, episodeRepo, seasonRepo)

	// Plan F: Admin & Operations services
	activitySvc := service.NewActivityService(activityRepo)
	defer activitySvc.Close()
	adminSvc := service.NewAdminService(db, userRepo, startTime, hwAccel, cfg.DatabasePath)
	webhookSvc := service.NewWebhookService(webhookRepo)
	notificationSvc := service.NewNotificationService(notificationRepo, userRepo, wsHub, slog.Default())
	notificationSvc.SetWebhookService(webhookSvc)
	librarySvc.SetNotificationService(notificationSvc)
	streamSvc.SetNotificationService(notificationSvc)
	if metadataSvc != nil {
		metadataSvc.SetNotificationService(notificationSvc)
	}

	// Handlers
	libraryHandler := handler.NewLibraryHandler(librarySvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)
	setupHandler := handler.NewSetupHandler(authSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(authSvc)
	profileHandler := handler.NewProfileHandler(authSvc, prefsRepo, userDataSvc)
	playbackHandler := handler.NewPlaybackHandler(mediaSvc, streamSvc, userDataSvc, subtitleSvc, audioTrackSvc, markerSvc, prefsRepo, appSettingsRepo)
	subtitleHandler := handler.NewSubtitleHandler(subtitleSvc, mediaFileRepo, appSettingsRepo, cfg.SubtitleCachePath)
	audioTrackHandler := handler.NewAudioTrackHandler(audioTrackSvc)
	// Settings handler
	builtinKeys := map[string]bool{
		"tmdb":   cfg.TMDbAPIKey != "",
		"omdb":   cfg.OMDbAPIKey != "",
		"tvdb":   cfg.TVDBAPIKey != "",
		"fanart": cfg.FanartAPIKey != "",
		"subdl":  cfg.SubdlAPIKey != "",
	}
	settingsHandler := handler.NewSettingsHandler(appSettingsRepo, builtinKeys)
	subtitleSearchSvc := service.NewSubtitleSearchService(mediaRepo, mediaFileRepo, subtitleRepo, appSettingsRepo, episodeRepo, seasonRepo, seriesRepo, cfg.SubtitleCachePath)
	subtitleSearchSvc.SetBuiltinSubdlKey(cfg.SubdlAPIKey)
	subtitleSearchSvc.SetNotificationService(notificationSvc)
	subtitleSearchHandler := handler.NewSubtitleSearchHandler(subtitleSearchSvc)
	pipeline.SetSubtitleAutoDownloader(subtitleSearchSvc)

	// Trickplay (Plan E Phase 03) — nil when disabled, handlers return 404 gracefully
	var trickplayGen *trickplay.Generator
	if cfg.TrickplayEnabled {
		if err := os.MkdirAll(cfg.TrickplayPath, 0755); err != nil {
			log.Fatalf("failed to create trickplay dir: %v", err)
		}
		trickplayGen = trickplay.New(cfg.TrickplayPath, cfg.TrickplayInterval)
		log.Printf("trickplay enabled (interval: %ds)", cfg.TrickplayInterval)
	}
	trickplayHandler := handler.NewTrickplayHandler(trickplayGen, streamSvc)
	imageHandler := handler.NewImageHandler()
	seriesHandler := handler.NewSeriesHandler(seriesRepo, seasonRepo, episodeRepo)
	imgStorage := storage.NewImageStorage(cfg.DataDir)
	metadataHandler := handler.NewMetadataHandler(mediaSvc, metadataSvc, imgStorage)
	activityHandler := handler.NewActivityHandler(activitySvc)
	adminHandler := handler.NewAdminHandler(adminSvc)
	webhookHandler := handler.NewWebhookHandler(webhookSvc)
	markerAdminHandler := handler.NewMarkerAdminHandler(markerSvc)
	notificationHandler := handler.NewNotificationHandler(notificationSvc)
	wsHandler := handler.NewWebSocketHandler(wsHub, jwtManager, slog.Default()) // NEW: marker admin handler for backfill

	// Router
	mux := http.NewServeMux()

	// Health check (public — used by Docker, load balancers, uptime monitors)
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Setup routes (public, only works before configured)
	mux.HandleFunc("GET /api/setup/status", setupHandler.Status)
	mux.HandleFunc("POST /api/setup", setupHandler.Setup)

	// Auth routes (public)
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)

	// Protected auth routes
	mux.HandleFunc("POST /api/auth/change-password", authHandler.ChangePassword)
	mux.HandleFunc("GET /api/auth/me", authHandler.Me)
	mux.HandleFunc("GET /api/auth/sessions", authHandler.ListSessions)
	mux.HandleFunc("DELETE /api/auth/sessions/{id}", authHandler.RevokeSession)

	// User management routes (admin only)
	mux.Handle("GET /api/users", middleware.RequireAdmin(http.HandlerFunc(userHandler.List)))
	mux.Handle("POST /api/users", middleware.RequireAdmin(http.HandlerFunc(userHandler.Create)))
	mux.Handle("PATCH /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(userHandler.Update)))
	mux.Handle("DELETE /api/users/{id}", middleware.RequireAdmin(http.HandlerFunc(userHandler.Delete)))
	mux.Handle("PUT /api/users/{id}/library-access", middleware.RequireAdmin(http.HandlerFunc(userHandler.SetLibraryAccess)))

	// Filesystem browser (admin only — used by Add Library UI)
	mux.Handle("GET /api/admin/fs/browse", middleware.RequireAdmin(http.HandlerFunc(handler.FSBrowse)))

	// Admin settings routes (admin only)
	mux.Handle("GET /api/admin/settings/opensubtitles", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetOpenSubtitles)))
	mux.Handle("PUT /api/admin/settings/opensubtitles", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateOpenSubtitles)))
	mux.Handle("GET /api/admin/settings/tmdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetTMDb)))
	mux.Handle("PUT /api/admin/settings/tmdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateTMDb)))
	mux.Handle("GET /api/admin/settings/omdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetOMDb)))
	mux.Handle("PUT /api/admin/settings/omdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateOMDb)))
	mux.Handle("GET /api/admin/settings/tvdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetTVDB)))
	mux.Handle("PUT /api/admin/settings/tvdb", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateTVDB)))
	mux.Handle("GET /api/admin/settings/fanart", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetFanart)))
	mux.Handle("PUT /api/admin/settings/fanart", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateFanart)))
	mux.Handle("GET /api/admin/settings/subdl", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetSubdl)))
	mux.Handle("PUT /api/admin/settings/subdl", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateSubdl)))
	mux.Handle("GET /api/admin/settings/deepl", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetDeepL)))
	mux.Handle("PUT /api/admin/settings/deepl", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateDeepL)))
	mux.Handle("GET /api/admin/settings/auto-subtitles", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetAutoSubtitles)))
	mux.Handle("PUT /api/admin/settings/auto-subtitles", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdateAutoSubtitles)))
	mux.Handle("GET /api/admin/settings/playback", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.GetPlayback)))
	mux.Handle("PUT /api/admin/settings/playback", middleware.RequireAdmin(http.HandlerFunc(settingsHandler.UpdatePlayback)))

	// Admin cinema mode settings
	mux.Handle("GET /api/admin/settings/cinema", middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enabled, _ := appSettingsRepo.Get(r.Context(), "cinema_mode_enabled")
		maxTrailers, _ := appSettingsRepo.Get(r.Context(), "cinema_max_trailers")
		introPath, _ := appSettingsRepo.Get(r.Context(), "cinema_intro_path")
		if maxTrailers == "" {
			maxTrailers = "2"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
			"enabled":      enabled == "true",
			"max_trailers": maxTrailers,
			"has_intro":    introPath != "",
		}})
	})))
	mux.Handle("PUT /api/admin/settings/cinema", middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Enabled     *bool   `json:"enabled"`
			MaxTrailers *string `json:"max_trailers"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"invalid body"}`)
			return
		}
		if body.Enabled != nil {
			val := "false"
			if *body.Enabled {
				val = "true"
			}
			appSettingsRepo.Set(r.Context(), "cinema_mode_enabled", val)
		}
		if body.MaxTrailers != nil {
			appSettingsRepo.Set(r.Context(), "cinema_max_trailers", *body.MaxTrailers)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"status": "updated"}})
	})))

	// Admin dashboard routes (Plan F)
	mux.Handle("GET /api/admin/activity", middleware.RequireAdmin(http.HandlerFunc(activityHandler.List)))
	mux.Handle("GET /api/admin/stats/playback", middleware.RequireAdmin(http.HandlerFunc(activityHandler.GetPlaybackStats)))
	mux.Handle("GET /api/admin/server", middleware.RequireAdmin(http.HandlerFunc(adminHandler.ServerInfo)))
	mux.Handle("GET /api/admin/stats/libraries", middleware.RequireAdmin(http.HandlerFunc(adminHandler.LibraryStats)))

	// Webhook routes (Plan F)
	mux.Handle("GET /api/admin/webhooks", middleware.RequireAdmin(http.HandlerFunc(webhookHandler.List)))
	mux.Handle("POST /api/admin/webhooks", middleware.RequireAdmin(http.HandlerFunc(webhookHandler.Create)))
	mux.Handle("PUT /api/admin/webhooks/{id}", middleware.RequireAdmin(http.HandlerFunc(webhookHandler.Update)))
	mux.Handle("DELETE /api/admin/webhooks/{id}", middleware.RequireAdmin(http.HandlerFunc(webhookHandler.Delete)))

	// Marker admin routes (Phase 04 - Fingerprint Backfill)
	mux.Handle("GET /api/admin/markers/detectors", middleware.RequireAdmin(http.HandlerFunc(markerAdminHandler.ListDetectors)))
	mux.Handle("POST /api/admin/markers/detect", middleware.RequireAdmin(http.HandlerFunc(markerAdminHandler.DetectWithDetector)))
	mux.Handle("POST /api/admin/markers/backfill", middleware.RequireAdmin(http.HandlerFunc(markerAdminHandler.BackfillMarkers)))

	// Library admin routes
	mux.Handle("POST /api/libraries", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Create)))
	mux.Handle("DELETE /api/libraries/{id}", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Delete)))
	mux.Handle("POST /api/libraries/{id}/scan", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.Scan)))
	mux.Handle("GET /api/libraries/{id}/scan-status", middleware.RequireAdmin(http.HandlerFunc(libraryHandler.ScanStatus)))

	// Profile routes (authenticated)
	mux.HandleFunc("GET /api/profile", profileHandler.GetProfile)      // GET current user profile
	mux.HandleFunc("PATCH /api/profile", profileHandler.UpdateProfile) // Update display_name
	mux.HandleFunc("GET /api/profile/preferences", profileHandler.GetPreferences)
	mux.HandleFunc("PUT /api/profile/preferences", profileHandler.UpdatePreferences)

	// User data routes (progress, favorites)
	mux.HandleFunc("GET /api/profile/progress/{mediaId}", profileHandler.GetProgress)
	mux.HandleFunc("PUT /api/profile/progress/{mediaId}", profileHandler.UpdateProgress)
	mux.HandleFunc("GET /api/profile/favorites", profileHandler.ListFavorites)
	mux.HandleFunc("POST /api/profile/favorites/{mediaId}", profileHandler.ToggleFavorite)
	mux.HandleFunc("GET /api/profile/recently-watched", profileHandler.ListRecentlyWatched)
	mux.HandleFunc("GET /api/profile/continue-watching", profileHandler.ContinueWatching)
	mux.HandleFunc("GET /api/profile/next-up", profileHandler.NextUp)
	mux.HandleFunc("DELETE /api/profile/progress/{mediaId}/dismiss", profileHandler.DismissProgress)

	// API routes - Libraries (read = authenticated, write = admin above)
	mux.HandleFunc("GET /api/libraries", libraryHandler.List)

	// API routes - Media
	mux.HandleFunc("GET /api/media", mediaHandler.List)
	mux.HandleFunc("GET /api/media/{id}", mediaHandler.Get)
	mux.HandleFunc("GET /api/media/{id}/files", mediaHandler.GetWithFiles)
	mux.HandleFunc("GET /api/media/{id}/versions", mediaHandler.GetVersions)

	// API routes - Folder Browse (DB-based, library-scoped, ACL-aware)
	// No library_id → show all accessible libraries as root folders
	// With library_id + path → browse inside that library
	mux.HandleFunc("GET /api/browse", func(w http.ResponseWriter, r *http.Request) {
		respondJSON := func(data any) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"data": data})
		}
		respondErr := func(status int, msg string) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			fmt.Fprintf(w, `{"error":%q}`, msg)
		}

		// Get user for ACL
		userID, isAdmin, _ := auth.UserFromContext(r.Context())

		// Get accessible library IDs for non-admin
		var accessIDs []int64
		if !isAdmin {
			var err error
			accessIDs, err = userRepo.GetLibraryIDs(r.Context(), userID)
			if err != nil {
				respondErr(http.StatusInternalServerError, "checking library access")
				return
			}
		}

		hasAccessTo := func(libID int64) bool {
			if isAdmin {
				return true
			}
			for _, id := range accessIDs {
				if id == libID {
					return true
				}
			}
			return false
		}

		libraryIDStr := r.URL.Query().Get("library_id")
		relativePath := r.URL.Query().Get("path")

		// Security: reject path traversal
		if strings.Contains(relativePath, "..") {
			respondErr(http.StatusBadRequest, "invalid path")
			return
		}

		// No library_id → show all accessible libraries as root folders (with posters)
		if libraryIDStr == "" || libraryIDStr == "0" {
			allLibs, err := libraryRepo.List(r.Context())
			if err != nil {
				respondErr(http.StatusInternalServerError, "listing libraries")
				return
			}
			folders := make([]repository.BrowseFolderItem, 0, len(allLibs))
			for _, lib := range allLibs {
				if !hasAccessTo(lib.ID) {
					continue
				}
				// Get poster from first media in this library
				var poster sql.NullString
				_ = db.QueryRowContext(r.Context(), `
					SELECT poster_path FROM media
					WHERE library_id = ? AND poster_path != ''
					ORDER BY sort_title LIMIT 1`, lib.ID).Scan(&poster)

				folders = append(folders, repository.BrowseFolderItem{
					Name:   lib.Name,
					Path:   fmt.Sprintf("lib:%d", lib.ID),
					Poster: poster.String,
				})
			}
			respondJSON(&repository.BrowseResult{
				Path: "", Parent: "", Folders: folders,
			})
			return
		}

		// Parse library_id
		libraryID, err := strconv.ParseInt(libraryIDStr, 10, 64)
		if err != nil || libraryID == 0 {
			respondErr(http.StatusBadRequest, "invalid library_id")
			return
		}

		// Validate library + access
		library, err := libraryRepo.GetByID(r.Context(), libraryID)
		if err != nil {
			respondErr(http.StatusNotFound, "library not found")
			return
		}
		if !hasAccessTo(libraryID) {
			respondErr(http.StatusForbidden, "no access to this library")
			return
		}

		// Multi-root: path="" with multiple roots → show roots as top-level folders
		if relativePath == "" && len(library.Paths) > 1 {
			folders := make([]repository.BrowseFolderItem, 0, len(library.Paths))
			nameCounts := map[string]int{}
			for i, p := range library.Paths {
				base := filepath.Base(p)
				nameCounts[base]++
				name := base
				if nameCounts[base] > 1 {
					name = fmt.Sprintf("%s-%d", base, nameCounts[base])
				}
				// Get poster from first media under this root
				var poster sql.NullString
				_ = db.QueryRowContext(r.Context(), `
					SELECT m.poster_path FROM media_files mf
					JOIN media m ON m.id = mf.media_id
					WHERE m.library_id = ? AND mf.file_path LIKE ? || '%' AND m.poster_path != ''
					ORDER BY m.sort_title LIMIT 1`, libraryID, p+"/").Scan(&poster)

				folders = append(folders, repository.BrowseFolderItem{
					Name:   name,
					Path:   fmt.Sprintf("root:%d", i),
					Poster: poster.String,
				})
			}
			respondJSON(&repository.BrowseResult{
				LibraryID: libraryID, Path: "", Parent: "", Folders: folders,
			})
			return
		}

		// Resolve absolute directory from library paths + relative path
		var rootPath, subPath string
		if len(library.Paths) == 1 {
			rootPath = library.Paths[0]
			subPath = relativePath
		} else if strings.HasPrefix(relativePath, "root:") {
			rest := relativePath[5:]
			var nStr string
			if slashIdx := strings.Index(rest, "/"); slashIdx > 0 {
				nStr = rest[:slashIdx]
				subPath = rest[slashIdx+1:]
			} else {
				nStr = rest
				subPath = ""
			}
			n, parseErr := strconv.Atoi(nStr)
			if parseErr != nil || n < 0 || n >= len(library.Paths) {
				respondErr(http.StatusBadRequest, "invalid root index")
				return
			}
			rootPath = library.Paths[n]
		} else {
			rootPath = library.Paths[0]
			subPath = relativePath
		}

		absDir := rootPath
		if subPath != "" {
			absDir = filepath.Join(rootPath, subPath)
		}

		result, err := mediaFileRepo.BrowseFolders(r.Context(), libraryID, absDir, relativePath)
		if err != nil {
			respondErr(http.StatusInternalServerError, err.Error())
			return
		}
		respondJSON(result)
	})

	// API routes - Genres
	mux.HandleFunc("GET /api/genres", func(w http.ResponseWriter, r *http.Request) {
		typeFilter := r.URL.Query().Get("type") // "movie" | "series" | "" (all)

		var genres []model.Genre
		var err error

		if typeFilter != "" {
			genres, err = genreRepo.ListWithFilter(r.Context(), typeFilter)
		} else {
			genres, err = genreRepo.List(r.Context())
		}

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": genres})
	})

	// API routes - Media genres & credits (for metadata editor)
	mux.HandleFunc("GET /api/media/{id}/genres", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		genres, err := genreRepo.ListByMediaID(r.Context(), id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": genres})
	})
	mux.HandleFunc("GET /api/media/{id}/credits", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		credits, err := personRepo.ListCreditsByMedia(r.Context(), id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": credits})
	})
	mux.HandleFunc("GET /api/series/{id}/genres", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		genres, err := genreRepo.ListBySeriesID(r.Context(), id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": genres})
	})
	mux.HandleFunc("GET /api/series/{id}/credits", func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
		credits, err := personRepo.ListCreditsBySeries(r.Context(), id)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": credits})
	})

	// API routes - Series
	mux.HandleFunc("GET /api/series", seriesHandler.ListSeries)
	mux.HandleFunc("GET /api/series/search", seriesHandler.SearchSeries)

	// API routes - Unified Search (media + series)
	mux.HandleFunc("GET /api/search", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"query required (use ?q=search_term)"}`)
			return
		}

		limit := 20
		if l := r.URL.Query().Get("limit"); l != "" {
			if n, err := strconv.Atoi(l); err == nil && n > 0 {
				limit = n
			}
		}

		// Search both media (movies only, not episodes) and series
		mediaResults, err := mediaRepo.ListFiltered(r.Context(), model.MediaListFilter{
			Search:    q,
			MediaType: "movie",
			Limit:     limit,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}

		seriesResults, err := seriesRepo.ListFiltered(r.Context(), model.SeriesListFilter{
			Search: q,
			Limit:  limit,
		})
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":%q}`, err.Error())
			return
		}

		result := model.SearchResult{
			Movies: mediaResults,
			Series: seriesResults,
		}
		if result.Movies == nil {
			result.Movies = []model.MediaListItem{}
		}
		if result.Series == nil {
			result.Series = []model.SeriesListItem{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": result})
	})
	mux.HandleFunc("GET /api/series/{id}", seriesHandler.GetSeries)
	mux.HandleFunc("GET /api/series/{id}/seasons", seriesHandler.ListSeasons)
	mux.HandleFunc("GET /api/series/{id}/seasons/{seasonId}/episodes", seriesHandler.ListEpisodes)

	// API routes - Metadata (admin only, nil-safe — handlers registered only when TMDb is configured)
	if metadataHandler != nil {
		mux.Handle("PUT /api/media/{id}/identify", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.Identify)))
		mux.Handle("POST /api/media/{id}/refresh", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.Refresh)))
		mux.Handle("POST /api/admin/metadata/refresh-ratings", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.BulkRefreshRatings)))
		mux.Handle("PATCH /api/media/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.EditMediaMetadata)))
		mux.Handle("PATCH /api/series/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.EditSeriesMetadata)))
		mux.Handle("PATCH /api/episodes/{id}/metadata", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.EditEpisodeMetadata)))
		mux.Handle("DELETE /api/media/{id}/metadata/lock", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.UnlockMediaMetadata)))
		mux.Handle("DELETE /api/series/{id}/metadata/lock", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.UnlockSeriesMetadata)))
		mux.Handle("POST /api/media/{id}/images", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.UploadMediaImage)))
		mux.Handle("POST /api/series/{id}/images", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.UploadSeriesImage)))
		mux.Handle("DELETE /api/media/{id}/images/{imageType}", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.DeleteMediaImage)))
		mux.Handle("DELETE /api/series/{id}/images/{imageType}", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.DeleteSeriesImage)))
		mux.Handle("POST /api/media/{id}/nfo", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.WriteMediaNFO)))
		mux.Handle("POST /api/series/{id}/nfo", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.WriteSeriesNFO)))
	}

	// API routes - Local images (public, no auth needed for cached images)
	if metadataHandler != nil {
		mux.HandleFunc("GET /api/images/local/{type}/{id}/{filename}", metadataHandler.ServeLocalImage)
	}

	// API routes - Stream URL (Jellyfin-style api_key for external players)
	mux.HandleFunc("POST /api/stream/{id}/url", func(w http.ResponseWriter, r *http.Request) {
		mediaID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"invalid id"}`)
			return
		}

		userID, isAdmin, ok := auth.UserFromContext(r.Context())
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, `{"error":"unauthorized"}`)
			return
		}

		apiKey := apiKeyStore.Generate(userID, isAdmin)

		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host

		directURL := fmt.Sprintf("%s://%s/api/stream/%d?api_key=%s", scheme, host, mediaID, apiKey)
		hlsURL := fmt.Sprintf("%s://%s/api/stream/%d/hls/master.m3u8?api_key=%s", scheme, host, mediaID, apiKey)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"direct_url": directURL,
				"hls_url":    hlsURL,
				"api_key":    apiKey,
				"expires_in": int(auth.StreamTokenExpiry.Seconds()),
			},
		})
	})

	// API routes - Cinema mode (trailers + intro before main video)
	mux.HandleFunc("GET /api/media/{id}/cinema", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"invalid id"}`)
			return
		}

		type cinemaItem struct {
			Type      string `json:"type"` // "intro" | "trailer" | "main"
			Title     string `json:"title"`
			URL       string `json:"url"`
			Duration  int    `json:"duration"` // seconds, 0 if unknown
			Skippable bool   `json:"skippable"`
		}

		var items []cinemaItem

		// 1. Cinema intro video (only if cinema mode enabled)
		cinemaEnabled, _ := appSettingsRepo.Get(r.Context(), "cinema_mode_enabled")
		introPath, _ := appSettingsRepo.Get(r.Context(), "cinema_intro_path")
		if introPath != "" && cinemaEnabled == "true" {
			items = append(items, cinemaItem{
				Type:      "intro",
				Title:     "Cinema Intro",
				URL:       "/api/cinema/intro",
				Skippable: true,
			})
		}

		// 2. Trailers from TMDb
		maxTrailers := 2
		if maxStr, _ := appSettingsRepo.Get(r.Context(), "cinema_max_trailers"); maxStr != "" {
			if n, err := strconv.Atoi(maxStr); err == nil && n >= 0 {
				maxTrailers = n
			}
		}

		if tmdbClient != nil && maxTrailers > 0 {
			media, err := mediaRepo.GetByID(r.Context(), id)
			if err == nil {
				var videos *tmdb.VideoList

				if media.MediaType == "episode" {
					// For episodes, fetch trailers from the parent series
					ep, epErr := episodeRepo.GetByMediaID(r.Context(), media.ID)
					if epErr == nil {
						series, sErr := seriesRepo.GetByID(r.Context(), ep.SeriesID)
						if sErr == nil && series.TmdbID != nil {
							tvDetails, tvErr := tmdbClient.GetTVDetails(r.Context(), int(*series.TmdbID))
							if tvErr == nil {
								videos = tvDetails.Videos
							}
						}
					}
				} else if media.TmdbID != nil {
					movieDetails, mErr := tmdbClient.GetMovieDetails(r.Context(), int(*media.TmdbID))
					if mErr == nil {
						videos = movieDetails.Videos
					}
				}

				if videos != nil {
					count := 0
					for _, v := range videos.Results {
						if count >= maxTrailers {
							break
						}
						if v.Site == "YouTube" && (v.Type == "Trailer" || v.Type == "Teaser") {
							items = append(items, cinemaItem{
								Type:      "trailer",
								Title:     v.Name,
								URL:       "https://www.youtube.com/embed/" + v.Key + "?autoplay=1&controls=0",
								Skippable: true,
							})
							count++
						}
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"items": items}})
	})

	// API routes - Cinema mode for series (trailers from TMDb TV)
	mux.HandleFunc("GET /api/series/{id}/cinema", func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"invalid id"}`)
			return
		}

		type cinemaItem struct {
			Type      string `json:"type"`
			Title     string `json:"title"`
			URL       string `json:"url"`
			Duration  int    `json:"duration"`
			Skippable bool   `json:"skippable"`
		}

		var items []cinemaItem

		maxTrailers := 2
		if maxStr, _ := appSettingsRepo.Get(r.Context(), "cinema_max_trailers"); maxStr != "" {
			if n, err := strconv.Atoi(maxStr); err == nil && n >= 0 {
				maxTrailers = n
			}
		}

		if tmdbClient != nil && maxTrailers > 0 {
			series, err := seriesRepo.GetByID(r.Context(), id)
			if err == nil && series.TmdbID != nil {
				tvDetails, tvErr := tmdbClient.GetTVDetails(r.Context(), int(*series.TmdbID))
				if tvErr == nil && tvDetails.Videos != nil {
					count := 0
					for _, v := range tvDetails.Videos.Results {
						if count >= maxTrailers {
							break
						}
						if v.Site == "YouTube" && (v.Type == "Trailer" || v.Type == "Teaser") {
							items = append(items, cinemaItem{
								Type:      "trailer",
								Title:     v.Name,
								URL:       "https://www.youtube.com/embed/" + v.Key + "?autoplay=1&controls=0",
								Skippable: true,
							})
							count++
						}
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"items": items}})
	})

	// API routes - Cinema intro video serve
	mux.HandleFunc("GET /api/cinema/intro", func(w http.ResponseWriter, r *http.Request) {
		introPath, _ := appSettingsRepo.Get(r.Context(), "cinema_intro_path")
		if introPath == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, introPath)
	})

	// API routes - Cinema intro upload (admin)
	mux.Handle("POST /api/admin/cinema/intro", middleware.RequireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 100<<20) // 100MB max for intro video
		if err := r.ParseMultipartForm(100 << 20); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"file too large (max 100MB)"}`)
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"missing file"}`)
			return
		}
		defer file.Close()

		// Save to data dir
		cinemaDir := cfg.DataDir + "/cinema"
		os.MkdirAll(cinemaDir, 0755)
		introPath := cinemaDir + "/intro.mp4"

		dst, err := os.Create(introPath)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"failed to save file"}`)
			return
		}
		defer dst.Close()

		io.Copy(dst, file)

		// Save path in settings
		appSettingsRepo.Set(r.Context(), "cinema_intro_path", introPath)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]string{"path": introPath}})
	})))

	// API routes - Streaming
	mux.HandleFunc("GET /api/stream/{id}", streamHandler.DirectPlay)
	mux.HandleFunc("GET /api/stream/{id}/hls/master.m3u8", streamHandler.HLSMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/abr.m3u8", streamHandler.HLSABRMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/{segment}", streamHandler.HLSSegment)

	// API routes - Trickplay (Plan E Phase 03)
	mux.HandleFunc("GET /api/media/{id}/trickplay/manifest.vtt", trickplayHandler.ServeVTT)
	mux.HandleFunc("GET /api/media/{id}/trickplay/{sprite}", trickplayHandler.ServeSprite)

	// API routes - Image Proxy (TMDb)
	mux.HandleFunc("GET /api/images/tmdb/{size}/{path...}", imageHandler.Serve)

	// API routes - Playback Decision
	mux.HandleFunc("POST /api/playback/{id}/info", playbackHandler.GetPlaybackInfo)
	mux.HandleFunc("GET /api/playback/capabilities", playbackHandler.GetClientCapabilities)

	// API routes - Subtitles (admin only for write, authenticated for read)
	mux.HandleFunc("GET /api/media-files/{media_file_id}/subtitles", subtitleHandler.ListByMediaFile)
	mux.HandleFunc("GET /api/media-files/{media_file_id}/subtitles/{subtitle_id}/serve", subtitleHandler.Serve)
	mux.Handle("POST /api/subtitles", middleware.RequireAdmin(http.HandlerFunc(subtitleHandler.Create)))
	mux.Handle("PATCH /api/subtitles/{id}", middleware.RequireAdmin(http.HandlerFunc(subtitleHandler.Update)))
	mux.Handle("DELETE /api/subtitles/{id}", middleware.RequireAdmin(http.HandlerFunc(subtitleHandler.Delete)))
	mux.Handle("POST /api/media-files/{media_file_id}/subtitles/{subtitle_id}/default", middleware.RequireAdmin(http.HandlerFunc(subtitleHandler.SetDefault)))
	mux.HandleFunc("POST /api/subtitles/{id}/translate", subtitleHandler.Translate)

	// API routes - Subtitle Search (external providers)
	mux.HandleFunc("GET /api/media/{id}/subtitles/search", subtitleSearchHandler.Search)
	mux.HandleFunc("POST /api/media/{id}/subtitles/download", subtitleSearchHandler.Download)

	// API routes - Notifications (WebSocket + REST)
	mux.HandleFunc("GET /api/ws", wsHandler.Handle)
	mux.HandleFunc("GET /api/notifications", notificationHandler.List)
	mux.HandleFunc("GET /api/notifications/unread-count", notificationHandler.UnreadCount)
	mux.HandleFunc("GET /api/notifications/types", notificationHandler.Types)
	mux.HandleFunc("PATCH /api/notifications/read", notificationHandler.MarkAsRead)
	mux.HandleFunc("PATCH /api/notifications/read-all", notificationHandler.MarkAllAsRead)
	mux.HandleFunc("DELETE /api/notifications/{id}", notificationHandler.DeleteOne)
	mux.HandleFunc("POST /api/notifications/delete", notificationHandler.Delete)

	// API routes - Audio Tracks (admin only for write, authenticated for read)
	mux.HandleFunc("GET /api/media-files/{media_file_id}/audio-tracks", audioTrackHandler.ListByMediaFile)
	mux.Handle("POST /api/audio-tracks", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Create)))
	mux.Handle("PATCH /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Update)))
	mux.Handle("DELETE /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Delete)))
	mux.Handle("POST /api/media-files/{media_file_id}/audio-tracks/{track_id}/default", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.SetDefault)))

	// Auth middleware with paths that don't require auth
	authMiddleware := middleware.RequireAuth(jwtManager, apiKeyStore,
		"/api/health",
		"/api/setup/status",
		"/api/setup",
		"/api/auth/login",
		"/api/auth/refresh",
		"/api/auth/logout",
		"/api/images/*",
	)

	// Session tracker middleware (per-session, not per-user)
	sessionTracker := middleware.SessionTracker(func(sessionID int64) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = sessionRepo.UpdateLastActive(ctx, sessionID)
	})

	// Middleware stack
	var h http.Handler = mux
	h = sessionTracker(h) // innermost: track session AFTER auth sets context
	h = authMiddleware(h) // auth: set user context
	h = middleware.CORS(cfg.CORSOrigin)(h)
	h = middleware.Logger(h)
	h = middleware.Recovery(h) // outermost

	// Scheduler (Plan F Phase 03)
	scheduler := service.NewScheduler()
	schedulerHandler := handler.NewSchedulerHandler(scheduler)

	// Register scheduler routes (admin only)
	mux.Handle("GET /api/admin/tasks", middleware.RequireAdmin(http.HandlerFunc(schedulerHandler.ListTasks)))
	mux.Handle("POST /api/admin/tasks/{name}/run", middleware.RequireAdmin(http.HandlerFunc(schedulerHandler.RunTask)))

	// File watcher (Phase 03) — watches library paths for new/removed video files
	if cfg.FileWatcherEnabled {
		fileWatcher, err := watcher.New(
			// onCreate: process the new file through the scan pipeline
			func(libraryID int64, path string) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()
				job, err := pipeline.CreateJob(ctx, libraryID)
				if err != nil {
					log.Printf("watcher: failed to create scan job for %s: %v", path, err)
					return
				}
				if err := pipeline.RunJob(ctx, job, false); err != nil {
					log.Printf("watcher: scan job failed for %s: %v", path, err)
				}
				// Notify users when watcher scan finds new files
				if job.NewFiles > 0 {
					libName := fmt.Sprintf("Library #%d", libraryID)
					if lib, err := libraryRepo.GetByID(context.Background(), libraryID); err == nil {
						libName = lib.Name
					}
					if err := notificationSvc.NotifyLibraryWatcher(context.Background(), libraryID, libName, job.NewFiles); err != nil {
						log.Printf("watcher: notify library %d: %v", libraryID, err)
					}
				}
			},
			// onRemove: mark file as missing via verifier
			func(libraryID int64, path string) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				existing, err := mediaFileRepo.FindByPath(ctx, path)
				if err != nil {
					return // file not in DB, nothing to do
				}
				if err := mediaFileRepo.MarkMissing(ctx, existing.ID); err != nil {
					log.Printf("watcher: failed to mark missing %s: %v", path, err)
				} else {
					log.Printf("watcher: marked missing %s", path)
				}
			},
		)
		if err != nil {
			log.Printf("file watcher: failed to initialize: %v", err)
		} else {
			// Add all existing libraries to the watcher
			libs, err := libraryRepo.List(context.Background())
			if err != nil {
				log.Printf("file watcher: failed to list libraries: %v", err)
			} else {
				for _, lib := range libs {
					if err := fileWatcher.AddLibrary(lib.ID, lib.Paths); err != nil {
						log.Printf("file watcher: failed to watch library %d: %v", lib.ID, err)
					}
				}
			}
			fileWatcher.Start()
			defer fileWatcher.Stop()
			log.Printf("file watcher enabled (%d libraries)", len(libs))
		}
	}

	// File verification — check for missing files at startup
	verifier := scanner.NewVerifier(mediaFileRepo)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		result, err := verifier.VerifyAll(ctx)
		if err != nil {
			log.Printf("startup verification: %v", err)
		} else if result.Missing > 0 {
			log.Printf("startup verification: %d missing files detected", result.Missing)
		}
	}()

	// Register scheduled tasks (Plan F)
	scheduler.Register("session-cleanup", 1*time.Hour, func(ctx context.Context) error {
		if err := refreshTokenRepo.DeleteExpired(ctx); err != nil {
			log.Printf("cleanup expired tokens: %v", err)
		}
		if err := sessionRepo.DeleteExpired(ctx); err != nil {
			log.Printf("cleanup expired sessions: %v", err)
		}
		return nil
	})
	scheduler.Register("missing-file-check", 6*time.Hour, func(ctx context.Context) error {
		_, err := verifier.VerifyAll(ctx)
		return err
	})
	scheduler.Register("transcode-cleanup", 24*time.Hour, func(ctx context.Context) error {
		return transcoderSvc.CleanupOlderThan(7 * 24 * time.Hour)
	})
	scheduler.Register("library-scan", 24*time.Hour, func(ctx context.Context) error {
		libs, err := libraryRepo.List(ctx)
		if err != nil {
			return fmt.Errorf("listing libraries: %w", err)
		}
		for _, lib := range libs {
			job, err := pipeline.CreateJob(ctx, lib.ID)
			if err != nil {
				log.Printf("scheduled scan: failed to create job for library %d: %v", lib.ID, err)
				continue
			}
			if err := pipeline.RunJob(ctx, job, false); err != nil {
				log.Printf("scheduled scan: library %d job %d failed: %v", lib.ID, job.ID, err)
			}
		}
		return nil
	})
	scheduler.Register("notification-cleanup", 24*time.Hour, func(ctx context.Context) error {
		// Delete notifications older than 30 days that are already read
		before := time.Now().Add(-30 * 24 * time.Hour)
		deleted, err := notificationRepo.DeleteOld(ctx, before)
		if err != nil {
			return fmt.Errorf("cleanup old notifications: %w", err)
		}
		if deleted > 0 {
			log.Printf("notification cleanup: deleted %d old notifications", deleted)
		}
		return nil
	})
	scheduler.Start()
	defer scheduler.Stop()

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Minute, // long for streaming
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("velox server listening on %s", cfg.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-done
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}
