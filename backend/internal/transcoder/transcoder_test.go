package transcoder

import (
	"strings"
	"testing"

	"github.com/thawng/velox/internal/hls"
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
	tcr := New("/tmp/velox-hls", "", 1, false)
	got := tcr.MasterPlaylistPath(42, "sess123", 7, 3, false, 120)

	if !strings.HasSuffix(got, "/42/sssess123_f7_sub3_off120000_master.m3u8") {
		t.Fatalf("unexpected master playlist path: %s", got)
	}
}

func TestCancelTranscodeByStreamSessionIDExcept(t *testing.T) {
	tcr := New("/tmp/velox-hls", "", 1, false)
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

func TestManagerCancelByStreamSessionID(t *testing.T) {
	mgr := &Manager{
		sessions: make(map[hls.SessionKey]*Session),
	}

	viewerA := hls.SessionKey{StreamSessionID: "viewer-a", MediaID: 42, FileID: 7}
	viewerAOtherQuality := hls.SessionKey{StreamSessionID: "viewer-a", MediaID: 42, FileID: 7, MaxHeight: 720}
	viewerB := hls.SessionKey{StreamSessionID: "viewer-b", MediaID: 42, FileID: 7}

	mgr.sessions[viewerA] = &Session{}
	mgr.sessions[viewerAOtherQuality] = &Session{}
	mgr.sessions[viewerB] = &Session{}

	killed := mgr.CancelByStreamSessionID("viewer-a")
	if killed != 2 {
		t.Fatalf("CancelByStreamSessionID killed %d sessions, want 2", killed)
	}
	if _, ok := mgr.sessions[viewerA]; ok {
		t.Fatalf("viewer-a session was not removed")
	}
	if _, ok := mgr.sessions[viewerAOtherQuality]; ok {
		t.Fatalf("viewer-a quality session was not removed")
	}
	if _, ok := mgr.sessions[viewerB]; !ok {
		t.Fatalf("viewer-b session should remain")
	}
}

func TestManagerCancelByMediaID(t *testing.T) {
	mgr := &Manager{
		sessions: make(map[hls.SessionKey]*Session),
	}

	media42 := hls.SessionKey{StreamSessionID: "viewer-a", MediaID: 42, FileID: 7}
	media42OtherViewer := hls.SessionKey{StreamSessionID: "viewer-b", MediaID: 42, FileID: 7}
	media43 := hls.SessionKey{StreamSessionID: "viewer-c", MediaID: 43, FileID: 8}

	mgr.sessions[media42] = &Session{}
	mgr.sessions[media42OtherViewer] = &Session{}
	mgr.sessions[media43] = &Session{}

	killed := mgr.CancelByMediaID(42)
	if killed != 2 {
		t.Fatalf("CancelByMediaID killed %d sessions, want 2", killed)
	}
	if _, ok := mgr.sessions[media42]; ok {
		t.Fatalf("media 42 session was not removed")
	}
	if _, ok := mgr.sessions[media42OtherViewer]; ok {
		t.Fatalf("media 42 second session was not removed")
	}
	if _, ok := mgr.sessions[media43]; !ok {
		t.Fatalf("media 43 session should remain")
	}
}
