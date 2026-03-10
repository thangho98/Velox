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

	// Repositories
	libraryRepo := repository.NewLibraryRepo(db)
	mediaRepo := repository.NewMediaRepo(db)
	progressRepo := repository.NewProgressRepo(db)

	// Services
	scannerSvc := scanner.New(mediaRepo, libraryRepo)
	transcoderSvc := transcoder.New(cfg.TranscodePath)
	librarySvc := service.NewLibraryService(libraryRepo, scannerSvc)
	mediaSvc := service.NewMediaService(mediaRepo)
	streamSvc := service.NewStreamService(mediaRepo, transcoderSvc)
	progressSvc := service.NewProgressService(progressRepo)

	// Router
	mux := http.NewServeMux()

	libraryHandler := handler.NewLibraryHandler(librarySvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)
	streamHandler := handler.NewStreamHandler(streamSvc)
	progressHandler := handler.NewProgressHandler(progressSvc)

	// API routes
	mux.HandleFunc("GET /api/libraries", libraryHandler.List)
	mux.HandleFunc("POST /api/libraries", libraryHandler.Create)
	mux.HandleFunc("DELETE /api/libraries/{id}", libraryHandler.Delete)
	mux.HandleFunc("POST /api/libraries/{id}/scan", libraryHandler.Scan)

	mux.HandleFunc("GET /api/media", mediaHandler.List)
	mux.HandleFunc("GET /api/media/{id}", mediaHandler.Get)

	mux.HandleFunc("GET /api/stream/{id}", streamHandler.DirectPlay)
	mux.HandleFunc("GET /api/stream/{id}/hls/master.m3u8", streamHandler.HLSMaster)
	mux.HandleFunc("GET /api/stream/{id}/hls/{segment}", streamHandler.HLSSegment)

	mux.HandleFunc("GET /api/progress/{mediaID}", progressHandler.Get)
	mux.HandleFunc("PUT /api/progress/{mediaID}", progressHandler.Update)

	// Middleware stack
	var h http.Handler = mux
	h = middleware.CORS(cfg.CORSOrigin)(h)
	h = middleware.Logger(h)
	h = middleware.Recovery(h)

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
