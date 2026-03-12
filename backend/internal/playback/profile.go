// Package playback provides playback decision engine and client capability profiles.
// It determines the optimal playback method (DirectPlay, DirectStream, Transcode)
// based on media file properties and client device capabilities.
package playback

// DeviceProfile defines what a client device can play
type DeviceProfile struct {
	Name                     string   `json:"name"`
	SupportedVideoCodecs     []string `json:"supported_video_codecs"`
	SupportedAudioCodecs     []string `json:"supported_audio_codecs"`
	SupportedContainers      []string `json:"supported_containers"`
	SupportedSubtitleFormats []string `json:"supported_subtitle_formats"`
	MaxWidth                 int      `json:"max_width"`          // 0 = unlimited
	MaxHeight                int      `json:"max_height"`         // e.g., 1080, 2160
	MaxBitrate               int      `json:"max_bitrate"`        // kbps, 0 = unlimited
	CanBurnSubtitles         bool     `json:"can_burn_subtitles"` // Server must burn subs
	SupportsHLS              bool     `json:"supports_hls"`
	SupportsWebM             bool     `json:"supports_webm"`
}

// Common codec constants
const (
	CodecH264   = "h264"
	CodecH265   = "hevc"
	CodecVP9    = "vp9"
	CodecVP8    = "vp8"
	CodecAV1    = "av1"
	CodecAAC    = "aac"
	CodecOpus   = "opus"
	CodecMP3    = "mp3"
	CodecFLAC   = "flac"
	CodecAC3    = "ac3"
	CodecEAC3   = "eac3"
	CodecDTS    = "dts"
	CodecTrueHD = "truehd"
)

// Common container constants
const (
	ContainerMP4  = "mp4"
	ContainerWebM = "webm"
	ContainerMKV  = "mkv"
	ContainerHLS  = "hls"
	ContainerMOV  = "mov"
)

// Subtitle format constants
const (
	SubtitleVTT    = "vtt"
	SubtitleSRT    = "srt"
	SubtitleASS    = "ass"
	SubtitlePGS    = "pgs"
	SubtitleVobSub = "vobsub"
)

// SupportsVideoCodec checks if codec is supported
func (p *DeviceProfile) SupportsVideoCodec(codec string) bool {
	for _, c := range p.SupportedVideoCodecs {
		if c == codec {
			return true
		}
	}
	return false
}

// SupportsAudioCodec checks if codec is supported
func (p *DeviceProfile) SupportsAudioCodec(codec string) bool {
	for _, c := range p.SupportedAudioCodecs {
		if c == codec {
			return true
		}
	}
	return false
}

// SupportsContainer checks if container is supported
func (p *DeviceProfile) SupportsContainer(container string) bool {
	for _, c := range p.SupportedContainers {
		if c == container {
			return true
		}
	}
	return false
}

// SupportsSubtitleFormat checks if subtitle format is supported (for text-based)
func (p *DeviceProfile) SupportsSubtitleFormat(format string) bool {
	for _, f := range p.SupportedSubtitleFormats {
		if f == format {
			return true
		}
	}
	return false
}

// CanPlayResolution checks if resolution is within limits
func (p *DeviceProfile) CanPlayResolution(width, height int) bool {
	if p.MaxWidth > 0 && width > p.MaxWidth {
		return false
	}
	if p.MaxHeight > 0 && height > p.MaxHeight {
		return false
	}
	return true
}

// CanPlayBitrate checks if bitrate is within limits
func (p *DeviceProfile) CanPlayBitrate(bitrate int) bool {
	if p.MaxBitrate > 0 && bitrate > p.MaxBitrate {
		return false
	}
	return true
}
