package transcoder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWriteMasterPlaylistWithAudio(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	tests := []struct {
		name         string
		variants     []AudioVariant
		audioPaths   map[int]string
		prefix       string
		wantContains []string
	}{
		{
			name:       "single audio track",
			variants:   []AudioVariant{{Language: "eng", Name: "English", StreamIndex: 1, IsDefault: true}},
			audioPaths: map[int]string{1: "/tmp/velox-hls/42/audio_1.m3u8"},
			prefix:     "f5_",
			wantContains: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:4",
				"#EXT-X-MEDIA:TYPE=AUDIO",
				`GROUP-ID="audio"`,
				`LANGUAGE="eng"`,
				`NAME="English"`,
				`DEFAULT=YES`,
				`AUTOSELECT=YES`,
				`URI="audio_1.m3u8"`,
				"#EXT-X-STREAM-INF:BANDWIDTH=4000000",
				"f5_video.m3u8",
			},
		},
		{
			name: "multiple audio tracks",
			variants: []AudioVariant{
				{Language: "eng", Name: "English", StreamIndex: 1, IsDefault: true},
				{Language: "jpn", Name: "Japanese", StreamIndex: 2, IsDefault: false},
				{Language: "fre", Name: "French", StreamIndex: 3, IsDefault: false},
			},
			audioPaths: map[int]string{
				1: "/tmp/velox-hls/42/audio_1.m3u8",
				2: "/tmp/velox-hls/42/audio_2.m3u8",
				3: "/tmp/velox-hls/42/audio_3.m3u8",
			},
			prefix: "f5_",
			wantContains: []string{
				"#EXTM3U",
				`LANGUAGE="eng"`,
				`DEFAULT=YES`,
				`LANGUAGE="jpn"`,
				`DEFAULT=NO`,
				`LANGUAGE="fre"`,
			},
		},
		{
			name:       "non-default audio",
			variants:   []AudioVariant{{Language: "jpn", Name: "Japanese", StreamIndex: 2, IsDefault: false}},
			audioPaths: map[int]string{2: "/tmp/velox-hls/42/audio_2.m3u8"},
			prefix:     "test_",
			wantContains: []string{
				`DEFAULT=NO`,
				`AUTOSELECT=NO`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tmpDir, err := os.MkdirTemp("", "velox-hls-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			masterPath := filepath.Join(tmpDir, "master.m3u8")

			err = tcr.writeMasterPlaylistWithAudio(masterPath, tt.variants, tt.audioPaths, tt.prefix)
			if err != nil {
				t.Fatalf("writeMasterPlaylistWithAudio() error = %v", err)
			}

			content, err := os.ReadFile(masterPath)
			if err != nil {
				t.Fatalf("failed to read master playlist: %v", err)
			}

			contentStr := string(content)
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("master playlist = %q, want to contain %q", contentStr, want)
				}
			}
		})
	}
}

func TestWriteABRMasterPlaylist(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	tests := []struct {
		name          string
		variants      []ABRVariant
		playlistNames []string
		wantContains  []string
	}{
		{
			name: "single variant",
			variants: []ABRVariant{
				{Height: 480, Bitrate: 1500, Bandwidth: 1_500_000},
			},
			playlistNames: []string{"f5_q480.m3u8"},
			wantContains: []string{
				"#EXTM3U",
				"#EXT-X-VERSION:4",
				"#EXT-X-STREAM-INF:BANDWIDTH=1500000",
				`RESOLUTION=854x480`,
				`CODECS="avc1.640028,mp4a.40.2"`,
				"f5_q480.m3u8",
			},
		},
		{
			name: "multiple variants",
			variants: []ABRVariant{
				{Height: 480, Bitrate: 1500, Bandwidth: 1_500_000},
				{Height: 720, Bitrate: 4000, Bandwidth: 4_000_000},
				{Height: 1080, Bitrate: 8000, Bandwidth: 8_000_000},
			},
			playlistNames: []string{"f5_q480.m3u8", "f5_q720.m3u8", "f5_q1080.m3u8"},
			wantContains: []string{
				"#EXTM3U",
				`RESOLUTION=854x480`,
				`RESOLUTION=1280x720`,
				`RESOLUTION=1920x1080`,
				"BANDWIDTH=1500000",
				"BANDWIDTH=4000000",
				"BANDWIDTH=8000000",
			},
		},
		{
			name: "4K variant",
			variants: []ABRVariant{
				{Height: 2160, Bitrate: 25000, Bandwidth: 25_000_000},
			},
			playlistNames: []string{"f5_q2160.m3u8"},
			wantContains: []string{
				`RESOLUTION=3840x2160`,
				"BANDWIDTH=25000000",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp dir
			tmpDir, err := os.MkdirTemp("", "velox-hls-test-*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			masterPath := filepath.Join(tmpDir, "abr_master.m3u8")

			err = tcr.writeABRMasterPlaylist(masterPath, tt.variants, tt.playlistNames)
			if err != nil {
				t.Fatalf("writeABRMasterPlaylist() error = %v", err)
			}

			content, err := os.ReadFile(masterPath)
			if err != nil {
				t.Fatalf("failed to read ABR master playlist: %v", err)
			}

			contentStr := string(content)
			for _, want := range tt.wantContains {
				if !strings.Contains(contentStr, want) {
					t.Errorf("ABR master playlist = %q, want to contain %q", contentStr, want)
				}
			}
		})
	}
}

// Test isHLSComplete tests the isHLSComplete function with temp files.
func TestIsHLSComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name:     "complete playlist",
			content:  "#EXTM3U\n#EXT-X-VERSION:4\n#EXT-X-ENDLIST\n",
			expected: true,
		},
		{
			name:     "missing ENDLIST",
			content:  "#EXTM3U\n#EXT-X-VERSION:4\n",
			expected: false,
		},
		{
			name:     "empty file",
			content:  "",
			expected: false,
		},
		{
			name:     "only ENDLIST",
			content:  "#EXT-X-ENDLIST\n",
			expected: true,
		},
		{
			name:     "ENLIST in middle not at end",
			content:  "#EXTM3U\n#EXT-X-ENDLIST\n#EXT-X-STREAM-INF\n",
			expected: true, // contains ENDLIST so it's "complete"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "hls-test-*.m3u8")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.WriteString(tt.content); err != nil {
				t.Fatalf("failed to write to temp file: %v", err)
			}
			tmpFile.Close()

			got := isHLSComplete(tmpFile.Name())
			if got != tt.expected {
				t.Errorf("isHLSComplete() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTranscoder_HLSDir(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	got := tcr.HLSDir(42)
	expected := "/tmp/velox-hls/42"
	if got != expected {
		t.Errorf("HLSDir(42) = %q, want %q", got, expected)
	}
}

func TestTranscoder_ABRMasterPath(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	got := tcr.ABRMasterPath(42, 5)
	expected := "/tmp/velox-hls/42/f5_abr_master.m3u8"
	if got != expected {
		t.Errorf("ABRMasterPath(42, 5) = %q, want %q", got, expected)
	}
}

func TestTranscoder_MasterPlaylistPath(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	tests := []struct {
		name        string
		mediaID     int64
		sessionID   string
		fileID      int64
		subIdx      int
		videoCopy   bool
		startOffset float64
		wantSuffix  string
	}{
		{
			name:        "basic",
			mediaID:     42,
			sessionID:   "",
			fileID:      5,
			subIdx:      -1,
			videoCopy:   false,
			startOffset: 0,
			wantSuffix:  "/42/f5_master.m3u8",
		},
		{
			name:        "with session",
			mediaID:     42,
			sessionID:   "sess123",
			fileID:      5,
			subIdx:      -1,
			videoCopy:   false,
			startOffset: 0,
			wantSuffix:  "/42/sssess123_f5_master.m3u8",
		},
		{
			name:        "video copy",
			mediaID:     42,
			sessionID:   "",
			fileID:      5,
			subIdx:      -1,
			videoCopy:   true,
			startOffset: 0,
			wantSuffix:  "/42/vcf5_master.m3u8",
		},
		{
			name:        "with subtitle",
			mediaID:     42,
			sessionID:   "",
			fileID:      5,
			subIdx:      2,
			videoCopy:   false,
			startOffset: 0,
			wantSuffix:  "/42/f5_sub2_master.m3u8",
		},
		{
			name:        "with seek offset",
			mediaID:     42,
			sessionID:   "",
			fileID:      5,
			subIdx:      -1,
			videoCopy:   false,
			startOffset: 120.5,
			wantSuffix:  "/42/f5_off120500_master.m3u8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tcr.MasterPlaylistPath(tt.mediaID, tt.sessionID, tt.fileID, tt.subIdx, tt.videoCopy, tt.startOffset)
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("MasterPlaylistPath() = %q, want to end with %q", got, tt.wantSuffix)
			}
		})
	}
}

func TestTranscoder_SegmentPath(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	got := tcr.SegmentPath(42, "seg_0000.m4s")
	expected := "/tmp/velox-hls/42/seg_0000.m4s"
	if got != expected {
		t.Errorf("SegmentPath(42, \"seg_0000.m4s\") = %q, want %q", got, expected)
	}
}

func TestTranscoder_ActiveCount(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	// Initially empty
	if count := tcr.ActiveCount(); count != 0 {
		t.Errorf("ActiveCount() = %d, want 0", count)
	}
}

func TestTranscoder_TryActiveCount(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	// Initially should be 0
	if count := tcr.TryActiveCount(); count != 0 {
		t.Errorf("TryActiveCount() = %d, want 0", count)
	}
}

func TestTranscoder_CleanupOlderThan(t *testing.T) {
	t.Parallel()

	// Create temp output dir
	tmpDir, err := os.MkdirTemp("", "velox-cleanup-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tcr := New(tmpDir, "", 1, false)

	// Create old directory
	oldDir := filepath.Join(tmpDir, "100")
	if err := os.MkdirAll(oldDir, 0755); err != nil {
		t.Fatalf("failed to create old dir: %v", err)
	}

	// Set mod time to 2 hours ago
	twoHoursAgo := time.Now().Add(-2 * time.Hour)
	os.Chtimes(oldDir, twoHoursAgo, twoHoursAgo)

	// Create recent directory
	recentDir := filepath.Join(tmpDir, "200")
	if err := os.MkdirAll(recentDir, 0755); err != nil {
		t.Fatalf("failed to create recent dir: %v", err)
	}

	// Cleanup files older than 1 hour
	err = tcr.CleanupOlderThan(1 * time.Hour)
	if err != nil {
		t.Errorf("CleanupOlderThan() error = %v", err)
	}

	// Old directory should be removed
	if _, err := os.Stat(oldDir); !os.IsNotExist(err) {
		t.Errorf("old dir %s still exists after cleanup", oldDir)
	}

	// Recent directory should still exist
	if _, err := os.Stat(recentDir); os.IsNotExist(err) {
		t.Errorf("recent dir %s was incorrectly removed", recentDir)
	}
}

func TestTranscoder_CancelTranscode(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	// Cancel with no active jobs should return 0
	killed := tcr.CancelTranscode(42)
	if killed != 0 {
		t.Errorf("CancelTranscode(42) = %d, want 0", killed)
	}
}

func TestTranscoder_CancelTranscodeByStreamSessionID(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	// Cancel with empty session ID should return 0
	killed := tcr.CancelTranscodeByStreamSessionID("")
	if killed != 0 {
		t.Errorf("CancelTranscodeByStreamSessionID(\"\") = %d, want 0", killed)
	}
}

func TestTranscoder_TouchJobByMediaID(t *testing.T) {
	t.Parallel()

	tcr := New("/tmp/velox-hls", "", 1, false)

	// Should not panic
	tcr.TouchJobByMediaID(42)
}

func TestSupportsSubtitleBurnIn(t *testing.T) {
	t.Parallel()

	// Just ensure the function doesn't panic
	_ = SupportsSubtitleBurnIn()
}
