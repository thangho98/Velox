package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/repository"
	"github.com/thawng/velox/internal/scanner"
	"github.com/thawng/velox/pkg/ffmpegbin"
)

type DownloadService struct {
	mediaRepo     *repository.MediaRepo
	mediaFileRepo *repository.MediaFileRepo
	streamSvc     *StreamService
	outputDir     string

	mu    sync.RWMutex
	tasks map[string]*model.DownloadTask

	queue chan *model.DownloadTask
}

func NewDownloadService(
	mediaRepo *repository.MediaRepo,
	mediaFileRepo *repository.MediaFileRepo,
	streamSvc *StreamService,
	outputDir string,
) *DownloadService {
	// Ensure directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("Failed to create download directory: %v", err)
	}

	svc := &DownloadService{
		mediaRepo:     mediaRepo,
		mediaFileRepo: mediaFileRepo,
		streamSvc:     streamSvc,
		outputDir:     outputDir,
		tasks:         make(map[string]*model.DownloadTask),
		queue:         make(chan *model.DownloadTask, 2000),
	}
	go svc.worker()
	return svc
}

func (s *DownloadService) ListTasks() []*model.DownloadTask {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*model.DownloadTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		list = append(list, t)
	}
	return list
}

func (s *DownloadService) Enqueue(ctx context.Context, mediaID int64) (*model.DownloadTask, error) {
	media, err := s.mediaRepo.GetByID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("media not found: %w", err)
	}

	mediaFile, err := s.mediaFileRepo.GetPrimaryByMediaID(ctx, mediaID)
	if err != nil {
		return nil, fmt.Errorf("no primary media file: %w", err)
	}

	// Make sure it's actually a cloud file
	if !strings.HasPrefix(mediaFile.FilePath, "ophim://") && !scanner.IsCloudPath(mediaFile.FilePath) {
		return nil, fmt.Errorf("this media is already local")
	}

	s.mu.Lock()
	for _, t := range s.tasks {
		if t.MediaID == mediaID && (t.Status == model.DownloadStatusPending || t.Status == model.DownloadStatusDownloading) {
			s.mu.Unlock()
			return t, nil
		}
	}

	taskID := fmt.Sprintf("dl-%d", time.Now().UnixNano())
	task := &model.DownloadTask{
		ID:          taskID,
		MediaID:     mediaID,
		MediaFileID: mediaFile.ID,
		Title:       media.Title,
		Status:      model.DownloadStatusPending,
		CreatedAt:   time.Now(),
	}

	s.tasks[taskID] = task
	s.mu.Unlock()

	s.queue <- task
	return task, nil
}

func (s *DownloadService) worker() {
	for task := range s.queue {
		s.processTask(task)
	}
}

func (s *DownloadService) updateTaskStatus(taskID string, progress float64, status model.DownloadStatus, errStr string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, exists := s.tasks[taskID]; exists {
		t.Progress = progress
		t.Status = status
		t.Error = errStr
	}
}

func (s *DownloadService) processTask(task *model.DownloadTask) {
	ctx := context.Background()
	s.updateTaskStatus(task.ID, 0, model.DownloadStatusDownloading, "")

	mediaFile, err := s.mediaFileRepo.GetByID(ctx, task.MediaFileID)
	if err != nil {
		s.updateTaskStatus(task.ID, 0, model.DownloadStatusFailed, err.Error())
		return
	}

	rawURL, err := s.streamSvc.ResolveFilePath(ctx, task.MediaID, mediaFile)
	if err != nil {
		s.updateTaskStatus(task.ID, 0, model.DownloadStatusFailed, "Url Resolution: "+err.Error())
		return
	}

	// Prepare output path
	cleanTitle := strings.ReplaceAll(task.Title, "/", "-")
	ext := ".mkv"
	if strings.HasPrefix(mediaFile.FilePath, "ophim://") || strings.HasSuffix(rawURL, ".m3u8") {
		ext = ".mp4"
	}
	outputFile := filepath.Join(s.outputDir, fmt.Sprintf("%s - %d%s", cleanTitle, task.MediaID, ext))

	log.Printf("Starting download for Media %d: %s -> %s", task.MediaID, rawURL, outputFile)

	if strings.HasSuffix(rawURL, ".m3u8") || strings.HasPrefix(mediaFile.FilePath, "ophim://") {
		err = s.downloadM3U8(ctx, rawURL, outputFile, task.ID)
	} else {
		err = s.downloadHTTP(ctx, rawURL, outputFile, task.ID)
	}

	if err != nil {
		s.updateTaskStatus(task.ID, task.Progress, model.DownloadStatusFailed, err.Error())
		log.Printf("Download failed %s: %v", task.ID, err)
		return
	}

	// Update DB to use the new Local file!
	if err := s.swapToLocalFile(ctx, task, outputFile); err != nil {
		s.updateTaskStatus(task.ID, task.Progress, model.DownloadStatusFailed, "DB Swap: "+err.Error())
		log.Printf("Download swap failed %s: %v", task.ID, err)
		return
	}
	s.updateTaskStatus(task.ID, 100, model.DownloadStatusCompleted, "")
}

func (s *DownloadService) swapToLocalFile(ctx context.Context, task *model.DownloadTask, newFilePath string) error {
	// Let's create a new primary file, and downgrade the old one so it's kept as fallback.
	existing, err := s.mediaFileRepo.FindByPath(ctx, newFilePath)
	if err == nil {
		if !existing.IsPrimary {
			if err := s.mediaFileRepo.SetPrimary(ctx, task.MediaID, existing.ID); err != nil {
				return fmt.Errorf("failed to upgrade existing local file to primary: %w", err)
			}
		}
		return nil
	}

	newMf := &model.MediaFile{
		MediaID:   task.MediaID,
		FilePath:  newFilePath,
		IsPrimary: false,
	}
	if err := s.mediaFileRepo.Create(ctx, newMf); err != nil {
		return fmt.Errorf("failed to create new local media file: %w", err)
	}

	if err := s.mediaFileRepo.SetPrimary(ctx, task.MediaID, newMf.ID); err != nil {
		return fmt.Errorf("failed to set new local file as primary: %w", err)
	}
	// Mark media as no longer cloud
	media, err := s.mediaRepo.GetByID(ctx, task.MediaID)
	if err == nil {
		_ = s.mediaRepo.Update(ctx, media)
	}
	return nil
}

func (s *DownloadService) downloadM3U8(ctx context.Context, m3u8URL string, destPath string, taskID string) error {
	// For m3u8, use ffmpeg
	// ffmpeg -i url -c copy -bsf:a aac_adtstoasc dest.mp4
	cmd := exec.CommandContext(ctx, ffmpegbin.FFmpeg(), "-y", "-i", m3u8URL, "-c", "copy", "-bsf:a", "aac_adtstoasc", destPath)
	// Progress parsing can be tricky, we'll just fake it or leave at 50%
	s.updateTaskStatus(taskID, 50, model.DownloadStatusDownloading, "")

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(out))
	}
	return nil
}

// downloadHTTP handles normal direct urls like Fshare
func (s *DownloadService) downloadHTTP(ctx context.Context, rawURL string, destPath string, taskID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	total := resp.ContentLength
	var downloaded int64

	buf := make([]byte, 1024*1024) // 1MB buffer
	lastUpdate := time.Now()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, wErr := f.Write(buf[:n])
			if wErr != nil {
				return wErr
			}
			downloaded += int64(n)

			// Update progress smoothly
			if time.Since(lastUpdate) > 2*time.Second && total > 0 {
				prog := float64(downloaded) / float64(total) * 100
				s.updateTaskStatus(taskID, prog, model.DownloadStatusDownloading, "")
				lastUpdate = time.Now()
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteDownloadedFile deletes local downloaded files for a given media ID
func (s *DownloadService) DeleteDownloadedFile(ctx context.Context, mediaID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Get the files associated with the mediaID
	files, err := s.mediaFileRepo.ListByMediaID(ctx, mediaID)
	if err != nil {
		return fmt.Errorf("list media files: %w", err)
	}

	deletedLocal := false
	for _, f := range files {
		// Only delete locally downloaded files inside s.outputDir
		if !scanner.IsCloudPath(f.FilePath) && strings.HasPrefix(f.FilePath, s.outputDir) {
			if err := os.Remove(f.FilePath); err != nil && !os.IsNotExist(err) {
				log.Printf("Failed to delete local media file %d: %v", f.ID, err)
			}
			if err := s.mediaFileRepo.Delete(ctx, f.ID); err != nil {
				log.Printf("Failed to delete media_files record %d: %v", f.ID, err)
			} else {
				deletedLocal = true
			}
		}
	}

	if !deletedLocal {
		return fmt.Errorf("no local downloaded file found to delete")
	}

	return nil
}
