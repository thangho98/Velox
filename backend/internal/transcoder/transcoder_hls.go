package transcoder

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/thawng/velox/internal/model"
)

// AudioVariant represents an audio track variant for HLS.
type AudioVariant struct {
	Language    string
	Name        string
	StreamIndex int
	IsDefault   bool
}

// GenerateHLS transcodes (or stream-copies) a video file into HLS segments.
// startOffset seeks the input before transcoding so the resulting HLS session
// starts near the user-requested timeline position.
// videoCopy=true: copies the video stream unchanged and only transcodes audio.
// For video copy, waits for the full transcode to finish (typically ~30-90s)
// so the complete playlist is available and seeking works instantly.
// For full transcode, returns as soon as the first segment is ready.
// Skips if already cached. Deduplicates concurrent requests for the same media.
func (t *Transcoder) GenerateHLS(mediaID int64, inputPath string, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	startOffset = normalizeStartOffset(startOffset)
	prefix := hlsPrefix(fileID, subtitleStreamIndex, videoCopy, startOffset)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")
	firstSeg := filepath.Join(dir, prefix+"seg_0000.ts")
	hdr := isHDRFile(inputPath)

	if videoCopy {
		if startOffset > 0 {
			// Seek-triggered video copy: return as soon as the first segment is
			// ready so the player can start playback immediately.  Subsequent
			// segments are served via WaitForSegment as FFmpeg produces them.
			return t.startHLSBackground(
				masterPath,
				firstSeg,
				inputPath,
				dir,
				prefix,
				subtitleStreamIndex,
				hdr,
				"",
				videoCopy,
				startOffset,
			)
		}
		// Initial load (no offset): wait for the full transcode so the
		// complete playlist is available and seeking works to any position.
		return t.startHLSAndWaitComplete(
			masterPath,
			firstSeg,
			inputPath,
			dir,
			prefix,
			subtitleStreamIndex,
			hdr,
			videoCopy,
			startOffset,
		)
	}
	return t.startHLSBackground(
		masterPath,
		firstSeg,
		inputPath,
		dir,
		prefix,
		subtitleStreamIndex,
		hdr,
		t.hwAccel,
		videoCopy,
		startOffset,
	)
}

// runHLSFFmpeg runs FFmpeg for single-quality HLS with the given encoder.
// videoCopy=true: copies the video stream unchanged (-c:v copy), transcodes audio only.
// hwAccel="" forces software encoding regardless of t.hwAccel.
func (t *Transcoder) runHLSFFmpeg(inputPath, dir, prefix string, siIdx int, hdr bool, hwAccel string, videoCopy bool, startOffset float64) error {
	masterPath := filepath.Join(dir, prefix+"master.m3u8")
	segPattern := filepath.Join(dir, prefix+"seg_%04d.ts")
	startOffset = normalizeStartOffset(startOffset)

	var args []string
	if videoCopy {
		// Video copy: no re-encode. Segment boundaries follow source keyframes.
		args = []string{"-hide_banner", "-loglevel", "warning",
			"-probesize", "50000000", "-analyzeduration", "100000000",
		}
		if startOffset > 0 {
			args = append(args, "-ss", fmt.Sprintf("%.3f", startOffset))
		}
		args = append(args,
			"-i", inputPath,
			"-map", "0:v:0", "-map", "0:a:0",
			"-sn",
			"-c:v", "copy",
			"-avoid_negative_ts", "make_zero",
			"-c:a", "aac", "-b:a", "192k", "-ac", "2",
			"-f", "hls",
			"-hls_time", "6",
			"-hls_list_size", "0",
			"-hls_playlist_type", "event",
			"-hls_segment_filename", segPattern,
			masterPath,
		)
	} else {
		args = []string{"-hide_banner", "-loglevel", "warning"}
		args = append(args, buildFFmpegInputArgs(hwAccel)...)
		if startOffset > 0 {
			args = append(args, "-ss", fmt.Sprintf("%.3f", startOffset))
		}
		args = append(args, "-i", inputPath)
		if siIdx >= 0 {
			args = append(args, buildImageSubtitleBurnInArgs(hwAccel, hdr, siIdx)...)
		} else {
			args = append(args, buildVideoEncodeArgs(hwAccel, hdr, siIdx, inputPath)...)
		}
		args = append(args,
			"-c:a", "aac", "-b:a", "128k", "-ac", "2",
			"-f", "hls",
			"-hls_time", "6",
			"-hls_list_size", "0",
			"-hls_playlist_type", "event",
			"-hls_segment_filename", segPattern,
			masterPath,
		)
	}

	log.Printf("transcoder: ffmpeg %s", strings.Join(args, " "))
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w — %s", err, stderr.String())
	}
	return nil
}

// GenerateHLSWithAudio generates HLS with multiple audio tracks using #EXT-X-MEDIA.
// Uses a single FFmpeg process for video + all audio outputs (one input read, one
// semaphore slot, video segments start immediately without waiting for audio).
// Falls back to simple HLS when <= 1 audio track.
// videoCopy=true: copies the video stream unchanged, transcodes audio only.
func (t *Transcoder) GenerateHLSWithAudio(mediaID int64, inputPath string, audioTracks []model.AudioTrack, fileID int64, subtitleStreamIndex int, videoCopy bool, startOffset float64) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	startOffset = normalizeStartOffset(startOffset)
	prefix := hlsPrefix(fileID, subtitleStreamIndex, videoCopy, startOffset)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	if len(audioTracks) <= 1 {
		return t.GenerateHLS(mediaID, inputPath, fileID, subtitleStreamIndex, videoCopy, startOffset)
	}

	// Check active map first (in-progress transcode).
	firstVideoSeg := filepath.Join(dir, prefix+"video_0000.ts")
	t.mu.Lock()
	if job, ok := t.active[masterPath]; ok {
		t.mu.Unlock()
		return t.waitForFirstSegment(job, firstVideoSeg)
	}
	// Full cache hit: master exists and video playlist is complete.
	videoPlaylist := filepath.Join(dir, prefix+"video.m3u8")
	if isHLSComplete(videoPlaylist) {
		t.mu.Unlock()
		return nil
	}
	job := &transcodeJob{done: make(chan struct{})}
	t.active[masterPath] = job
	t.mu.Unlock()

	hdr := isHDRFile(inputPath)

	// Build variants list (fast, synchronous).
	variants := make([]AudioVariant, 0, len(audioTracks))
	for _, track := range audioTracks {
		name := track.Title
		if name == "" {
			name = track.Language
			if name == "" {
				name = fmt.Sprintf("Audio %d", track.StreamIndex)
			}
		}
		name = strings.ReplaceAll(name, `"`, `'`)
		variants = append(variants, AudioVariant{
			Language:    track.Language,
			Name:        name,
			StreamIndex: track.StreamIndex,
			IsDefault:   track.IsDefault,
		})
	}

	// Audio playlists are namespaced with the same prefix as video to avoid
	// cache collisions between different file/subtitle/videoCopy variants.
	audioPlaylistPaths := make(map[int]string)
	for _, v := range variants {
		audioPlaylistPaths[v.StreamIndex] = filepath.Join(dir, fmt.Sprintf("%saudio_%d.m3u8", prefix, v.StreamIndex))
	}

	// Write master.m3u8 BEFORE encoding starts — it only contains URIs that
	// reference the audio and video playlists. Clients fetch those sub-playlists
	// on demand; WaitForSegment handles the case where they aren't ready yet.
	if err := t.writeMasterPlaylistWithAudio(masterPath, variants, audioPlaylistPaths, prefix); err != nil {
		t.mu.Lock()
		job.err = err
		close(job.done)
		delete(t.active, masterPath)
		t.mu.Unlock()
		return err
	}

	// Single FFmpeg process: video + all audio outputs in one pass.
	// Video copy is lightweight (no encoding) — skip semaphore to avoid
	// blocking on heavy transcode jobs that hold all slots.
	go func() {
		if !videoCopy {
			release := t.acquireSlot()
			defer release()
		}

		job.err = t.runMultiOutputHLS(
			inputPath,
			dir,
			prefix,
			variants,
			subtitleStreamIndex,
			hdr,
			t.hwAccel,
			videoCopy,
			startOffset,
		)
		if job.err != nil && t.hwAccel != "" && !videoCopy {
			log.Printf("transcoder: HW encode failed for multi-audio (%v), retrying with software", job.err)
			job.err = t.runMultiOutputHLS(
				inputPath,
				dir,
				prefix,
				variants,
				subtitleStreamIndex,
				hdr,
				"",
				videoCopy,
				startOffset,
			)
		}

		if job.err != nil {
			log.Printf("transcoder: background multi-audio transcode failed for %s: %v", masterPath, job.err)
		}

		t.mu.Lock()
		close(job.done)
		delete(t.active, masterPath)
		t.mu.Unlock()
	}()

	return t.waitForFirstSegment(job, firstVideoSeg)
}

// runMultiOutputHLS runs a single FFmpeg process that produces video HLS + N audio
// HLS outputs simultaneously. This avoids reading the input file N+1 times and
// ensures video encoding starts immediately (no audio-blocking-video bottleneck).
func (t *Transcoder) runMultiOutputHLS(inputPath, dir, prefix string, variants []AudioVariant, siIdx int, hdr bool, hwAccel string, videoCopy bool, startOffset float64) error {
	args := []string{"-hide_banner", "-loglevel", "warning"}
	startOffset = normalizeStartOffset(startOffset)
	if videoCopy {
		args = append(args, ffmpegInputProbeArgs()...)
	} else {
		args = append(args, buildFFmpegInputArgs(hwAccel)...)
	}
	if startOffset > 0 {
		args = append(args, "-ss", fmt.Sprintf("%.3f", startOffset))
	}
	args = append(args, "-i", inputPath)

	// Global option for video copy mode.
	if videoCopy {
		args = append(args, "-avoid_negative_ts", "make_zero")
	}

	// ── Video output ────────────────────────────────────────────────────────
	if videoCopy {
		args = append(args, "-map", "0:v:0", "-c:v", "copy")
	} else if siIdx >= 0 {
		args = append(args, buildImageSubtitleBurnInVideoOnlyArgs(hwAccel, hdr, siIdx)...)
	} else {
		args = append(args, "-map", "0:v:0")
		args = append(args, buildVideoEncodeArgs(hwAccel, hdr, siIdx, inputPath)...)
	}
	args = append(args,
		"-an",
		"-f", "hls", "-hls_time", "6",
		"-hls_list_size", "0", "-hls_playlist_type", "event",
		"-hls_segment_filename", filepath.Join(dir, prefix+"video_%04d.ts"),
		filepath.Join(dir, prefix+"video.m3u8"),
	)

	// ── Audio outputs (one per track, CPU-only AAC) ─────────────────────────
	for _, v := range variants {
		args = append(args,
			"-map", fmt.Sprintf("0:%d", v.StreamIndex),
			"-c:a", "aac", "-b:a", "128k", "-ac", "2",
			"-f", "hls", "-hls_time", "6",
			"-hls_list_size", "0", "-hls_playlist_type", "event",
			"-hls_segment_filename", filepath.Join(dir, fmt.Sprintf("%saudio_%d_%%04d.ts", prefix, v.StreamIndex)),
			filepath.Join(dir, fmt.Sprintf("%saudio_%d.m3u8", prefix, v.StreamIndex)),
		)
	}

	log.Printf("transcoder: ffmpeg %s", strings.Join(args, " "))
	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("transcode multi-output HLS: %w — %s", err, stderr.String())
	}
	return nil
}

// startHLSBackground launches FFmpeg for a single-stream HLS transcode in the
// background. Returns as soon as the first .ts segment exists (or an error).
// Deduplicates: a second call for the same masterPath joins the existing job.
func (t *Transcoder) startHLSBackground(masterPath, firstSeg, inputPath, dir, prefix string, siIdx int, hdr bool, hwAccel string, videoCopy bool, startOffset float64) error {
	// Check active map first (in-progress transcode).
	t.mu.Lock()
	if job, ok := t.active[masterPath]; ok {
		t.mu.Unlock()
		return t.waitForFirstSegment(job, firstSeg)
	}
	// Full cache hit: complete playlist on disk and not currently running.
	if isHLSComplete(masterPath) {
		t.mu.Unlock()
		return nil
	}
	// Start new background transcode.
	job := &transcodeJob{done: make(chan struct{})}
	t.active[masterPath] = job
	t.mu.Unlock()

	go func() {
		// Video copy is lightweight (no encoding) — skip semaphore to avoid
		// blocking on heavy transcode jobs that hold all slots.
		if !videoCopy {
			release := t.tryAcquireSlot(30 * time.Second)
			if release == nil {
				// All transcode slots busy — fail fast instead of blocking for minutes.
				t.mu.Lock()
				job.err = fmt.Errorf("all transcode slots busy, try again later")
				close(job.done)
				delete(t.active, masterPath)
				t.mu.Unlock()
				return
			}
			defer release()
		}

		err := t.runHLSFFmpeg(inputPath, dir, prefix, siIdx, hdr, hwAccel, videoCopy, startOffset)
		// Only retry with software encoder when HW encoding was attempted (not for video copy).
		if err != nil && hwAccel != "" && !videoCopy {
			log.Printf("transcoder: HW encode failed (%v), retrying with software", err)
			err = t.runHLSFFmpeg(inputPath, dir, prefix, siIdx, hdr, "", videoCopy, startOffset)
		}

		if err != nil {
			log.Printf("transcoder: background transcode failed for %s: %v", masterPath, err)
		}

		// Signal completion BEFORE removing from active map so WaitForSegment
		// can observe the done channel and fail fast instead of timing out.
		t.mu.Lock()
		job.err = err
		close(job.done)
		delete(t.active, masterPath)
		t.mu.Unlock()
	}()

	return t.waitForFirstSegment(job, firstSeg)
}

// startHLSAndWaitComplete is like startHLSBackground but waits for the full
// transcode to finish instead of returning after the first segment.
// Used for video copy mode where the transcode is fast (~30-90s) and we want
// the complete playlist available immediately so seeking works to any position.
func (t *Transcoder) startHLSAndWaitComplete(masterPath, firstSeg, inputPath, dir, prefix string, siIdx int, hdr bool, videoCopy bool, startOffset float64) error {
	t.mu.Lock()
	if job, ok := t.active[masterPath]; ok {
		t.mu.Unlock()
		// Another goroutine is already transcoding — wait for it to finish.
		<-job.done
		return job.err
	}
	if isHLSComplete(masterPath) {
		t.mu.Unlock()
		return nil
	}
	job := &transcodeJob{done: make(chan struct{})}
	t.active[masterPath] = job
	t.mu.Unlock()

	// Run synchronously — video copy is lightweight (no video encoding)
	// and skips the semaphore, so it won't block other transcodes.
	err := t.runHLSFFmpeg(inputPath, dir, prefix, siIdx, hdr, "", videoCopy, startOffset)
	if err != nil {
		log.Printf("transcoder: video copy transcode failed for %s: %v", masterPath, err)
	}

	t.mu.Lock()
	job.err = err
	close(job.done)
	delete(t.active, masterPath)
	t.mu.Unlock()

	return err
}

// writeMasterPlaylistWithAudio creates the master playlist with #EXT-X-MEDIA tags.
func (t *Transcoder) writeMasterPlaylistWithAudio(masterPath string, variants []AudioVariant, audioPaths map[int]string, prefix string) error {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:4\n")

	groupID := "audio"
	for _, v := range variants {
		// AUTOSELECT must be YES when DEFAULT=YES (RFC 8216 §4.3.4.1).
		yesNo := "NO"
		if v.IsDefault {
			yesNo = "YES"
		}
		sb.WriteString(fmt.Sprintf(
			"#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"%s\",LANGUAGE=\"%s\",NAME=\"%s\",DEFAULT=%s,AUTOSELECT=%s,URI=\"%s\"\n",
			groupID,
			v.Language,
			v.Name,
			yesNo,
			yesNo,
			filepath.Base(audioPaths[v.StreamIndex]),
		))
	}

	// 4 Mbps is a realistic estimate for CRF=22 at typical 1080p content.
	sb.WriteString(fmt.Sprintf(
		"#EXT-X-STREAM-INF:BANDWIDTH=4000000,AUDIO=\"%s\"\n",
		groupID,
	))
	sb.WriteString(prefix + "video.m3u8\n")

	return os.WriteFile(masterPath, []byte(sb.String()), 0644)
}
