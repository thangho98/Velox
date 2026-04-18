package scanner

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/thawng/velox/internal/cloudstorage"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/ffprobe"
	"github.com/thawng/velox/pkg/nameparser"
)

// VideoExtensions is the canonical set of supported video file extensions.
var VideoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	".ts": true, ".mpg": true, ".mpeg": true, ".m2ts": true,
}

// IsVideoFile checks if a file path has a video extension.
func IsVideoFile(path string) bool {
	return VideoExtensions[strings.ToLower(filepath.Ext(path))]
}

// CloudProbeFunc resolves a cloud native ID to an HTTP download URL then runs ffprobe on it.
// The provider is passed in so the caller can reuse the same authenticated session.
type CloudProbeFunc func(ctx context.Context, provider cloudstorage.Provider, nativeID string) (*ffprobe.ProbeResult, error)

type Scanner struct {
	db             *sql.DB
	mediaRepo      *repository.MediaRepo
	mediaFileRepo  *repository.MediaFileRepo
	libraryRepo    *repository.LibraryRepo
	subtitleRepo   *repository.SubtitleRepo
	audioTrackRepo *repository.AudioTrackRepo
	metadataSvc    MetadataMatcher
	subtitleDL     SubtitleAutoDownloader
	cloudProbeFn   CloudProbeFunc
}

func New(db *sql.DB, mediaRepo *repository.MediaRepo, mediaFileRepo *repository.MediaFileRepo, libraryRepo *repository.LibraryRepo) *Scanner {
	return &Scanner{
		db:            db,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		libraryRepo:   libraryRepo,
	}
}

// SetMetadataMatcher attaches an optional metadata enrichment service.
func (s *Scanner) SetMetadataMatcher(m MetadataMatcher) {
	s.metadataSvc = m
}

// SetSubtitleAutoDownloader attaches an optional subtitle auto-download service.
func (s *Scanner) SetSubtitleAutoDownloader(dl SubtitleAutoDownloader) {
	s.subtitleDL = dl
}

// SetCloudProbe attaches an optional cloud probe function for extracting metadata from cloud files during scan.
func (s *Scanner) SetCloudProbe(fn CloudProbeFunc, subtitleRepo *repository.SubtitleRepo, audioTrackRepo *repository.AudioTrackRepo) {
	s.cloudProbeFn = fn
	s.subtitleRepo = subtitleRepo
	s.audioTrackRepo = audioTrackRepo
}

// ScanLibrary walks the library path and indexes all video files.
func (s *Scanner) ScanLibrary(ctx context.Context, libraryID int64) error {
	lib, err := s.libraryRepo.GetByID(ctx, libraryID)
	if err != nil {
		return err
	}

	if len(lib.Paths) == 0 {
		return fmt.Errorf("library %d has no configured paths", libraryID)
	}
	return filepath.WalkDir(lib.Paths[0], func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		if d.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !VideoExtensions[ext] {
			return nil
		}

		// Check if file already exists
		_, err = s.mediaFileRepo.FindByPath(ctx, path)
		if err == nil {
			return nil // already exists
		}
		if !errors.Is(err, repository.ErrNotFound) {
			return fmt.Errorf("checking existing file %s: %w", path, err)
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		probe, err := ffprobe.Probe(path)
		if err != nil {
			log.Printf("ffprobe failed for %s: %v", path, err)
			return nil // skip files that can't be probed
		}

		title := strings.TrimSuffix(filepath.Base(path), ext)

		if err := s.persistMedia(ctx, lib.ID, title, path, info.Size(), probe); err != nil {
			log.Printf("failed to index %s: %v", path, err)
		} else {
			log.Printf("indexed: %s", title)
		}

		return nil
	})
}

// ScanCloudLibrary indexes a cloud-backed library using the given Provider.
// The caller is responsible for constructing the Provider (driver + credentials);
// rootFolderID is typically obtained from Provider.ParseFolderURL(library.SourceURL).
//
// Unlike ScanLibrary, no ffprobe runs: ExoPlayer probes runtime on the client.
// media_files rows get path="{provider_type}://{native_id}" so the stream
// handler can route playback back to the correct driver.
func (s *Scanner) ScanCloudLibrary(ctx context.Context, libraryID int64, provider cloudstorage.Provider, rootFolderID string, force bool) (int, int, int, error) {
	lib, err := s.libraryRepo.GetByID(ctx, libraryID)
	if err != nil {
		return 0, 0, 0, err
	}
	if !lib.IsCloudLibrary() {
		return 0, 0, 0, fmt.Errorf("library %d is not cloud-backed", libraryID)
	}

	var totalFiles, newFiles, errCount int

	walker := NewCloudWalker(provider,
		func(ctx context.Context, item cloudstorage.Item, storagePath string, relativePath string) error {
			totalFiles++

			// Dedupe — file already indexed.
			var existing *model.MediaFile
			var checkErr error
			if provType, nativeID, ok := ParseCloudPath(storagePath); ok {
				existing, checkErr = s.mediaFileRepo.FindCloudFile(ctx, provType, nativeID)
			} else {
				existing, checkErr = s.mediaFileRepo.FindByPath(ctx, storagePath)
			}

			if checkErr == nil && existing != nil {
				if force {
					parsed := nameparser.ParseWithParents(relativePath)
					displayTitle := parsed.Title
					if parsed.MediaType == "episode" && parsed.EpisodeTitle != "" {
						displayTitle = parsed.Title + " - " + parsed.EpisodeTitle
					}

					// Re-parse filename and update database title
					if err := s.mediaRepo.UpdateTitle(ctx, existing.MediaID, displayTitle); err != nil {
						log.Printf("cloud force: update title error: %v", err)
					}

					// Re-trigger metadata hook
					if s.metadataSvc != nil {
						fullMedia, _ := s.mediaRepo.GetByID(ctx, existing.MediaID)
						if fullMedia != nil {
							var metaErr error
							if lib.Type == model.LibraryTypeAnime {
								metaErr = s.metadataSvc.MatchAndPersistAnime(ctx, fullMedia, parsed, storagePath, lib.ID, true)
							} else {
								switch parsed.MediaType {
								case "episode":
									metaErr = s.metadataSvc.MatchAndPersistEpisode(ctx, fullMedia, parsed, storagePath, lib.ID, true)
								default:
									metaErr = s.metadataSvc.MatchAndPersistMovie(ctx, fullMedia, parsed, storagePath, true)
								}
							}
							if metaErr != nil {
								log.Printf("cloud force meta error %s: %v", storagePath, metaErr)
							} else {
								log.Printf("cloud force meta updated: %s", displayTitle)
							}
						}
					}

					// Auto-download subtitles on force refresh
					if s.subtitleDL != nil {
						if dlErr := s.subtitleDL.AutoDownload(ctx, existing.MediaID, existing.ID); dlErr != nil {
							log.Printf("cloud check auto-sub error for %s: %v", storagePath, dlErr)
						}
					}
				}
				return nil
			} else if !errors.Is(err, repository.ErrNotFound) {
				errCount++
				return fmt.Errorf("check existing %s: %w", storagePath, err)
			}

			parsed := nameparser.ParseWithParents(relativePath)
			displayTitle := parsed.Title
			if parsed.MediaType == "episode" && parsed.EpisodeTitle != "" {
				displayTitle = parsed.Title + " - " + parsed.EpisodeTitle
			}

			if err := s.persistCloudMedia(ctx, lib.ID, displayTitle, storagePath, item, parsed, provider); err != nil {
				log.Printf("cloud index failed for %s: %v", storagePath, err)
				errCount++
				return nil // non-fatal: skip file, continue walk
			}
			newFiles++
			log.Printf("cloud indexed: %s (%s)", displayTitle, storagePath)
			return nil
		},
		func(folderID string, walkErr error) {
			errCount++
			log.Printf("cloud walk: folder %s: %v", folderID, walkErr)
		},
	)

	err = walker.Walk(ctx, rootFolderID)
	return totalFiles, newFiles, errCount, err
}

// persistCloudMedia creates a media + media_file row pair for a cloud file.
// Runtime-probeable fields (duration, codec, HDR) stay empty/zero; ExoPlayer
// detects them on stream open. Safe because migration 035 keeps IsHDR runtime
// for cloud paths (no ffprobe result ever reaches the DB here).
func (s *Scanner) persistCloudMedia(ctx context.Context, libraryID int64, title, storagePath string, item cloudstorage.Item, parsed nameparser.ParsedMedia, provider cloudstorage.Provider) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	mediaRepo := s.mediaRepo.WithTx(tx)
	mediaFileRepo := s.mediaFileRepo.WithTx(tx)

	media := &model.Media{
		LibraryID: libraryID,
		MediaType: parsed.MediaType,
		Title:     title,
		SortTitle: title,
	}
	if err := mediaRepo.Create(ctx, media); err != nil {
		return fmt.Errorf("creating media: %w", err)
	}

	mf := &model.MediaFile{
		MediaID:   media.ID,
		FilePath:  storagePath, // "fshare://linkcode"
		FileSize:  item.Size,
		IsPrimary: true,
	}
	if err := mediaFileRepo.Create(ctx, mf); err != nil {
		return fmt.Errorf("creating media file: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing tx: %w", err)
	}

	// Metadata enrichment (non-critical, outside transaction)
	if s.metadataSvc != nil {
		fullMedia, err := s.mediaRepo.GetByID(ctx, media.ID)
		if err != nil {
			log.Printf("Failed to get media %d for metadata: %v", media.ID, err)
		} else {
			var metaErr error
			if lib, libErr := s.libraryRepo.GetByID(ctx, libraryID); libErr == nil && lib.Type == model.LibraryTypeAnime {
				metaErr = s.metadataSvc.MatchAndPersistAnime(ctx, fullMedia, parsed, storagePath, libraryID, false)
			} else {
				switch parsed.MediaType {
				case "episode":
					metaErr = s.metadataSvc.MatchAndPersistEpisode(ctx, fullMedia, parsed, storagePath, libraryID, false)
				default:
					metaErr = s.metadataSvc.MatchAndPersistMovie(ctx, fullMedia, parsed, storagePath, false)
				}
			}
			if metaErr != nil {
				log.Printf("Metadata match for %s: %v", storagePath, metaErr)
			}
		}
	}

	// Auto-download subtitles (non-critical)
	if s.subtitleDL != nil {
		if dlErr := s.subtitleDL.AutoDownload(ctx, media.ID, mf.ID); dlErr != nil {
			log.Printf("cloud auto-sub error for %s: %v", storagePath, dlErr)
		}
	}

	// Cloud probe: resolve → ffprobe → persist embedded subtitles + audio tracks + media file metadata.
	// Runs sequentially per file to avoid FShare rate limiting.
	if s.cloudProbeFn != nil {
		_, nativeID, _ := ParseCloudPath(storagePath)
		probeCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
		probe, probeErr := s.cloudProbeFn(probeCtx, provider, nativeID)
		cancel()
		if probeErr != nil {
			log.Printf("cloud probe skipped for %s: %v", storagePath, probeErr)
		} else {
			// Update media file with probed metadata
			mf.VideoCodec = probe.VideoCodec
			mf.VideoProfile = probe.VideoProfile
			mf.VideoLevel = probe.VideoLevel
			mf.VideoFPS = probe.VideoFPS
			mf.AudioCodec = probe.AudioCodec
			mf.Container = probe.Container
			if probe.Width > 0 && probe.Height > 0 {
				mf.Width = probe.Width
				mf.Height = probe.Height
			}
			if probe.Duration > 0 {
				mf.Duration = probe.Duration
			}
			mf.Bitrate = int(probe.Bitrate)
			mf.IsHDR = probe.IsHDR
			mf.DVProfile = probe.DVProfile
			mf.ColorTransfer = probe.ColorTransfer
			mf.ColorPrimaries = probe.ColorPrimaries
			tx, err := s.db.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("cloud probe: failed to begin tx: %v", err)
				return nil
			}
			defer tx.Rollback()

			mfRepo := s.mediaFileRepo.WithTx(tx)
			if err := mfRepo.Update(ctx, mf); err != nil {
				log.Printf("cloud probe: update media file %d: %v", mf.ID, err)
				return nil
			}

			// Persist audio tracks
			if s.audioTrackRepo != nil {
				atRepo := s.audioTrackRepo.WithTx(tx)
				for _, t := range probe.AudioTracks {
					track := &model.AudioTrack{
						MediaFileID:   mf.ID,
						StreamIndex:   t.StreamIndex,
						Language:      t.Language,
						Codec:         t.Codec,
						Channels:      t.Channels,
						ChannelLayout: t.ChannelLayout,
						Bitrate:       int(t.Bitrate),
						SampleRate:    int(t.SampleRate),
						Title:         t.Title,
						IsDefault:     t.IsDefault,
					}
					if err := atRepo.Create(ctx, track); err != nil {
						log.Printf("cloud probe: audio track %d: %v", t.StreamIndex, err)
						return nil
					}
				}
			}

			// Persist embedded subtitles
			if s.subtitleRepo != nil {
				subRepo := s.subtitleRepo.WithTx(tx)
				for _, sub := range probe.Subtitles {
					subtitle := &model.Subtitle{
						MediaFileID: mf.ID,
						Language:    sub.Language,
						Codec:       sub.Codec,
						Title:       sub.Title,
						IsEmbedded:  true,
						StreamIndex: sub.StreamIndex,
						IsForced:    sub.IsForced,
						IsDefault:   sub.IsDefault,
						IsSDH:       sub.IsSDH,
					}
					if err := subRepo.Create(ctx, subtitle); err != nil {
						log.Printf("cloud probe: subtitle %d: %v", sub.StreamIndex, err)
						return nil
					}
				}
			}

			if err := tx.Commit(); err != nil {
				log.Printf("cloud probe: tx commit failed: %v", err)
			}

			log.Printf("cloud probed: %s (%dx%d, %d subs, %d audio)", storagePath, mf.Width, mf.Height, len(probe.Subtitles), len(probe.AudioTracks))
		}
	}

	return nil
}

// persistMedia creates media + media_file inside a transaction to prevent orphan rows.
func (s *Scanner) persistMedia(ctx context.Context, libraryID int64, title, path string, fileSize int64, probe *ffprobe.ProbeResult) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	mediaRepo := s.mediaRepo.WithTx(tx)
	mediaFileRepo := s.mediaFileRepo.WithTx(tx)

	media := &model.Media{
		LibraryID: libraryID,
		MediaType: "movie", // TODO: detect episode from filename
		Title:     title,
		SortTitle: title,
	}
	if err := mediaRepo.Create(ctx, media); err != nil {
		return fmt.Errorf("creating media: %w", err)
	}

	mf := &model.MediaFile{
		MediaID:      media.ID,
		FilePath:     path,
		FileSize:     fileSize,
		Duration:     probe.Duration,
		Width:        probe.Width,
		Height:       probe.Height,
		VideoCodec:   probe.VideoCodec,
		VideoProfile: probe.VideoProfile,
		VideoLevel:   probe.VideoLevel,
		VideoFPS:     probe.VideoFPS,
		AudioCodec:   probe.AudioCodec,
		Container:    probe.Container,
		Bitrate:      int(probe.Bitrate),
		IsPrimary:    true,
		Fingerprint:  "", // TODO: implement fingerprint in Phase 03
	}
	if err := mediaFileRepo.Create(ctx, mf); err != nil {
		return fmt.Errorf("creating media file: %w", err)
	}

	return tx.Commit()
}
