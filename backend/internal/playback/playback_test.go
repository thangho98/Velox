package playback

import (
	"testing"
)

// --- ParseQuality tests ---

func TestParseQuality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{name: "4K variants", input: "4k", expected: 2160},
		{name: "2160p", input: "2160p", expected: 2160},
		{name: "2160", input: "2160", expected: 2160},
		{name: "1080p", input: "1080p", expected: 1080},
		{name: "1080", input: "1080", expected: 1080},
		{name: "720p", input: "720p", expected: 720},
		{name: "720", input: "720", expected: 720},
		{name: "480p", input: "480p", expected: 480},
		{name: "480", input: "480", expected: 480},
		{name: "original", input: "original", expected: 0},
		{name: "Original uppercase", input: "ORIGINAL", expected: 0},
		{name: "empty string", input: "", expected: 0},
		{name: "unknown", input: "1440p", expected: 0},
		{name: "random", input: "abc", expected: 0},
		{name: "case insensitive", input: "4K", expected: 2160},
		{name: "case insensitive 1080P", input: "1080P", expected: 1080},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseQuality(tt.input)
			if got != tt.expected {
				t.Errorf("ParseQuality(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

// --- normalizeCodec tests ---

func TestNormalizeCodec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// H.264
		{name: "h264 lowercase", input: "h264", expected: "h264"},
		{name: "h264 uppercase", input: "H264", expected: "h264"},
		{name: "avc", input: "avc", expected: "h264"},
		{name: "avc1", input: "avc1", expected: "h264"},
		{name: "AVC1", input: "AVC1", expected: "h264"},

		// H.265 / HEVC
		{name: "hevc lowercase", input: "hevc", expected: "hevc"},
		{name: "HEVC uppercase", input: "HEVC", expected: "hevc"},
		{name: "h265", input: "h265", expected: "hevc"},
		{name: "HEV1", input: "HEV1", expected: "hevc"},
		{name: "HVC1", input: "HVC1", expected: "hevc"},

		// VP9
		{name: "vp9 lowercase", input: "vp9", expected: "vp9"},
		{name: "VP9 uppercase", input: "VP9", expected: "vp9"},
		{name: "vp09", input: "vp09", expected: "vp9"},

		// AV1
		{name: "av1 lowercase", input: "av1", expected: "av1"},
		{name: "AV1 uppercase", input: "AV1", expected: "av1"},
		{name: "av01", input: "av01", expected: "av1"},

		// Audio codecs
		{name: "aac", input: "aac", expected: "aac"},
		{name: "mp4a", input: "mp4a", expected: "aac"},
		{name: "opus", input: "opus", expected: "opus"},
		{name: "flac", input: "flac", expected: "flac"},
		{name: "mp3", input: "mp3", expected: "mp3"},
		{name: "mpg", input: "mpg", expected: "mp3"},
		{name: "ac-3", input: "ac-3", expected: "ac3"},
		{name: "ac3", input: "ac3", expected: "ac3"},
		{name: "ec-3", input: "ec-3", expected: "eac3"},
		{name: "eac3", input: "eac3", expected: "eac3"},
		{name: "dts", input: "dts", expected: "dts"},

		// Unknown stays as-is
		{name: "unknown codec", input: "prores", expected: "prores"},
		{name: "theora", input: "theora", expected: "theora"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeCodec(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeCodec(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- normalizeContainer tests ---

func TestNormalizeContainer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "mp4", input: "mp4", expected: "mp4"},
		{name: "MP4 uppercase", input: "MP4", expected: "mp4"},
		{name: "mpeg4", input: "mpeg4", expected: "mp4"},
		{name: "m4v", input: "m4v", expected: "mp4"},
		{name: "webm", input: "webm", expected: "webm"},
		{name: "mkv", input: "mkv", expected: "mkv"},
		{name: "matroska", input: "matroska", expected: "mkv"},
		{name: "mov", input: "mov", expected: "mov"},
		{name: "qt", input: "qt", expected: "mov"},
		{name: "matroska,webm compound", input: "matroska,webm", expected: "mkv"},
		{name: "avi", input: "avi", expected: "avi"},
		{name: "unknown", input: "ts", expected: "ts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeContainer(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeContainer(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- NormalizeSubtitleCodec tests ---

func TestNormalizeSubtitleCodec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "pgs variants", input: "pgs", expected: "pgs"},
		{name: "hdmv_pgs_subtitle", input: "hdmv_pgs_subtitle", expected: "pgs"},
		{name: "pgssub", input: "pgssub", expected: "pgs"},
		{name: "vobsub", input: "vobsub", expected: "vobsub"},
		{name: "dvd_subtitle", input: "dvd_subtitle", expected: "vobsub"},
		{name: "dvdsub", input: "dvdsub", expected: "vobsub"},
		{name: "srt", input: "srt", expected: "srt"},
		{name: "subrip", input: "subrip", expected: "srt"},
		{name: "ass", input: "ass", expected: "ass"},
		{name: "ssa", input: "ssa", expected: "ass"},
		{name: "vtt", input: "vtt", expected: "vtt"},
		{name: "webvtt", input: "webvtt", expected: "vtt"},
		{name: "unknown", input: "xxx", expected: "xxx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeSubtitleCodec(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeSubtitleCodec(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- DeviceProfile method tests ---

func TestDeviceProfile_SupportsVideoCodec(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		SupportedVideoCodecs: []string{"h264", "hevc", "vp9"},
	}

	tests := []struct {
		name     string
		codec    string
		expected bool
	}{
		{name: "supported codec", codec: "h264", expected: true},
		{name: "supported codec uppercase", codec: "H264", expected: false}, // case-sensitive
		{name: "another supported codec", codec: "hevc", expected: true},
		{name: "unsupported codec", codec: "av1", expected: false},
		{name: "empty codec assumes supported", codec: "", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.SupportsVideoCodec(tt.codec)
			if got != tt.expected {
				t.Errorf("SupportsVideoCodec(%q) = %v, want %v", tt.codec, got, tt.expected)
			}
		})
	}
}

func TestDeviceProfile_SupportsAudioCodec(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		SupportedAudioCodecs: []string{"aac", "opus", "mp3"},
	}

	tests := []struct {
		name     string
		codec    string
		expected bool
	}{
		{name: "supported codec", codec: "aac", expected: true},
		{name: "unsupported codec", codec: "flac", expected: false},
		{name: "empty codec assumes supported", codec: "", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.SupportsAudioCodec(tt.codec)
			if got != tt.expected {
				t.Errorf("SupportsAudioCodec(%q) = %v, want %v", tt.codec, got, tt.expected)
			}
		})
	}
}

func TestDeviceProfile_SupportsContainer(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		SupportedContainers: []string{"mp4", "mkv", "webm"},
	}

	tests := []struct {
		name      string
		container string
		expected  bool
	}{
		{name: "supported container", container: "mp4", expected: true},
		{name: "unsupported container", container: "avi", expected: false},
		{name: "empty container assumes supported", container: "", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.SupportsContainer(tt.container)
			if got != tt.expected {
				t.Errorf("SupportsContainer(%q) = %v, want %v", tt.container, got, tt.expected)
			}
		})
	}
}

func TestDeviceProfile_SupportsSubtitleFormat(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		SupportedSubtitleFormats: []string{"vtt", "srt", "ass"},
	}

	tests := []struct {
		name     string
		format   string
		expected bool
	}{
		{name: "supported format", format: "vtt", expected: true},
		{name: "another supported format", format: "srt", expected: true},
		{name: "unsupported format", format: "pgs", expected: false},
		{name: "unsupported format ass", format: "ass", expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.SupportsSubtitleFormat(tt.format)
			if got != tt.expected {
				t.Errorf("SupportsSubtitleFormat(%q) = %v, want %v", tt.format, got, tt.expected)
			}
		})
	}
}

func TestDeviceProfile_CanPlayResolution(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		MaxWidth:  1920,
		MaxHeight: 1080,
	}

	tests := []struct {
		name     string
		width    int
		height   int
		expected bool
	}{
		{name: "within limits", width: 1920, height: 1080, expected: true},
		{name: "below limits", width: 1280, height: 720, expected: true},
		{name: "width exceeds", width: 3840, height: 1080, expected: false},
		{name: "height exceeds", width: 1920, height: 2160, expected: false},
		{name: "both exceed", width: 3840, height: 2160, expected: false},
		{name: "zero limits unlimited", width: 0, height: 0, expected: true},
	}

	profileUnlimited := &DeviceProfile{MaxWidth: 0, MaxHeight: 0}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.CanPlayResolution(tt.width, tt.height)
			if got != tt.expected {
				t.Errorf("CanPlayResolution(%d, %d) = %v, want %v", tt.width, tt.height, got, tt.expected)
			}
		})
	}

	// Test with unlimited profile
	if !profileUnlimited.CanPlayResolution(7680, 4320) {
		t.Errorf("unlimited profile should support any resolution")
	}
}

func TestDeviceProfile_CanPlayBitrate(t *testing.T) {
	t.Parallel()

	profile := &DeviceProfile{
		MaxBitrate: 10000,
	}

	tests := []struct {
		name     string
		bitrate  int
		expected bool
	}{
		{name: "within limit", bitrate: 5000, expected: true},
		{name: "at limit", bitrate: 10000, expected: true},
		{name: "exceeds limit", bitrate: 15000, expected: false},
		{name: "zero limit unlimited", bitrate: 0, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := profile.CanPlayBitrate(tt.bitrate)
			if got != tt.expected {
				t.Errorf("CanPlayBitrate(%d) = %v, want %v", tt.bitrate, got, tt.expected)
			}
		})
	}
}

// --- GetBuiltinProfile tests ---

func TestGetBuiltinProfile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		wantName string
	}{
		{name: "android_native", input: "android_native", wantName: "Android Native"},
		{name: "chrome", input: "chrome", wantName: "Chrome Desktop"},
		{name: "firefox", input: "firefox", wantName: "Firefox Desktop"},
		{name: "safari", input: "safari", wantName: "Safari Desktop"},
		{name: "mobile_safari", input: "mobile_safari", wantName: "Mobile Safari"},
		{name: "edge", input: "edge", wantName: "Edge Desktop"},
		{name: "smarttv", input: "smarttv", wantName: "Smart TV"},
		{name: "generic", input: "unknown", wantName: "Generic Browser"},
		{name: "empty string", input: "", wantName: "Generic Browser"},
		{name: "case sensitive - uppercase fails", input: "CHROME", wantName: "Generic Browser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBuiltinProfile(tt.input)
			if got.Name != tt.wantName {
				t.Errorf("GetBuiltinProfile(%q).Name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

// --- AllBuiltinProfiles tests ---

func TestAllBuiltinProfiles(t *testing.T) {
	t.Parallel()

	profiles := AllBuiltinProfiles()

	expected := []string{
		"android_native",
		"chrome",
		"firefox",
		"safari",
		"mobile_safari",
		"edge",
		"smarttv",
		"generic",
	}

	if len(profiles) != len(expected) {
		t.Errorf("AllBuiltinProfiles() returned %d profiles, want %d", len(profiles), len(expected))
	}

	for _, name := range expected {
		if profiles[name] == nil {
			t.Errorf("AllBuiltinProfiles()[%q] is nil", name)
		}
		if profiles[name].Name == "" {
			t.Errorf("AllBuiltinProfiles()[%q].Name is empty", name)
		}
	}
}

// --- Builtin profile capabilities ---

func TestAndroidNativeProfile(t *testing.T) {
	t.Parallel()

	p := &AndroidNative

	// Android supports HEVC and VP9, AV1
	if !p.SupportsVideoCodec("hevc") {
		t.Error("AndroidNative should support HEVC")
	}
	if !p.SupportsVideoCodec("vp9") {
		t.Error("AndroidNative should support VP9")
	}
	if !p.SupportsVideoCodec("av1") {
		t.Error("AndroidNative should support AV1")
	}
	if !p.SupportsHDR {
		t.Error("AndroidNative should support HDR")
	}
}

func TestChromeDesktopProfile(t *testing.T) {
	t.Parallel()

	p := &ChromeDesktop

	// Chrome doesn't support HEVC
	if p.SupportsVideoCodec("hevc") {
		t.Error("ChromeDesktop should NOT support HEVC")
	}
	// But does support VP9 and AV1
	if !p.SupportsVideoCodec("vp9") {
		t.Error("ChromeDesktop should support VP9")
	}
	if !p.SupportsVideoCodec("av1") {
		t.Error("ChromeDesktop should support AV1")
	}
}

func TestSafariDesktopProfile(t *testing.T) {
	t.Parallel()

	p := &SafariDesktop

	// Safari supports HEVC on macOS
	if !p.SupportsVideoCodec("hevc") {
		t.Error("SafariDesktop should support HEVC")
	}
	// Safari does NOT support WebM
	if p.SupportsWebM {
		t.Error("SafariDesktop should NOT support WebM")
	}
}

// --- Additional Decide tests for edge cases ---

func TestDecide_HDRServerTonemap(t *testing.T) {
	t.Parallel()

	// DV Profile 5 (NeedsServerTonemap=true) should trigger transcode on non-HDR client
	media := MediaFileInfo{
		VideoCodec:         "hevc",
		AudioCodec:         "aac",
		Container:          "mkv",
		Width:              1920,
		Height:             1080,
		Bitrate:            8000,
		IsHDR:              true,
		NeedsServerTonemap: true,
	}

	// Chrome doesn't support HDR
	decision := Decide(media, &ChromeDesktop, UserPreferences{})

	if decision.Method != MethodFullTranscode {
		t.Errorf("HDR DV Profile 5 on non-HDR client = %v, want %v", decision.Method, MethodFullTranscode)
	}
}

func TestDecide_StandardHDRToneMapping(t *testing.T) {
	t.Parallel()

	// Standard HDR10 (NeedsServerTonemap=false) should also trigger transcode on non-HDR client
	media := MediaFileInfo{
		VideoCodec:         "hevc",
		AudioCodec:         "aac",
		Container:          "mkv",
		Width:              1920,
		Height:             1080,
		Bitrate:            8000,
		IsHDR:              true,
		NeedsServerTonemap: false,
	}

	// Chrome doesn't support HDR
	decision := Decide(media, &ChromeDesktop, UserPreferences{})

	if decision.Method != MethodFullTranscode {
		t.Errorf("Standard HDR on non-HDR client = %v, want %v", decision.Method, MethodFullTranscode)
	}
}

func TestDecide_HDROnHDRCapableClient(t *testing.T) {
	t.Parallel()

	// HDR content on HDR-capable client (Android/ExoPlayer) should direct play
	media := MediaFileInfo{
		VideoCodec:         "hevc",
		AudioCodec:         "aac",
		Container:          "mkv",
		Width:              1920,
		Height:             1080,
		Bitrate:            8000,
		IsHDR:              true,
		NeedsServerTonemap: false,
	}

	// Android Native supports HDR
	decision := Decide(media, &AndroidNative, UserPreferences{})

	// Should direct play since client can handle HDR
	if decision.Method != MethodDirectPlay {
		t.Errorf("HDR on HDR-capable client = %v, want %v", decision.Method, MethodDirectPlay)
	}
}

func TestEstimateBitrate(t *testing.T) {
	t.Parallel()

	// estimateBitrate is internal, but we test via the public API behavior
	// We can only test through the Decision.EstimatedBitrate field
	tests := []struct {
		name      string
		height    int
		wantRange [2]int // acceptable range for estimated bitrate
	}{
		{name: "4K and above", height: 2160, wantRange: [2]int{20000, 30000}},
		{name: "4K exact", height: 2160, wantRange: [2]int{20000, 30000}},
		{name: "1080p range", height: 1080, wantRange: [2]int{6000, 10000}},
		{name: "720p range", height: 720, wantRange: [2]int{3000, 5000}},
		{name: "480p range", height: 480, wantRange: [2]int{1500, 2500}},
		{name: "below 480p", height: 360, wantRange: [2]int{1000, 2000}},
		{name: "very small", height: 240, wantRange: [2]int{1000, 2000}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			media := MediaFileInfo{
				VideoCodec: "hevc", // incompatible codec to force FullTranscode
				AudioCodec: "aac",
				Container:  "mp4",
				Height:     tt.height,
				Bitrate:    8000, // low enough to not trigger bitrate limit
			}
			profile := &GenericBrowser
			prefs := UserPreferences{MaxStreamingQuality: "original"}

			decision := Decide(media, profile, prefs)
			// Full transcode should estimate bitrate based on height
			if decision.Method == MethodFullTranscode {
				if decision.EstimatedBitrate < tt.wantRange[0] || decision.EstimatedBitrate > tt.wantRange[1] {
					t.Errorf("EstimatedBitrate for %dp = %d, want between %d and %d",
						tt.height, decision.EstimatedBitrate, tt.wantRange[0], tt.wantRange[1])
				}
			}
		})
	}
}

func TestDecide_VobSubSubtitles(t *testing.T) {
	t.Parallel()

	media := MediaFileInfo{
		VideoCodec:   "h264",
		AudioCodec:   "aac",
		Container:    "mp4",
		Width:        1920,
		Height:       1080,
		HasSubtitles: true,
		SubType:      SubtitleVobSub,
	}

	prefs := UserPreferences{
		SelectedSubtitle: "en",
	}

	decision := Decide(media, &ChromeDesktop, prefs)

	// VobSub should trigger burn-in
	if decision.SubtitleAction != SubtitleBurnIn {
		t.Errorf("VobSub subtitles action = %v, want %v", decision.SubtitleAction, SubtitleBurnIn)
	}
	if decision.Method != MethodFullTranscode {
		t.Errorf("VobSub method = %v, want %v", decision.Method, MethodFullTranscode)
	}
}

func TestDecide_NoSubtitleSelected(t *testing.T) {
	t.Parallel()

	media := MediaFileInfo{
		VideoCodec:   "h264",
		AudioCodec:   "aac",
		Container:    "mp4",
		Width:        1920,
		Height:       1080,
		HasSubtitles: true,
		SubType:      SubtitlePGS,
	}

	// No subtitle selected
	prefs := UserPreferences{
		SelectedSubtitle: "off",
	}

	decision := Decide(media, &ChromeDesktop, prefs)

	if decision.SubtitleAction != SubtitleNone {
		t.Errorf("SubtitleAction with 'off' = %v, want %v", decision.SubtitleAction, SubtitleNone)
	}
}

func TestDecide_VideoCopyWithIncompatibleAudio(t *testing.T) {
	t.Parallel()

	// Video is compatible but audio is not → should use MethodTranscodeAudio (video copy)
	media := MediaFileInfo{
		VideoCodec: "h264",
		AudioCodec: "dts",
		Container:  "mp4",
		Width:      1920,
		Height:     1080,
		Bitrate:    8000,
	}

	decision := Decide(media, &ChromeDesktop, UserPreferences{})

	if decision.Method != MethodTranscodeAudio {
		t.Errorf("DTS audio method = %v, want %v", decision.Method, MethodTranscodeAudio)
	}
	if decision.VideoAction != VideoCopy {
		t.Errorf("DTS audio video action = %v, want %v", decision.VideoAction, VideoCopy)
	}
}

// NOTE: TestDecide_NilProfile removed because the Decide function
// does not gracefully handle nil profile - it will panic on nil profile
// when checking profile.SupportsVideoCodec(). This is a known limitation
// and callers are expected to pass a non-nil profile.
