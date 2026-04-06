package playback

import "testing"

func TestHWAccelSupportedRequiresCorrectFFmpegCapabilities(t *testing.T) {
	caps := ffmpegCapabilities{
		hwaccels: "hardware acceleration methods:\nvaapi\nqsv\nvideotoolbox\ncuda\n",
		encoders: "encoders:\n h264_vaapi\n h264_qsv\n h264_videotoolbox\n h264_nvenc\n h264_amf\n",
	}

	tests := []struct {
		accel string
		want  bool
	}{
		{accel: "vaapi", want: true},
		{accel: "qsv", want: true},
		{accel: "videotoolbox", want: true},
		{accel: "nvenc", want: true},
		{accel: "amf", want: true},
		{accel: "unknown", want: false},
	}

	for _, tt := range tests {
		if got := hwAccelSupported(tt.accel, caps); got != tt.want {
			t.Fatalf("hwAccelSupported(%q) = %v, want %v", tt.accel, got, tt.want)
		}
	}
}

func TestHWAccelSupportedNVENCAndAMFUseEncoderDiscovery(t *testing.T) {
	caps := ffmpegCapabilities{
		hwaccels: "hardware acceleration methods:\nvaapi\nqsv\nvideotoolbox\ncuda\n",
		encoders: "encoders:\n h264_nvenc\n h264_amf\n",
	}

	if !hwAccelSupported("nvenc", caps) {
		t.Fatal("expected nvenc support to come from encoder discovery")
	}
	if !hwAccelSupported("amf", caps) {
		t.Fatal("expected amf support to come from encoder discovery")
	}
}

func TestHWAccelSupportedVAAPIRequiresEncoderAndHwaccel(t *testing.T) {
	capsMissingEncoder := ffmpegCapabilities{
		hwaccels: "hardware acceleration methods:\nvaapi\n",
		encoders: "encoders:\n",
	}
	if hwAccelSupported("vaapi", capsMissingEncoder) {
		t.Fatal("expected vaapi to require h264_vaapi encoder support")
	}

	capsMissingHWAccel := ffmpegCapabilities{
		hwaccels: "hardware acceleration methods:\n",
		encoders: "encoders:\n h264_vaapi\n",
	}
	if hwAccelSupported("vaapi", capsMissingHWAccel) {
		t.Fatal("expected vaapi to require vaapi hwaccel support")
	}
}
