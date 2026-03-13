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
	"github.com/thawng/velox/pkg/nameparser"
)

// MetadataMatcher is an optional interface for metadata enrichment (TMDb).
// When non-nil, the pipeline calls it after persisting new media.
type MetadataMatcher interface {
	MatchAndPersistMovie(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, force bool) error
	MatchAndPersistEpisode(ctx context.Context, media *model.Media, parsed nameparser.ParsedMedia, filePath string, libraryID int64, force bool) error
}

// Pipeline handles the media scanning process
type Pipeline struct {
	db             *sql.DB
	libraryRepo    *repository.LibraryRepo
	mediaRepo      *repository.MediaRepo
	mediaFileRepo  *repository.MediaFileRepo
	seriesRepo     *repository.SeriesRepo
	seasonRepo     *repository.SeasonRepo
	episodeRepo    *repository.EpisodeRepo
	scanJobRepo    *repository.ScanJobRepo
	subtitleRepo   *repository.SubtitleRepo
	audioTrackRepo *repository.AudioTrackRepo
	metadataSvc    MetadataMatcher // nil = skip metadata enrichment
}

// NewPipeline creates a new scan pipeline
func NewPipeline(
	db *sql.DB,
	libraryRepo *repository.LibraryRepo,
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	seriesRepo *repository.SeriesRepo,
	seasonRepo *repository.SeasonRepo,
	episodeRepo *repository.EpisodeRepo,
	scanJobRepo *repository.ScanJobRepo,
	subtitleRepo *repository.SubtitleRepo,
	audioTrackRepo *repository.AudioTrackRepo,
) *Pipeline {
	return &Pipeline{
		db:             db,
		libraryRepo:    libraryRepo,
		mediaRepo:      mediaRepo,
		mediaFileRepo:  mediaFileRepo,
		seriesRepo:     seriesRepo,
		seasonRepo:     seasonRepo,
		episodeRepo:    episodeRepo,
		scanJobRepo:    scanJobRepo,
		subtitleRepo:   subtitleRepo,
		audioTrackRepo: audioTrackRepo,
	}
}

// SetMetadataMatcher attaches an optional metadata enrichment service.
func (p *Pipeline) SetMetadataMatcher(m MetadataMatcher) {
	p.metadataSvc = m
}

// ScanContext holds state during a scan
type ScanContext struct {
	LibraryID int64
	JobID     int64
	Force     bool // Force re-parse titles for existing files
	ctx       context.Context
}

// CreateJob creates a queued scan job for a library.
// This is synchronous and fast — safe to call from an HTTP handler.
func (p *Pipeline) CreateJob(ctx context.Context, libraryID int64) (*model.ScanJob, error) {
	job := &model.ScanJob{
		LibraryID: libraryID,
		Status:    "queued",
	}
	if err := p.scanJobRepo.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("creating scan job: %w", err)
	}
	return job, nil
}

// RunJob executes a scan for an already-created job. This is blocking
// and should be called from a goroutine for async operation.
func (p *Pipeline) RunJob(ctx context.Context, job *model.ScanJob, force bool) error {
	if err := p.scanJobRepo.Start(ctx, job.ID); err != nil {
		return fmt.Errorf("starting scan job: %w", err)
	}

	scanCtx := &ScanContext{
		LibraryID: job.LibraryID,
		JobID:     job.ID,
		Force:     force,
		ctx:       ctx,
	}

	// Discover phase
	files, err := p.discover(scanCtx)
	if err != nil {
		p.scanJobRepo.Fail(ctx, job.ID, err.Error())
		return err
	}

	job.TotalFiles = len(files)

	var newFiles, errors int
	var errorLog string

	// Process each file
	for i, path := range files {
		isNew, err := p.processFile(scanCtx, path)
		if err != nil {
			errors++
			if errorLog != "" {
				errorLog += "\n"
			}
			errorLog += fmt.Sprintf("%s: %v", path, err)
			log.Printf("Scan error for %s: %v", path, err)
		} else if isNew {
			newFiles++
		}

		// Update progress every 10 files
		if i%10 == 0 {
			p.scanJobRepo.UpdateProgress(ctx, job.ID, i+1)
		}
	}

	if err := p.scanJobRepo.Complete(ctx, job.ID, job.TotalFiles, newFiles, errors, errorLog); err != nil {
		return fmt.Errorf("completing scan job: %w", err)
	}

	return nil
}

// Run creates and executes a scan job synchronously (blocking).
// Prefer CreateJob + RunJob for async usage.
func (p *Pipeline) Run(ctx context.Context, libraryID int64) (*model.ScanJob, error) {
	job, err := p.CreateJob(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	if err := p.RunJob(ctx, job, false); err != nil {
		return job, err
	}
	return job, nil
}

// discover finds all video files in the library (across all configured paths).
func (p *Pipeline) discover(scanCtx *ScanContext) ([]string, error) {
	lib, err := p.libraryRepo.GetByID(scanCtx.ctx, scanCtx.LibraryID)
	if err != nil {
		return nil, fmt.Errorf("getting library: %w", err)
	}

	var files []string
	for _, root := range lib.Paths {
		if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // skip inaccessible paths
			}
			if d.IsDir() {
				return nil
			}
			if isVideoFile(path) {
				files = append(files, path)
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("walking %q: %w", root, err)
		}
	}

	return files, nil
}

// processFile processes a single video file through the pipeline.
// Returns (true, nil) if new content was created, (false, nil) if file was already known.
func (p *Pipeline) processFile(scanCtx *ScanContext, path string) (bool, error) {
	// Step 1: Compute fingerprint
	fp, err := ComputeFingerprint(path)
	if err != nil {
		return false, fmt.Errorf("computing fingerprint: %w", err)
	}

	// Step 2: Check if file exists by fingerprint (rename detection)
	existingFile, err := p.mediaFileRepo.FindByFingerprint(scanCtx.ctx, fp)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("checking fingerprint: %w", err)
	}
	if err == nil && existingFile != nil {
		// File was renamed/moved - update path
		if existingFile.FilePath != path {
			log.Printf("Detected rename: %s -> %s", existingFile.FilePath, path)
			if err := p.mediaFileRepo.UpdatePath(scanCtx.ctx, existingFile.ID, path); err != nil {
				return false, fmt.Errorf("updating file path: %w", err)
			}
		}
		if scanCtx.Force {
			if err := p.refreshTitle(scanCtx.ctx, existingFile.MediaID, path); err != nil {
				return false, fmt.Errorf("refreshing title: %w", err)
			}
		}
		return false, nil // Already known file
	}

	// Step 3: Check if path exists (re-scanning)
	var replaceFileID int64
	var replaceMediaID int64
	replacePrimary := true // new files default to primary
	existingByPath, err := p.mediaFileRepo.FindByPath(scanCtx.ctx, path)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, fmt.Errorf("checking path: %w", err)
	}
	if err == nil && existingByPath != nil {
		// File at this path already known
		info, _ := os.Stat(path)
		if info != nil && existingByPath.FileSize == info.Size() && existingByPath.Fingerprint == fp {
			// Same size + same fingerprint = truly unchanged
			if scanCtx.Force {
				if err := p.refreshTitle(scanCtx.ctx, existingByPath.MediaID, path); err != nil {
					return false, fmt.Errorf("refreshing title: %w", err)
				}
			}
			return false, nil
		}
		// File was replaced (different size or different fingerprint) — defer deletion to persist()
		if info != nil {
			log.Printf("File replaced at %s (size %d->%d, fp changed=%v), re-indexing",
				path, existingByPath.FileSize, info.Size(), existingByPath.Fingerprint != fp)
		}
		replaceFileID = existingByPath.ID
		replaceMediaID = existingByPath.MediaID
		replacePrimary = existingByPath.IsPrimary
	}

	// Step 4: Probe with ffprobe
	probe, err := ffprobe.Probe(path)
	if err != nil {
		return false, fmt.Errorf("probing file: %w", err)
	}

	// Step 5: Parse filename
	parsed := nameparser.ParseWithParents(path)

	// Step 6: Create media and media_file (atomically replaces old record if needed)
	if err := p.persist(scanCtx, path, fp, probe, parsed, replaceFileID, replaceMediaID, replacePrimary); err != nil {
		return false, fmt.Errorf("persisting media: %w", err)
	}

	return true, nil
}

// persist saves media data to database inside a transaction to prevent orphan rows.
// If replaceFileID > 0, the old media_file (and orphaned media) is deleted atomically
// within the same transaction, so no data is lost if the insert fails.
func (p *Pipeline) persist(scanCtx *ScanContext, path string, fingerprint string, probe *ffprobe.ProbeResult, parsed nameparser.ParsedMedia, replaceFileID, replaceMediaID int64, isPrimary bool) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	ctx := scanCtx.ctx

	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Create transaction-scoped repos
	mediaRepo := p.mediaRepo.WithTx(tx)
	mediaFileRepo := p.mediaFileRepo.WithTx(tx)
	audioTrackRepo := p.audioTrackRepo.WithTx(tx)
	subtitleRepo := p.subtitleRepo.WithTx(tx)

	// Determine the media ID for the new file
	var mediaID int64

	if replaceFileID > 0 {
		// Delete old media_file (cascades subtitles + audio_tracks via ON DELETE CASCADE)
		if err := mediaFileRepo.Delete(ctx, replaceFileID); err != nil {
			return fmt.Errorf("deleting replaced file: %w", err)
		}
		// Always reuse the existing media row to preserve user_data, episodes, genres, credits.
		// The new media_file will be attached to the same logical media item.
		mediaID = replaceMediaID
	}

	if mediaID == 0 {
		// Truly new file — create a new Media (logical item)
		// Build display title: for episodes, combine series name + episode name
		displayTitle := parsed.Title
		if parsed.MediaType == "episode" && parsed.EpisodeTitle != "" {
			displayTitle = parsed.Title + " - " + parsed.EpisodeTitle
		}
		media := &model.Media{
			LibraryID:   scanCtx.LibraryID,
			MediaType:   parsed.MediaType,
			Title:       displayTitle,
			SortTitle:   displayTitle,
			ReleaseDate: "", // Will be filled by TMDb matcher in Phase 04
		}
		if err := mediaRepo.Create(ctx, media); err != nil {
			return fmt.Errorf("creating media: %w", err)
		}
		mediaID = media.ID
	}

	// Create MediaFile (physical file)
	mediaFile := &model.MediaFile{
		MediaID:     mediaID,
		FilePath:    path,
		FileSize:    info.Size(),
		Duration:    probe.Duration,
		Width:       probe.Width,
		Height:      probe.Height,
		VideoCodec:  probe.VideoCodec,
		AudioCodec:  probe.AudioCodec,
		Container:   probe.Container,
		Bitrate:     int(probe.Bitrate),
		Fingerprint: fingerprint,
		IsPrimary:   isPrimary,
	}

	if err := mediaFileRepo.Create(ctx, mediaFile); err != nil {
		return fmt.Errorf("creating media file: %w", err)
	}

	// Save audio tracks
	for _, track := range probe.AudioTracks {
		at := &model.AudioTrack{
			MediaFileID:   mediaFile.ID,
			StreamIndex:   track.StreamIndex,
			Codec:         track.Codec,
			Language:      track.Language,
			Channels:      track.Channels,
			ChannelLayout: track.ChannelLayout,
			Bitrate:       track.Bitrate,
			Title:         track.Title,
			IsDefault:     track.IsDefault,
		}
		if err := audioTrackRepo.Create(ctx, at); err != nil {
			return fmt.Errorf("saving audio track %d: %w", track.StreamIndex, err)
		}
	}

	// Save embedded subtitles
	for _, sub := range probe.Subtitles {
		subtitle := &model.Subtitle{
			MediaFileID: mediaFile.ID,
			Language:    sub.Language,
			Codec:       sub.Codec,
			Title:       sub.Title,
			IsEmbedded:  true,
			StreamIndex: sub.StreamIndex,
			IsForced:    sub.IsForced,
			IsDefault:   sub.IsDefault,
			IsSDH:       sub.IsSDH,
		}
		if err := subtitleRepo.Create(ctx, subtitle); err != nil {
			return fmt.Errorf("saving subtitle %d: %w", sub.StreamIndex, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	// External subtitles are scanned outside the transaction —
	// they are non-critical and won't leave orphan media rows on failure.
	if err := p.scanExternalSubtitles(ctx, mediaFile.ID, path); err != nil {
		log.Printf("Failed to scan external subtitles: %v", err)
	}

	// Phase 04: TMDb metadata enrichment (non-critical, outside transaction)
	if p.metadataSvc != nil {
		media, err := p.mediaRepo.GetByID(ctx, mediaID)
		if err != nil {
			log.Printf("Failed to get media %d for metadata: %v", mediaID, err)
		} else {
			var metaErr error
			switch parsed.MediaType {
			case "episode":
				metaErr = p.metadataSvc.MatchAndPersistEpisode(ctx, media, parsed, path, scanCtx.LibraryID, false)
			default:
				metaErr = p.metadataSvc.MatchAndPersistMovie(ctx, media, parsed, path, false)
			}
			if metaErr != nil {
				log.Printf("Metadata match for %s: %v", path, metaErr)
			}
		}
	}

	return nil
}

// refreshTitle re-parses the filename and updates the media title if it changed.
func (p *Pipeline) refreshTitle(ctx context.Context, mediaID int64, path string) error {
	parsed := nameparser.ParseWithParents(path)
	displayTitle := parsed.Title
	if parsed.MediaType == "episode" && parsed.EpisodeTitle != "" {
		displayTitle = parsed.Title + " - " + parsed.EpisodeTitle
	}
	return p.mediaRepo.UpdateTitle(ctx, mediaID, displayTitle)
}

// scanExternalSubtitles looks for sidecar subtitle files
func (p *Pipeline) scanExternalSubtitles(ctx context.Context, mediaFileID int64, videoPath string) error {
	dir := filepath.Dir(videoPath)
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	subtitleExts := []string{".srt", ".vtt", ".ass", ".ssa", ".sub"}

	for _, ext := range subtitleExts {
		// Direct match: video.srt
		subPath := filepath.Join(dir, base+ext)
		if _, err := os.Stat(subPath); err == nil {
			p.addExternalSubtitle(ctx, mediaFileID, subPath)
		}

		// Language match: video.en.srt, video.vi.srt
		// This is simplified - full implementation would parse language codes
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasPrefix(name, base+".") && strings.HasSuffix(strings.ToLower(name), ext) {
				subPath := filepath.Join(dir, name)
				p.addExternalSubtitle(ctx, mediaFileID, subPath)
			}
		}
	}

	return nil
}

func (p *Pipeline) addExternalSubtitle(ctx context.Context, mediaFileID int64, subPath string) {
	// Parse language from filename
	lang := parseSubtitleLanguage(subPath)
	lowerPath := strings.ToLower(subPath)
	forced := strings.Contains(lowerPath, ".forced.")
	sdh := strings.Contains(lowerPath, ".sdh.") || strings.Contains(lowerPath, ".hi.")

	sub := &model.Subtitle{
		MediaFileID: mediaFileID,
		Language:    lang,
		IsEmbedded:  false,
		StreamIndex: -1,
		FilePath:    subPath,
		IsForced:    forced,
		IsSDH:       sdh,
	}

	if err := p.subtitleRepo.Create(ctx, sub); err != nil {
		log.Printf("Failed to save external subtitle: %v", err)
	}
}

// isVideoFile checks if a file is a video file (delegates to package-level IsVideoFile).
func isVideoFile(path string) bool {
	return IsVideoFile(path)
}

// parseSubtitleLanguage extracts language code from subtitle filename
func parseSubtitleLanguage(path string) string {
	// Common patterns: video.en.srt, video.vi.forced.srt
	base := filepath.Base(path)
	parts := strings.Split(base, ".")

	for i := 1; i < len(parts)-1; i++ {
		part := strings.ToLower(parts[i])
		// Skip common non-language parts
		if part == "forced" || part == "sdh" || part == "hi" || part == "default" {
			continue
		}
		// Check if it looks like a language code (2-3 letters)
		if len(part) >= 2 && len(part) <= 3 {
			return part
		}
	}

	return "" // Unknown
}
