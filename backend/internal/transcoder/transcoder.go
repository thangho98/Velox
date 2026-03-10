package transcoder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Transcoder struct {
	outputDir string
}

func New(outputDir string) *Transcoder {
	return &Transcoder{outputDir: outputDir}
}

// HLSDir returns the directory where HLS segments for a media item are stored.
func (t *Transcoder) HLSDir(mediaID int64) string {
	return filepath.Join(t.outputDir, fmt.Sprintf("%d", mediaID))
}

// GenerateHLS transcodes a video file into HLS segments.
func (t *Transcoder) GenerateHLS(mediaID int64, inputPath string) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	masterPath := filepath.Join(dir, "master.m3u8")

	// Check if already transcoded
	if _, err := os.Stat(masterPath); err == nil {
		return nil
	}

	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "22",
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(dir, "seg_%04d.ts"),
		masterPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SegmentPath returns the full path to an HLS segment file.
func (t *Transcoder) SegmentPath(mediaID int64, segment string) string {
	return filepath.Join(t.HLSDir(mediaID), segment)
}

// Clean removes transcoded files for a media item.
func (t *Transcoder) Clean(mediaID int64) error {
	return os.RemoveAll(t.HLSDir(mediaID))
}
