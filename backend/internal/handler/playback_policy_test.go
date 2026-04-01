package handler

import (
	"strings"
	"testing"

	"github.com/thawng/velox/internal/playback"
)

func TestApplyAdminPlaybackPolicyUsesFullTranscodeForBrowserAudioMismatch(t *testing.T) {
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

	if got.Method != playback.MethodFullTranscode {
		t.Fatalf("Method = %q, want %q", got.Method, playback.MethodFullTranscode)
	}
	if got.VideoAction != playback.VideoTranscode {
		t.Fatalf("VideoAction = %q, want %q", got.VideoAction, playback.VideoTranscode)
	}
	if got.AudioAction != playback.AudioTranscode {
		t.Fatalf("AudioAction = %q, want %q", got.AudioAction, playback.AudioTranscode)
	}
	if !strings.Contains(got.Reason, "reliable HLS seeking") {
		t.Fatalf("Reason = %q, want reliable seek hint", got.Reason)
	}
}
