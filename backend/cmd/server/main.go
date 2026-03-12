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
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/transcoder"
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

	// Services
	pipeline := scanner.NewPipeline(
		db, libraryRepo, mediaRepo, mediaFileRepo,
		seriesRepo, seasonRepo, episodeRepo,
		scanJobRepo, subtitleRepo, audioTrackRepo,
	)
	transcoderSvc := transcoder.New(cfg.TranscodePath)
	librarySvc := service.NewLibraryService(libraryRepo, scanJobRepo, pipeline)
	mediaSvc := service.NewMediaService(mediaRepo, mediaFileRepo)
	streamSvc := service.NewStreamService(mediaFileRepo, audioTrackRepo, transcoderSvc)
	authSvc := service.NewAuthService(userRepo, refreshTokenRepo, sessionRepo, jwtManager, db)
	userDataSvc := service.NewUserDataService(userDataRepo)
	subtitleSvc := service.NewSubtitleService(subtitleRepo)
	audioTrackSvc := service.NewAudioTrackService(audioTrackRepo)

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

	// API routes - Libraries (read = authenticated, write = admin above)
	mux.HandleFunc("GET /api/libraries", libraryHandler.List)

	// API routes - Media
	mux.HandleFunc("GET /api/media", mediaHandler.List)
	mux.HandleFunc("GET /api/media/{id}", mediaHandler.Get)
	mux.HandleFunc("GET /api/media/{id}/files", mediaHandler.GetWithFiles)
	mux.HandleFunc("GET /api/media/{id}/versions", mediaHandler.GetVersions)

	// API routes - Streaming
	mux.HandleFunc("GET /api/stream/{id}", streamHandler.DirectPlay)
	mux.HandleFunc("GET /api/stream/{id}/hls/master.m3u8", streamHandler.HLSMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/{segment}", streamHandler.HLSSegment)

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

	// Cleanup expired tokens/sessions every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := refreshTokenRepo.DeleteExpired(ctx); err != nil {
				log.Printf("cleanup expired tokens: %v", err)
			}
			if err := sessionRepo.DeleteExpired(ctx); err != nil {
				log.Printf("cleanup expired sessions: %v", err)
			}
			cancel()
		}
	}()

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
