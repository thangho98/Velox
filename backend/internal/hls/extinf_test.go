package hls

import (
	"strings"
	"testing"
)

func TestParseMediaPlaylist(t *testing.T) {
	playlist := `
#EXTM3U
#EXT-X-VERSION:6
#EXT-X-TARGETDURATION:6
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PLAYLIST-TYPE:EVENT
#EXT-X-MAP:URI="v2_ss123_video_-1.mp4"
#EXTINF:6.000000,
v2_ss123_video_0.m4s
#EXTINF:5.808000,
v2_ss123_video_1.m4s
#EXTINF:6.123000,
` // simulate a partial write (race condition)

	extmaps, err := ParseMediaPlaylist([]byte(playlist))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if len(extmaps) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(extmaps))
	}
	if extmaps[0] != 6.000000 {
		t.Errorf("Expected 6.0 for seg 0, got %f", extmaps[0])
	}
	if extmaps[1] != 5.808000 {
		t.Errorf("Expected 5.808 for seg 1, got %f", extmaps[1])
	}
	// The partial #EXTINF:6.123000 should be ignored because no filename follows it
}

func TestMergeExtinf(t *testing.T) {
	existing := map[int]float64{
		0: 6.0,
		1: 6.0,
		2: 6.0,
	}
	newMap := map[int]float64{
		2: 5.5, // simulate an overwrite
		3: 6.0,
		4: 6.0,
	}

	MergeExtinf(existing, newMap)

	if len(existing) != 5 {
		t.Errorf("Expected 5 segments, got %d", len(existing))
	}
	if existing[2] != 5.5 {
		t.Errorf("Expected segment 2 to be overwritten to 5.5, got %f", existing[2])
	}
	if existing[4] != 6.0 {
		t.Errorf("Expected segment 4 to be 6.0")
	}
}

func TestRenderMediaPlaylist_EventMode(t *testing.T) {
	m := map[int]float64{
		10: 6.0,
		11: 5.5,
		12: 6.0,
	}

	opts := RenderOpts{
		ExtinfMap:       m,
		SessionStartSeg: 10,
		TotalDuration:   120.0, // many segments left
		SegLength:       6.0,
		Prefix:          "pre_",
		Kind:            "video",
	}

	res := RenderMediaPlaylist(opts)

	if !strings.Contains(res, "#EXT-X-TARGETDURATION:6") {
		t.Error("Missing TARGETDURATION")
	}
	if !strings.Contains(res, "#EXT-X-MEDIA-SEQUENCE:10") {
		t.Error("Missing MEDIA-SEQUENCE 10")
	}
	if !strings.Contains(res, "#EXT-X-START:TIME-OFFSET=0") {
		t.Error("Missing TIME-OFFSET")
	}
	if !strings.Contains(res, "pre_video_-1.mp4") {
		t.Error("Missing init mp4")
	}
	if !strings.Contains(res, "pre_video_10.m4s") {
		t.Error("Missing seg 10")
	}
	if !strings.Contains(res, "pre_video_11.m4s") {
		t.Error("Missing seg 11")
	}
	if !strings.Contains(res, "pre_video_12.m4s") {
		t.Error("Missing seg 12")
	}
	if !strings.Contains(res, "#EXT-X-PLAYLIST-TYPE:EVENT") {
		t.Error("Missing PLAYLIST-TYPE EVENT")
	}
	if strings.Contains(res, "#EXT-X-ENDLIST") {
		t.Error("Should not have ENDLIST")
	}
}

func TestRenderMediaPlaylist_VODComplete(t *testing.T) {
	opts := RenderOpts{
		ExtinfMap: map[int]float64{
			0: 6.0,
			1: 6.0,
			2: 3.5, // total 15.5
		},
		SessionStartSeg: 0,
		TotalDuration:   15.5,
		SegLength:       6.0,
		Prefix:          "pre_",
		Kind:            "video",
	}

	// lastSegNum will be 15.5 / 6.0 = 2.
	// contiguousEnd is 2. it matches >= lastSegNum
	res := RenderMediaPlaylist(opts)

	if !strings.Contains(res, "#EXT-X-ENDLIST") {
		t.Error("Missing ENDLIST for complete playlist")
	}
	if strings.Contains(res, "#EXT-X-PLAYLIST-TYPE:EVENT") {
		t.Error("Should not have EVENT if complete")
	}
}

func TestRenderMediaPlaylist_GapAvoidance(t *testing.T) {
	m := map[int]float64{
		0: 6.0,
		1: 6.0,
		2: 6.0,
		5: 6.0, // Gap between 2 and 5
		6: 6.0,
	}

	opts := RenderOpts{
		ExtinfMap:       m,
		SessionStartSeg: 0,
		TotalDuration:   100.0,
		SegLength:       6.0,
		Prefix:          "pre_",
		Kind:            "video",
	}

	res := RenderMediaPlaylist(opts)

	if strings.Contains(res, "pre_video_5.m4s") {
		t.Error("Should not render segment 5 because it's past the gap")
	}
	if !strings.Contains(res, "pre_video_2.m4s") {
		t.Error("Should render up to segment 2")
	}
}
