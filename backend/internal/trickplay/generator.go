package trickplay

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/thawng/velox/pkg/ffmpegbin"
	"github.com/thawng/velox/pkg/ffprobe"
)

const (
	tileWidth       = 320                    // px per thumbnail frame
	tileHeight      = 180                    // px per thumbnail frame (16:9)
	tileColumns     = 10                     // tiles per row in sprite sheet
	tileRows        = 10                     // rows per sprite sheet
	framesPerSprite = tileColumns * tileRows // 100 frames per sprite sheet
)

// Generator creates trickplay sprite sheets and VTT manifests for media items.
type Generator struct {
	outputDir  string
	interval   int      // seconds between thumbnail frames
	hwAccel    string   // hardware accelerator ("vaapi", "videotoolbox", "nvenc", etc.)
	inProgress sync.Map // mediaID (int64) → struct{}; prevents duplicate concurrent generation
}

// New creates a Generator.
// interval: seconds between thumbnail frames (default 10 if <= 0).
// hwAccel: hardware accelerator to use for decoding (e.g., "vaapi", "" for software).
func New(outputDir string, interval int, hwAccel string) *Generator {
	if interval <= 0 {
		interval = 10
	}
	return &Generator{outputDir: outputDir, interval: interval, hwAccel: hwAccel}
}

// MediaDir returns the trickplay directory for a media item.
func (g *Generator) MediaDir(mediaID int64) string {
	return filepath.Join(g.outputDir, fmt.Sprintf("%d", mediaID))
}

// VTTPath returns the path to the VTT manifest for a media item.
func (g *Generator) VTTPath(mediaID int64) string {
	return filepath.Join(g.MediaDir(mediaID), "manifest.vtt")
}

// SpritePath returns the path to a sprite sheet file (1-based index).
func (g *Generator) SpritePath(mediaID int64, index int) string {
	return filepath.Join(g.MediaDir(mediaID), fmt.Sprintf("sprite_%d.jpg", index))
}

// IsDone reports whether trickplay has been fully generated for this media item.
func (g *Generator) IsDone(mediaID int64) bool {
	_, err := os.Stat(g.VTTPath(mediaID))
	return err == nil
}

// GenerateAsync starts trickplay generation in a background goroutine.
// Returns immediately. No-op if already in progress.
func (g *Generator) GenerateAsync(mediaID int64, inputPath string, durationSec int) {
	// LoadOrStore is the gating mechanism. Only the goroutine that wins the store
	// proceeds; others return immediately without launching a duplicate goroutine.
	if _, loaded := g.inProgress.LoadOrStore(mediaID, struct{}{}); loaded {
		return
	}
	go func() {
		defer g.inProgress.Delete(mediaID)
		// Re-check IsDone inside the goroutine: generation may have completed in
		// the window between the caller's entry and this goroutine starting.
		if g.IsDone(mediaID) {
			return
		}
		if err := g.Generate(mediaID, inputPath, durationSec); err != nil {
			log.Printf("trickplay: generate media %d: %v", mediaID, err)
		}
	}()
}

// Generate generates sprite sheets and VTT manifest for a media file.
// Runs FFmpeg synchronously; intended to be called from a goroutine.
// Uses seeking + frame extraction to handle problematic codecs (e.g., Dolby Vision).
func (g *Generator) Generate(mediaID int64, inputPath string, durationSec int) error {
	dir := g.MediaDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	// Detect HDR content to enable tone mapping.
	isHDR := isHDRFile(inputPath)

	// One frame every interval seconds; at least one frame for very short files.
	totalFrames := durationSec / g.interval
	if totalFrames == 0 {
		totalFrames = 1
	}

	// Cap total frames to avoid excessive generation time.
	// For very long videos, this means less coverage but faster generation.
	if totalFrames > framesPerSprite*10 {
		totalFrames = framesPerSprite * 10
	}

	// Use seeking-based extraction which works with problematic codecs like Dolby Vision.
	// The fps=1/N + tile filter fails for DOV content; seeking + frame extraction works.
	frames, err := g.extractFramesWithSeek(mediaID, inputPath, durationSec, totalFrames, isHDR)
	if err != nil {
		return fmt.Errorf("extract frames: %w", err)
	}

	if len(frames) == 0 {
		return fmt.Errorf("no frames extracted")
	}

	// Group frames into sprite sheets using tile filter.
	// We write frames to a temp list file, then use ffmpeg with concat demuxer + tile.
	if err := g.createSpriteSheets(mediaID, frames); err != nil {
		return fmt.Errorf("create sprite sheets: %w", err)
	}

	// Verify at least one sprite was produced
	if _, err := os.Stat(g.SpritePath(mediaID, 1)); err != nil {
		return fmt.Errorf("no sprites generated: %w", err)
	}

	// Cleanup frames directory now that sprites are created to save disk space.
	framesDir := filepath.Join(g.MediaDir(mediaID), "frames")
	if err := os.RemoveAll(framesDir); err != nil {
		log.Printf("trickplay: warning: failed to cleanup frames dir for media %d: %v", mediaID, err)
	}

	// Use actual frame count (len(frames)) in case some extractions failed.
	// Note: if extraction failed for some frames, VTT entries will reference
	// existing frames only (sorted by filename), so timestamps may not be exact.
	actualFrameCount := len(frames)
	vtt := g.buildVTTFromFrames(mediaID, frames, g.interval)
	if err := os.WriteFile(g.VTTPath(mediaID), []byte(vtt), 0644); err != nil {
		return fmt.Errorf("write vtt: %w", err)
	}

	spriteCount := int(math.Ceil(float64(actualFrameCount) / float64(framesPerSprite)))
	log.Printf("trickplay: generated %d frames for media %d (%d sprites)",
		actualFrameCount, mediaID, spriteCount)
	return nil
}

// extractFramesWithSeek extracts frames at evenly-spaced timestamps using seeking.
// This approach works with codecs like Dolby Vision that fail with fps filter + tile.
// Frames are stored in persistent storage so they survive container restarts.
func (g *Generator) extractFramesWithSeek(mediaID int64, inputPath string, durationSec, totalFrames int, isHDR bool) ([]string, error) {
	// Use persistent storage instead of /tmp so frames survive container restarts.
	framesDir := filepath.Join(g.MediaDir(mediaID), "frames")
	if err := os.MkdirAll(framesDir, 0755); err != nil {
		return nil, fmt.Errorf("mkdir frames dir: %w", err)
	}

	// Build list of timestamps (in seconds) for frame extraction.
	var timestamps []int
	interval := durationSec / totalFrames
	if interval < 1 {
		interval = 1
	}
	for i := 0; i < totalFrames; i++ {
		ts := i * interval
		if ts >= durationSec {
			break
		}
		timestamps = append(timestamps, ts)
	}

	// Extract frames at each timestamp, skipping already-extracted frames (resume support).
	for i, ts := range timestamps {
		outputPath := filepath.Join(framesDir, fmt.Sprintf("frame_%05d.jpg", i))
		// Skip if frame already exists (allows resume after container restart).
		if _, err := os.Stat(outputPath); err == nil {
			continue
		}
		if err := g.extractFrameAt(inputPath, ts, outputPath, isHDR); err != nil {
			log.Printf("trickplay: warning: failed to extract frame at %ds for media %d: %v", ts, mediaID, err)
			continue
		}
	}

	// Collect all extracted frame paths.
	entries, err := os.ReadDir(framesDir)
	if err != nil {
		return nil, fmt.Errorf("read frames dir: %w", err)
	}
	var frames []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jpg") {
			frames = append(frames, filepath.Join(framesDir, entry.Name()))
		}
	}
	// Sort frames by name to ensure correct order.
	sort.Strings(frames)

	return frames, nil
}

// extractFrameAt extracts a single frame from inputPath at the given timestamp (seconds).
// Uses seeking to avoid decoding the entire video.
// For HDR content, applies tone mapping to convert to SDR.
func (g *Generator) extractFrameAt(inputPath string, timestampSec int, outputPath string, isHDR bool) error {
	args := []string{"-hide_banner", "-loglevel", "error"}
	if g.hwAccel != "" {
		switch g.hwAccel {
		case "vaapi":
			args = append(args, "-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128")
		case "videotoolbox":
			args = append(args, "-hwaccel", "videotoolbox")
		case "nvenc":
			args = append(args, "-hwaccel", "cuda")
		case "qsv":
			args = append(args, "-hwaccel", "qsv")
		}
	}
	args = append(args, "-ss", fmt.Sprintf("%d", timestampSec))
	args = append(args, "-i", inputPath)

	// Build video filter chain.
	vf := g.buildFrameFilter(isHDR)
	args = append(args, "-vf", vf)
	args = append(args, "-frames:v", "1", "-q:v", "5", "-y", outputPath)

	cmd := exec.Command(ffmpegbin.FFmpeg(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg: %w — %s", err, stderr.String())
	}
	return nil
}

// buildFrameFilter returns the FFmpeg video filter chain for frame extraction.
// For HDR content, includes tone mapping based on hardware accelerator.
func (g *Generator) buildFrameFilter(isHDR bool) string {
	if !isHDR {
		return fmt.Sprintf("scale=%d:%d", tileWidth, tileHeight)
	}

	switch g.hwAccel {
	case "vaapi":
		// VAAPI HDR: upload to GPU, tone map to SDR (NV12), then hardware scale.
		// Order: format -> hwupload -> tonemap_vaapi -> scale_vaapi
		return fmt.Sprintf("format=nv12,hwupload,tonemap_vaapi=format=nv12,scale_vaapi=w=%d:h=%d:format=nv12", tileWidth, tileHeight)
	case "nvenc", "cuda", "amf":
		// NVIDIA GPU HDR via OpenCL: zscale to linear BT.709, tonemap to SDR, then scale.
		return fmt.Sprintf("zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=opencl:format=nv12,scale=%d:%d", tileWidth, tileHeight)
	case "videotoolbox":
		// macOS VideoToolbox: no HW tonemap, use software zscale+hable.
		return fmt.Sprintf("zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable,zscale=t=bt709:m=bt709,format=yuv420p,scale=%d:%d", tileWidth, tileHeight)
	default:
		// Software HDR tonemap fallback using zscale + hable.
		return fmt.Sprintf("zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable,zscale=t=bt709:m=bt709,format=yuv420p,scale=%d:%d", tileWidth, tileHeight)
	}
}

// isHDRFile returns true if the input file uses HDR color transfer (PQ/SMPTE2084)
// or BT.2020 color primaries.
func isHDRFile(inputPath string) bool {
	return ffprobe.IsHDRLike(inputPath)
}

// createSpriteSheets combines extracted frames into sprite sheets using ffmpeg's
// concat demuxer + tile filter.
func (g *Generator) createSpriteSheets(mediaID int64, frames []string) error {
	if len(frames) == 0 {
		return fmt.Errorf("no frames provided")
	}

	dir := g.MediaDir(mediaID)
	spritePattern := filepath.Join(dir, "sprite_%d.jpg")

	// If we have exactly framesPerSprite frames (100), we can tile directly.
	if len(frames) >= framesPerSprite {
		if err := g.tileFrames(frames[:framesPerSprite], spritePattern); err != nil {
			return fmt.Errorf("tile frames: %w", err)
		}
	}

	// If we have more frames, create multiple sprite sheets.
	sheets := (len(frames) + framesPerSprite - 1) / framesPerSprite
	for sheet := 1; sheet < sheets; sheet++ {
		start := sheet * framesPerSprite
		end := start + framesPerSprite
		if start >= len(frames) {
			break
		}
		if end > len(frames) {
			end = len(frames)
		}
		sheetFrames := frames[start:end]
		outputPath := fmt.Sprintf("%s/sprite_%d.jpg", dir, sheet+1)
		if err := g.tileFrames(sheetFrames, outputPath); err != nil {
			return fmt.Errorf("tile sheet %d: %w", sheet+1, err)
		}
	}

	return nil
}

// tileFrames combines a list of frame paths into a single sprite sheet image using ffmpeg.
func (g *Generator) tileFrames(frames []string, outputPath string) error {
	if len(frames) == 0 {
		return nil
	}

	// Create a list file for ffmpeg concat demuxer.
	listFile, err := os.CreateTemp(os.TempDir(), "trickplay-frames-*.txt")
	if err != nil {
		return fmt.Errorf("create list file: %w", err)
	}
	listFilePath := listFile.Name()
	defer os.Remove(listFilePath)

	for _, frame := range frames {
		if _, err := listFile.WriteString(fmt.Sprintf("file '%s'\n", frame)); err != nil {
			return fmt.Errorf("write list file: %w", err)
		}
	}
	listFile.Close()

	// Calculate how many frames fit in the tile grid.
	cols := int(math.Ceil(math.Sqrt(float64(len(frames)))))
	if cols > tileColumns {
		cols = tileColumns
	}
	rows := (len(frames) + cols - 1) / cols

	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-f", "concat", "-safe", "0",
		"-i", listFilePath,
		"-vf", fmt.Sprintf("scale=%d:%d,tile=%dx%d", tileWidth, tileHeight, cols, rows),
		"-q:v", "5", "-y", outputPath,
	}

	cmd := exec.Command(ffmpegbin.FFmpeg(), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg tile: %w — %s", err, stderr.String())
	}
	return nil
}

// buildVTTFromFrames builds VTT using actual extracted frames (sorted by filename).
// Frame names are "frame_%05d.jpg" where the index corresponds to the original timestamp.
// Missing frames (extraction failures) are skipped, so timestamps may not be evenly spaced.
func (g *Generator) buildVTTFromFrames(mediaID int64, frames []string, interval int) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	for i, framePath := range frames {
		// Extract frame index from filename (e.g., "frame_00051.jpg" -> 51).
		filename := filepath.Base(framePath)
		var frameIdx int
		if _, err := fmt.Sscanf(filename, "frame_%d.jpg", &frameIdx); err != nil {
			log.Printf("trickplay: warning: could not parse frame index from %s: %v", filename, err)
			continue
		}

		startSec := frameIdx * interval
		endSec := startSec + interval

		// Sprite sheet index (1-based) and position within the sheet.
		spriteIndex := i/framesPerSprite + 1
		frameInSprite := i % framesPerSprite
		col := frameInSprite % tileColumns
		row := frameInSprite / tileColumns
		x := col * tileWidth
		y := row * tileHeight

		sb.WriteString(vttTimestamp(startSec) + " --> " + vttTimestamp(endSec) + "\n")
		fmt.Fprintf(&sb,
			"/api/media/%d/trickplay/sprite_%d.jpg#xywh=%d,%d,%d,%d\n\n",
			mediaID, spriteIndex, x, y, tileWidth, tileHeight)
	}

	return sb.String()
}

// vttTimestamp formats seconds as HH:MM:SS.000 for WebVTT cues.
func vttTimestamp(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d.000", h, m, s)
}
