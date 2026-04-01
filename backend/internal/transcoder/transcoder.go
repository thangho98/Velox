package transcoder

import (
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// transcodeJob tracks a background FFmpeg transcode.
// Multiple HTTP requests waiting for the same transcode all share one job.
type transcodeJob struct {
	done chan struct{} // closed (not sent) when FFmpeg exits
	err  error         // set before done is closed
}

// Transcoder manages FFmpeg-based HLS transcoding and remuxing.
type Transcoder struct {
	outputDir string
	hwAccel   string        // resolved HW accel type ("videotoolbox", "nvenc", "vaapi", "qsv", or "")
	semaphore chan struct{} // limits concurrent FFmpeg transcode jobs
	mu        sync.Mutex
	active    map[string]*transcodeJob // masterPath → in-progress job
}

// New creates a Transcoder.
// hwAccel: resolved hardware accelerator (never "auto"; use playback.DetectHWAccel first).
// maxConcurrent: max simultaneous FFmpeg transcode jobs (>= 1).
func New(outputDir string, hwAccel string, maxConcurrent int) *Transcoder {
	if maxConcurrent <= 0 {
		maxConcurrent = 2
	}
	sem := make(chan struct{}, maxConcurrent)
	for i := 0; i < maxConcurrent; i++ {
		sem <- struct{}{}
	}
	if !hasSubtitlesFilter {
		log.Println("WARN: FFmpeg missing 'subtitles' filter (libass not linked) — subtitle burn-in disabled, using client-side rendering")
	}
	return &Transcoder{
		outputDir: outputDir,
		hwAccel:   hwAccel,
		semaphore: sem,
		active:    make(map[string]*transcodeJob),
	}
}

// ActiveCount returns the number of in-progress realtime transcodes.
func (t *Transcoder) ActiveCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.active)
}

// TryActiveCount returns the number of in-progress realtime transcodes,
// or 0 if the mutex is currently held (non-blocking).
func (t *Transcoder) TryActiveCount() int {
	if !t.mu.TryLock() {
		return 0
	}
	defer t.mu.Unlock()
	return len(t.active)
}

// acquireSlot blocks until a transcode slot is available.
// The returned function must be deferred to release the slot.
func (t *Transcoder) acquireSlot() func() {
	<-t.semaphore
	return func() { t.semaphore <- struct{}{} }
}

// tryAcquireSlotIfSpare acquires a slot only if at least 2 slots are free,
// ensuring at least 1 slot remains available for realtime playback requests.
// Used by background tasks (ABR) that should never starve active users.
func (t *Transcoder) tryAcquireSlotIfSpare() func() {
	if len(t.semaphore) < 2 {
		return nil
	}
	select {
	case <-t.semaphore:
		return func() { t.semaphore <- struct{}{} }
	default:
		return nil
	}
}

// tryAcquireSlot tries to acquire a transcode slot within the given timeout.
// Returns a release function on success, or nil if the timeout expired.
// A timeout of 0 means non-blocking: return immediately if no slot is free.
func (t *Transcoder) tryAcquireSlot(timeout time.Duration) func() {
	if timeout <= 0 {
		select {
		case <-t.semaphore:
			return func() { t.semaphore <- struct{}{} }
		default:
			return nil
		}
	}
	select {
	case <-t.semaphore:
		return func() { t.semaphore <- struct{}{} }
	case <-time.After(timeout):
		return nil
	}
}

// HLSDir returns the directory where HLS segments for a media item are stored.
func (t *Transcoder) HLSDir(mediaID int64) string {
	return filepath.Join(t.outputDir, fmt.Sprintf("%d", mediaID))
}

func normalizeStartOffset(startOffset float64) float64 {
	if startOffset <= 0.25 {
		return 0
	}
	return math.Round(startOffset*1000) / 1000
}

func startOffsetMillis(startOffset float64) int64 {
	normalized := normalizeStartOffset(startOffset)
	if normalized <= 0 {
		return 0
	}
	return int64(math.Round(normalized * 1000))
}

// hlsPrefix returns the filename prefix used for HLS output files.
// Encodes file version, subtitle burn-in index, seek offset, and video copy
// mode so each unique combination gets its own cached playlist.
// Example: fileID=5, siIdx=2 → "f5_sub2_"
// Example: fileID=5, startOffset=3480 → "f5_off3480000_"
// Example: fileID=5, videoCopy=true → "vcf5_"
func hlsPrefix(fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) string {
	var prefix string
	if videoCopy {
		prefix += "vc"
	}
	if fileID > 0 {
		prefix += fmt.Sprintf("f%d_", fileID)
	}
	if subtitleStreamIndex >= 0 {
		prefix += fmt.Sprintf("sub%d_", subtitleStreamIndex)
	}
	if offsetMillis := startOffsetMillis(startOffset); offsetMillis > 0 {
		prefix += fmt.Sprintf("off%d_", offsetMillis)
	}
	return prefix
}

// MasterPlaylistPath returns the expected path to the master playlist for the
// given (mediaID, fileID, subtitleStreamIndex, videoCopy, startOffset)
// combination.
// Used by StreamService to retrieve the correct playlist path after transcoding.
func (t *Transcoder) MasterPlaylistPath(mediaID, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) string {
	return filepath.Join(
		t.HLSDir(mediaID),
		hlsPrefix(fileID, subtitleStreamIndex, videoCopy, startOffset)+"master.m3u8",
	)
}

// ABRMasterPath returns the path to the adaptive bitrate master playlist.
func (t *Transcoder) ABRMasterPath(mediaID, fileID int64) string {
	return filepath.Join(t.HLSDir(mediaID), fmt.Sprintf("f%d_abr_master.m3u8", fileID))
}

// ABRCached reports whether the ABR master playlist already exists on disk.
func (t *Transcoder) ABRCached(mediaID, fileID int64) bool {
	_, err := os.Stat(t.ABRMasterPath(mediaID, fileID))
	return err == nil
}

// isHLSComplete reports whether a media playlist has been fully written
// (i.e. FFmpeg added #EXT-X-ENDLIST at the end).
func isHLSComplete(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("transcoder: isHLSComplete(%s): read error: %v", path, err)
		return false
	}
	complete := strings.Contains(string(data), "#EXT-X-ENDLIST")
	if !complete {
		log.Printf("transcoder: isHLSComplete(%s): playlist exists (%d bytes) but no ENDLIST — will re-transcode", path, len(data))
	} else {
		log.Printf("transcoder: isHLSComplete(%s): cache hit (%d bytes)", path, len(data))
	}
	return complete
}

// waitForFirstSegment polls until segPath exists OR the job finishes.
// Returns nil as soon as the first segment appears (FFmpeg still running in background).
// Returns job.err if FFmpeg exits before the segment appears.
func (t *Transcoder) waitForFirstSegment(job *transcodeJob, segPath string) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(3 * time.Minute)

	for {
		select {
		case <-job.done:
			// FFmpeg exited — either finished OK (all segments written) or failed.
			return job.err
		case <-ticker.C:
			if _, err := os.Stat(segPath); err == nil {
				return nil // first segment ready; FFmpeg continues in background
			}
		case <-timeout:
			return fmt.Errorf("transcode start timeout: first segment not ready after 3 minutes")
		}
	}
}

// SegmentPath returns the full path to an HLS segment file.
func (t *Transcoder) SegmentPath(mediaID int64, segment string) string {
	return filepath.Join(t.HLSDir(mediaID), segment)
}

// WaitForSegment waits up to timeout for a segment file to appear on disk.
// Returns true if the path exists, false if timed out or all relevant jobs
// in the same media directory have finished without producing it.
func (t *Transcoder) WaitForSegment(path string, timeout time.Duration) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	segDir := filepath.Dir(path)
	if !t.hasActiveJobInDir(segDir) {
		return false
	}

	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				return true
			}
			if !t.hasActiveJobInDir(segDir) {
				_, err := os.Stat(path)
				return err == nil
			}
		case <-deadline:
			return false
		}
	}
}

func (t *Transcoder) hasActiveJobInDir(dir string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for masterPath := range t.active {
		if filepath.Dir(masterPath) == dir {
			return true
		}
	}
	return false
}

// RemuxToWriter remuxes a video file to fragmented MP4 and writes to w.
// Used for DirectStream: container-only operation, no codec transcoding.
func (t *Transcoder) RemuxToWriter(inputPath string, w io.Writer) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c", "copy",
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov",
		"pipe:1",
	)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean removes transcoded files for a media item.
func (t *Transcoder) Clean(mediaID int64) error {
	return os.RemoveAll(t.HLSDir(mediaID))
}

// CleanupOlderThan removes transcode directories that haven't been modified
// within the given age duration.
func (t *Transcoder) CleanupOlderThan(age time.Duration) error {
	entries, err := os.ReadDir(t.outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading transcode dir: %w", err)
	}

	cutoff := time.Now().Add(-age)
	var removed int

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			dirPath := filepath.Join(t.outputDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("transcode cleanup: failed to remove %s: %v", dirPath, err)
			} else {
				removed++
			}
		}
	}

	if removed > 0 {
		log.Printf("transcode cleanup: removed %d stale directories", removed)
	}
	return nil
}
