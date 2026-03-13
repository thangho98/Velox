package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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
	"github.com/thawng/velox/internal/transcoder"
	"github.com/thawng/velox/internal/trickplay"
	"github.com/thawng/velox/internal/watcher"
	"github.com/thawng/velox/pkg/fanart"
	"github.com/thawng/velox/pkg/omdb"
	"github.com/thawng/velox/pkg/thetvdb"
	"github.com/thawng/velox/pkg/tmdb"
	"github.com/thawng/velox/pkg/tvmaze"
)

func main() {
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

	// Services
	pipeline := scanner.NewPipeline(
		db, libraryRepo, mediaRepo, mediaFileRepo,
		seriesRepo, seasonRepo, episodeRepo,
		scanJobRepo, subtitleRepo, audioTrackRepo,
	)
	// TMDb metadata enrichment (Phase 04)
	appSettingsRepo := repository.NewAppSettingsRepo(db)
	tmdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingTMDbAPIKey)
	if tmdbAPIKey == "" {
		tmdbAPIKey = tmdb.DefaultAPIKey
		log.Println("TMDb using built-in API key (override in Settings → Metadata)")
	} else {
		log.Println("TMDb using custom API key from settings")
	}
	tmdbClient := tmdb.New(tmdbAPIKey)
	metadataSvc := service.NewMetadataService(tmdbClient, mediaRepo, mediaFileRepo, seriesRepo, seasonRepo, episodeRepo, genreRepo, personRepo)
	if metadataSvc != nil {
		pipeline.SetMetadataMatcher(metadataSvc)
		log.Println("TMDb metadata enrichment enabled")
	}

	// OMDb ratings enrichment
	omdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingOMDbAPIKey)
	if omdbAPIKey == "" {
		omdbAPIKey = omdb.DefaultAPIKey
		log.Println("OMDb using built-in API key (override in Settings → Metadata)")
	} else {
		log.Println("OMDb using custom API key from settings")
	}
	omdbClient := omdb.New(omdbAPIKey)
	if metadataSvc != nil {
		metadataSvc.SetOMDbClient(omdbClient)
		log.Println("OMDb ratings enrichment enabled (IMDb, Rotten Tomatoes, Metacritic)")
	}

	// TheTVDB metadata enrichment
	tvdbAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingTVDBAPIKey)
	if tvdbAPIKey == "" {
		tvdbAPIKey = thetvdb.DefaultAPIKey
		log.Println("TheTVDB using built-in API key (override in Settings → Metadata)")
	} else {
		log.Println("TheTVDB using custom API key from settings")
	}
	tvdbClient := thetvdb.New(tvdbAPIKey)
	if metadataSvc != nil {
		metadataSvc.SetTVDBClient(tvdbClient)
		log.Println("TheTVDB metadata enrichment enabled")
	}

	// Fanart.tv artwork enrichment
	fanartAPIKey, _ := appSettingsRepo.Get(context.Background(), model.SettingFanartAPIKey)
	if fanartAPIKey == "" {
		fanartAPIKey = fanart.DefaultAPIKey
		log.Println("fanart.tv using built-in API key (override in Settings → Metadata)")
	} else {
		log.Println("fanart.tv using custom API key from settings")
	}
	fanartClient := fanart.New(fanartAPIKey)
	if metadataSvc != nil {
		metadataSvc.SetFanartClient(fanartClient)
		log.Println("fanart.tv artwork enrichment enabled (logos, thumbs)")
	}

	// TVmaze TV enrichment (free, no API key)
	tvmazeClient := tvmaze.New()
	if metadataSvc != nil {
		metadataSvc.SetTVmazeClient(tvmazeClient)
		log.Println("TVmaze enrichment enabled (network, schedule, ID cross-reference)")
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
	subtitleSvc := service.NewSubtitleService(subtitleRepo)
	audioTrackSvc := service.NewAudioTrackService(audioTrackRepo)

	// Plan F: Admin & Operations services
	activitySvc := service.NewActivityService(activityRepo)
	defer activitySvc.Close()
	adminSvc := service.NewAdminService(db, userRepo, startTime, hwAccel, cfg.DatabasePath)
	webhookSvc := service.NewWebhookService(webhookRepo)

	// Handlers
	libraryHandler := handler.NewLibraryHandler(librarySvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)
	setupHandler := handler.NewSetupHandler(authSvc)
	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(authSvc)
	profileHandler := handler.NewProfileHandler(authSvc, prefsRepo, userDataSvc)
	playbackHandler := handler.NewPlaybackHandler(mediaSvc, streamSvc, userDataSvc, subtitleSvc, audioTrackSvc, prefsRepo)
	subtitleHandler := handler.NewSubtitleHandler(subtitleSvc, mediaFileRepo, cfg.SubtitleCachePath)
	audioTrackHandler := handler.NewAudioTrackHandler(audioTrackSvc)
	settingsHandler := handler.NewSettingsHandler(appSettingsRepo)
	subtitleSearchSvc := service.NewSubtitleSearchService(mediaRepo, mediaFileRepo, subtitleRepo, appSettingsRepo, cfg.SubtitleCachePath)
	subtitleSearchHandler := handler.NewSubtitleSearchHandler(subtitleSearchSvc)

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
	metadataHandler := handler.NewMetadataHandler(mediaSvc, metadataSvc)
	activityHandler := handler.NewActivityHandler(activitySvc)
	adminHandler := handler.NewAdminHandler(adminSvc)
	webhookHandler := handler.NewWebhookHandler(webhookSvc)

	// Router
	mux := http.NewServeMux()

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

	// API routes - Series
	mux.HandleFunc("GET /api/series", seriesHandler.ListSeries)
	mux.HandleFunc("GET /api/series/search", seriesHandler.SearchSeries)
	mux.HandleFunc("GET /api/series/{id}", seriesHandler.GetSeries)
	mux.HandleFunc("GET /api/series/{id}/seasons", seriesHandler.ListSeasons)
	mux.HandleFunc("GET /api/series/{id}/seasons/{seasonId}/episodes", seriesHandler.ListEpisodes)

	// API routes - Metadata (admin only, nil-safe — handlers registered only when TMDb is configured)
	if metadataHandler != nil {
		mux.Handle("PUT /api/media/{id}/identify", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.Identify)))
		mux.Handle("POST /api/media/{id}/refresh", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.Refresh)))
		mux.Handle("POST /api/admin/metadata/refresh-ratings", middleware.RequireAdmin(http.HandlerFunc(metadataHandler.BulkRefreshRatings)))
	}

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

	// API routes - Subtitle Search (external providers)
	mux.HandleFunc("GET /api/media/{id}/subtitles/search", subtitleSearchHandler.Search)
	mux.HandleFunc("POST /api/media/{id}/subtitles/download", subtitleSearchHandler.Download)

	// API routes - Audio Tracks (admin only for write, authenticated for read)
	mux.HandleFunc("GET /api/media-files/{media_file_id}/audio-tracks", audioTrackHandler.ListByMediaFile)
	mux.Handle("POST /api/audio-tracks", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Create)))
	mux.Handle("PATCH /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Update)))
	mux.Handle("DELETE /api/audio-tracks/{id}", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.Delete)))
	mux.Handle("POST /api/media-files/{media_file_id}/audio-tracks/{track_id}/default", middleware.RequireAdmin(http.HandlerFunc(audioTrackHandler.SetDefault)))

	// Auth middleware with paths that don't require auth
	authMiddleware := middleware.RequireAuth(jwtManager,
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
