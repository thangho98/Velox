package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/thawng/velox/internal/database"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/service"
)

// batchAutoDownload iterates all media files and runs AutoDownload for missing subtitles.
func batchAutoDownload(dbPath, subtitleDir string, langs []string, dryRun bool, titleFilter string, limitN int) {
	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	// Set auto_sub_languages
	appSettingsRepo := repository.NewAppSettingsRepo(db)
	langStr := strings.Join(langs, ",")
	if err := appSettingsRepo.Set(context.Background(), model.SettingAutoSubLanguages, langStr); err != nil {
		log.Fatalf("setting auto_sub_languages: %v", err)
	}
	log.Printf("auto_sub_languages set to %q", langStr)

	// Build service
	mediaRepo := repository.NewMediaRepo(db)
	mfRepo := repository.NewMediaFileRepo(db)
	subtitleRepo := repository.NewSubtitleRepo(db)
	episodeRepo := repository.NewEpisodeRepo(db)
	seasonRepo := repository.NewSeasonRepo(db)
	seriesRepo := repository.NewSeriesRepo(db)

	svc := service.NewSubtitleSearchService(
		mediaRepo, mfRepo, subtitleRepo, appSettingsRepo,
		episodeRepo, seasonRepo, seriesRepo, subtitleDir,
	)

	// Query all media + primary files
	type mediaFile struct {
		mediaID   int64
		fileID    int64
		title     string
		mediaType string
	}

	query := `
		SELECT m.id, mf.id, m.title, m.media_type
		FROM media m
		JOIN media_files mf ON mf.media_id = m.id AND mf.is_primary = 1
	`
	var args []any
	if titleFilter != "" {
		query += " WHERE LOWER(m.title) LIKE ?"
		args = append(args, "%"+strings.ToLower(titleFilter)+"%")
	}
	query += " ORDER BY m.id"
	if limitN > 0 {
		query += fmt.Sprintf(" LIMIT %d", limitN)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		log.Fatalf("querying media: %v", err)
	}

	var items []mediaFile
	for rows.Next() {
		var mf mediaFile
		if err := rows.Scan(&mf.mediaID, &mf.fileID, &mf.title, &mf.mediaType); err != nil {
			log.Fatalf("scanning row: %v", err)
		}
		items = append(items, mf)
	}
	rows.Close()

	log.Printf("found %d media files to process", len(items))

	if dryRun {
		// Just check what's missing
		for _, mf := range items {
			missing := getMissingLangs(db, mf.fileID, langs)
			if len(missing) > 0 {
				fmt.Printf("[%d] %s — missing: %s\n", mf.mediaID, mf.title, strings.Join(missing, ", "))
			}
		}
		return
	}

	// Process each media
	downloaded := 0
	skipped := 0
	failed := 0

	for i, mf := range items {
		missing := getMissingLangs(db, mf.fileID, langs)
		if len(missing) == 0 {
			skipped++
			continue
		}

		fmt.Printf("[%d/%d] %s — downloading %s...\n", i+1, len(items), mf.title, strings.Join(missing, ", "))

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		err := svc.AutoDownload(ctx, mf.mediaID, mf.fileID)
		cancel()

		if err != nil {
			log.Printf("  ERROR: %v", err)
			failed++
		} else {
			downloaded++
		}

		// Small delay to be nice to providers
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Printf("\n=== Done ===\n")
	fmt.Printf("Processed: %d, Downloaded: %d, Skipped (already have): %d, Failed: %d\n",
		len(items), downloaded, skipped, failed)
}

func getMissingLangs(db *sql.DB, fileID int64, langs []string) []string {
	var missing []string
	for _, lang := range langs {
		var count int
		// Only count text-based subs (srt, ass, vtt, subrip, webvtt).
		// PGS/VOBSUB are image-based and can't be used with Direct Play.
		err := db.QueryRow(`
			SELECT COUNT(*) FROM subtitles
			WHERE media_file_id = ?
			AND LOWER(language) IN (?, ?, ?)
			AND LOWER(codec) IN ('subrip', 'srt', 'ass', 'ssa', 'webvtt', 'vtt', 'mov_text')
		`, fileID, lang, longLang(lang), lang[:2]).Scan(&count)
		if err != nil || count == 0 {
			missing = append(missing, lang)
		}
	}
	return missing
}

func longLang(iso string) string {
	m := map[string]string{
		"en": "eng", "vi": "vie", "fr": "fre", "de": "ger",
		"es": "spa", "ja": "jpn", "ko": "kor", "zh": "chi",
	}
	if v, ok := m[iso]; ok {
		return v
	}
	return iso
}
