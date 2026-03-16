package ffprobe

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// ProbeResult contains basic media file metadata
type ProbeResult struct {
	Duration     float64
	Width        int
	Height       int
	VideoCodec   string
	VideoProfile string  // e.g. "High", "Main 10"
	VideoLevel   int     // e.g. 40, 51
	VideoFPS     float64 // frames per second from r_frame_rate
	AudioCodec   string
	Container    string
	Bitrate      int64
	HasSub       bool
	AudioTracks  []AudioTrackInfo
	Subtitles    []SubtitleInfo
	Chapters     []ChapterInfo // NEW: chapter markers for intro/credits skip
}

// AudioTrackInfo contains detailed audio track metadata
type AudioTrackInfo struct {
	StreamIndex   int
	Codec         string
	Language      string
	Channels      int
	ChannelLayout string
	Bitrate       int
	SampleRate    int
	Title         string
	IsDefault     bool
}

// SubtitleInfo contains subtitle track metadata
type SubtitleInfo struct {
	StreamIndex int
	Codec       string
	Language    string
	Title       string
	IsForced    bool
	IsDefault   bool
	IsSDH       bool // Hearing impaired
}

// ChapterInfo contains chapter metadata from ffprobe
type ChapterInfo struct {
	ID        int
	StartTime float64 // seconds
	EndTime   float64 // seconds
	Title     string
}

// DetailedProbeResult contains full ffprobe output
type DetailedProbeResult struct {
	Format   FormatInfo   `json:"format"`
	Streams  []StreamInfo `json:"streams"`
	Chapters []Chapter    `json:"chapters"`
}

// FormatInfo from ffprobe
type FormatInfo struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
	Name     string `json:"format_name"`
}

// StreamInfo from ffprobe
type StreamInfo struct {
	Index         int         `json:"index"`
	CodecType     string      `json:"codec_type"`
	CodecName     string      `json:"codec_name"`
	Profile       string      `json:"profile"`
	Level         int         `json:"level"`
	Width         int         `json:"width"`
	Height        int         `json:"height"`
	RFrameRate    string      `json:"r_frame_rate"`
	AvgFrameRate  string      `json:"avg_frame_rate"`
	Channels      int         `json:"channels"`
	ChannelLayout string      `json:"channel_layout"`
	BitRate       string      `json:"bit_rate"`
	SampleRate    string      `json:"sample_rate"`
	Tags          StreamTags  `json:"tags"`
	Disposition   Disposition `json:"disposition"`
}

// StreamTags from ffprobe
type StreamTags struct {
	Language    string `json:"language"`
	Title       string `json:"title"`
	HandlerName string `json:"handler_name"`
}

// Disposition from ffprobe (track flags)
type Disposition struct {
	Default         int `json:"default"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
}

// Chapter from ffprobe chapters output
type Chapter struct {
	ID        int         `json:"id"`
	TimeBase  string      `json:"time_base"`
	Start     int64       `json:"start"`
	End       int64       `json:"end"`
	StartTime string      `json:"start_time"`
	EndTime   string      `json:"end_time"`
	Tags      ChapterTags `json:"tags"`
}

// ChapterTags contains chapter metadata
type ChapterTags struct {
	Title string `json:"title"`
}

// Probe runs ffprobe on the given file and returns parsed metadata.
func Probe(path string) (*ProbeResult, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-show_chapters", // NEW: extract chapter markers
		path,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe error: %w", err)
	}

	var detailed DetailedProbeResult
	if err := json.Unmarshal(out, &detailed); err != nil {
		return nil, fmt.Errorf("parse ffprobe output: %w", err)
	}

	r := &ProbeResult{
		Container: detailed.Format.Name,
		Chapters:  parseChapters(detailed.Chapters), // NEW: parse chapters
	}

	if v, err := strconv.ParseFloat(detailed.Format.Duration, 64); err != nil {
		if detailed.Format.Duration != "" {
			log.Printf("ffprobe: unparseable duration %q for %s", detailed.Format.Duration, path)
		}
	} else {
		r.Duration = v
	}

	if v, err := strconv.ParseInt(detailed.Format.BitRate, 10, 64); err != nil {
		if detailed.Format.BitRate != "" {
			log.Printf("ffprobe: unparseable bitrate %q for %s", detailed.Format.BitRate, path)
		}
	} else {
		r.Bitrate = v
	}

	var firstAudioCodec string

	for _, s := range detailed.Streams {
		switch s.CodecType {
		case "video":
			if r.VideoCodec == "" {
				r.VideoCodec = s.CodecName
				r.Width = s.Width
				r.Height = s.Height
				r.VideoProfile = s.Profile
				r.VideoLevel = s.Level
				r.VideoFPS = parseFrameRate(s.RFrameRate)
				if r.VideoFPS == 0 {
					r.VideoFPS = parseFrameRate(s.AvgFrameRate)
				}
			}
		case "audio":
			if firstAudioCodec == "" {
				firstAudioCodec = s.CodecName
			}

			isDefault := s.Disposition.Default == 1
			if isDefault && r.AudioCodec == "" {
				r.AudioCodec = s.CodecName
			}

			bitrate, err := strconv.Atoi(s.BitRate)
			if err != nil && s.BitRate != "" {
				log.Printf("ffprobe: unparseable audio bitrate %q (stream %d) for %s", s.BitRate, s.Index, path)
			}

			title := s.Tags.Title
			if title == "" {
				title = s.Tags.HandlerName
			}

			sampleRate, _ := strconv.Atoi(s.SampleRate)

			track := AudioTrackInfo{
				StreamIndex:   s.Index,
				Codec:         s.CodecName,
				Language:      s.Tags.Language,
				Channels:      s.Channels,
				ChannelLayout: s.ChannelLayout,
				Bitrate:       bitrate,
				SampleRate:    sampleRate,
				Title:         title,
				IsDefault:     isDefault,
			}
			r.AudioTracks = append(r.AudioTracks, track)

		case "subtitle":
			r.HasSub = true

			isDefault := s.Disposition.Default == 1
			isForced := s.Disposition.Forced == 1
			isSDH := s.Disposition.HearingImpaired == 1

			subTitle := s.Tags.Title
			if subTitle == "" {
				subTitle = s.Tags.HandlerName
			}

			sub := SubtitleInfo{
				StreamIndex: s.Index,
				Codec:       s.CodecName,
				Language:    s.Tags.Language,
				Title:       subTitle,
				IsForced:    isForced,
				IsDefault:   isDefault,
				IsSDH:       isSDH,
			}
			r.Subtitles = append(r.Subtitles, sub)
		}
	}

	// Fallback: if no audio stream had the default disposition, use the first one
	if r.AudioCodec == "" && firstAudioCodec != "" {
		r.AudioCodec = firstAudioCodec
	}

	return r, nil
}

// IsTextBasedSubtitle returns true if the codec is text-based (can be extracted to VTT)
func IsTextBasedSubtitle(codec string) bool {
	textCodecs := []string{"subrip", "ass", "ssa", "webvtt", "mov_text"}
	for _, c := range textCodecs {
		if codec == c {
			return true
		}
	}
	return false
}

// parseFrameRate parses ffprobe frame rate strings like "24000/1001" or "24" into float64.
func parseFrameRate(rate string) float64 {
	if rate == "" || rate == "0/0" {
		return 0
	}
	parts := strings.SplitN(rate, "/", 2)
	if len(parts) == 2 {
		num, err1 := strconv.ParseFloat(parts[0], 64)
		den, err2 := strconv.ParseFloat(parts[1], 64)
		if err1 == nil && err2 == nil && den > 0 {
			return num / den
		}
	}
	v, err := strconv.ParseFloat(rate, 64)
	if err == nil {
		return v
	}
	return 0
}

// IsImageBasedSubtitle returns true if the codec is image-based (requires burn-in)
func IsImageBasedSubtitle(codec string) bool {
	imageCodecs := []string{"hdmv_pgs_subtitle", "dvd_subtitle", "dvb_subtitle", "xsub"}
	for _, c := range imageCodecs {
		if codec == c {
			return true
		}
	}
	return false
}

// parseChapters converts ffprobe chapter data to ChapterInfo
func parseChapters(chapters []Chapter) []ChapterInfo {
	if len(chapters) == 0 {
		return nil
	}

	result := make([]ChapterInfo, 0, len(chapters))
	for _, c := range chapters {
		info := ChapterInfo{
			ID:    c.ID,
			Title: c.Tags.Title,
		}

		// Parse start_time and end_time (in seconds)
		if c.StartTime != "" {
			info.StartTime, _ = strconv.ParseFloat(c.StartTime, 64)
		} else if c.TimeBase != "" && c.Start > 0 {
			// Fallback: calculate from time_base and start ticks
			info.StartTime = parseTimeBase(c.TimeBase, c.Start)
		}
		if c.EndTime != "" {
			info.EndTime, _ = strconv.ParseFloat(c.EndTime, 64)
		} else if c.TimeBase != "" && c.End > 0 {
			info.EndTime = parseTimeBase(c.TimeBase, c.End)
		}

		result = append(result, info)
	}
	return result
}

// parseTimeBase converts time base (e.g., "1/1000") and ticks to seconds
func parseTimeBase(timeBase string, ticks int64) float64 {
	parts := strings.SplitN(timeBase, "/", 2)
	if len(parts) != 2 {
		return 0
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return 0
	}
	return float64(ticks) * num / den
}
