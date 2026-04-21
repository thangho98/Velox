package transcoder

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/logger"
	"github.com/thawng/velox/pkg/ffmpegbin"
)

// transcodeJob tracks a background FFmpeg transcode.
// Multiple HTTP requests waiting for the same transcode all share one job.
type transcodeJob struct {
	done            chan struct{}      // closed (not sent) when FFmpeg exits
	err             error              // set before done is closed
	cancel          context.CancelFunc // cancels the FFmpeg context
	mediaID         int64              // media this job belongs to
	streamSessionID string             // viewer-scoped playback session
	lastActivity    time.Time          // last time a segment was requested
}

// staleSessionTimeout is how long a transcode session can go without any
// segment requests before it is considered abandoned and killed.
const staleSessionTimeout = 5 * time.Minute

// Transcoder manages FFmpeg-based HLS transcoding and remuxing.
type Transcoder struct {
	outputDir       string
	hwAccel         string        // resolved HW accel type ("videotoolbox", "nvenc", "vaapi", "qsv", "amf", or "")
	enableHwTonemap bool          // feature flag for HW HDR tonemapping
	semaphore       chan struct{} // limits concurrent FFmpeg transcode jobs
	mu              sync.Mutex
	active          map[string]*transcodeJob // masterPath → in-progress job
	stopCh          chan struct{}            // closed when the transcoder is closed
	log             *slog.Logger
}

// New creates a Transcoder.
// hwAccel: resolved hardware accelerator (never "auto"; use playback.DetectHWAccel first).
// maxConcurrent: max simultaneous FFmpeg transcode jobs (>= 1).
func New(outputDir string, hwAccel string, maxConcurrent int, enableHwTonemap bool) *Transcoder {
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
	t := &Transcoder{
		outputDir:       outputDir,
		hwAccel:         hwAccel,
		enableHwTonemap: enableHwTonemap,
		semaphore:       sem,
		active:          make(map[string]*transcodeJob),
		stopCh:          make(chan struct{}),
		log:             logger.New("transcoder"),
	}
	go t.cleanupStaleJobs()
	return t
}

// Close stops the cleanup goroutine. It does not cancel active transcodes.
func (t *Transcoder) Close() {
	close(t.stopCh)
}

// TouchJob updates the last activity time for an active transcode job.
// Called by the segment server when serving segments to detect abandoned sessions.
func (t *Transcoder) TouchJob(masterPath string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if job, ok := t.active[masterPath]; ok {
		job.lastActivity = time.Now()
	}
}

// TouchJobByStreamSessionID updates the last activity time for all active
// transcode jobs for a given viewer-scoped stream session.
func (t *Transcoder) TouchJobByStreamSessionID(streamSessionID string) {
	if streamSessionID == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for _, job := range t.active {
		if job.streamSessionID == streamSessionID {
			job.lastActivity = now
		}
	}
}

// TouchJobByMediaID updates the last activity time for all active transcode jobs
// for a given media ID. Called by the segment server when serving segments.
func (t *Transcoder) TouchJobByMediaID(mediaID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for _, job := range t.active {
		if job.mediaID == mediaID {
			job.lastActivity = now
		}
	}
}

// cleanupStaleJobs periodically cancels transcode sessions that have not served
// any segment within staleSessionTimeout. This handles the case where a client
// disconnects without calling DELETE /stream/{id}/session.
func (t *Transcoder) cleanupStaleJobs() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.cancelStaleJobs()
		}
	}
}

// cancelStaleJobs kills transcode jobs that have not had any activity since staleSessionTimeout.
func (t *Transcoder) cancelStaleJobs() {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	for path, job := range t.active {
		if now.Sub(job.lastActivity) > staleSessionTimeout {
			log.Printf(
				"transcoder: killing stale transcode for media %d session %s (%s), last activity %v ago",
				job.mediaID,
				job.streamSessionID,
				path,
				now.Sub(job.lastActivity).Round(time.Second),
			)
			job.cancel()
		}
	}
}

// CancelTranscode kills all active FFmpeg processes for a given media ID.
// Called when a user stops playback or switches media/quality.
func (t *Transcoder) CancelTranscode(mediaID int64) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	killed := 0
	for path, job := range t.active {
		if job.mediaID == mediaID && job.cancel != nil {
			log.Printf("transcoder: cancelling transcode for media %d (%s)", mediaID, path)
			job.cancel()
			killed++
		}
	}
	return killed
}

// CancelTranscodeByStreamSessionID kills all active FFmpeg processes for a
// given viewer-scoped stream session.
func (t *Transcoder) CancelTranscodeByStreamSessionID(streamSessionID string) int {
	if streamSessionID == "" {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	killed := 0
	for path, job := range t.active {
		if job.streamSessionID == streamSessionID && job.cancel != nil {
			log.Printf("transcoder: cancelling stream session %s (%s)", streamSessionID, path)
			job.cancel()
			killed++
		}
	}
	return killed
}

// CancelTranscodeByStreamSessionIDExcept kills all active FFmpeg processes for a
// stream session except the currently requested master playlist.
func (t *Transcoder) CancelTranscodeByStreamSessionIDExcept(streamSessionID, keepMasterPath string) int {
	if streamSessionID == "" {
		return 0
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	killed := 0
	for path, job := range t.active {
		if job.streamSessionID != streamSessionID || path == keepMasterPath || job.cancel == nil {
			continue
		}
		log.Printf("transcoder: cancelling superseded stream session %s (%s)", streamSessionID, path)
		job.cancel()
		killed++
	}
	return killed
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
func hlsPrefix(streamSessionID string, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) string {
	var prefix string
	if streamSessionID != "" {
		prefix += fmt.Sprintf("ss%s_", streamSessionID)
	}
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
func (t *Transcoder) MasterPlaylistPath(mediaID int64, streamSessionID string, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) string {
	return filepath.Join(
		t.HLSDir(mediaID),
		hlsPrefix(streamSessionID, fileID, subtitleStreamIndex, videoCopy, startOffset)+"master.m3u8",
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
	t.log.Debug("waiting for first segment",
		"segment", filepath.Base(segPath),
		"media_id", job.mediaID,
		"session", job.streamSessionID,
	)
	start := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(3 * time.Minute)

	for {
		select {
		case <-job.done:
			// FFmpeg exited — either finished OK (all segments written) or failed.
			if job.err != nil {
				t.log.Error("FFmpeg exited before first segment",
					"media_id", job.mediaID,
					"session", job.streamSessionID,
					"waited", time.Since(start),
					"error", job.err,
				)
			} else {
				t.log.Debug("FFmpeg completed (all segments written)",
					"media_id", job.mediaID,
					"session", job.streamSessionID,
					"waited", time.Since(start),
				)
			}
			return job.err
		case <-ticker.C:
			if _, err := os.Stat(segPath); err == nil {
				t.log.Debug("first segment ready",
					"segment", filepath.Base(segPath),
					"media_id", job.mediaID,
					"session", job.streamSessionID,
					"waited", time.Since(start),
				)
				return nil // first segment ready; FFmpeg continues in background
			}
		case <-timeout:
			t.log.Error("first segment timeout after 3min",
				"segment", filepath.Base(segPath),
				"media_id", job.mediaID,
				"session", job.streamSessionID,
			)
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
func (t *Transcoder) WaitForSegment(streamSessionID, path string, timeout time.Duration) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	segDir := filepath.Dir(path)
	segName := filepath.Base(path)
	sessionScoped := streamSessionID != ""
	if sessionScoped {
		if !t.hasActiveJobForStreamSession(streamSessionID) {
			t.log.Debug("WaitForSegment — no active job for session",
				"segment", segName,
				"session", streamSessionID,
			)
			return false
		}
	} else if !t.hasActiveJobInDir(segDir) {
		t.log.Debug("WaitForSegment — no active job in dir",
			"segment", segName,
		)
		return false
	}

	start := time.Now()
	t.log.Debug("WaitForSegment — polling",
		"segment", segName,
		"session", streamSessionID,
		"timeout", timeout,
	)
	ticker := time.NewTicker(300 * time.Millisecond)
	defer ticker.Stop()
	deadline := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if _, err := os.Stat(path); err == nil {
				t.log.Debug("WaitForSegment — found",
					"segment", segName,
					"session", streamSessionID,
					"waited", time.Since(start),
				)
				return true
			}
			active := t.hasActiveJobInDir(segDir)
			if sessionScoped {
				active = t.hasActiveJobForStreamSession(streamSessionID)
			}
			if !active {
				exists := false
				if _, err := os.Stat(path); err == nil {
					exists = true
				}
				t.log.Debug("WaitForSegment — job finished",
					"segment", segName,
					"session", streamSessionID,
					"found", exists,
					"waited", time.Since(start),
				)
				return exists
			}
		case <-deadline:
			t.log.Warn("WaitForSegment — timeout",
				"segment", segName,
				"session", streamSessionID,
				"waited", time.Since(start),
			)
			return false
		}
	}
}

func (t *Transcoder) hasActiveJobForStreamSession(streamSessionID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, job := range t.active {
		if job.streamSessionID == streamSessionID {
			return true
		}
	}
	return false
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

// RemuxSelectedAudioToWriter remuxes a video file to fragmented MP4 and writes to w.
// If audioStreamIndex >= 0, it keeps the first video stream and the selected audio stream.
// Otherwise it copies all default stream mappings.
func (t *Transcoder) RemuxSelectedAudioToWriter(inputPath string, audioStreamIndex int, w io.Writer) error {
	args := []string{"-i", inputPath}
	if audioStreamIndex >= 0 {
		args = append(args,
			"-map", "0:v:0",
			"-map", fmt.Sprintf("0:%d", audioStreamIndex),
			"-c:v", "copy",
			"-c:a", "copy",
		)
	} else {
		args = append(args, "-c", "copy")
	}
	args = append(args,
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof+delay_moov",
		"pipe:1",
	)

	cmd := exec.Command(ffmpegbin.FFmpeg(), args...)
	cmd.Stdout = w
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RemuxToWriter remuxes a video file to fragmented MP4 and writes to w.
// Used for DirectStream: container-only operation, no codec transcoding.
func (t *Transcoder) RemuxToWriter(inputPath string, w io.Writer) error {
	return t.RemuxSelectedAudioToWriter(inputPath, -1, w)
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
