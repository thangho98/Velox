package hls

import (
	"strings"
	"testing"

	"github.com/thawng/velox/internal/model"
)

func TestGenerateMasterPlaylist(t *testing.T) {
	prefix := "test_prefix_"
	videoBandwidth := 10000000

	audioTracks := []model.AudioTrack{
		{
			Title:     "English (Stereo)",
			Language:  "eng",
			IsDefault: true,
		},
		{
			Language:  "vie",
			IsDefault: false,
		},
	}

	master := GenerateMasterPlaylist(audioTracks, prefix, videoBandwidth)

	if !strings.Contains(master, "#EXTM3U") {
		t.Error("Missing EXTM3U")
	}

	if !strings.Contains(master, "#EXT-X-VERSION:6") {
		t.Error("Missing VERSION 6")
	}

	// Audio tracks
	if !strings.Contains(master, "URI=\"test_prefix_audio0.m3u8\"") {
		t.Error("Missing audio0 URI")
	}
	if !strings.Contains(master, "DEFAULT=YES") {
		t.Error("Missing DEFAULT=YES for audio0")
	}
	if !strings.Contains(master, "URI=\"test_prefix_audio1.m3u8\"") {
		t.Error("Missing audio1 URI")
	}
	if !strings.Contains(master, "NAME=\"Track 2\"") { // Fallback title
		t.Error("Missing fallback name for audio1")
	}

	// Video track
	if !strings.Contains(master, "BANDWIDTH=10000000") {
		t.Error("Missing bandwidth")
	}
	if !strings.Contains(master, "AUDIO=\"audio\"") {
		t.Error("Missing audio group reference in video stream info")
	}
	if !strings.Contains(master, "test_prefix_video.m3u8") {
		t.Error("Missing video playlist URI")
	}
}

func TestGenerateMasterPlaylist_NoAudio(t *testing.T) {
	master := GenerateMasterPlaylist(nil, "pre_", 5000000)

	if strings.Contains(master, "AUDIO=\"audio\"") {
		t.Error("Should not have AUDIO reference when no audio tracks")
	}
	if !strings.Contains(master, "pre_video.m3u8") {
		t.Error("Missing video playlist URI")
	}
}
