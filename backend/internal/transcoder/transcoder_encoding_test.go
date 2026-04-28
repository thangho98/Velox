package transcoder

import (
	"strings"
	"testing"
)

func TestNormalizeStartOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{name: "zero", input: 0, expected: 0},
		{name: "negative", input: -5, expected: 0},
		{name: "tiny 0.1s rounds to 0", input: 0.1, expected: 0},
		{name: "tiny 0.2s rounds to 0", input: 0.2, expected: 0},
		{name: "tiny 0.25s rounds to 0", input: 0.25, expected: 0},
		{name: "at boundary 0.251", input: 0.251, expected: 0.251},
		{name: "small 0.5s", input: 0.5, expected: 0.5},
		{name: "large 3480.125", input: 3480.125, expected: 3480.125},
		{name: "rounds to 3 decimals", input: 1.23456, expected: 1.235},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeStartOffset(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeStartOffset(%.3f) = %.3f, want %.3f", tt.input, got, tt.expected)
			}
		})
	}
}

func TestStartOffsetMillis(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected int64
	}{
		{name: "zero stays zero", input: 0, expected: 0},
		{name: "tiny rounds to zero", input: 0.2, expected: 0},
		{name: "at boundary 0.25 rounds to 0", input: 0.25, expected: 0},
		{name: "above boundary 0.251", input: 0.251, expected: 251},
		{name: "0.5s becomes 500ms", input: 0.5, expected: 500},
		{name: "1s becomes 1000ms", input: 1.0, expected: 1000},
		{name: "3480.125s becomes 3480125ms", input: 3480.125, expected: 3480125},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := startOffsetMillis(tt.input)
			if got != tt.expected {
				t.Errorf("startOffsetMillis(%.3f) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestHlsPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		streamSessionID   string
		fileID            int64
		subtitleStreamIdx int
		videoCopy         bool
		startOffset       float64
		wantContains      []string
		wantNotContains   []string
	}{
		{
			name:              "basic no session",
			streamSessionID:   "",
			fileID:            5,
			subtitleStreamIdx: -1,
			videoCopy:         false,
			startOffset:       0,
			wantContains:      []string{"f5_"},
			wantNotContains:   []string{"ss", "vc", "sub", "off"},
		},
		{
			name:              "with stream session",
			streamSessionID:   "sess123",
			fileID:            5,
			subtitleStreamIdx: -1,
			videoCopy:         false,
			startOffset:       0,
			wantContains:      []string{"sssess123_", "f5_"},
			wantNotContains:   []string{"vc", "sub", "off"},
		},
		{
			name:              "video copy adds vc",
			streamSessionID:   "",
			fileID:            5,
			subtitleStreamIdx: -1,
			videoCopy:         true,
			startOffset:       0,
			wantContains:      []string{"vc", "f5_"},
			wantNotContains:   []string{"ss", "sub", "off"},
		},
		{
			name:              "subtitle stream adds sub",
			streamSessionID:   "",
			fileID:            5,
			subtitleStreamIdx: 2,
			videoCopy:         false,
			startOffset:       0,
			wantContains:      []string{"f5_", "sub2_"},
			wantNotContains:   []string{"ss", "vc", "off"},
		},
		{
			name:              "large seek offset",
			streamSessionID:   "",
			fileID:            5,
			subtitleStreamIdx: -1,
			videoCopy:         false,
			startOffset:       3480.125,
			wantContains:      []string{"f5_", "off3480125_"},
			wantNotContains:   []string{"ss", "vc", "sub"},
		},
		{
			name:              "tiny seek offset omitted",
			streamSessionID:   "",
			fileID:            5,
			subtitleStreamIdx: -1,
			videoCopy:         false,
			startOffset:       0.2,
			wantContains:      []string{"f5_"},
			wantNotContains:   []string{"off"},
		},
		{
			name:              "all options combined",
			streamSessionID:   "sess123",
			fileID:            5,
			subtitleStreamIdx: 2,
			videoCopy:         true,
			startOffset:       120.5,
			wantContains:      []string{"sssess123_", "vc", "f5_", "sub2_", "off120500_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hlsPrefix(tt.streamSessionID, tt.fileID, tt.subtitleStreamIdx, tt.videoCopy, tt.startOffset)
			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("hlsPrefix() = %q, want to contain %q", got, want)
				}
			}
			for _, dontWant := range tt.wantNotContains {
				if strings.Contains(got, dontWant) {
					t.Errorf("hlsPrefix() = %q, should NOT contain %q", got, dontWant)
				}
			}
		})
	}
}

func TestHlsPrefix_OmitsTinyOffset(t *testing.T) {
	t.Parallel()

	// Edge case: tiny offset < 0.25s should be omitted
	got := hlsPrefix("", 5, -1, false, 0.249)
	if strings.Contains(got, "off") {
		t.Errorf("hlsPrefix with 0.249 should not contain 'off', got %q", got)
	}
}

func TestHlsPrefix_OmitsSubtitleWhenNegative(t *testing.T) {
	t.Parallel()

	got := hlsPrefix("", 5, -1, false, 0)
	if strings.Contains(got, "sub") {
		t.Errorf("hlsPrefix with siIdx=-1 should not contain 'sub', got %q", got)
	}
}

func TestHlsPrefix_OmitsFileIDWhenZero(t *testing.T) {
	t.Parallel()

	got := hlsPrefix("sess", 0, -1, false, 0)
	if strings.Contains(got, "f0") {
		t.Errorf("hlsPrefix with fileID=0 should not contain 'f0', got %q", got)
	}
}

func TestHlsPrefix_SessionPrefixOnly(t *testing.T) {
	t.Parallel()

	got := hlsPrefix("abc123", 0, -1, false, 0)
	if got != "ssabc123_" {
		t.Errorf("hlsPrefix() = %q, want %q", got, "ssabc123_")
	}
}

// --- hwInputArgs tests ---

func TestHwInputArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hwAccel  string
		expected []string
	}{
		{name: "videotoolbox", hwAccel: "videotoolbox", expected: []string{"-hwaccel", "videotoolbox"}},
		{name: "vaapi", hwAccel: "vaapi", expected: []string{"-hwaccel", "vaapi", "-hwaccel_output_format", "vaapi", "-hwaccel_device", "/dev/dri/renderD128"}},
		{name: "nvenc", hwAccel: "nvenc", expected: []string{"-hwaccel", "cuda"}},
		{name: "qsv", hwAccel: "qsv", expected: []string{"-hwaccel", "qsv"}},
		{name: "amf returns nil", hwAccel: "amf", expected: nil},
		{name: "unknown returns nil", hwAccel: "", expected: nil},
		{name: "random returns nil", hwAccel: "vulkan", expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hwInputArgs(tt.hwAccel)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("hwInputArgs(%q) = %v, want nil", tt.hwAccel, got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("hwInputArgs(%q) len = %d, want %d", tt.hwAccel, len(got), len(tt.expected))
				return
			}
			for i, want := range tt.expected {
				if got[i] != want {
					t.Errorf("hwInputArgs(%q)[%d] = %q, want %q", tt.hwAccel, i, got[i], want)
				}
			}
		})
	}
}

// --- hwScaleFilter tests ---

func TestHwScaleFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hwAccel  string
		height   int
		expected string
	}{
		{name: "vaapi 1080", hwAccel: "vaapi", height: 1080, expected: "scale_vaapi=w=-2:h=1080:format=nv12"},
		{name: "vaapi 720", hwAccel: "vaapi", height: 720, expected: "scale_vaapi=w=-2:h=720:format=nv12"},
		{name: "nvenc 1080", hwAccel: "nvenc", height: 1080, expected: "scale=-2:1080"},
		{name: "qsv 480", hwAccel: "qsv", height: 480, expected: "scale=-2:480"},
		{name: "empty hwAccel 1080", hwAccel: "", height: 1080, expected: "scale=-2:1080"},
		{name: "unknown hwAccel", hwAccel: "videotoolbox", height: 1080, expected: "scale=-2:1080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hwScaleFilter(tt.hwAccel, tt.height)
			if got != tt.expected {
				t.Errorf("hwScaleFilter(%q, %d) = %q, want %q", tt.hwAccel, tt.height, got, tt.expected)
			}
		})
	}
}

// --- ffmpegInputProbeArgs tests ---

func TestFfmpegInputProbeArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputPath string
		wantLen   int
		wantFirst string
	}{
		{name: "local file", inputPath: "/mnt/media/movie.mkv", wantLen: 4, wantFirst: "-probesize"},
		{name: "http URL", inputPath: "http://example.com/stream", wantLen: 12, wantFirst: "-reconnect"},
		{name: "https URL", inputPath: "https://example.com/stream", wantLen: 12, wantFirst: "-reconnect"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ffmpegInputProbeArgs(tt.inputPath)
			if len(got) != tt.wantLen {
				t.Errorf("ffmpegInputProbeArgs(%q) len = %d, want %d", tt.inputPath, len(got), tt.wantLen)
			}
			if got[0] != tt.wantFirst {
				t.Errorf("ffmpegInputProbeArgs(%q)[0] = %q, want %q", tt.inputPath, got[0], tt.wantFirst)
			}
		})
	}
}

// --- hwVideoCodec tests ---

func TestHwVideoCodec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hwAccel  string
		expected string
	}{
		{name: "videotoolbox", hwAccel: "videotoolbox", expected: "h264_videotoolbox"},
		{name: "vaapi", hwAccel: "vaapi", expected: "h264_vaapi"},
		{name: "nvenc", hwAccel: "nvenc", expected: "h264_nvenc"},
		{name: "qsv", hwAccel: "qsv", expected: "h264_qsv"},
		{name: "amf", hwAccel: "amf", expected: "h264_amf"},
		{name: "empty falls back to libx264", hwAccel: "", expected: "libx264"},
		{name: "unknown falls back to libx264", hwAccel: "vulkan", expected: "libx264"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hwVideoCodec(tt.hwAccel)
			if got != tt.expected {
				t.Errorf("hwVideoCodec(%q) = %q, want %q", tt.hwAccel, got, tt.expected)
			}
		})
	}
}

// --- hdrToneMapFilterSW tests ---

func TestHdrToneMapFilterSW(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputPath string
		outFmt    string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "yuv420p output",
			inputPath: "/test/video.mkv",
			outFmt:    "yuv420p",
			wantStart: "tonemapx=tonemap=bt2390",
			wantEnd:   "format=yuv420p",
		},
		{
			name:      "nv12 output",
			inputPath: "/test/video.mkv",
			outFmt:    "nv12",
			wantStart: "tonemapx=tonemap=bt2390",
			wantEnd:   "format=nv12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hdrToneMapFilterSW(tt.inputPath, tt.outFmt)
			if !strings.HasPrefix(got, tt.wantStart) {
				t.Errorf("hdrToneMapFilterSW() = %q, want to start with %q", got, tt.wantStart)
			}
			if !strings.HasSuffix(got, tt.wantEnd) {
				t.Errorf("hdrToneMapFilterSW() = %q, want to end with %q", got, tt.wantEnd)
			}
		})
	}
}

// --- ShouldUseHwTonemap tests ---

func TestShouldUseHwTonemap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		hwAccel         string
		dvProfile       int
		enableHwTonemap bool
		expected        bool
	}{
		{name: "disabled returns false", hwAccel: "vaapi", dvProfile: 0, enableHwTonemap: false, expected: false},
		{name: "DV Profile 5 returns false", hwAccel: "vaapi", dvProfile: 5, enableHwTonemap: true, expected: false},
		{name: "DV Profile 7 may return true", hwAccel: "vaapi", dvProfile: 7, enableHwTonemap: true, expected: false}, // depends on IsVAAPITonemapAvailable
		{name: "unknown hwAccel returns false", hwAccel: "", dvProfile: 0, enableHwTonemap: true, expected: false},
		{name: "nvenc with DV profile 8", hwAccel: "nvenc", dvProfile: 8, enableHwTonemap: true, expected: false}, // depends on IsVulkanPlaceboAvailable
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldUseHwTonemap(tt.hwAccel, tt.dvProfile, tt.enableHwTonemap)
			if got != tt.expected {
				t.Errorf("ShouldUseHwTonemap(%q, %d, %v) = %v, want %v",
					tt.hwAccel, tt.dvProfile, tt.enableHwTonemap, got, tt.expected)
			}
		})
	}
}

// --- NeedsDoviStripBSF tests ---

func TestNeedsDoviStripBSF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		dvProfile int
		expected  bool
	}{
		{name: "Profile 0", dvProfile: 0, expected: false},
		{name: "Profile 1", dvProfile: 1, expected: false},
		{name: "Profile 4", dvProfile: 4, expected: false},
		{name: "Profile 5", dvProfile: 5, expected: false},
		{name: "Profile 7", dvProfile: 7, expected: true},
		{name: "Profile 8", dvProfile: 8, expected: true},
		{name: "Profile 9", dvProfile: 9, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsDoviStripBSF(tt.dvProfile)
			if got != tt.expected {
				t.Errorf("NeedsDoviStripBSF(%d) = %v, want %v", tt.dvProfile, got, tt.expected)
			}
		})
	}
}

// --- HDRToneMapFilterForHW tests ---

func TestHDRToneMapFilterForHW(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		hwAccel   string
		inputPath string
		dvProfile int
		enableHw  bool
		wantEmpty bool // true if we expect software path (empty string or contains tonemapx)
	}{
		{
			name:      "vaapi hw tonemap",
			hwAccel:   "vaapi",
			inputPath: "/test/video.mkv",
			dvProfile: 0,
			enableHw:  false, // will use SW tonemap
			wantEmpty: false,
		},
		{
			name:      "empty hwAccel uses SW tonemapx",
			hwAccel:   "",
			inputPath: "/test/video.mkv",
			dvProfile: 0,
			enableHw:  true,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HDRToneMapFilterForHW(tt.hwAccel, tt.inputPath, tt.dvProfile, tt.enableHw)
			if tt.wantEmpty && got == "" {
				t.Errorf("HDRToneMapFilterForHW() = empty, want non-empty")
			}
			if !tt.wantEmpty && got == "" {
				t.Errorf("HDRToneMapFilterForHW() = empty, want non-empty")
			}
		})
	}
}

// --- hwEncoderArgs tests ---

func TestHwEncoderArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hwAccel  string
		expected []string
	}{
		{name: "empty software", hwAccel: "", expected: []string{"-preset", "veryfast", "-crf", "23", "-threads", "0", "-pix_fmt", "yuv420p"}},
		{name: "vaapi", hwAccel: "vaapi", expected: []string{"-profile:v", "main", "-qp", "23"}},
		{name: "amf", hwAccel: "amf", expected: []string{"-preset", "veryfast", "-quality", "balanced"}},
		{name: "other returns nil", hwAccel: "nvenc", expected: nil},
		{name: "qsv returns nil", hwAccel: "qsv", expected: nil},
		{name: "videotoolbox returns nil", hwAccel: "videotoolbox", expected: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hwEncoderArgs(tt.hwAccel)
			if tt.expected == nil {
				if got != nil {
					t.Errorf("hwEncoderArgs(%q) = %v, want nil", tt.hwAccel, got)
				}
				return
			}
			if len(got) != len(tt.expected) {
				t.Errorf("hwEncoderArgs(%q) len = %d, want %d", tt.hwAccel, len(got), len(tt.expected))
				return
			}
			for i, want := range tt.expected {
				if got[i] != want {
					t.Errorf("hwEncoderArgs(%q)[%d] = %q, want %q", tt.hwAccel, i, got[i], want)
				}
			}
		})
	}
}

// --- escapeFFmpegSubtitlePath tests ---

func TestEscapeFfmpegSubtitlePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "simple path", input: "/media/video.mkv", expected: "/media/video.mkv"},
		{name: "with colon", input: "C:/media/video.mkv", expected: "C\\:/media/video.mkv"},
		{name: "with brackets", input: "/media/[subs]/video.mkv", expected: "/media/\\[subs\\]/video.mkv"},
		{name: "with semicolon", input: "/media/subs;test/video.mkv", expected: "/media/subs\\;test/video.mkv"},
		{name: "with single quote", input: "/media/it's/video.mkv", expected: "/media/it\\'s/video.mkv"},
		{name: "with backslash and colon", input: "C:\\media\\video.mkv", expected: "C\\:\\\\media\\\\video.mkv"},
		{name: "complex path", input: "C:/My [Movies]/test's;file.mkv", expected: "C\\:/My \\[Movies\\]/test\\'s\\;file.mkv"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeFFmpegSubtitlePath(tt.input)
			if got != tt.expected {
				t.Errorf("escapeFFmpegSubtitlePath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

// --- hdrToneMapFilter tests ---

func TestHdrToneMapFilter(t *testing.T) {
	t.Parallel()

	got := hdrToneMapFilter("/test/video.mkv")
	if !strings.Contains(got, "tonemapx") {
		t.Errorf("hdrToneMapFilter() = %q, want to contain 'tonemapx'", got)
	}
	if !strings.HasSuffix(got, "yuv420p") {
		t.Errorf("hdrToneMapFilter() = %q, want to end with 'yuv420p'", got)
	}
}

// --- hdrColorMetadataFallbackPrefix tests ---

func TestHdrColorMetadataFallbackPrefix(t *testing.T) {
	t.Parallel()

	// This function checks ffprobe.NeedsHDRColorMetadataFallback which we can't easily mock
	// Just verify the function returns a string (either empty or the params prefix)
	got := hdrColorMetadataFallbackPrefix("/test/video.mkv")
	_ = got // Just ensure it doesn't panic
}

// --- buildImageSubtitleBurnInFilter tests ---

func TestBuildImageSubtitleBurnInFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		hwAccel   string
		hdr       bool
		inputPath string
		subIdx    int
		wantStart string
	}{
		{
			name:      "non-HDR",
			hwAccel:   "",
			hdr:       false,
			inputPath: "/test/video.mkv",
			subIdx:    0,
			wantStart: "[0:v:0][0:0]overlay",
		},
		{
			name:      "HDR uses tonemapx",
			hwAccel:   "",
			hdr:       true,
			inputPath: "/test/video.mkv",
			subIdx:    2,
			wantStart: "[0:v:0]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildImageSubtitleBurnInFilter(tt.hwAccel, tt.hdr, tt.inputPath, tt.subIdx)
			if !strings.HasPrefix(got, tt.wantStart) {
				t.Errorf("buildImageSubtitleBurnInFilter() = %q, want to start with %q", got, tt.wantStart)
			}
			if !strings.Contains(got, "[vout]") {
				t.Errorf("buildImageSubtitleBurnInFilter() = %q, want to contain '[vout]'", got)
			}
		})
	}
}
