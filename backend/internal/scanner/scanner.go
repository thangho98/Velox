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

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/ffprobe"
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

type Scanner struct {
	db            *sql.DB
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	libraryRepo   *repository.LibraryRepo
}

func New(db *sql.DB, mediaRepo *repository.MediaRepo, mediaFileRepo *repository.MediaFileRepo, libraryRepo *repository.LibraryRepo) *Scanner {
	return &Scanner{
		db:            db,
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		libraryRepo:   libraryRepo,
	}
}

// ScanLibrary walks the library path and indexes all video files.
func (s *Scanner) ScanLibrary(ctx context.Context, libraryID int64) error {
	lib, err := s.libraryRepo.GetByID(ctx, libraryID)
	if err != nil {
		return err
	}

	return filepath.WalkDir(lib.Path, func(path string, d os.DirEntry, err error) error {
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
		if !errors.Is(err, sql.ErrNoRows) {
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
		MediaID:     media.ID,
		FilePath:    path,
		FileSize:    fileSize,
		Duration:    probe.Duration,
		Width:       probe.Width,
		Height:      probe.Height,
		VideoCodec:  probe.VideoCodec,
		AudioCodec:  probe.AudioCodec,
		Container:   probe.Container,
		Bitrate:     int(probe.Bitrate),
		IsPrimary:   true,
		Fingerprint: "", // TODO: implement fingerprint in Phase 03
	}
	if err := mediaFileRepo.Create(ctx, mf); err != nil {
		return fmt.Errorf("creating media file: %w", err)
	}

	return tx.Commit()
}
