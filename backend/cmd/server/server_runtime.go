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

	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/watcher"
)

func (app *serverApp) startBackgroundServices() func() {
	fileWatcher := app.startFileWatcher()
	app.startStartupVerification()
	app.registerScheduledTasks()
	app.services.scheduler.Start()
	app.services.pretranscode.Start()
	go app.services.imagemeta.Worker().Run(context.Background())

	return func() {
		app.services.pretranscode.Stop()
		app.services.scheduler.Stop()
		if fileWatcher != nil {
			fileWatcher.Stop()
		}
	}
}

func (app *serverApp) startFileWatcher() *watcher.Watcher {
	if !app.cfg.FileWatcherEnabled {
		return nil
	}

	fileWatcher, err := watcher.New(
		func(libraryID int64, path string) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			job, err := app.services.pipeline.CreateJob(ctx, libraryID)
			if err != nil {
				log.Printf("watcher: failed to create scan job for %s: %v", path, err)
				return
			}
			if err := app.services.pipeline.RunJob(ctx, job, false); err != nil {
				log.Printf("watcher: scan job failed for %s: %v", path, err)
			}

			if job.NewFiles <= 0 {
				return
			}

			libraryName := fmt.Sprintf("Library #%d", libraryID)
			if library, err := app.repos.library.GetByID(context.Background(), libraryID); err == nil {
				libraryName = library.Name
			}
			if err := app.services.notification.NotifyLibraryWatcher(context.Background(), libraryID, libraryName, job.NewFiles); err != nil {
				log.Printf("watcher: notify library %d: %v", libraryID, err)
			}
		},
		func(_ int64, path string) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			existing, err := app.repos.mediaFile.FindByPath(ctx, path)
			if err != nil {
				return
			}
			if err := app.repos.mediaFile.MarkMissing(ctx, existing.ID); err != nil {
				log.Printf("watcher: failed to mark missing %s: %v", path, err)
				return
			}
			log.Printf("watcher: marked missing %s", path)
		},
	)
	if err != nil {
		log.Printf("file watcher: failed to initialize: %v", err)
		return nil
	}

	libraries, err := app.repos.library.List(context.Background())
	if err != nil {
		log.Printf("file watcher: failed to list libraries: %v", err)
	} else {
		for _, library := range libraries {
			if err := fileWatcher.AddLibrary(library.ID, library.Paths); err != nil {
				log.Printf("file watcher: failed to watch library %d: %v", library.ID, err)
			}
		}
	}

	fileWatcher.Start()
	log.Printf("file watcher enabled (%d libraries)", len(libraries))

	return fileWatcher
}

func (app *serverApp) startStartupVerification() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		result, err := app.services.verifier.VerifyAll(ctx)
		if err != nil {
			log.Printf("startup verification: %v", err)
			return
		}
		if result.Missing > 0 {
			log.Printf("startup verification: %d missing files detected", result.Missing)
		}
	}()
}

func (app *serverApp) registerScheduledTasks() {
	app.services.scheduler.Register("session-cleanup", 1*time.Hour, func(ctx context.Context) error {
		if err := app.repos.refreshToken.DeleteExpired(ctx); err != nil {
			log.Printf("cleanup expired tokens: %v", err)
		}
		if err := app.repos.session.DeleteExpired(ctx); err != nil {
			log.Printf("cleanup expired sessions: %v", err)
		}
		return nil
	})

	app.services.scheduler.Register("missing-file-check", 6*time.Hour, func(ctx context.Context) error {
		_, err := app.services.verifier.VerifyAll(ctx)
		return err
	})

	app.services.scheduler.Register("transcode-cleanup", 1*time.Hour, func(ctx context.Context) error {
		return app.services.transcoder.CleanupOlderThan(2 * time.Hour)
	})

	app.services.scheduler.Register("library-scan", 24*time.Hour, func(ctx context.Context) error {
		libraries, err := app.repos.library.List(ctx)
		if err != nil {
			return fmt.Errorf("listing libraries: %w", err)
		}

		for _, library := range libraries {
			job, err := app.services.pipeline.CreateJob(ctx, library.ID)
			if err != nil {
				log.Printf("scheduled scan: failed to create job for library %d: %v", library.ID, err)
				continue
			}
			if err := app.services.pipeline.RunJob(ctx, job, false); err != nil {
				log.Printf("scheduled scan: library %d job %d failed: %v", library.ID, job.ID, err)
			}
		}

		return nil
	})

	app.services.scheduler.Register("notification-cleanup", 24*time.Hour, func(ctx context.Context) error {
		before := time.Now().Add(-30 * 24 * time.Hour)
		deleted, err := app.repos.notification.DeleteOld(ctx, before)
		if err != nil {
			return fmt.Errorf("cleanup old notifications: %w", err)
		}
		if deleted > 0 {
			log.Printf("notification cleanup: deleted %d old notifications", deleted)
		}
		return nil
	})

	app.services.scheduler.Register("subtitle-auto-translate", 6*time.Hour, func(ctx context.Context) error {
		return app.services.subtitle.AutoTranslateAll(ctx)
	})

	app.services.scheduler.Register("subtitle-auto-download", 24*time.Hour, func(ctx context.Context) error {
		if app.services.subtitleSearch == nil {
			return nil
		}
		limit := 100
		offset := 0
		for {
			files, err := app.repos.mediaFile.ListAllPaginated(ctx, limit, offset)
			if err != nil {
				return fmt.Errorf("listing media files: %w", err)
			}
			if len(files) == 0 {
				break
			}
			for _, f := range files {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				_ = app.services.subtitleSearch.AutoDownload(ctx, f.MediaID, f.ID)
			}
			offset += limit
		}
		return nil
	})

	// Cloud provider session refresh (Plan W). fshare TTL is ~30min; refresh
	// tick every 5min to catch providers in the "expiring within 5min" window.
	if app.cloudRegistry != nil {
		refreshSvc := service.NewProviderRefreshService(
			app.repos.storageProvider,
			app.cloudRegistry,
			app.cfg.Cloud.SecretKey,
			nil, // use default logger
		)
		app.services.scheduler.Register("cloud-provider-refresh", 5*time.Minute, func(ctx context.Context) error {
			return refreshSvc.RefreshExpiring(ctx, 5*time.Minute)
		})

		// Backfill cloud media metadata: probe unindexed cloud files for codec,
		// resolution, audio tracks, and embedded subtitles. Runs sequentially
		// to avoid FShare rate limiting. Only processes files with empty video_codec.
		app.services.scheduler.Register("cloud-media-probe", 6*time.Hour, func(ctx context.Context) error {
			const batchSize = 50
			probed := 0
			for {
				// Always query from offset 0: successfully probed files drop out of the
				// result set (video_codec becomes non-empty), so the window shifts naturally.
				// Only increment offset by the number of failures to skip past stuck files.
				files, err := app.repos.mediaFile.ListUnprobedCloud(ctx, batchSize, 0)
				if err != nil {
					return fmt.Errorf("listing unprobed cloud files: %w", err)
				}
				if len(files) == 0 {
					break
				}
				batchFailed := 0
				for i := range files {
					if ctx.Err() != nil {
						return ctx.Err()
					}
					if err := app.services.stream.ProbeAndUpdateCloudMetadata(ctx, &files[i]); err != nil {
						log.Printf("cloud-media-probe: file %d: %v", files[i].ID, err)
						batchFailed++
						continue
					}
					probed++
				}
				// If every file in the batch failed, stop to avoid infinite loop.
				if batchFailed == len(files) {
					log.Printf("cloud-media-probe: entire batch failed (%d files), stopping", batchFailed)
					break
				}
			}
			if probed > 0 {
				log.Printf("cloud-media-probe: probed %d files", probed)
			}
			return nil
		})
	}
}

func serve(server *http.Server) error {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(done)

	errCh := make(chan error, 1)
	go func() {
		log.Printf("velox server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err, ok := <-errCh:
		if ok && err != nil {
			return err
		}
		return nil
	case <-done:
	}

	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	if err, ok := <-errCh; ok && err != nil {
		return err
	}

	log.Println("server stopped")
	return nil
}
