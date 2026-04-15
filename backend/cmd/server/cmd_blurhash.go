package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/thawng/velox/internal/config"
	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service/imagemeta"
	"github.com/thawng/velox/internal/storage"
	"golang.org/x/time/rate"
)

func runBlurhash() {
	if len(os.Args) < 3 || os.Args[2] != "backfill" {
		fmt.Fprintln(os.Stderr, "Usage: velox blurhash backfill [--force] [--limit=N]")
		os.Exit(1)
	}

	force := false
	limit := 0
	for _, arg := range os.Args[3:] {
		if arg == "--force" {
			force = true
		} else if strings.HasPrefix(arg, "--limit=") {
			l, err := strconv.Atoi(strings.TrimPrefix(arg, "--limit="))
			if err == nil {
				limit = l
			}
		}
	}

	cfg := config.Load()
	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize dependencies
	metaRepo := repository.NewImageMetadataRepo(db)
	imgStorage := storage.NewImageStorage(cfg.DataDir)
	svc := imagemeta.NewService(metaRepo, imgStorage, http.DefaultClient, rate.NewLimiter(20, 1))

	query := `
		SELECT DISTINCT path FROM (
			SELECT poster_path as path FROM media WHERE poster_path != ''
			UNION SELECT backdrop_path as path FROM media WHERE backdrop_path != ''
			UNION SELECT logo_path as path FROM media WHERE logo_path != ''
			UNION SELECT thumb_path as path FROM media WHERE thumb_path != ''
			UNION SELECT poster_path as path FROM series WHERE poster_path != ''
			UNION SELECT backdrop_path as path FROM series WHERE backdrop_path != ''
			UNION SELECT logo_path as path FROM series WHERE logo_path != ''
			UNION SELECT thumb_path as path FROM series WHERE thumb_path != ''
		) WHERE path IS NOT NULL AND path != ''
	`

	if !force {
		// exclude already computed ones
		query += ` AND path NOT IN (SELECT path FROM image_metadata)`
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	log.Printf("Finding images to compute...")
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("failed to query paths: %v", err)
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err == nil {
			paths = append(paths, p)
		}
	}

	total := len(paths)
	log.Printf("Found %d pending images for blurhash computation.", total)
	if total == 0 {
		return
	}

	startTime := time.Now()
	computed := 0
	failed := 0
	ctx := context.Background()

	// Batch sizes of 50 for progress logging
	batchSize := 50
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}
		batch := paths[i:end]

		errs := svc.ComputeBatch(ctx, batch)

		batchFailed := len(errs)
		batchSuccess := len(batch) - batchFailed

		computed += batchSuccess
		failed += batchFailed

		if batchFailed > 0 {
			for p, err := range errs {
				log.Printf("  error %s: %v", p, err)
			}
		}

		elapsed := time.Since(startTime).Round(time.Second)
		log.Printf("blurhash: [%d/%d] computed=%d failed=%d elapsed=%s", end, total, computed, failed, elapsed)
	}

	log.Printf("Done. Processed %d images. computed=%d failed=%d duration=%s", total, computed, failed, time.Since(startTime).Round(time.Second))
}
