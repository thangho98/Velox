package ffprobe

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// ProbeResult contains basic media file metadata
type ProbeResult struct {
	Duration    float64
	Width       int
	Height      int
	VideoCodec  string
	AudioCodec  string
	Container   string
	Bitrate     int64
	HasSub      bool
	AudioTracks []AudioTrackInfo
	Subtitles   []SubtitleInfo
}

// AudioTrackInfo contains detailed audio track metadata
type AudioTrackInfo struct {
	StreamIndex   int
	Codec         string
	Language      string
	Channels      int
	ChannelLayout string
	Bitrate       int
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

// DetailedProbeResult contains full ffprobe output
type DetailedProbeResult struct {
	Format  FormatInfo   `json:"format"`
	Streams []StreamInfo `json:"streams"`
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
	Width         int         `json:"width"`
	Height        int         `json:"height"`
	Channels      int         `json:"channels"`
	ChannelLayout string      `json:"channel_layout"`
	BitRate       string      `json:"bit_rate"`
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

// Probe runs ffprobe on the given file and returns parsed metadata.
func Probe(path string) (*ProbeResult, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
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

			track := AudioTrackInfo{
				StreamIndex:   s.Index,
				Codec:         s.CodecName,
				Language:      s.Tags.Language,
				Channels:      s.Channels,
				ChannelLayout: s.ChannelLayout,
				Bitrate:       bitrate,
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
