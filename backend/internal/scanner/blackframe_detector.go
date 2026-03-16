package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os/exec"
	"regexp"
	"sort"
	"strconv"

	"github.com/thawng/velox/internal/repository"
)

// SilenceRange represents a detected silence period.
type SilenceRange struct {
	Start float64
	End   float64
}

// BlackFrameDetector uses FFmpeg blackframe + silencedetect filters to detect credits.
// Only detects credits (not intro). Acts as fallback after chromaprint.
type BlackFrameDetector struct {
	markerRepo    *repository.MediaMarkerRepo
	mediaFileRepo *repository.MediaFileRepo
}

// NewBlackFrameDetector creates a new black frame + silence detector.
func NewBlackFrameDetector(markerRepo *repository.MediaMarkerRepo, mediaFileRepo *repository.MediaFileRepo) *BlackFrameDetector {
	return &BlackFrameDetector{
		markerRepo:    markerRepo,
		mediaFileRepo: mediaFileRepo,
	}
}

func (d *BlackFrameDetector) Name() string        { return "blackframe" }
func (d *BlackFrameDetector) Confidence() float64 { return 0.65 }

// Detect finds credits boundaries using black frame + silence analysis.
func (d *BlackFrameDetector) Detect(ctx context.Context, fileID int64, filePath string) ([]DetectedMarker, error) {
	// Skip if credits marker already exists from higher priority source
	existing, err := d.markerRepo.GetByMediaFileID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("checking existing markers: %w", err)
	}
	for _, m := range existing {
		if m.MarkerType == "credits" && (m.Source == "manual" || m.Source == "chapter") {
			slog.Debug("blackframe: skipping, higher priority credits exist", "file_id", fileID)
			return nil, nil
		}
	}

	// Get file duration
	file, err := d.mediaFileRepo.GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("getting media file: %w", err)
	}

	if file.Duration < 120 {
		return nil, nil // Too short to have meaningful credits
	}

	// Analyze last 10 minutes
	searchDuration := math.Min(600, file.Duration)
	searchStart := math.Max(0, file.Duration-searchDuration)

	// Run both analyses
	blackFrames, err := detectBlackFrames(ctx, filePath, searchStart, searchDuration)
	if err != nil {
		slog.Warn("blackframe: detection failed", "file_id", fileID, "error", err)
		blackFrames = nil
	}

	silences, err := detectSilence(ctx, filePath, searchStart, searchDuration)
	if err != nil {
		slog.Warn("blackframe: silence detection failed", "file_id", fileID, "error", err)
		silences = nil
	}

	// Find credits boundary
	creditsStart, confidence := findCreditsBoundary(blackFrames, silences, searchStart, file.Duration)
	if creditsStart <= 0 {
		slog.Debug("blackframe: no credits boundary found", "file_id", fileID)
		return nil, nil
	}

	// Validate: credits should be at least 15s and at most 450s
	creditsDur := file.Duration - creditsStart
	if creditsDur < 15 || creditsDur > 450 {
		return nil, nil
	}

	slog.Info("blackframe: detected credits",
		"file_id", fileID,
		"start", creditsStart,
		"end", file.Duration,
		"confidence", confidence,
	)

	return []DetectedMarker{
		{
			Type:       "credits",
			StartSec:   creditsStart,
			EndSec:     file.Duration,
			Source:     "fingerprint", // Reuse existing source to avoid schema change
			Confidence: confidence,
			Label:      "blackframe+silence",
		},
	}, nil
}

// detectBlackFrames runs FFmpeg blackframe filter and returns timestamps of black frames.
func detectBlackFrames(ctx context.Context, filePath string, startSec, durationSec float64) ([]float64, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", strconv.FormatFloat(startSec, 'f', 2, 64),
		"-i", filePath,
		"-t", strconv.FormatFloat(durationSec, 'f', 2, 64),
		"-an", "-dn", "-sn",
		"-vf", "blackframe=amount=50:threshold=32",
		"-f", "null", "-",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		// FFmpeg returns non-zero on some edge cases, try to parse anyway
		if len(out) == 0 {
			return nil, fmt.Errorf("ffmpeg blackframe: %w", err)
		}
	}

	return parseBlackFrameOutput(string(out)), nil
}

// detectSilence runs FFmpeg silencedetect filter and returns silence ranges.
func detectSilence(ctx context.Context, filePath string, startSec, durationSec float64) ([]SilenceRange, error) {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", strconv.FormatFloat(startSec, 'f', 2, 64),
		"-i", filePath,
		"-t", strconv.FormatFloat(durationSec, 'f', 2, 64),
		"-vn", "-dn", "-sn",
		"-af", "silencedetect=noise=-50dB:duration=0.5",
		"-f", "null", "-",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		if len(out) == 0 {
			return nil, fmt.Errorf("ffmpeg silencedetect: %w", err)
		}
	}

	return parseSilenceDetectOutput(string(out)), nil
}

var blackFrameRe = regexp.MustCompile(`\[blackframe[^\]]*\].*t:\s*([\d.]+)`)
var silenceStartRe = regexp.MustCompile(`silence_start:\s*([\d.]+)`)
var silenceEndRe = regexp.MustCompile(`silence_end:\s*([\d.]+)`)

// parseBlackFrameOutput extracts timestamps from FFmpeg blackframe filter output.
func parseBlackFrameOutput(output string) []float64 {
	var timestamps []float64
	for _, match := range blackFrameRe.FindAllStringSubmatch(output, -1) {
		if t, err := strconv.ParseFloat(match[1], 64); err == nil {
			timestamps = append(timestamps, t)
		}
	}
	return timestamps
}

// parseSilenceDetectOutput extracts silence ranges from FFmpeg silencedetect output.
func parseSilenceDetectOutput(output string) []SilenceRange {
	starts := silenceStartRe.FindAllStringSubmatch(output, -1)
	ends := silenceEndRe.FindAllStringSubmatch(output, -1)

	var ranges []SilenceRange
	for i := 0; i < len(starts) && i < len(ends); i++ {
		s, err1 := strconv.ParseFloat(starts[i][1], 64)
		e, err2 := strconv.ParseFloat(ends[i][1], 64)
		if err1 == nil && err2 == nil && e > s {
			ranges = append(ranges, SilenceRange{Start: s, End: e})
		}
	}

	// Handle trailing silence_start without matching end
	if len(starts) > len(ends) {
		s, err := strconv.ParseFloat(starts[len(starts)-1][1], 64)
		if err == nil {
			ranges = append(ranges, SilenceRange{Start: s, End: s + 1})
		}
	}

	return ranges
}

// findCreditsBoundary finds the credits start point using combined black frame + silence heuristic.
// Returns (absoluteTimestamp, confidence) or (0, 0) if not found.
func findCreditsBoundary(blackFrames []float64, silences []SilenceRange, regionStart, totalDuration float64) (float64, float64) {
	const tolerance = 5.0 // seconds — black frame and silence must co-occur within this window

	if len(blackFrames) == 0 && len(silences) == 0 {
		return 0, 0
	}

	// Sort black frames chronologically
	sort.Float64s(blackFrames)

	// Strategy 1: Find co-occurring black frame cluster + silence (high confidence)
	if len(blackFrames) > 0 && len(silences) > 0 {
		// Find clusters of 3+ black frames within a 10-second window
		clusters := findBlackFrameClusters(blackFrames, 10, 3)

		for _, clusterTime := range clusters {
			absTime := regionStart + clusterTime
			// Check if any silence overlaps within tolerance
			for _, s := range silences {
				absSilenceStart := regionStart + s.Start
				if math.Abs(absTime-absSilenceStart) <= tolerance {
					// Must be in last 30% of the file
					if absTime > totalDuration*0.7 {
						return absTime, 0.75
					}
				}
			}
		}
	}

	// Strategy 2: Black frame cluster alone (lower confidence)
	if len(blackFrames) > 0 {
		clusters := findBlackFrameClusters(blackFrames, 10, 3)
		for _, clusterTime := range clusters {
			absTime := regionStart + clusterTime
			if absTime > totalDuration*0.7 {
				return absTime, 0.55
			}
		}
	}

	return 0, 0
}

// findBlackFrameClusters finds timestamps where N+ black frames occur within windowSec.
func findBlackFrameClusters(frames []float64, windowSec float64, minCount int) []float64 {
	var clusters []float64
	for i := 0; i <= len(frames)-minCount; i++ {
		windowEnd := frames[i] + windowSec
		count := 0
		for j := i; j < len(frames) && frames[j] <= windowEnd; j++ {
			count++
		}
		if count >= minCount {
			clusters = append(clusters, frames[i])
			// Skip past this cluster
			for i+1 < len(frames) && frames[i+1] <= windowEnd {
				i++
			}
		}
	}
	return clusters
}
