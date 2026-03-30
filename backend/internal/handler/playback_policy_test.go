package handler

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/playback"
	"github.com/thawng/velox/internal/repository"
)

func setupPlaybackPolicyRepo(t *testing.T, playbackMode string) *repository.AppSettingsRepo {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:?_foreign_keys=on")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(`
		CREATE TABLE app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT '',
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		t.Fatalf("create app_settings table: %v", err)
	}

	repo := repository.NewAppSettingsRepo(db)
	if err := repo.Set(context.Background(), model.SettingPlaybackMode, playbackMode); err != nil {
		t.Fatalf("seed playback mode: %v", err)
	}
	return repo
}

func TestApplyAdminPlaybackPolicyUsesFullTranscodeForBrowserAudioMismatch(t *testing.T) {
	repo := setupPlaybackPolicyRepo(t, "direct_play")
	profile := &playback.DeviceProfile{
		Name:                 "Chrome",
		SupportedVideoCodecs: []string{playback.CodecH264},
		SupportedAudioCodecs: []string{playback.CodecAAC},
		SupportedContainers:  []string{playback.ContainerMP4, playback.ContainerHLS},
		SupportsHLS:          true,
	}
	decision := playback.PlaybackDecision{
		Method:           playback.MethodDirectPlay,
		VideoAction:      playback.VideoCopy,
		AudioAction:      playback.AudioCopy,
		SubtitleAction:   playback.SubtitleNone,
		EstimatedBitrate: 8000,
	}

	got := applyAdminPlaybackPolicy(context.Background(), repo, decision, profile, playback.MediaFileInfo{
		VideoCodec: playback.CodecH264,
		AudioCodec: playback.CodecDTS,
		Height:     1080,
		Bitrate:    8000,
	})

	if got.Method != playback.MethodFullTranscode {
		t.Fatalf("Method = %q, want %q", got.Method, playback.MethodFullTranscode)
	}
	if got.VideoAction != playback.VideoTranscode {
		t.Fatalf("VideoAction = %q, want %q", got.VideoAction, playback.VideoTranscode)
	}
	if got.AudioAction != playback.AudioTranscode {
		t.Fatalf("AudioAction = %q, want %q", got.AudioAction, playback.AudioTranscode)
	}
	if !strings.Contains(got.Reason, "reliable HLS seeking") {
		t.Fatalf("Reason = %q, want reliable seek hint", got.Reason)
	}
}
