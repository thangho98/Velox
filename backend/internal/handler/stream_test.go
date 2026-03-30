package handler

import (
	"net/url"
	"strings"
	"testing"

	"github.com/thawng/velox/internal/playback"
)

func TestBuildHLSRedirectURLPreservesAuthAndPlaybackQuery(t *testing.T) {
	original := url.Values{
		"api_key": {"stream-key"},
		"token":   {"abc123"},
		"at":      {"11"},
		"sub":     {"vi"},
		"start":   {"3480.5"},
	}

	got := buildHLSRedirectURL(323, 323, original)

	if !strings.HasPrefix(got, "/api/stream/323/hls/master.m3u8?") {
		t.Fatalf("unexpected redirect path: %s", got)
	}
	if !strings.Contains(got, "fid=323") {
		t.Fatalf("redirect missing fid: %s", got)
	}
	if !strings.Contains(got, "api_key=stream-key") {
		t.Fatalf("redirect missing api_key: %s", got)
	}
	if !strings.Contains(got, "token=abc123") {
		t.Fatalf("redirect missing token: %s", got)
	}
	if !strings.Contains(got, "at=11") {
		t.Fatalf("redirect missing audio track: %s", got)
	}
	if !strings.Contains(got, "start=3480.5") {
		t.Fatalf("redirect missing start offset: %s", got)
	}
	if strings.Contains(got, "sub=vi") {
		t.Fatalf("redirect should not carry subtitle language hint: %s", got)
	}
}

func TestRewriteHLSPlaylistPropagatesToken(t *testing.T) {
	content := []byte("#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=4000000\nf323_q1080.m3u8\n#EXT-X-MEDIA:TYPE=AUDIO,URI=\"audio_1.m3u8\"\nf323_seg_0001.ts\n")
	query := url.Values{
		"api_key": {"stream-key"},
		"token":   {"abc123"},
		"at":      {"11"},
		"si":      {"2"},
		"start":   {"120"},
	}

	got := string(rewriteHLSPlaylist(content, query))

	for _, want := range []string{
		"f323_q1080.m3u8?api_key=stream-key&at=11&si=2&start=120&token=abc123",
		"URI=\"audio_1.m3u8?api_key=stream-key&at=11&si=2&start=120&token=abc123\"",
		"f323_seg_0001.ts?api_key=stream-key&token=abc123",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("rewritten playlist missing %q in %q", want, got)
		}
	}
}

func TestParseStartOffset(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want float64
	}{
		{name: "empty", raw: "", want: 0},
		{name: "invalid", raw: "abc", want: 0},
		{name: "negative", raw: "-5", want: 0},
		{name: "valid", raw: "123.5", want: 123.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseStartOffset(tt.raw); got != tt.want {
				t.Fatalf("parseStartOffset(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestExplicitPlaybackMethod(t *testing.T) {
	tests := []struct {
		raw  string
		want string
	}{
		{raw: "direct", want: string(playback.MethodDirectPlay)},
		{raw: "directstream", want: string(playback.MethodDirectStream)},
		{raw: "unknown", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			if got := string(explicitPlaybackMethod(tt.raw)); got != tt.want {
				t.Fatalf("explicitPlaybackMethod(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}
