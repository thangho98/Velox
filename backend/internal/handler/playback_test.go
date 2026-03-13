package handler

import (
	"reflect"
	"testing"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
)

func TestApplyClientCapabilityOverrides(t *testing.T) {
	base := &playback.DeviceProfile{
		Name:                 "base",
		SupportedVideoCodecs: []string{playback.CodecH264, playback.CodecVP9},
		SupportedAudioCodecs: []string{playback.CodecAAC, playback.CodecOpus},
		SupportedContainers:  []string{playback.ContainerMP4, playback.ContainerMKV, playback.ContainerHLS},
		MaxHeight:            2160,
	}

	overridden := applyClientCapabilityOverrides(base, PlaybackInfoRequest{
		VideoCodecs: []string{" H264 "},
		AudioCodecs: []string{"AAC"},
		Containers:  []string{"mp4", "hls"},
		MaxHeight:   1080,
	})

	if overridden == base {
		t.Fatal("expected cloned profile, got original pointer")
	}
	if got, want := overridden.MaxHeight, 1080; got != want {
		t.Fatalf("MaxHeight = %d, want %d", got, want)
	}
	if got, want := overridden.SupportedContainers, []string{"mp4", "hls"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedContainers = %v, want %v", got, want)
	}
	if got, want := overridden.SupportedVideoCodecs, []string{"h264"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedVideoCodecs = %v, want %v", got, want)
	}
	if got, want := overridden.SupportedAudioCodecs, []string{"aac"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedAudioCodecs = %v, want %v", got, want)
	}
	if got, want := base.SupportedContainers, []string{playback.ContainerMP4, playback.ContainerMKV, playback.ContainerHLS}; !reflect.DeepEqual(got, want) {
		t.Fatalf("base profile mutated: containers = %v, want %v", got, want)
	}
}

func TestResolveSelectedAudioTrackID(t *testing.T) {
	audioTracks := []model.AudioTrack{
		{ID: 10, IsDefault: true},
		{ID: 11, IsDefault: false},
	}

	tests := []struct {
		name       string
		requested  int
		wantResult int
	}{
		{name: "none selected", requested: 0, wantResult: 0},
		{name: "invalid stale id", requested: 999, wantResult: 0},
		{name: "default track ignored", requested: 10, wantResult: 0},
		{name: "non-default track kept", requested: 11, wantResult: 11},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveSelectedAudioTrackID(tt.requested, audioTracks)
			if got != tt.wantResult {
				t.Fatalf("resolveSelectedAudioTrackID(%d) = %d, want %d", tt.requested, got, tt.wantResult)
			}
		})
	}
}

func TestPlaybackModeQuery(t *testing.T) {
	if got := playbackModeQuery(playback.MethodDirectPlay); got != "direct" {
		t.Fatalf("playbackModeQuery(DirectPlay) = %q, want %q", got, "direct")
	}
	if got := playbackModeQuery(playback.MethodDirectStream); got != "directstream" {
		t.Fatalf("playbackModeQuery(DirectStream) = %q, want %q", got, "directstream")
	}
}

func TestFindSubtitleByLanguageNormalizesCodes(t *testing.T) {
	subtitles := []model.Subtitle{
		{ID: 1, Language: "vie"},
		{ID: 2, Language: "eng"},
	}

	if got := findSubtitleByLanguage(subtitles, "en"); got == nil || got.ID != 2 {
		t.Fatalf("findSubtitleByLanguage(en) = %+v, want subtitle ID 2", got)
	}
	if got := findSubtitleByLanguage(subtitles, "vi"); got == nil || got.ID != 1 {
		t.Fatalf("findSubtitleByLanguage(vi) = %+v, want subtitle ID 1", got)
	}
}

func TestFindSubtitleByLanguagePrefersTextOverImage(t *testing.T) {
	subtitles := []model.Subtitle{
		{ID: 1, Language: "eng", Codec: "hdmv_pgs_subtitle"},
		{ID: 2, Language: "eng", Codec: "subrip"},
	}

	got := findSubtitleByLanguage(subtitles, "en")
	if got == nil || got.ID != 2 {
		t.Fatalf("findSubtitleByLanguage(en) = %+v, want text subtitle ID 2", got)
	}
}

func TestFilterPlayableSubtitlesHidesImageTracksWithoutBurnInSupport(t *testing.T) {
	subtitles := []model.Subtitle{
		{ID: 1, Language: "eng", Codec: "hdmv_pgs_subtitle"},
		{ID: 2, Language: "eng", Codec: "subrip"},
		{ID: 3, Language: "vie", Codec: "webvtt"},
	}

	got := filterPlayableSubtitles(subtitles, false)
	want := []model.Subtitle{
		{ID: 2, Language: "eng", Codec: "subrip"},
		{ID: 3, Language: "vie", Codec: "webvtt"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterPlayableSubtitles(..., false) = %+v, want %+v", got, want)
	}
}
