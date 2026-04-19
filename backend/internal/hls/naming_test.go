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
		AudioTrackID:      77,
		VideoCopy:         false,
		MaxHeight:         1080,
	}

	prefix := BuildPrefix(key)
	expected := "v2_ssabcd123_10_f20_si-1_at77_vc0_h1080_"
	if prefix != expected {
		t.Errorf("Expected %s, got %s", expected, prefix)
	}

	key.VideoCopy = true
	prefix = BuildPrefix(key)
	expected = "v2_ssabcd123_10_f20_si-1_at77_vc1_h1080_"
	if prefix != expected {
		t.Errorf("Expected %s, got %s", expected, prefix)
	}
}

func TestParseFilenamePlaylist(t *testing.T) {
	filename := "v2_sssess123_1371_f1371_si-1_at363_vc1_h0_video.m3u8"

	key, kind, segNum, err := ParseFilename(filename)
	if err != nil {
		t.Fatalf("ParseFilename returned error: %v", err)
	}
	if key.StreamSessionID != "sess123" {
		t.Fatalf("StreamSessionID = %q, want %q", key.StreamSessionID, "sess123")
	}
	if key.MediaID != 1371 || key.FileID != 1371 {
		t.Fatalf("media/file = %d/%d, want 1371/1371", key.MediaID, key.FileID)
	}
	if key.SubtitleStreamIdx != -1 {
		t.Fatalf("SubtitleStreamIdx = %d, want -1", key.SubtitleStreamIdx)
	}
	if key.AudioTrackID != 363 {
		t.Fatalf("AudioTrackID = %d, want 363", key.AudioTrackID)
	}
	if !key.VideoCopy {
		t.Fatal("VideoCopy = false, want true")
	}
	if key.MaxHeight != 0 {
		t.Fatalf("MaxHeight = %d, want 0", key.MaxHeight)
	}
	if kind != "video" {
		t.Fatalf("kind = %q, want %q", kind, "video")
	}
	if segNum != 0 {
		t.Fatalf("segNum = %d, want 0 for playlist", segNum)
	}
}

func TestParseFilenameSegment(t *testing.T) {
	filename := "v2_sssess123_1371_f1371_si-1_at363_vc0_h720_audio0_12.m4s"

	key, kind, segNum, err := ParseFilename(filename)
	if err != nil {
		t.Fatalf("ParseFilename returned error: %v", err)
	}
	if key.AudioTrackID != 363 {
		t.Fatalf("AudioTrackID = %d, want 363", key.AudioTrackID)
	}
	if key.VideoCopy {
		t.Fatal("VideoCopy = true, want false")
	}
	if key.MaxHeight != 720 {
		t.Fatalf("MaxHeight = %d, want 720", key.MaxHeight)
	}
	if kind != "audio0" {
		t.Fatalf("kind = %q, want %q", kind, "audio0")
	}
	if segNum != 12 {
		t.Fatalf("segNum = %d, want 12", segNum)
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
