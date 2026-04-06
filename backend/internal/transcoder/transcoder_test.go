package transcoder

import (
	"strings"
	"testing"
)

func TestHlsPrefixIncludesSeekOffset(t *testing.T) {
	got := hlsPrefix("sess123", 5, 2, true, 3480.125)

	for _, want := range []string{"sssess123_", "vc", "f5_", "sub2_", "off3480125_"} {
		if !strings.Contains(got, want) {
			t.Fatalf("hlsPrefix missing %q in %q", want, got)
		}
	}
}

func TestHlsPrefixOmitsTinySeekOffset(t *testing.T) {
	got := hlsPrefix("sess123", 5, -1, false, 0.2)
	if strings.Contains(got, "off") {
		t.Fatalf("expected tiny seek offset to be ignored, got %q", got)
	}
}

func TestMasterPlaylistPathIncludesSeekOffsetPrefix(t *testing.T) {
	tcr := New("/tmp/velox-hls", "", 1)
	got := tcr.MasterPlaylistPath(42, "sess123", 7, 3, false, 120)

	if !strings.HasSuffix(got, "/42/sssess123_f7_sub3_off120000_master.m3u8") {
		t.Fatalf("unexpected master playlist path: %s", got)
	}
}

func TestCancelTranscodeByStreamSessionIDExcept(t *testing.T) {
	tcr := New("/tmp/velox-hls", "", 1)
	cancelled := make([]string, 0, 2)
	cancelFor := func(name string) func() {
		return func() {
			cancelled = append(cancelled, name)
		}
	}

	keepPath := "/tmp/velox-hls/42/sskeep_master.m3u8"
	oldPath := "/tmp/velox-hls/42/ssold_master.m3u8"
	otherPath := "/tmp/velox-hls/42/ssother_master.m3u8"

	tcr.active[keepPath] = &transcodeJob{streamSessionID: "viewer-a", cancel: cancelFor("keep")}
	tcr.active[oldPath] = &transcodeJob{streamSessionID: "viewer-a", cancel: cancelFor("old")}
	tcr.active[otherPath] = &transcodeJob{streamSessionID: "viewer-b", cancel: cancelFor("other")}

	killed := tcr.CancelTranscodeByStreamSessionIDExcept("viewer-a", keepPath)
	if killed != 1 {
		t.Fatalf("CancelTranscodeByStreamSessionIDExcept killed %d jobs, want 1", killed)
	}
	if len(cancelled) != 1 || cancelled[0] != "old" {
		t.Fatalf("unexpected cancelled jobs: %#v", cancelled)
	}
}
