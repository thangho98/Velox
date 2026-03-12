package transcoder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/thawng/velox/internal/model"
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

// GenerateHLS transcodes a video file into HLS segments.
// fileID: the actual media_files.id being transcoded; used as part of the cache key so
// different file versions of the same media don't collide.
// subtitleStreamIndex: if >= 0, burn-in the subtitle at that absolute stream index.
func (t *Transcoder) GenerateHLS(mediaID int64, inputPath string, fileID int64, subtitleStreamIndex int) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	// Check if already transcoded
	if _, err := os.Stat(masterPath); err == nil {
		return nil
	}

	args := []string{
		"-i", inputPath,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "22",
	}
	if subtitleStreamIndex >= 0 {
		escaped := strings.ReplaceAll(inputPath, ":", "\\:")
		args = append(args, "-vf", fmt.Sprintf("subtitles='%s':si=%d", escaped, subtitleStreamIndex))
	}
	args = append(args,
		"-c:a", "aac",
		"-b:a", "128k",
		"-ac", "2",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(dir, prefix+"seg_%04d.ts"),
		masterPath,
	)

	cmd := exec.Command("ffmpeg", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// AudioVariant represents an audio track variant for HLS
type AudioVariant struct {
	Language    string
	Name        string
	StreamIndex int
	IsDefault   bool
}

// GenerateHLSWithAudio generates HLS with multiple audio tracks using #EXT-X-MEDIA.
// Creates separate audio playlists and a master playlist that references them.
// fileID: the actual media_files.id being transcoded; part of the cache key.
// subtitleStreamIndex: if >= 0, burn-in the subtitle at that absolute stream index into the video.
func (t *Transcoder) GenerateHLSWithAudio(mediaID int64, inputPath string, audioTracks []model.AudioTrack, fileID int64, subtitleStreamIndex int) error {
	dir := t.HLSDir(mediaID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	prefix := hlsPrefix(fileID, subtitleStreamIndex)
	masterPath := filepath.Join(dir, prefix+"master.m3u8")

	// Check if already transcoded
	if _, err := os.Stat(masterPath); err == nil {
		return nil
	}

	// If no multiple audio tracks, fall back to simple HLS
	if len(audioTracks) <= 1 {
		return t.GenerateHLS(mediaID, inputPath, fileID, subtitleStreamIndex)
	}

	// Build variants for HLS master playlist
	variants := make([]AudioVariant, 0, len(audioTracks))
	for _, track := range audioTracks {
		name := track.Title
		if name == "" {
			name = track.Language
			if name == "" {
				name = fmt.Sprintf("Audio %d", track.StreamIndex)
			}
		}
		// B3: Strip double quotes — they would break the M3U8 quoted-string format.
		name = strings.ReplaceAll(name, `"`, `'`)
		variants = append(variants, AudioVariant{
			Language:    track.Language,
			Name:        name,
			StreamIndex: track.StreamIndex,
			IsDefault:   track.IsDefault,
		})
	}

	// Generate separate audio playlists for each track
	audioPlaylistPaths := make(map[int]string) // stream index -> playlist path
	for _, variant := range variants {
		audioPlaylist := filepath.Join(dir, fmt.Sprintf("audio_%d.m3u8", variant.StreamIndex))
		audioPlaylistPaths[variant.StreamIndex] = audioPlaylist

		// Transcode audio-only stream.
		// B1: Use absolute stream index (0:N) not relative audio index (0:a:N).
		// AudioTrack.StreamIndex is the absolute FFprobe stream index (e.g. 1 for the
		// first audio stream in a file where stream 0 is video).
		cmd := exec.Command("ffmpeg",
			"-i", inputPath,
			"-map", fmt.Sprintf("0:%d", variant.StreamIndex),
			"-c:a", "aac",
			"-b:a", "128k",
			"-ac", "2",
			"-f", "hls",
			"-hls_time", "6",
			"-hls_list_size", "0",
			"-hls_segment_filename", filepath.Join(dir, fmt.Sprintf("audio_%d_%%04d.ts", variant.StreamIndex)),
			audioPlaylist,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to transcode audio stream %d: %w", variant.StreamIndex, err)
		}
	}

	// Generate video-only playlist (no audio), optionally with subtitle burn-in
	videoPlaylist := filepath.Join(dir, prefix+"video.m3u8")
	videoArgs := []string{
		"-i", inputPath,
		"-map", "0:v:0", // First video stream only
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "22",
	}
	if subtitleStreamIndex >= 0 {
		escaped := strings.ReplaceAll(inputPath, ":", "\\:")
		videoArgs = append(videoArgs, "-vf", fmt.Sprintf("subtitles='%s':si=%d", escaped, subtitleStreamIndex))
	}
	videoArgs = append(videoArgs,
		"-an", // No audio
		"-f", "hls",
		"-hls_time", "6",
		"-hls_list_size", "0",
		"-hls_segment_filename", filepath.Join(dir, prefix+"video_%04d.ts"),
		videoPlaylist,
	)
	cmd := exec.Command("ffmpeg", videoArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to transcode video: %w", err)
	}

	// Build master playlist with #EXT-X-MEDIA
	return t.writeMasterPlaylistWithAudio(masterPath, variants, audioPlaylistPaths, prefix)
}

// writeMasterPlaylistWithAudio creates the master playlist with #EXT-X-MEDIA tags
func (t *Transcoder) writeMasterPlaylistWithAudio(masterPath string, variants []AudioVariant, audioPaths map[int]string, prefix string) error {
	var sb strings.Builder
	sb.WriteString("#EXTM3U\n")
	sb.WriteString("#EXT-X-VERSION:4\n")

	// Write #EXT-X-MEDIA entries for each audio track
	groupID := "audio"
	for _, v := range variants {
		// B4: AUTOSELECT must be YES when DEFAULT=YES (RFC 8216 §4.3.4.1).
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

	// Write video stream info with audio group.
	// B2: 4 Mbps is a realistic estimate for CRF=22 at typical 1080p content.
	sb.WriteString(fmt.Sprintf(
		"#EXT-X-STREAM-INF:BANDWIDTH=4000000,AUDIO=\"%s\"\n",
		groupID,
	))
	sb.WriteString(prefix + "video.m3u8\n")

	return os.WriteFile(masterPath, []byte(sb.String()), 0644)
}

// SegmentPath returns the full path to an HLS segment file.
func (t *Transcoder) SegmentPath(mediaID int64, segment string) string {
	return filepath.Join(t.HLSDir(mediaID), segment)
}

// RemuxToWriter remuxes a video file to fragmented MP4 and writes to w.
// Used for DirectStream: container is incompatible but codecs are compatible.
// No codec transcoding occurs — this is a fast container-only operation.
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
