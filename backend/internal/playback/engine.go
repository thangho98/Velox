package playback

import (
	"fmt"
	"strings"
)

// PlaybackMethod defines how to serve the media
type PlaybackMethod string

const (
	MethodDirectPlay     PlaybackMethod = "DirectPlay"
	MethodDirectStream   PlaybackMethod = "DirectStream" // Remux only
	MethodTranscodeAudio PlaybackMethod = "TranscodeAudio"
	MethodFullTranscode  PlaybackMethod = "FullTranscode"
)

// VideoAction defines what to do with video stream
type VideoAction string

const (
	VideoCopy      VideoAction = "Copy"
	VideoTranscode VideoAction = "Transcode"
)

// AudioAction defines what to do with audio stream
type AudioAction string

const (
	AudioCopy      AudioAction = "Copy"
	AudioTranscode AudioAction = "Transcode"
)

// SubtitleAction defines how to handle subtitles
type SubtitleAction string

const (
	SubtitleNone   SubtitleAction = "None"
	SubtitleCopy   SubtitleAction = "Copy"   // Text-based, served externally
	SubtitleBurnIn SubtitleAction = "BurnIn" // Image-based, requires transcode
)

// PlaybackDecision is the result of the decision engine
type PlaybackDecision struct {
	Method              PlaybackMethod `json:"method"`
	VideoAction         VideoAction    `json:"video_action"`
	AudioAction         AudioAction    `json:"audio_action"`
	SubtitleAction      SubtitleAction `json:"subtitle_action"`
	SubtitleStreamIndex int            `json:"subtitle_stream_index,omitempty"` // absolute stream index; set by caller when SubtitleBurnIn
	VideoCodec          string         `json:"video_codec,omitempty"`
	AudioCodec          string         `json:"audio_codec,omitempty"`
	Container           string         `json:"container,omitempty"`
	EstimatedBitrate    int            `json:"estimated_bitrate,omitempty"` // kbps
	Reason              string         `json:"reason"`                      // Why this decision was made
}

// MediaFileInfo represents media file metadata for decision making
type MediaFileInfo struct {
	ID           int    `json:"id"`
	Path         string `json:"path"`
	VideoCodec   string `json:"video_codec"`
	AudioCodec   string `json:"audio_codec"`
	Container    string `json:"container"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Duration     int    `json:"duration"` // seconds
	Bitrate      int    `json:"bitrate"`  // kbps
	HasSubtitles bool   `json:"has_subtitles"`
	SubType      string `json:"subtitle_type,omitempty"` // srt, vtt, pgs, etc.
}

// UserPreferences for playback
type UserPreferences struct {
	MaxStreamingQuality string `json:"max_streaming_quality"` // original, 4k, 1080p, 720p, 480p
	PreferDirectPlay    bool   `json:"prefer_direct_play"`
	SelectedAudioTrack  int    `json:"selected_audio_track"` // 0 = default
	SelectedSubtitle    string `json:"selected_subtitle"`    // language code or "off"
}

// Engine decides the best playback method
func Decide(media MediaFileInfo, profile *DeviceProfile, prefs UserPreferences) PlaybackDecision {
	decision := PlaybackDecision{
		VideoAction:    VideoCopy,
		AudioAction:    AudioCopy,
		SubtitleAction: SubtitleNone,
		Container:      ContainerMP4, // Default output container
	}

	// Normalize codec names
	videoCodec := normalizeCodec(media.VideoCodec)
	audioCodec := normalizeCodec(media.AudioCodec)
	container := normalizeContainer(media.Container)

	// Decide video path (resolution → codec → container → bitrate, in priority order)
	needsVideoTranscode := false

	maxHeight := parseQuality(prefs.MaxStreamingQuality)
	if maxHeight > 0 && media.Height > maxHeight {
		needsVideoTranscode = true
		decision.VideoAction = VideoTranscode
		decision.VideoCodec = selectBestVideoCodec(profile)
		decision.Method = MethodFullTranscode
		decision.Reason = fmt.Sprintf("Resolution %dp exceeds max %dp", media.Height, maxHeight)
		decision.EstimatedBitrate = estimateBitrate(maxHeight)
	} else if !profile.SupportsVideoCodec(videoCodec) {
		needsVideoTranscode = true
		decision.VideoAction = VideoTranscode
		decision.VideoCodec = selectBestVideoCodec(profile)
		decision.Method = MethodFullTranscode
		decision.Reason = fmt.Sprintf("Video codec %s not supported", videoCodec)
		decision.EstimatedBitrate = estimateBitrate(media.Height)
	} else if profile.MaxBitrate > 0 && media.Bitrate > profile.MaxBitrate {
		needsVideoTranscode = true
		decision.VideoAction = VideoTranscode
		decision.VideoCodec = selectBestVideoCodec(profile)
		decision.Method = MethodFullTranscode
		decision.Reason = fmt.Sprintf("Bitrate %d kbps exceeds max %d kbps", media.Bitrate, profile.MaxBitrate)
		decision.EstimatedBitrate = profile.MaxBitrate
	} else if !profile.SupportsContainer(container) {
		// Video codec OK but container needs remux
		decision.Method = MethodDirectStream
		decision.Reason = fmt.Sprintf("Container %s not supported, remuxing", container)
		decision.EstimatedBitrate = media.Bitrate
	} else {
		decision.Method = MethodDirectPlay
		decision.Reason = "Direct play compatible"
		decision.EstimatedBitrate = media.Bitrate
	}

	// Check if audio codec is supported
	if !profile.SupportsAudioCodec(audioCodec) {
		decision.AudioAction = AudioTranscode
		decision.AudioCodec = selectBestAudioCodec(profile)

		// Upgrade method only if video isn't already being transcoded
		if !needsVideoTranscode {
			if decision.Method == MethodDirectPlay {
				decision.Method = MethodTranscodeAudio
				decision.Reason = fmt.Sprintf("Audio codec %s not supported, transcoding audio", audioCodec)
			}
			// DirectStream + audio transcode still uses HLS pipeline → FullTranscode
			if decision.Method == MethodDirectStream {
				decision.Method = MethodFullTranscode
				decision.VideoAction = VideoTranscode
				decision.VideoCodec = selectBestVideoCodec(profile)
				decision.Reason += fmt.Sprintf(" + audio codec %s not supported", audioCodec)
			}
		}
	}

	// Check subtitles
	if media.HasSubtitles && prefs.SelectedSubtitle != "" && prefs.SelectedSubtitle != "off" {
		if media.SubType == SubtitlePGS || media.SubType == SubtitleVobSub {
			// Image-based subs require burn-in
			decision.SubtitleAction = SubtitleBurnIn
			// Force full transcode if not already
			if decision.Method == MethodDirectPlay || decision.Method == MethodDirectStream || decision.Method == MethodTranscodeAudio {
				decision.Method = MethodFullTranscode
				decision.VideoAction = VideoTranscode
				decision.VideoCodec = selectBestVideoCodec(profile)
				decision.EstimatedBitrate = estimateBitrate(media.Height)
				decision.Reason += " + image subtitles require burn-in"
			}
		} else {
			decision.SubtitleAction = SubtitleCopy
		}
	}

	return decision
}

// NormalizeSubtitleCodec normalizes FFprobe subtitle codec names to constants
func NormalizeSubtitleCodec(codec string) string {
	switch strings.ToLower(codec) {
	case "pgs", "hdmv_pgs_subtitle", "pgssub":
		return SubtitlePGS
	case "vobsub", "dvd_subtitle", "dvdsub":
		return SubtitleVobSub
	case "srt", "subrip":
		return SubtitleSRT
	case "ass", "ssa":
		return SubtitleASS
	case "vtt", "webvtt":
		return SubtitleVTT
	default:
		return strings.ToLower(codec)
	}
}

// normalizeCodec normalizes codec names
func normalizeCodec(codec string) string {
	codec = strings.ToLower(codec)
	switch codec {
	case "h264", "avc", "avc1":
		return CodecH264
	case "h265", "hevc", "hev1", "hvc1":
		return CodecH265
	case "vp09", "vp9":
		return CodecVP9
	case "av01", "av1":
		return CodecAV1
	case "aac", "mp4a":
		return CodecAAC
	case "opus":
		return CodecOpus
	case "flac":
		return CodecFLAC
	case "mp3", "mpg":
		return CodecMP3
	case "ac-3", "ac3":
		return CodecAC3
	case "ec-3", "eac3":
		return CodecEAC3
	case "dts":
		return CodecDTS
	default:
		return codec
	}
}

// normalizeContainer normalizes container names
func normalizeContainer(container string) string {
	container = strings.ToLower(container)
	switch container {
	case "mp4", "mpeg4", "m4v":
		return ContainerMP4
	case "webm":
		return ContainerWebM
	case "mkv", "matroska":
		return ContainerMKV
	case "mov", "qt":
		return ContainerMOV
	default:
		return container
	}
}

// selectBestVideoCodec chooses best codec for client
func selectBestVideoCodec(profile *DeviceProfile) string {
	if profile == nil {
		return CodecH264 // Safe fallback
	}

	// Prefer AV1 if supported (efficient)
	if profile.SupportsVideoCodec(CodecAV1) {
		return CodecAV1
	}
	// Then VP9
	if profile.SupportsVideoCodec(CodecVP9) {
		return CodecVP9
	}
	// HEVC for Apple devices
	if profile.SupportsVideoCodec(CodecH265) {
		return CodecH265
	}
	// Fallback to H.264 (universal)
	return CodecH264
}

// selectBestAudioCodec chooses best audio codec for client
func selectBestAudioCodec(profile *DeviceProfile) string {
	if profile == nil {
		return CodecAAC
	}

	// AAC is most compatible
	if profile.SupportsAudioCodec(CodecAAC) {
		return CodecAAC
	}
	// Opus for web browsers
	if profile.SupportsAudioCodec(CodecOpus) {
		return CodecOpus
	}
	// MP3 as fallback
	if profile.SupportsAudioCodec(CodecMP3) {
		return CodecMP3
	}
	return CodecAAC
}

// parseQuality converts quality string to max height
func parseQuality(quality string) int {
	switch strings.ToLower(quality) {
	case "4k", "2160p", "2160":
		return 2160
	case "1080p", "1080":
		return 1080
	case "720p", "720":
		return 720
	case "480p", "480":
		return 480
	case "original":
		return 0 // No limit
	default:
		return 0
	}
}

// estimateBitrate estimates bitrate for given resolution
func estimateBitrate(height int) int {
	switch {
	case height >= 2160:
		return 25000 // 4K
	case height >= 1080:
		return 8000 // 1080p
	case height >= 720:
		return 4000 // 720p
	case height >= 480:
		return 2000 // 480p
	default:
		return 1500
	}
}
