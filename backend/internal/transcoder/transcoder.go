package transcoder

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/thawng/velox/internal/model"
)

// ABRVariant describes a single quality level for adaptive bitrate HLS.
type ABRVariant struct {
	Height    int // target video height (px)
	Bitrate   int // target video bitrate (kbps)
	Bandwidth int // bits/s used in master playlist BANDWIDTH attribute
}

// defaultABRVariants are the standard quality ladder, lowest to highest.
var defaultABRVariants = []ABRVariant{
	{Height: 480, Bitrate: 1500, Bandwidth: 1_500_000},
	{Height: 720, Bitrate: 4000, Bandwidth: 4_000_000},
	{Height: 1080, Bitrate: 8000, Bandwidth: 8_000_000},
}

// hasSubtitlesFilter is set once at init — true when FFmpeg was built with libass.
var hasSubtitlesFilter = detectSubtitlesFilter()

// SupportsSubtitleBurnIn reports whether the local FFmpeg build can burn subtitles.
func SupportsSubtitleBurnIn() bool {
	return hasSubtitlesFilter
}

func detectSubtitlesFilter() bool {
	out, err := exec.Command("ffmpeg", "-filters").CombinedOutput()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), "subtitles")
}

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

// isHLSComplete reports whether a media playlist has been fully written
// (i.e. FFmpeg added #EXT-X-ENDLIST at the end).
func isHLSComplete(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "#EXT-X-ENDLIST")
}

// waitForFirstSegment polls until segPath exists OR the job finishes.
// Returns nil as soon as the first segment appears (FFmpeg still running in background).
// Returns job.err if FFmpeg exits before the segment appears.
func (t *Transcoder) waitForFirstSegment(job *transcodeJob, segPath string) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(10 * time.Minute)

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

// startHLSBackground launches FFmpeg for a single-stream HLS transcode in the
// background. Returns as soon as the first .ts segment exists (or an error).
// Deduplicates: a second call for the same masterPath joins the existing job.
func (t *Transcoder) startHLSBackground(masterPath, firstSeg, inputPath, dir, prefix string, siIdx int, hdr bool, hwAccel string, videoCopy bool) error {
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
			release := t.acquireSlot()
			defer release()
		}

		err := t.runHLSFFmpeg(inputPath, dir, prefix, siIdx, hdr, hwAccel, videoCopy)
		// Only retry with software encoder when HW encoding was attempted (not for video copy).
		if err != nil && hwAccel != "" && !videoCopy {
			log.Printf("transcoder: HW encode failed (%v), retrying with software", err)
			err = t.runHLSFFmpeg(inputPath, dir, prefix, siIdx, hdr, "", videoCopy)
		}

		t.mu.Lock()
		job.err = err
		delete(t.active, masterPath)
		t.mu.Unlock()

		if err != nil {
			log.Printf("transcoder: background transcode failed for %s: %v", masterPath, err)
		}
		close(job.done)
	}()

	return t.waitForFirstSegment(job, firstSeg)
}

// acquireSlot blocks until a transcode slot is available.
// The returned function must be deferred to release the slot.
func (t *Transcoder) acquireSlot() func() {
	<-t.semaphore
	return func() { t.semaphore <- struct{}{} }
}

// HLSDir returns the directory where HLS segments for a media item are stored.
func (t *Transcoder) HLSDir(mediaID int64) string {
	return filepath.Join(t.outputDir, fmt.Sprintf("%d", mediaID))
}

// hlsPrefix returns the filename prefix used for HLS output files.
// Encodes file version, subtitle burn-in index, and video copy mode so each
// unique (file, subtitle, videoCopy) combination gets its own cached playlist.
// Example: fileID=5, siIdx=2 → "f5_sub2_"
// Example: fileID=5, videoCopy=true → "vcf5_"
func hlsPrefix(fileID int64, subtitleStreamIndex int, videoCopy bool) string {
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
	return prefix
}

// MasterPlaylistPath returns the expected path to the master playlist for the
// given (mediaID, fileID, subtitleStreamIndex, videoCopy) combination.
// Used by StreamService to retrieve the correct playlist path after transcoding.
func (t *Transcoder) MasterPlaylistPath(mediaID, fileID int64, subtitleStreamIndex int, videoCopy bool) string {
	return filepath.Join(t.HLSDir(mediaID), hlsPrefix(fileID, subtitleStreamIndex, videoCopy)+"master.m3u8")
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

// GenerateHLS transcodes (or stream-copies) a video file into HLS segments.
// videoCopy=true: copies the video stream unchanged and only transcodes audio.
// Returns as soon as the first segment is ready (FFmpeg continues in background).
// Skips if already cached. Deduplicates concurrent requests for the same media.
func (t *Transcoder) GenerateHLS(mediaID int64, inputPath string, fileID int64, subtitleStreamIndex int, videoCopy bool) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex, videoCopy)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")
	firstSeg := filepath.Join(dir, prefix+"seg_0000.ts")
	hdr := isHDRFile(inputPath)

	return t.startHLSBackground(masterPath, firstSeg, inputPath, dir, prefix, subtitleStreamIndex, hdr, t.hwAccel, videoCopy)
}

// runHLSFFmpeg runs FFmpeg for single-quality HLS with the given encoder.
// videoCopy=true: copies the video stream unchanged (-c:v copy), transcodes audio only.
// hwAccel="" forces software encoding regardless of t.hwAccel.
func (t *Transcoder) runHLSFFmpeg(inputPath, dir, prefix string, siIdx int, hdr bool, hwAccel string, videoCopy bool) error {
	masterPath := filepath.Join(dir, prefix+"master.m3u8")
	segPattern := filepath.Join(dir, prefix+"seg_%04d.ts")

	var args []string
	if videoCopy {
		// Video copy: no re-encode. Segment boundaries follow source keyframes.
		args = []string{"-hide_banner", "-loglevel", "warning",
			"-probesize", "50000000", "-analyzeduration", "100000000",
			"-i", inputPath,
			"-map", "0:v:0", "-map", "0:a:0?",
			"-c:v", "copy",
			"-avoid_negative_ts", "make_zero",
			"-c:a", "aac", "-b:a", "192k", "-ac", "2",
			"-f", "hls",
			"-hls_time", "6",
			"-hls_list_size", "0",
			"-hls_segment_filename", segPattern,
			masterPath,
		}
	} else {
		args = []string{"-hide_banner", "-loglevel", "warning"}
		args = append(args, buildFFmpegInputArgs(hwAccel)...)
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

// AudioVariant represents an audio track variant for HLS.
type AudioVariant struct {
	Language    string
	Name        string
	StreamIndex int
	IsDefault   bool
}

// GenerateHLSWithAudio generates HLS with multiple audio tracks using #EXT-X-MEDIA.
// Creates separate audio playlists and a master playlist that references them.
// Falls back to simple HLS when <= 1 audio track.
// videoCopy=true: copies the video stream unchanged, transcodes audio only.
func (t *Transcoder) GenerateHLSWithAudio(mediaID int64, inputPath string, audioTracks []model.AudioTrack, fileID int64, subtitleStreamIndex int, videoCopy bool) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex, videoCopy)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	if len(audioTracks) <= 1 {
		return t.GenerateHLS(mediaID, inputPath, fileID, subtitleStreamIndex, videoCopy)
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
		// Strip double quotes — they would break the M3U8 quoted-string format.
		name = strings.ReplaceAll(name, `"`, `'`)
		variants = append(variants, AudioVariant{
			Language:    track.Language,
			Name:        name,
			StreamIndex: track.StreamIndex,
			IsDefault:   track.IsDefault,
		})
	}

	// Run all encoding in a background goroutine so the HTTP handler can return
	// as soon as the first video segment is ready.
	go func() {
		release := t.acquireSlot()
		defer release()

		job.err = func() error {
			// --- Audio tracks (audio-only encode, much faster than video) ---
			audioPlaylistPaths := make(map[int]string)
			for _, v := range variants {
				audioPlaylist := filepath.Join(dir, fmt.Sprintf("audio_%d.m3u8", v.StreamIndex))
				audioPlaylistPaths[v.StreamIndex] = audioPlaylist

				cmd := exec.Command("ffmpeg",
					"-hide_banner", "-loglevel", "warning",
					"-probesize", "50000000",
					"-analyzeduration", "100000000",
					"-i", inputPath,
					"-map", fmt.Sprintf("0:%d", v.StreamIndex),
					"-c:a", "aac", "-b:a", "128k", "-ac", "2",
					"-f", "hls",
					"-hls_time", "6",
					"-hls_list_size", "0",
					"-hls_segment_filename", filepath.Join(dir, fmt.Sprintf("audio_%d_%%04d.ts", v.StreamIndex)),
					audioPlaylist,
				)
				var stderr bytes.Buffer
				cmd.Stderr = &stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("transcode audio stream %d: %w — %s", v.StreamIndex, err, stderr.String())
				}
			}

			// Write master.m3u8 before video starts so it is available as soon
			// as the first video segment appears (caller polls for that segment).
			if err := t.writeMasterPlaylistWithAudio(masterPath, variants, audioPlaylistPaths, prefix); err != nil {
				return err
			}

			// --- Video track (the slow part — runs fully in background) ---
			vp := filepath.Join(dir, prefix+"video.m3u8")
			segFile := filepath.Join(dir, prefix+"video_%04d.ts")
			var videoArgs []string
			if videoCopy {
				videoArgs = []string{
					"-hide_banner", "-loglevel", "warning",
					"-probesize", "50000000", "-analyzeduration", "100000000",
					"-i", inputPath,
					"-map", "0:v:0", "-c:v", "copy",
					"-avoid_negative_ts", "make_zero",
					"-an", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
					"-hls_segment_filename", segFile,
					vp,
				}
			} else {
				videoArgs = []string{"-hide_banner", "-loglevel", "warning"}
				videoArgs = append(videoArgs, buildFFmpegInputArgs(t.hwAccel)...)
				videoArgs = append(videoArgs, "-i", inputPath)
				if subtitleStreamIndex >= 0 {
					videoArgs = append(videoArgs, buildImageSubtitleBurnInVideoOnlyArgs(t.hwAccel, hdr, subtitleStreamIndex)...)
				} else {
					videoArgs = append(videoArgs, "-map", "0:v:0")
					videoArgs = append(videoArgs, buildVideoEncodeArgs(t.hwAccel, hdr, subtitleStreamIndex, inputPath)...)
				}
				videoArgs = append(videoArgs,
					"-an", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
					"-hls_segment_filename", segFile,
					vp,
				)
			}
			cmd := exec.Command("ffmpeg", videoArgs...)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				if t.hwAccel != "" && !videoCopy {
					log.Printf("transcoder: HW encode failed for video track (%v), retrying with software", err)
					swArgs := []string{
						"-hide_banner", "-loglevel", "warning",
						"-probesize", "50000000",
						"-analyzeduration", "100000000",
						"-i", inputPath,
					}
					if subtitleStreamIndex >= 0 {
						swArgs = append(swArgs, buildImageSubtitleBurnInVideoOnlyArgs("", hdr, subtitleStreamIndex)...)
					} else {
						swArgs = append(swArgs, "-map", "0:v:0")
						swArgs = append(swArgs, buildVideoEncodeArgs("", hdr, subtitleStreamIndex, inputPath)...)
					}
					swArgs = append(swArgs,
						"-an", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
						"-hls_segment_filename", segFile,
						vp,
					)
					cmdSW := exec.Command("ffmpeg", swArgs...)
					var stderr2 bytes.Buffer
					cmdSW.Stderr = &stderr2
					if err2 := cmdSW.Run(); err2 != nil {
						return fmt.Errorf("transcode video: %w — %s", err2, stderr2.String())
					}
				} else {
					return fmt.Errorf("transcode video: %w — %s", err, stderr.String())
				}
			}
			return nil
		}()

		t.mu.Lock()
		delete(t.active, masterPath)
		t.mu.Unlock()
		if job.err != nil {
			log.Printf("transcoder: background multi-audio transcode failed for %s: %v", masterPath, job.err)
		}
		close(job.done)
	}()

	return t.waitForFirstSegment(job, firstVideoSeg)
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

// GenerateABRHLS generates multi-quality adaptive bitrate HLS variants.
// Only generates qualities at or below sourceHeight. Always generates at least
// one variant. Skips if already cached. Deduplicates concurrent requests.
func (t *Transcoder) GenerateABRHLS(mediaID int64, inputPath string, sourceHeight int, fileID int64) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	masterPath := t.ABRMasterPath(mediaID, fileID)

	// Dedup: if another goroutine is already generating this ABR set, wait for it.
	t.mu.Lock()
	if job, ok := t.active[masterPath]; ok {
		t.mu.Unlock()
		<-job.done
		return job.err
	}
	// Cached on disk.
	if _, err := os.Stat(masterPath); err == nil {
		t.mu.Unlock()
		return nil
	}
	job := &transcodeJob{done: make(chan struct{})}
	t.active[masterPath] = job
	t.mu.Unlock()

	go func() {
		job.err = t.generateABRVariants(dir, masterPath, inputPath, sourceHeight, fileID)

		t.mu.Lock()
		delete(t.active, masterPath)
		t.mu.Unlock()
		if job.err != nil {
			log.Printf("transcoder: ABR generation failed for %s: %v", masterPath, job.err)
		}
		close(job.done)
	}()

	<-job.done
	return job.err
}

// generateABRVariants encodes all ABR quality levels sequentially.
// Each variant acquires its own semaphore slot so other transcode jobs
// (e.g. a different media item) can interleave.
func (t *Transcoder) generateABRVariants(dir, masterPath, inputPath string, sourceHeight int, fileID int64) error {
	// Select variants at or below source resolution.
	var variants []ABRVariant
	for _, v := range defaultABRVariants {
		if v.Height <= sourceHeight {
			variants = append(variants, v)
		}
	}
	if len(variants) == 0 {
		// Source is lower than 480p; encode at source height to avoid upscaling.
		h := sourceHeight
		if h <= 0 {
			h = defaultABRVariants[0].Height
		}
		base := defaultABRVariants[0]
		scaledBitrate := base.Bitrate * h * h / (base.Height * base.Height)
		if scaledBitrate < 200 {
			scaledBitrate = 200
		}
		variants = []ABRVariant{{
			Height:    h,
			Bitrate:   scaledBitrate,
			Bandwidth: scaledBitrate * 1000,
		}}
	}

	playlistNames := make([]string, 0, len(variants))
	for _, v := range variants {
		name := fmt.Sprintf("f%d_q%d.m3u8", fileID, v.Height)
		playlistPath := filepath.Join(dir, name)
		segPattern := filepath.Join(dir, fmt.Sprintf("f%d_q%d_seg_%%04d.ts", fileID, v.Height))

		if _, err := os.Stat(playlistPath); err != nil {
			release := t.acquireSlot()
			// ABR uses software encoding — HW encoders on low-VRAM iGPUs
			// (e.g. 64MB shared) OOM when combining scale + bitrate control.
			err := t.generateABRVariant(inputPath, playlistPath, segPattern, v, "")
			release()
			if err != nil {
				return fmt.Errorf("generate %dp variant: %w", v.Height, err)
			}
		}
		playlistNames = append(playlistNames, name)
	}

	return t.writeABRMasterPlaylist(masterPath, variants, playlistNames)
}

// generateABRVariant runs FFmpeg for a single ABR quality variant.
// hwAccel="" forces software encoding.
func (t *Transcoder) generateABRVariant(inputPath, playlistPath, segPattern string, v ABRVariant, hwAccel string) error {
	bitrateStr := fmt.Sprintf("%dk", v.Bitrate)
	maxrateStr := fmt.Sprintf("%dk", int(float64(v.Bitrate)*1.2))
	bufsizeStr := fmt.Sprintf("%dk", v.Bitrate*2)

	args := []string{"-hide_banner", "-loglevel", "warning"}
	args = append(args, buildFFmpegInputArgs(hwAccel)...)
	args = append(args, "-i", inputPath)
	args = append(args,
		"-vf", hwScaleFilter(hwAccel, v.Height),
		"-c:v", hwVideoCodec(hwAccel),
	)
	switch hwAccel {
	case "":
		args = append(args, "-preset", "veryfast", "-profile:v", "high", "-level", "4.1", "-threads", "0")
		args = append(args, "-pix_fmt", "yuv420p")
	case "vaapi":
		args = append(args, "-profile:v", "main")
	}
	args = append(args,
		"-b:v", bitrateStr,
		"-maxrate", maxrateStr,
		"-bufsize", bufsizeStr,
		"-c:a", "aac", "-b:a", "128k", "-ac", "2",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_filename", segPattern,
		playlistPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg %dp: %w — %s", v.Height, err, stderr.String())
	}
	return nil
}

// writeABRMasterPlaylist writes master.m3u8 with #EXT-X-STREAM-INF per quality level.
func (t *Transcoder) writeABRMasterPlaylist(masterPath string, variants []ABRVariant, playlistNames []string) error {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:4\n")

	for i, v := range variants {
		// Approximate width for 16:9 content, rounded to nearest even number.
		width := (v.Height*16/9 + 1) &^ 1
		sb.WriteString(fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d,CODECS=\"avc1.640028,mp4a.40.2\"\n",
			v.Bandwidth, width, v.Height,
		))
		sb.WriteString(playlistNames[i] + "\n")
	}

	return os.WriteFile(masterPath, []byte(sb.String()), 0644)
}

// SegmentPath returns the full path to an HLS segment file.
func (t *Transcoder) SegmentPath(mediaID int64, segment string) string {
	return filepath.Join(t.HLSDir(mediaID), segment)
}

// WaitForSegment waits up to timeout for a segment file to appear on disk.
// Returns true if the segment exists, false if timed out.
// Used when FFmpeg is still writing segments and the player requests one
// that hasn't been flushed yet — avoids 404 spam from the client.
func (t *Transcoder) WaitForSegment(path string, timeout time.Duration) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}

	// Only wait if there's an active transcode that could produce this segment.
	t.mu.Lock()
	hasActive := len(t.active) > 0
	t.mu.Unlock()
	if !hasActive {
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
		case <-deadline:
			return false
		}
	}
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

// --- HW accel helpers ---

// hwInputArgs returns FFmpeg input-side args for the given hardware accelerator.
func hwInputArgs(hwAccel string) []string {
	switch hwAccel {
	case "videotoolbox":
		return []string{"-hwaccel", "videotoolbox"}
	case "vaapi":
		return []string{"-hwaccel", "vaapi", "-hwaccel_output_format", "vaapi", "-hwaccel_device", "/dev/dri/renderD128"}
	case "nvenc":
		return []string{"-hwaccel", "cuda"}
	case "qsv":
		return []string{"-hwaccel", "qsv"}
	}
	return nil
}

// hwScaleFilter returns the appropriate scale filter for the given HW accelerator.
// For VAAPI, forces NV12 output format so h264_vaapi can encode 10-bit sources
// (e.g. HEVC P010) that would otherwise fail with "No usable encoding profile".
func hwScaleFilter(hwAccel string, height int) string {
	switch hwAccel {
	case "vaapi":
		return fmt.Sprintf("scale_vaapi=w=-2:h=%d:format=nv12", height)
	default:
		return fmt.Sprintf("scale=-2:%d", height)
	}
}

// ffmpegInputProbeArgs increases demux probing so image-based subtitle streams
// like PGS are discovered reliably before we attempt burn-in.
func ffmpegInputProbeArgs() []string {
	return []string{
		"-probesize", "50000000",
		"-analyzeduration", "100000000",
	}
}

func buildFFmpegInputArgs(hwAccel string) []string {
	args := ffmpegInputProbeArgs()
	args = append(args, hwInputArgs(hwAccel)...)
	return args
}

// hwVideoCodec returns the FFmpeg video encoder for the given HW accelerator.
// Falls back to libx264 when hwAccel is empty.
func hwVideoCodec(hwAccel string) string {
	switch hwAccel {
	case "videotoolbox":
		return "h264_videotoolbox"
	case "vaapi":
		return "h264_vaapi"
	case "nvenc":
		return "h264_nvenc"
	case "qsv":
		return "h264_qsv"
	}
	return "libx264"
}

// buildVideoEncodeArgs builds -vf + -c:v args for single-quality HLS encoding.
// Handles HW encoder selection, HDR→SDR tone mapping, and subtitle burn-in.
func buildVideoEncodeArgs(hwAccel string, hdr bool, siIdx int, inputPath string) []string {
	var filters []string
	if hdr {
		filters = append(filters, hdrToneMapFilter())
	}
	if siIdx >= 0 && hasSubtitlesFilter {
		escaped := escapeFFmpegSubtitlePath(inputPath)
		filters = append(filters, fmt.Sprintf("subtitles=filename='%s':si=%d", escaped, siIdx))
	}
	// VAAPI: force NV12 surface format so h264_vaapi can encode 10-bit sources.
	if hwAccel == "vaapi" && len(filters) == 0 {
		filters = append(filters, "scale_vaapi=format=nv12")
	}

	var args []string
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-c:v", hwVideoCodec(hwAccel))
	args = append(args, hwEncoderArgs(hwAccel)...)
	return args
}

// buildImageSubtitleBurnInArgs burns a bitmap subtitle stream (PGS/VobSub) into
// the primary video using filter_complex overlay. The selected subtitle stream is
// referenced by absolute stream index on input 0.
func buildImageSubtitleBurnInArgs(hwAccel string, hdr bool, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hdr, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-map", "0:a:0?",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
	return args
}

// buildImageSubtitleBurnInVideoOnlyArgs is the same burn-in path as
// buildImageSubtitleBurnInArgs, but only maps the filtered video output.
func buildImageSubtitleBurnInVideoOnlyArgs(hwAccel string, hdr bool, subtitleStreamIndex int) []string {
	complexFilter := buildImageSubtitleBurnInFilter(hdr, subtitleStreamIndex)
	args := []string{
		"-filter_complex", complexFilter,
		"-map", "[vout]",
		"-c:v", hwVideoCodec(hwAccel),
	}
	args = append(args, hwEncoderArgs(hwAccel)...)
	return args
}

// hwEncoderArgs returns encoder-specific args (profile, quality, pixel format)
// for the given HW accelerator. Centralizes encoder tuning so all encode paths
// use consistent settings.
func hwEncoderArgs(hwAccel string) []string {
	switch hwAccel {
	case "":
		return []string{"-preset", "veryfast", "-crf", "23", "-threads", "0", "-pix_fmt", "yuv420p"}
	case "vaapi":
		return []string{"-profile:v", "main", "-qp", "23"}
	default:
		return nil
	}
}

func buildImageSubtitleBurnInFilter(hdr bool, subtitleStreamIndex int) string {
	if hdr {
		return fmt.Sprintf("[0:v:0]%s[base];[base][0:%d]overlay[vout]", hdrToneMapFilter(), subtitleStreamIndex)
	}
	return fmt.Sprintf("[0:v:0][0:%d]overlay[vout]", subtitleStreamIndex)
}

// escapeFFmpegSubtitlePath escapes a file path for FFmpeg's subtitles filter.
// The subtitles filter uses libass which requires escaping at two levels:
//  1. Filter option level: : ; [ ] ' \
//  2. The path is NOT wrapped in quotes — all special chars are backslash-escaped.
func escapeFFmpegSubtitlePath(path string) string {
	// Order matters: escape backslash first
	path = strings.ReplaceAll(path, `\`, `\\`)
	path = strings.ReplaceAll(path, `'`, `\'`)
	path = strings.ReplaceAll(path, `:`, `\:`)
	path = strings.ReplaceAll(path, `[`, `\[`)
	path = strings.ReplaceAll(path, `]`, `\]`)
	path = strings.ReplaceAll(path, `;`, `\;`)
	return path
}

// --- HDR detection ---

// isHDRFile returns true if the input file's primary video stream uses HDR
// color transfer (PQ/SMPTE2084) or BT.2020 color primaries.
func isHDRFile(inputPath string) bool {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-select_streams", "v:0",
		"-show_entries", "stream=color_transfer,color_primaries",
		"-of", "default=noprint_wrappers=1",
		inputPath,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("transcoder: HDR probe failed for %q: %v", inputPath, err)
		return false
	}
	lower := strings.ToLower(out.String())
	return strings.Contains(lower, "smpte2084") || strings.Contains(lower, "bt2020")
}

// hdrToneMapFilter returns the FFmpeg -vf filter chain for HDR→SDR tone mapping.
// Output is SDR BT.709 in yuv420p, suitable for H.264/H.265 HLS streaming.
func hdrToneMapFilter() string {
	return "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable,zscale=t=bt709:m=bt709,format=yuv420p"
}
