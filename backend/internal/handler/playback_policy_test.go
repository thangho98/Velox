package handler

import (
	"strings"
	"testing"

	"github.com/thawng/velox/internal/playback"
)

func TestApplyAdminPlaybackPolicyUsesTranscodeAudioForBrowserAudioMismatch(t *testing.T) {
	// Jellyfin-style: audio-only transcode even on browser/HLS when video is compatible.
	// Does NOT force full transcode just because of audio codec mismatch.
	profile := &playback.DeviceProfile{
		Name:                 "Chrome",
		SupportedVideoCodecs: []string{playback.CodecH264},
		SupportedAudioCodecs: []string{playback.CodecAAC},
		SupportedContainers:  []string{playback.ContainerMP4, playback.ContainerHLS},
		SupportsHLS:          true,
	}
	decision := playback.PlaybackDecision{
		Method:           playback.MethodDirectPlay,
		VideoAction:      playback.VideoCopy,
		AudioAction:      playback.AudioCopy,
		SubtitleAction:   playback.SubtitleNone,
		EstimatedBitrate: 8000,
	}

	got := applyAdminPlaybackPolicy("direct_play", decision, profile, playback.MediaFileInfo{
		VideoCodec: playback.CodecH264,
		AudioCodec: playback.CodecDTS,
		Height:     1080,
		Bitrate:    8000,
	})

	if got.Method != playback.MethodTranscodeAudio {
		t.Fatalf("Method = %q, want %q", got.Method, playback.MethodTranscodeAudio)
	}
	if got.VideoAction != playback.VideoCopy {
		t.Fatalf("VideoAction = %q, want %q", got.VideoAction, playback.VideoCopy)
	}
	if got.AudioAction != playback.AudioTranscode {
		t.Fatalf("AudioAction = %q, want %q", got.AudioAction, playback.AudioTranscode)
	}
	if !strings.Contains(got.Reason, "audio transcode") {
		t.Fatalf("Reason = %q, want audio transcode hint", got.Reason)
	}
}
