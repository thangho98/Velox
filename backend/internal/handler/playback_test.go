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

func TestResolvePlaybackAudioTrackAutoSelectsCompatibleAlternative(t *testing.T) {
	audioTracks := []model.AudioTrack{
		{ID: 10, Language: "eng", Codec: "dts", IsDefault: true},
		{ID: 11, Language: "eng", Codec: "aac", IsDefault: false},
	}

	selected, effectiveID, autoSelected := resolvePlaybackAudioTrack(0, "", audioTracks, &playback.ChromeDesktop)
	if selected == nil {
		t.Fatal("selected track = nil, want AAC alternative")
	}
	if got, want := selected.Codec, "aac"; got != want {
		t.Fatalf("selected codec = %q, want %q", got, want)
	}
	if got, want := effectiveID, 11; got != want {
		t.Fatalf("effectiveID = %d, want %d", got, want)
	}
	if !autoSelected {
		t.Fatal("autoSelected = false, want true")
	}
}

func TestResolvePlaybackAudioTrackPrefersSameLanguageAlternative(t *testing.T) {
	audioTracks := []model.AudioTrack{
		{ID: 10, Language: "jpn", Codec: "dts", IsDefault: true},
		{ID: 11, Language: "eng", Codec: "aac", IsDefault: false},
		{ID: 12, Language: "jpn", Codec: "aac", IsDefault: false},
	}

	selected, effectiveID, autoSelected := resolvePlaybackAudioTrack(0, "", audioTracks, &playback.ChromeDesktop)
	if selected == nil {
		t.Fatal("selected track = nil, want same-language AAC alternative")
	}
	if got, want := selected.ID, int64(12); got != want {
		t.Fatalf("selected ID = %d, want %d", got, want)
	}
	if got, want := effectiveID, 12; got != want {
		t.Fatalf("effectiveID = %d, want %d", got, want)
	}
	if !autoSelected {
		t.Fatal("autoSelected = false, want true")
	}
}

func TestResolvePlaybackAudioTrackHonorsExplicitSelection(t *testing.T) {
	audioTracks := []model.AudioTrack{
		{ID: 10, Language: "eng", Codec: "dts", IsDefault: true},
		{ID: 11, Language: "eng", Codec: "aac", IsDefault: false},
	}

	selected, effectiveID, autoSelected := resolvePlaybackAudioTrack(11, "", audioTracks, &playback.ChromeDesktop)
	if selected == nil {
		t.Fatal("selected track = nil, want explicit track")
	}
	if got, want := selected.ID, int64(11); got != want {
		t.Fatalf("selected ID = %d, want %d", got, want)
	}
	if got, want := effectiveID, 11; got != want {
		t.Fatalf("effectiveID = %d, want %d", got, want)
	}
	if autoSelected {
		t.Fatal("autoSelected = true, want false for explicit selection")
	}
}

func TestAdjustPlaybackDecisionForSelectedAudioTrackUsesDirectStream(t *testing.T) {
	decision := playback.PlaybackDecision{
		Method:         playback.MethodDirectPlay,
		VideoAction:    playback.VideoCopy,
		AudioAction:    playback.AudioCopy,
		SubtitleAction: playback.SubtitleNone,
		Reason:         "Direct play compatible",
	}
	selected := &model.AudioTrack{ID: 11, Codec: "aac", IsDefault: false}

	got := adjustPlaybackDecisionForSelectedAudioTrack(decision, selected, 11, true)
	if got.Method != playback.MethodDirectStream {
		t.Fatalf("Method = %q, want %q", got.Method, playback.MethodDirectStream)
	}
	if got.VideoAction != playback.VideoCopy {
		t.Fatalf("VideoAction = %q, want %q", got.VideoAction, playback.VideoCopy)
	}
	if got.AudioAction != playback.AudioCopy {
		t.Fatalf("AudioAction = %q, want %q", got.AudioAction, playback.AudioCopy)
	}
	if got.Reason == decision.Reason {
		t.Fatalf("Reason = %q, want reason to mention selected compatible audio track", got.Reason)
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

func TestFindSubtitleByID(t *testing.T) {
	subtitles := []model.Subtitle{
		{ID: 11, Language: "eng", Codec: "subrip"},
		{ID: 12, Language: "eng", Codec: "subrip"},
	}

	if got := findSubtitleByID(subtitles, 12); got == nil || got.ID != 12 {
		t.Fatalf("findSubtitleByID(12) = %+v, want subtitle ID 12", got)
	}
	if got := findSubtitleByID(subtitles, 999); got != nil {
		t.Fatalf("findSubtitleByID(999) = %+v, want nil", got)
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
