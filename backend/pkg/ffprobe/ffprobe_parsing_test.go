package ffprobe

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsTextBasedSubtitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		codec  string
		isText bool
	}{
		{"subrip", true},
		{"ass", true},
		{"ssa", true},
		{"webvtt", true},
		{"mov_text", true},
		{"hdmv_pgs_subtitle", false},
		{"dvd_subtitle", false},
		{"eac3", false},
		{"h264", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.codec, func(t *testing.T) {
			t.Parallel()
			got := IsTextBasedSubtitle(tt.codec)
			assert.Equal(t, tt.isText, got)
		})
	}
}

func TestIsImageBasedSubtitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		codec   string
		isImage bool
	}{
		{"hdmv_pgs_subtitle", true},
		{"dvd_subtitle", true},
		{"dvb_subtitle", true},
		{"xsub", true},
		{"subrip", false},
		{"ass", false},
		{"eac3", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.codec, func(t *testing.T) {
			t.Parallel()
			got := IsImageBasedSubtitle(tt.codec)
			assert.Equal(t, tt.isImage, got)
		})
	}
}

func TestParseFrameRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		rate string
		want float64
	}{
		{"24000/1001", 23.976},
		{"25/1", 25.0},
		{"30000/1001", 29.97},
		{"50/1", 50.0},
		{"24/1", 24.0},
		{"0/0", 0},
		{"", 0},
		{"invalid", 0},
		{"23.976", 23.976},
		{"30", 30.0},
	}

	for _, tt := range tests {
		t.Run(tt.rate, func(t *testing.T) {
			t.Parallel()
			got := parseFrameRate(tt.rate)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestParseTimeBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		timeBase string
		ticks    int64
		want     float64
	}{
		{"1/1000", 5000, 5.0},
		{"1/44100", 44100, 1.0},
		{"1/1000", 0, 0},
		{"invalid", 1000, 0},
		{"1/0", 1000, 0},
		{"", 1000, 0},
	}

	for _, tt := range tests {
		t.Run(tt.timeBase, func(t *testing.T) {
			t.Parallel()
			got := parseTimeBase(tt.timeBase, tt.ticks)
			assert.InDelta(t, tt.want, got, 0.01)
		})
	}
}

func TestIsUnknownOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value     string
		isUnknown bool
	}{
		{"", true},
		{"unknown", true},
		{"UNKNOWN", true},
		{"Unknown", true},
		{"unknown ", true},
		{" reserved", true},
		{"RESERVED", true},
		{"bt2020", false},
		{"bt709", false},
		{"smpte2084", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			got := isUnknownOrEmpty(tt.value)
			assert.Equal(t, tt.isUnknown, got)
		})
	}
}

func TestParseChapters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		chapters []Chapter
		wantLen  int
	}{
		{
			name:     "empty_chapters",
			chapters: []Chapter{},
			wantLen:  0,
		},
		{
			name:     "nil_chapters",
			chapters: nil,
			wantLen:  0,
		},
		{
			name: "single_chapter",
			chapters: []Chapter{
				{
					ID:        0,
					StartTime: "10.5",
					EndTime:   "60.0",
					Tags:      ChapterTags{Title: "Intro"},
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple_chapters",
			chapters: []Chapter{
				{ID: 0, StartTime: "0", EndTime: "60", Tags: ChapterTags{Title: "Chapter 1"}},
				{ID: 1, StartTime: "60", EndTime: "120", Tags: ChapterTags{Title: "Chapter 2"}},
				{ID: 2, StartTime: "120", EndTime: "180", Tags: ChapterTags{Title: "Chapter 3"}},
			},
			wantLen: 3,
		},
		{
			name: "chapter_with_timebase_fallback",
			chapters: []Chapter{
				{
					ID:       0,
					TimeBase: "1/1000",
					Start:    10000,
					End:      60000,
					Tags:     ChapterTags{Title: "Chapter 1"},
				},
			},
			wantLen: 1,
		},
		{
			name: "chapter_with_empty_title",
			chapters: []Chapter{
				{ID: 0, StartTime: "0", EndTime: "60", Tags: ChapterTags{Title: ""}},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := parseChapters(tt.chapters)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

func TestParseChapters_Values(t *testing.T) {
	t.Parallel()

	chapters := []Chapter{
		{
			ID:        0,
			StartTime: "10.500",
			EndTime:   "60.000",
			Tags:      ChapterTags{Title: "Intro"},
		},
	}

	got := parseChapters(chapters)
	if len(got) > 0 {
		assert.Equal(t, 0, got[0].ID)
		assert.Equal(t, "Intro", got[0].Title)
		assert.InDelta(t, 10.5, got[0].StartTime, 0.001)
		assert.InDelta(t, 60.0, got[0].EndTime, 0.001)
	}
}

func TestParseChapters_TimeBaseFallback(t *testing.T) {
	t.Parallel()

	chapters := []Chapter{
		{
			ID:       0,
			TimeBase: "1/1000",
			Start:    10500,
			End:      60000,
			Tags:     ChapterTags{Title: "Chapter 1"},
		},
	}

	got := parseChapters(chapters)
	require.Len(t, got, 1)
	assert.InDelta(t, 10.5, got[0].StartTime, 0.001)
	assert.InDelta(t, 60.0, got[0].EndTime, 0.001)
}
