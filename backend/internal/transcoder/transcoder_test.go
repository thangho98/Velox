package transcoder

import (
	"strings"
	"testing"
)

func TestHlsPrefixIncludesSeekOffset(t *testing.T) {
	got := hlsPrefix(5, 2, true, 3480.125)

	for _, want := range []string{"vc", "f5_", "sub2_", "off3480125_"} {
		if !strings.Contains(got, want) {
			t.Fatalf("hlsPrefix missing %q in %q", want, got)
		}
	}
}

func TestHlsPrefixOmitsTinySeekOffset(t *testing.T) {
	got := hlsPrefix(5, -1, false, 0.2)
	if strings.Contains(got, "off") {
		t.Fatalf("expected tiny seek offset to be ignored, got %q", got)
	}
}

func TestMasterPlaylistPathIncludesSeekOffsetPrefix(t *testing.T) {
	tcr := New("/tmp/velox-hls", "", 1)
	got := tcr.MasterPlaylistPath(42, 7, 3, false, 120)

	if !strings.HasSuffix(got, "/42/f7_sub3_off120000_master.m3u8") {
		t.Fatalf("unexpected master playlist path: %s", got)
	}
}
