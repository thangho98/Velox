package scanner

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/pkg/ffprobe"
)

var videoExtensions = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
	".wmv": true, ".flv": true, ".webm": true, ".m4v": true,
	".ts": true, ".mpg": true, ".mpeg": true,
}

type Scanner struct {
	mediaRepo   *repository.MediaRepo
	libraryRepo *repository.LibraryRepo
}

func New(mediaRepo *repository.MediaRepo, libraryRepo *repository.LibraryRepo) *Scanner {
	return &Scanner{mediaRepo: mediaRepo, libraryRepo: libraryRepo}
}

// ScanLibrary walks the library path and indexes all video files.
func (s *Scanner) ScanLibrary(libraryID int64) error {
	lib, err := s.libraryRepo.GetByID(libraryID)
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
		if !videoExtensions[ext] {
			return nil
		}

		exists, err := s.mediaRepo.ExistsByPath(path)
		if err != nil {
			return err
		}
		if exists {
			return nil
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

		m := &model.Media{
			LibraryID:   libraryID,
			Title:       title,
			FilePath:    path,
			Duration:    probe.Duration,
			Size:        info.Size(),
			Width:       probe.Width,
			Height:      probe.Height,
			VideoCodec:  probe.VideoCodec,
			AudioCodec:  probe.AudioCodec,
			Container:   probe.Container,
			Bitrate:     probe.Bitrate,
			HasSubtitle: probe.HasSub,
		}

		if err := s.mediaRepo.Upsert(m); err != nil {
			log.Printf("failed to save media %s: %v", path, err)
		} else {
			log.Printf("indexed: %s", title)
		}

		return nil
	})
}
