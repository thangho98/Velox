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

// Transcoder manages FFmpeg-based HLS transcoding and remuxing.
type Transcoder struct {
	outputDir string
	hwAccel   string        // resolved HW accel type ("videotoolbox", "nvenc", "vaapi", "qsv", or "")
	semaphore chan struct{} // limits concurrent FFmpeg transcode jobs
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
	return &Transcoder{
		outputDir: outputDir,
		hwAccel:   hwAccel,
		semaphore: sem,
	}
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
// Encodes both the file version and subtitle burn-in index so each unique
// (file, subtitle) combination gets its own cached playlist.
// Example: fileID=5, siIdx=2 → "f5_sub2_"
func hlsPrefix(fileID int64, subtitleStreamIndex int) string {
	var prefix string
	if fileID > 0 {
		prefix += fmt.Sprintf("f%d_", fileID)
	}
	if subtitleStreamIndex >= 0 {
		prefix += fmt.Sprintf("sub%d_", subtitleStreamIndex)
	}
	return prefix
}

// MasterPlaylistPath returns the expected path to the master playlist for the
// given (mediaID, fileID, subtitleStreamIndex) combination.
// Used by StreamService to retrieve the correct playlist path after transcoding.
func (t *Transcoder) MasterPlaylistPath(mediaID, fileID int64, subtitleStreamIndex int) string {
	return filepath.Join(t.HLSDir(mediaID), hlsPrefix(fileID, subtitleStreamIndex)+"master.m3u8")
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

// GenerateHLS transcodes a video file into HLS segments.
// Skips if already cached. Acquires a semaphore slot for the duration.
// On HW encoder failure, automatically retries with the software encoder.
func (t *Transcoder) GenerateHLS(mediaID int64, inputPath string, fileID int64, subtitleStreamIndex int) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	if _, err := os.Stat(masterPath); err == nil {
		return nil // cached
	}

	release := t.acquireSlot()
	defer release()

	hdr := isHDRFile(inputPath)
	if err := t.runHLSFFmpeg(inputPath, dir, prefix, subtitleStreamIndex, hdr, t.hwAccel); err != nil {
		if t.hwAccel != "" {
			log.Printf("transcoder: HW encode failed (%v), retrying with software", err)
			return t.runHLSFFmpeg(inputPath, dir, prefix, subtitleStreamIndex, hdr, "")
		}
		return err
	}
	return nil
}

// runHLSFFmpeg runs FFmpeg for single-quality HLS with the given encoder.
// hwAccel="" forces software encoding regardless of t.hwAccel.
func (t *Transcoder) runHLSFFmpeg(inputPath, dir, prefix string, siIdx int, hdr bool, hwAccel string) error {
	masterPath := filepath.Join(dir, prefix+"master.m3u8")
	segPattern := filepath.Join(dir, prefix+"seg_%04d.ts")

	args := []string{"-hide_banner", "-loglevel", "warning"}
	args = append(args, hwInputArgs(hwAccel)...)
	args = append(args, "-i", inputPath)
	args = append(args, buildVideoEncodeArgs(hwAccel, hdr, siIdx, inputPath)...)
	args = append(args,
		"-c:a", "aac", "-b:a", "128k", "-ac", "2",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", segPattern,
		masterPath,
	)

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
func (t *Transcoder) GenerateHLSWithAudio(mediaID int64, inputPath string, audioTracks []model.AudioTrack, fileID int64, subtitleStreamIndex int) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	if _, err := os.Stat(masterPath); err == nil {
		return nil // cached
	}

	if len(audioTracks) <= 1 {
		return t.GenerateHLS(mediaID, inputPath, fileID, subtitleStreamIndex)
	}

	release := t.acquireSlot()
	defer release()

	hdr := isHDRFile(inputPath)

	// Build variants list
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

	// Generate per-track audio playlists
	audioPlaylistPaths := make(map[int]string)
	for _, v := range variants {
		audioPlaylist := filepath.Join(dir, fmt.Sprintf("audio_%d.m3u8", v.StreamIndex))
		audioPlaylistPaths[v.StreamIndex] = audioPlaylist

		// Use absolute stream index (0:N), not relative audio index (0:a:N).
		cmd := exec.Command("ffmpeg",
			"-hide_banner", "-loglevel", "warning",
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

	// Generate video-only playlist (no audio), optionally with subtitle burn-in
	videoPlaylist := filepath.Join(dir, prefix+"video.m3u8")
	videoArgs := []string{"-hide_banner", "-loglevel", "warning"}
	videoArgs = append(videoArgs, hwInputArgs(t.hwAccel)...)
	videoArgs = append(videoArgs, "-i", inputPath, "-map", "0:v:0")
	videoArgs = append(videoArgs, buildVideoEncodeArgs(t.hwAccel, hdr, subtitleStreamIndex, inputPath)...)
	videoArgs = append(videoArgs,
		"-an",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(dir, prefix+"video_%04d.ts"),
		videoPlaylist,
	)

	cmd := exec.Command("ffmpeg", videoArgs...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if t.hwAccel != "" {
			log.Printf("transcoder: HW encode failed for video track (%v), retrying with software", err)
			swArgs := []string{"-hide_banner", "-loglevel", "warning", "-i", inputPath, "-map", "0:v:0"}
			swArgs = append(swArgs, buildVideoEncodeArgs("", hdr, subtitleStreamIndex, inputPath)...)
			swArgs = append(swArgs,
				"-an", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0",
				"-hls_segment_filename", filepath.Join(dir, prefix+"video_%04d.ts"),
				videoPlaylist,
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

	return t.writeMasterPlaylistWithAudio(masterPath, variants, audioPlaylistPaths, prefix)
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
// one variant. Skips if already cached.
func (t *Transcoder) GenerateABRHLS(mediaID int64, inputPath string, sourceHeight int, fileID int64) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	masterPath := t.ABRMasterPath(mediaID, fileID)
	if _, err := os.Stat(masterPath); err == nil {
		return nil // cached
	}

	release := t.acquireSlot()
	defer release()

	// Select variants at or below source resolution.
	var variants []ABRVariant
	for _, v := range defaultABRVariants {
		if v.Height <= sourceHeight {
			variants = append(variants, v)
		}
	}
	if len(variants) == 0 {
		// Source is lower than 480p; encode at source height to avoid upscaling.
		// Scale bitrate proportionally to resolution area vs 480p baseline.
		h := sourceHeight
		if h <= 0 {
			h = defaultABRVariants[0].Height
		}
		base := defaultABRVariants[0]
		scaledBitrate := base.Bitrate * h * h / (base.Height * base.Height)
		if scaledBitrate < 200 {
			scaledBitrate = 200 // floor: avoid unusably low bitrate
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
			if err := t.generateABRVariant(inputPath, playlistPath, segPattern, v, t.hwAccel); err != nil {
				if t.hwAccel != "" {
					log.Printf("transcoder: HW encode failed for %dp ABR (%v), retrying with software", v.Height, err)
					if err2 := t.generateABRVariant(inputPath, playlistPath, segPattern, v, ""); err2 != nil {
						return fmt.Errorf("generate %dp variant: %w", v.Height, err2)
					}
				} else {
					return fmt.Errorf("generate %dp variant: %w", v.Height, err)
				}
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
	args = append(args, hwInputArgs(hwAccel)...)
	args = append(args, "-i", inputPath)
	args = append(args,
		"-vf", fmt.Sprintf("scale=-2:%d", v.Height),
		"-c:v", hwVideoCodec(hwAccel),
	)
	if hwAccel == "" {
		args = append(args, "-preset", "fast", "-profile:v", "high", "-level", "4.1")
	}
	args = append(args,
		"-b:v", bitrateStr,
		"-maxrate", maxrateStr,
		"-bufsize", bufsizeStr,
		"-pix_fmt", "yuv420p",
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
		return []string{"-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128"}
	case "nvenc":
		return []string{"-hwaccel", "cuda"}
	case "qsv":
		return []string{"-hwaccel", "qsv"}
	}
	return nil
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
	if siIdx >= 0 {
		escaped := escapeFFmpegSubtitlePath(inputPath)
		filters = append(filters, fmt.Sprintf("subtitles='%s':si=%d", escaped, siIdx))
	}

	var args []string
	if len(filters) > 0 {
		args = append(args, "-vf", strings.Join(filters, ","))
	}

	args = append(args, "-c:v", hwVideoCodec(hwAccel))
	if hwAccel == "" {
		args = append(args, "-preset", "fast", "-crf", "22")
	}
	args = append(args, "-pix_fmt", "yuv420p")
	return args
}

// escapeFFmpegSubtitlePath escapes a file path for use in FFmpeg's subtitles filter.
// The path is wrapped in single quotes in the filter string; only backslash and
// single-quote characters need to be escaped inside the single-quoted section.
func escapeFFmpegSubtitlePath(path string) string {
	path = strings.ReplaceAll(path, `\`, `\\`)
	path = strings.ReplaceAll(path, `'`, `\'`)
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
