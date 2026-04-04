package trickplay

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
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
	inProgress sync.Map // mediaID (int64) → struct{}; prevents duplicate concurrent generations
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
// Falls back to software decode if hardware acceleration fails to produce output.
func (g *Generator) Generate(mediaID int64, inputPath string, durationSec int) error {
	dir := g.MediaDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %s: %w", dir, err)
	}

	// One frame every interval seconds; at least one frame for very short files.
	totalFrames := durationSec / g.interval
	if totalFrames == 0 {
		totalFrames = 1
	}

	// ffmpeg tile filter produces one output image per sprite sheet of framesPerSprite tiles.
	// sprite_%d.jpg uses 1-based numbering by default with the tile filter.
	spritePattern := filepath.Join(dir, "sprite_%d.jpg")
	vfFilter := fmt.Sprintf(
		"fps=1/%d,scale=%d:%d,tile=%dx%d",
		g.interval, tileWidth, tileHeight, tileColumns, tileRows,
	)

	var lastErr error
	softwareFallback := false

	// Try hardware-accelerated decode first (if configured), then fall back to software.
	// Some HDR/DV profiles can't be decoded by VAAPI hardware.
	tryGenerate := func(hwAccel string) error {
		args := []string{"-hide_banner", "-loglevel", "error"}
		switch hwAccel {
		case "vaapi":
			args = append(args, "-hwaccel", "vaapi", "-hwaccel_device", "/dev/dri/renderD128")
		case "videotoolbox":
			args = append(args, "-hwaccel", "videotoolbox")
		case "nvenc":
			args = append(args, "-hwaccel", "cuda")
		case "qsv":
			args = append(args, "-hwaccel", "qsv")
		case "amf":
			args = append(args, "-hwaccel", "amf")
		}
		args = append(args, "-i", inputPath, "-vf", vfFilter, "-q:v", "5", "-y", spritePattern)

		cmd := exec.Command("ffmpeg", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("ffmpeg sprite generation: %w — %s", err, stderr.String())
		}
		return nil
	}

	// Try with hardware acceleration first (if available)
	if g.hwAccel != "" {
		lastErr = tryGenerate(g.hwAccel)
		if lastErr == nil {
			// Verify at least one sprite was produced
			if _, err := os.Stat(g.SpritePath(mediaID, 1)); err != nil {
				log.Printf("trickplay: hwaccel %s produced no sprites for media %d, falling back to software",
					g.hwAccel, mediaID)
				lastErr = err
				softwareFallback = true
			}
		} else {
			log.Printf("trickplay: hwaccel %s failed for media %d: %v, trying software decode",
				g.hwAccel, mediaID, lastErr)
			softwareFallback = true
		}
	}

	// Fall back to software decode if hw accel failed or produced no output
	if softwareFallback || g.hwAccel == "" {
		// Clean up any partial output from failed hwaccel attempt
		if softwareFallback {
			for i := 1; ; i++ {
				if os.Remove(g.SpritePath(mediaID, i)) != nil {
					break
				}
			}
		}
		lastErr = tryGenerate("")
	}

	if lastErr != nil {
		return fmt.Errorf("software decode fallback: %w", lastErr)
	}

	// Verify at least one sprite was produced
	if _, err := os.Stat(g.SpritePath(mediaID, 1)); err != nil {
		return fmt.Errorf("no sprites generated: ffmpeg produced no output files")
	}

	vtt := g.buildVTT(mediaID, totalFrames)
	if err := os.WriteFile(g.VTTPath(mediaID), []byte(vtt), 0644); err != nil {
		return fmt.Errorf("write vtt: %w", err)
	}

	spriteCount := int(math.Ceil(float64(totalFrames) / float64(framesPerSprite)))
	accelStr := ""
	if g.hwAccel != "" && !softwareFallback {
		accelStr = fmt.Sprintf(" (hw: %s)", g.hwAccel)
	} else if softwareFallback {
		accelStr = " (software fallback)"
	}
	log.Printf("trickplay: generated %d frames for media %d (%d sprites)%s",
		totalFrames, mediaID, spriteCount, accelStr)
	return nil
}

// buildVTT returns the WebVTT manifest mapping timestamps to sprite coordinates.
// Sprite files are referenced by their API URL (/api/media/{id}/trickplay/sprite_N.jpg).
func (g *Generator) buildVTT(mediaID int64, totalFrames int) string {
	var sb strings.Builder
	sb.WriteString("WEBVTT\n\n")

	for i := 0; i < totalFrames; i++ {
		startSec := i * g.interval
		endSec := startSec + g.interval

		// Sprite sheet index (1-based) and position within the sheet.
		spriteIndex := i/framesPerSprite + 1
		frameInSprite := i % framesPerSprite
		col := frameInSprite % tileColumns
		row := frameInSprite / tileColumns
		x := col * tileWidth
		y := row * tileHeight

		sb.WriteString(vttTimestamp(startSec) + " --> " + vttTimestamp(endSec) + "\n")
		sb.WriteString(fmt.Sprintf(
			"/api/media/%d/trickplay/sprite_%d.jpg#xywh=%d,%d,%d,%d\n\n",
			mediaID, spriteIndex, x, y, tileWidth, tileHeight,
		))
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
