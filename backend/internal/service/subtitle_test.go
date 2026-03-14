package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/thawng/velox/internal/model"
)

func TestRankSubtitlesForMediaFilePrefersClosestExternalMatch(t *testing.T) {
	dir := t.TempDir()

	goodPath := filepath.Join(dir, "subdl__good.srt")
	badPath := filepath.Join(dir, "subdl__bad.srt")

	goodContent := "1\n00:00:05,088 --> 00:00:06,463\nHey, new wallet?\n\n2\n00:22:09,453 --> 00:22:11,454\n[English - US - SDH]\n"
	badContent := "1\n00:00:02,252 --> 00:00:05,546\nOkay, so we went to the beach...\n\n2\n00:22:40,984 --> 00:22:42,985\n[English - US - SDH]\n"

	if err := os.WriteFile(goodPath, []byte(goodContent), 0o644); err != nil {
		t.Fatalf("write good subtitle: %v", err)
	}
	if err := os.WriteFile(badPath, []byte(badContent), 0o644); err != nil {
		t.Fatalf("write bad subtitle: %v", err)
	}

	mediaFile := &model.MediaFile{
		FilePath: "/media/Friends.S04E04.The.One.With.The.Ballroom.Dancing.1080p.BluRay.REMUX.AVC.DD.5.1-EPSiLON_Vietsub.mkv",
		Duration: 1336.768,
	}
	subtitles := []model.Subtitle{
		{ID: 627, Language: "en", Codec: "srt", FilePath: badPath, Title: "English (subdl)"},
		{ID: 626, Language: "en", Codec: "srt", FilePath: goodPath, Title: "English (subdl)"},
	}

	ranked := rankSubtitlesForMediaFile(subtitles, mediaFile)
	if len(ranked) != 2 {
		t.Fatalf("expected 2 subtitles, got %d", len(ranked))
	}
	if ranked[0].ID != 626 {
		t.Fatalf("expected best-matching subtitle first, got ID %d", ranked[0].ID)
	}
}

func TestRankSubtitlesForMediaFileKeepsDefaultFirstAcrossLanguages(t *testing.T) {
	mediaFile := &model.MediaFile{FilePath: "/media/example.mkv", Duration: 120}
	subtitles := []model.Subtitle{
		{ID: 1, Language: "eng", Codec: "srt"},
		{ID: 2, Language: "vie", Codec: "srt", IsDefault: true},
	}

	ranked := rankSubtitlesForMediaFile(subtitles, mediaFile)
	if ranked[0].ID != 2 {
		t.Fatalf("expected default subtitle to stay first, got ID %d", ranked[0].ID)
	}
}

func TestFilterMalformedExternalSubtitlesDropsHTMLPayloads(t *testing.T) {
	dir := t.TempDir()

	htmlPath := filepath.Join(dir, "fake.srt")
	goodPath := filepath.Join(dir, "real.srt")

	htmlContent := "<!DOCTYPE html><html><head><title>Subscene</title></head><body>blocked</body></html>"
	goodContent := "1\n00:00:05,000 --> 00:00:06,000\nHello\n"

	if err := os.WriteFile(htmlPath, []byte(htmlContent), 0o644); err != nil {
		t.Fatalf("write fake subtitle: %v", err)
	}
	if err := os.WriteFile(goodPath, []byte(goodContent), 0o644); err != nil {
		t.Fatalf("write real subtitle: %v", err)
	}

	filtered := filterMalformedExternalSubtitles([]model.Subtitle{
		{ID: 1, Language: "en", Codec: "srt", FilePath: htmlPath},
		{ID: 2, Language: "en", Codec: "srt", FilePath: goodPath},
		{ID: 3, Language: "vie", Codec: "subrip", IsEmbedded: true},
	})

	if len(filtered) != 2 {
		t.Fatalf("expected 2 subtitles after filtering, got %d", len(filtered))
	}
	if filtered[0].ID != 2 || filtered[1].ID != 3 {
		t.Fatalf("unexpected subtitles after filtering: %+v", filtered)
	}
}
