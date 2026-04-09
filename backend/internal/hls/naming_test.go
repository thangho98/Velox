package hls

import (
	"testing"
)

func TestBuildPrefix(t *testing.T) {
	key := SessionKey{
		StreamSessionID:   "abcd123",
		MediaID:           10,
		FileID:            20,
		SubtitleStreamIdx: -1,
		VideoCopy:         false,
		MaxHeight:         1080,
	}

	prefix := BuildPrefix(key)
	expected := "v2_ssabcd123_10_f20_si-1_vc0_h1080_"
	if prefix != expected {
		t.Errorf("Expected %s, got %s", expected, prefix)
	}

	key.VideoCopy = true
	prefix = BuildPrefix(key)
	expected = "v2_ssabcd123_10_f20_si-1_vc1_h1080_"
	if prefix != expected {
		t.Errorf("Expected %s, got %s", expected, prefix)
	}
}

func TestSegmentTimeRangeAndNumber(t *testing.T) {
	segLength := 6.0

	if n := SegmentNumber(0, segLength); n != 0 {
		t.Errorf("Expected 0, got %d", n)
	}
	if n := SegmentNumber(5.99, segLength); n != 0 {
		t.Errorf("Expected 0, got %d", n)
	}
	if n := SegmentNumber(6.0, segLength); n != 1 {
		t.Errorf("Expected 1, got %d", n)
	}
	if n := SegmentNumber(13.2, segLength); n != 2 {
		t.Errorf("Expected 2, got %d", n)
	}

	s, e := SegmentTimeRange(0, segLength)
	if s != 0 || e != 6.0 {
		t.Errorf("Expected 0-6.0, got %f-%f", s, e)
	}
	s, e = SegmentTimeRange(1, segLength)
	if s != 6.0 || e != 12.0 {
		t.Errorf("Expected 6.0-12.0, got %f-%f", s, e)
	}
}

func TestFilenames(t *testing.T) {
	prefix := "pre_"

	f := SegmentFilename(prefix, "video", 5)
	if f != "pre_video_5.m4s" {
		t.Errorf("SegmentFilename mismatch: %s", f)
	}

	f = InitFilename(prefix, "audio0")
	if f != "pre_audio0_-1.mp4" {
		t.Errorf("InitFilename mismatch: %s", f)
	}

	f = MediaPlaylistFilename(prefix, "video")
	if f != "pre_video.m3u8" {
		t.Errorf("MediaPlaylistFilename mismatch: %s", f)
	}
}
