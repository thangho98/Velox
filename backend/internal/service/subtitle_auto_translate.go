package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/thawng/velox/internal/model"
)

// AutoTranslateAll checks all media files and translates missing subtitles
// according to the admin configuration.
func (s *SubtitleService) AutoTranslateAll(ctx context.Context) error {
	if s.settingsRepo == nil || s.mediaFileRepo == nil {
		return fmt.Errorf("missing repositories for subtitle auto-translation")
	}

	// 1. Check if feature is enabled
	vals, err := s.settingsRepo.GetMulti(
		ctx,
		model.SettingAutoTranslateEnabled,
		model.SettingAutoTranslateLanguages,
	)
	if err != nil {
		return fmt.Errorf("failed to fetch auto-translate settings: %w", err)
	}

	enabled := vals[model.SettingAutoTranslateEnabled]
	if enabled != "true" && enabled != "1" {
		return nil // Not enabled
	}

	targetsStr := vals[model.SettingAutoTranslateLanguages]
	if targetsStr == "" {
		return nil // No languages configured
	}

	var targetLangs []string
	for _, l := range strings.Split(targetsStr, ",") {
		lang := strings.TrimSpace(l)
		if lang != "" {
			targetLangs = append(targetLangs, lang)
		}
	}

	if len(targetLangs) == 0 {
		return nil
	}

	slog.Info("running subtitle auto-translation", "targets", targetLangs)

	// 2. Iterate paginated over all media files
	limit := 100
	offset := 0
	count := 0
	translatedCount := 0

	for {
		// Stop if context cancelled
		if err := ctx.Err(); err != nil {
			slog.Info("auto-translation stopped by context", "processed", count)
			break
		}

		mediaFiles, err := s.mediaFileRepo.ListAllPaginated(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("listing media files: %w", err)
		}

		if len(mediaFiles) == 0 {
			break // Done
		}

		for _, mf := range mediaFiles {
			// Stop if context cancelled
			if err := ctx.Err(); err != nil {
				break
			}

			count++
			if mf.LastVerifiedAt == nil {
				continue
			}

			// Get subtitles for media
			subs, err := s.subtitleRepo.ListByMediaFileID(ctx, mf.ID)
			if err != nil || len(subs) == 0 {
				continue // If no subtitles exist at all -> skip it (can't translate logic)
			}

			// Find what languages exist
			existingLangs := make(map[string]bool)
			// Also find a good source subtitle (e.g. English, or just the first valid one)
			var sourceSub *model.Subtitle

			for _, sub := range subs {
				lang := strings.ToLower(sub.Language)
				existingLangs[lang] = true

				// Prefer english as source if possible
				if sourceSub == nil {
					sourceSub = &sub // copy first sub
				} else if lang == "en" {
					englishSub := sub // local copy
					sourceSub = &englishSub
				}
			}

			if sourceSub == nil {
				continue
			}

			// Check missing targets
			for _, target := range targetLangs {
				target = strings.ToLower(target)

				// Same language
				if strings.ToLower(sourceSub.Language) == target {
					continue
				}

				if !existingLangs[target] {
					// Missing language => Translate
					slog.Info("auto-translate: translating missing subtitle",
						"media_file_id", mf.ID,
						"source_lang", sourceSub.Language,
						"target", target,
					)

					_, err := s.Translate(ctx, sourceSub.ID, target)
					if err != nil {
						slog.Warn("auto-translate: failed to translate",
							"subtitle_id", sourceSub.ID,
							"target", target,
							"error", err,
						)
					} else {
						translatedCount++
					}
				}
			}
		}

		offset += limit
	}

	slog.Info("subtitle auto-translation finished", "scanned_files", count, "translated_subs", translatedCount)
	return nil
}
