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
	job := &transcodeJob{done: make(chan struct{}), lastActivity: time.Now()}
	t.active[masterPath] = job
	t.mu.Unlock()

	go func() {
		job.err = t.generateABRVariants(dir, masterPath, inputPath, sourceHeight, fileID)

		if job.err != nil {
			log.Printf("transcoder: ABR generation failed for %s: %v", masterPath, job.err)
		}

		t.mu.Lock()
		close(job.done)
		delete(t.active, masterPath)
		t.mu.Unlock()
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
			// ABR is background work — only proceed if a slot is free AND at least
			// one slot remains for realtime playback requests from active users.
			release := t.tryAcquireSlotIfSpare()
			if release == nil {
				return fmt.Errorf("generate %dp variant: all transcode slots busy", v.Height)
			}
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
