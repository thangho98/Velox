// Package subtitle provides subtitle extraction and conversion utilities.
package subtitle

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// ExtractSubtitle extracts an embedded subtitle stream from a video file and
// writes it as a WebVTT file. Results are cached on disk: if the output file
// already exists it is returned immediately without re-running FFmpeg.
//
// videoPath:   absolute path to the source video file
// streamIndex: absolute FFprobe stream index (e.g. 2 for the third stream in
//
//	the container — same value stored in model.Subtitle.StreamIndex)
//
// outputDir:   directory where the .vtt file will be written (created if needed)
//
// Returns the path to the extracted .vtt file.
func ExtractSubtitle(videoPath string, streamIndex int, outputDir string) (string, error) {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("creating subtitle cache dir: %w", err)
	}

	outputPath := filepath.Join(outputDir, strconv.Itoa(streamIndex)+".vtt")

	// Cache hit: skip FFmpeg if file already exists.
	if _, err := os.Stat(outputPath); err == nil {
		return outputPath, nil
	}

	// Use absolute stream index (-map 0:N) — consistent with how AudioTrack and
	// Subtitle rows are populated by the scanner (absolute FFprobe stream index).
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-map", fmt.Sprintf("0:%d", streamIndex),
		"-c:s", "webvtt",
		"-y",
		outputPath,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Remove any partial file so a retry can succeed.
		os.Remove(outputPath)
		return "", fmt.Errorf("extracting subtitle stream %d from %q: %w", streamIndex, videoPath, err)
	}

	return outputPath, nil
}
