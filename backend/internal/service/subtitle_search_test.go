package service

import (
	"reflect"
	"testing"

	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/pkg/subprovider"
)

func TestFilterAndRankSubtitleSearchResultsKeepsOnlyExactEpisodeMatches(t *testing.T) {
	results := []subprovider.Result{
		{
			Provider:   "subdl",
			ExternalID: "/subtitle/2508531-697600.zip",
			Title:      "Friends.S04.720p.BluRay.x264-PSYCHD",
			Language:   "en",
			Format:     "srt",
		},
		{
			Provider:   "subdl",
			ExternalID: "/subtitle/2508529-649677.zip",
			Title:      "Friends.S04E02.720p.BluRay.x264-PSYCHD",
			Language:   "en",
			Format:     "srt",
		},
	}

	got := filterAndRankSubtitleSearchResults(results, &episodeInfo{seasonNumber: 4, episodeNumber: 2}, "Friends.S04E02.The.One.With.The.Cat.1080p.BluRay.REMUX.AVC.DD.5.1-EPSiLON_Vietsub")
	want := []subprovider.Result{results[1]}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filterAndRankSubtitleSearchResults(...) = %+v, want %+v", got, want)
	}
}

func TestFilterAndRankSubtitleSearchResultsRanksBestReleaseOverlapFirst(t *testing.T) {
	results := []subprovider.Result{
		{
			Provider:        "subdl",
			ExternalID:      "second",
			Title:           "Friends.S04E02.720p.BluRay.x264-PSYCHD",
			Language:        "en",
			Format:          "srt",
			HearingImpaired: true,
		},
		{
			Provider:        "opensubtitles",
			ExternalID:      "first",
			Title:           "Friends.S04E02.The.One.With.The.Cat.1080p.BluRay.REMUX.AVC.DD.5.1-EPSiLON",
			Language:        "en",
			Format:          "srt",
			HearingImpaired: false,
		},
	}

	got := filterAndRankSubtitleSearchResults(results, &episodeInfo{seasonNumber: 4, episodeNumber: 2}, "Friends.S04E02.The.One.With.The.Cat.1080p.BluRay.REMUX.AVC.DD.5.1-EPSiLON_Vietsub")
	if len(got) != 2 {
		t.Fatalf("expected 2 exact matches, got %d", len(got))
	}
	if got[0].ExternalID != "first" {
		t.Fatalf("expected best release overlap first, got %+v", got)
	}
}

func TestBuildSubdlSearchParamsPrefersCanonicalIDsForEpisodes(t *testing.T) {
	media := &model.Media{
		Title:       "Friends - The One With The Cat",
		ReleaseDate: "1997-09-25",
	}

	params := buildSubdlSearchParams(
		media,
		&episodeInfo{
			seriesTitle:   "Friends",
			seriesTmdbID:  1668,
			seriesImdbID:  "tt0108778",
			seasonNumber:  4,
			episodeNumber: 2,
		},
		"en",
		"",
	)

	if params.FileName != "" {
		t.Fatalf("expected primary params to omit file_name when canonical IDs exist, got %q", params.FileName)
	}
	if params.ImdbID != "tt0108778" || params.TmdbID != 1668 {
		t.Fatalf("expected canonical IDs, got %+v", params)
	}
	if params.SeasonNumber != 4 || params.EpisodeNumber != 2 || params.Type != "tv" {
		t.Fatalf("expected episode metadata, got %+v", params)
	}
}

func TestBuildSubdlSearchParamsUsesFileNameWithoutCanonicalIDs(t *testing.T) {
	media := &model.Media{
		Title:       "Friends - The One With The Cat",
		ReleaseDate: "1997-09-25",
	}

	params := buildSubdlSearchParams(media, &episodeInfo{
		seriesTitle:   "Friends",
		seasonNumber:  4,
		episodeNumber: 2,
	}, "en", "Friends.S04E02.720p.BluRay.x264-PSYCHD")

	if params.FileName != "Friends.S04E02.720p.BluRay.x264-PSYCHD" {
		t.Fatalf("expected filename fallback without canonical IDs, got %+v", params)
	}
}
