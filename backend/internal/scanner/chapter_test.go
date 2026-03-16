package scanner

import (
	"testing"

	"github.com/thawng/velox/pkg/ffprobe"
)

func TestClassifyChapter(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		// Intro — standard
		{name: "exact intro", title: "Intro", expected: "intro"},
		{name: "opening", title: "Opening", expected: "intro"},
		{name: "opening credits", title: "Opening Credits", expected: "intro"},
		{name: "theme", title: "Theme", expected: "intro"},
		{name: "title sequence", title: "Title Sequence", expected: "intro"},
		{name: "previously on", title: "Previously On", expected: "intro"},
		{name: "recap", title: "Recap", expected: "intro"},
		{name: "cold open", title: "Cold Open", expected: "intro"},
		{name: "intro with punctuation", title: "* Intro *", expected: "intro"},
		{name: "intro mixed case", title: "OPENING CREDITS", expected: "intro"},
		{name: "intro with extra spaces", title: "  Opening  Credits  ", expected: "intro"},

		// Intro — anime
		{name: "anime OP", title: "OP", expected: "intro"},
		{name: "anime OP1", title: "OP1", expected: "intro"},
		{name: "anime OP 2", title: "OP 2", expected: "intro"},
		{name: "anime OP9", title: "OP9", expected: "intro"},

		// Credits — standard
		{name: "exact credits", title: "Credits", expected: "credits"},
		{name: "end credits", title: "End Credits", expected: "credits"},
		{name: "closing credits", title: "Closing Credits", expected: "credits"},
		{name: "end titles", title: "End Titles", expected: "credits"},
		{name: "closing titles", title: "Closing Titles", expected: "credits"},
		{name: "outro", title: "Outro", expected: "credits"},
		{name: "ending", title: "Ending", expected: "credits"},
		{name: "credits mixed case", title: "END CREDITS", expected: "credits"},

		// Credits — anime
		{name: "anime ED", title: "ED", expected: "credits"},
		{name: "anime ED1", title: "ED1", expected: "credits"},
		{name: "anime ED 3", title: "ED 3", expected: "credits"},

		// Non-matching — must NOT match
		{name: "chapter number", title: "Chapter 1", expected: ""},
		{name: "act one", title: "Act One", expected: ""},
		{name: "scene 5", title: "Scene 5", expected: ""},
		{name: "empty", title: "", expected: ""},
		{name: "intermission", title: "Intermission", expected: ""},
		{name: "introduction to (negative)", title: "Introduction to the Dark Side", expected: ""},
		{name: "editorial (negative)", title: "Editorial", expected: ""},
		{name: "credited cast (negative)", title: "Credited Cast", expected: ""},
		{name: "opening statement (negative)", title: "Opening Statement of Defense", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyChapter(tt.title)
			if got != tt.expected {
				t.Errorf("classifyChapter(%q) = %q, want %q", tt.title, got, tt.expected)
			}
		})
	}
}

func TestNormalizeChapterTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "lowercase", input: "INTRO", expected: "intro"},
		{name: "trim spaces", input: "  Intro  ", expected: "intro"},
		{name: "collapse spaces", input: "Opening  Credits", expected: "opening credits"},
		{name: "remove punctuation", input: "* Intro *", expected: "intro"},
		{name: "remove brackets", input: "[Opening]", expected: "opening"},
		{name: "remove dashes", input: "- Credits -", expected: "credits"},
		{name: "empty string", input: "", expected: ""},
		{name: "only punctuation", input: "***", expected: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeChapterTitle(tt.input)
			if got != tt.expected {
				t.Errorf("normalizeChapterTitle(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsValidSegment(t *testing.T) {
	tests := []struct {
		name       string
		start      float64
		end        float64
		markerType string
		expected   bool
	}{
		// Intro timing: 15-120s, start <= 15min
		{name: "valid intro 90s", start: 0, end: 90, markerType: "intro", expected: true},
		{name: "intro exactly 15s", start: 0, end: 15, markerType: "intro", expected: true},
		{name: "intro exactly 120s", start: 0, end: 120, markerType: "intro", expected: true},
		{name: "intro too short 14s", start: 0, end: 14, markerType: "intro", expected: false},
		{name: "intro too long 121s", start: 0, end: 121, markerType: "intro", expected: false},
		{name: "intro at 15min boundary", start: 900, end: 990, markerType: "intro", expected: true},
		{name: "intro too late 16min", start: 960, end: 1050, markerType: "intro", expected: false},

		// Credits timing: 15-450s, no start constraint
		{name: "valid credits 80s", start: 2500, end: 2580, markerType: "credits", expected: true},
		{name: "credits exactly 15s", start: 2500, end: 2515, markerType: "credits", expected: true},
		{name: "credits exactly 450s", start: 2000, end: 2450, markerType: "credits", expected: true},
		{name: "credits too short 14s", start: 2500, end: 2514, markerType: "credits", expected: false},
		{name: "credits too long 451s", start: 2000, end: 2451, markerType: "credits", expected: false},
		{name: "credits late is OK", start: 3500, end: 3600, markerType: "credits", expected: true},

		// Invalid ranges
		{name: "end <= start", start: 80, end: 80, markerType: "intro", expected: false},
		{name: "reversed", start: 80, end: 10, markerType: "intro", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidSegment(tt.start, tt.end, tt.markerType)
			if got != tt.expected {
				t.Errorf("isValidSegment(%.0f, %.0f, %q) = %v, want %v",
					tt.start, tt.end, tt.markerType, got, tt.expected)
			}
		})
	}
}

func TestExtractChapterMarkers_Named(t *testing.T) {
	tests := []struct {
		name     string
		chapters []ffprobe.ChapterInfo
		wantLen  int
		wantType string
	}{
		{
			name:     "empty chapters",
			chapters: nil,
			wantLen:  0,
		},
		{
			name: "one intro chapter",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 85, Title: "Intro"},
			},
			wantLen:  1,
			wantType: "intro",
		},
		{
			name: "anime OP + ED",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 90, Title: "OP"},
				{ID: 1, StartTime: 90, EndTime: 1300, Title: "Part A"},
				{ID: 2, StartTime: 1300, EndTime: 1390, Title: "ED"},
			},
			wantLen: 2,
		},
		{
			name: "intro + credits",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 85, Title: "Opening"},
				{ID: 1, StartTime: 85, EndTime: 2500, Title: "Chapter 1"},
				{ID: 2, StartTime: 2500, EndTime: 2580, Title: "End Credits"},
			},
			wantLen: 2,
		},
		{
			name: "skip too-short intro (14s)",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 14, Title: "Intro"},
			},
			wantLen: 0,
		},
		{
			name: "skip non-matching titles",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 300, Title: "Chapter 1"},
				{ID: 1, StartTime: 300, EndTime: 600, Title: "Act Two"},
			},
			wantLen: 0, // Falls through to unnamed heuristic but "Act Two" is named
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markers := ExtractChapterMarkers(tt.chapters)
			if len(markers) != tt.wantLen {
				t.Errorf("got %d markers, want %d", len(markers), tt.wantLen)
			}
			if tt.wantType != "" && len(markers) > 0 && markers[0].Type != tt.wantType {
				t.Errorf("first marker type = %q, want %q", markers[0].Type, tt.wantType)
			}
			for _, m := range markers {
				if m.Source != "chapter" {
					t.Errorf("marker source = %q, want 'chapter'", m.Source)
				}
			}
		})
	}
}

func TestExtractChapterMarkers_UnnamedHeuristic(t *testing.T) {
	tests := []struct {
		name       string
		chapters   []ffprobe.ChapterInfo
		wantLen    int
		wantTypes  []string
		wantConfid float64
	}{
		{
			name: "BluRay Friends pattern — intro + credits",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 98.515, Title: "Chapter 1"},
				{ID: 1, StartTime: 98.515, EndTime: 682.307, Title: "Chapter 2"},
				{ID: 2, StartTime: 682.307, EndTime: 1278.277, Title: "Chapter 3"},
				{ID: 3, StartTime: 1278.277, EndTime: 1336.352, Title: "Chapter 4"},
			},
			wantLen:   2,
			wantTypes: []string{"intro", "credits"},
		},
		{
			name: "first chapter too long (3 min) — no intro",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 180, Title: "Chapter 1"},
				{ID: 1, StartTime: 180, EndTime: 1200, Title: "Chapter 2"},
				{ID: 2, StartTime: 1200, EndTime: 1260, Title: "Chapter 3"},
			},
			wantLen:   1, // Only credits detected
			wantTypes: []string{"credits"},
		},
		{
			name: "last chapter too long (3 min) — no credits",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 60, Title: "Chapter 1"},
				{ID: 1, StartTime: 60, EndTime: 1200, Title: "Chapter 2"},
				{ID: 2, StartTime: 1200, EndTime: 1380, Title: "Chapter 3"},
			},
			wantLen:   1, // Only intro detected
			wantTypes: []string{"intro"},
		},
		{
			name: "last chapter not late enough (<85%) — no credits",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 60, Title: "Chapter 1"},
				{ID: 1, StartTime: 60, EndTime: 600, Title: "Chapter 2"},
				{ID: 2, StartTime: 600, EndTime: 660, Title: "Chapter 3"}, // 660s total, 600/660 = 90.9% OK
			},
			wantLen:   2,
			wantTypes: []string{"intro", "credits"},
		},
		{
			name: "has named non-matching chapters — skip heuristic",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 60, Title: "Prologue"},
				{ID: 1, StartTime: 60, EndTime: 1200, Title: "Main Story"},
				{ID: 2, StartTime: 1200, EndTime: 1260, Title: "Epilogue"},
			},
			wantLen: 0, // Named chapters exist but none match → don't guess
		},
		{
			name: "no titles at all — heuristic applies",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 90, Title: ""},
				{ID: 1, StartTime: 90, EndTime: 1200, Title: ""},
				{ID: 2, StartTime: 1200, EndTime: 1260, Title: ""},
			},
			wantLen:   2,
			wantTypes: []string{"intro", "credits"},
		},
		{
			name: "first chapter too short (14s) — no intro",
			chapters: []ffprobe.ChapterInfo{
				{ID: 0, StartTime: 0, EndTime: 14, Title: "Chapter 1"},
				{ID: 1, StartTime: 14, EndTime: 1200, Title: "Chapter 2"},
			},
			wantLen: 0,
		},
		{
			name:     "single chapter — not enough for heuristic",
			chapters: []ffprobe.ChapterInfo{{ID: 0, StartTime: 0, EndTime: 1200, Title: ""}},
			wantLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markers := ExtractChapterMarkers(tt.chapters)
			if len(markers) != tt.wantLen {
				t.Errorf("got %d markers, want %d", len(markers), tt.wantLen)
				for i, m := range markers {
					t.Logf("  marker[%d]: type=%s start=%.1f end=%.1f conf=%.2f", i, m.Type, m.StartSec, m.EndSec, m.Confidence)
				}
				return
			}
			for i, wantType := range tt.wantTypes {
				if i < len(markers) && markers[i].Type != wantType {
					t.Errorf("marker[%d] type = %q, want %q", i, markers[i].Type, wantType)
				}
			}
			if tt.wantConfid > 0 {
				for _, m := range markers {
					if m.Confidence != tt.wantConfid {
						t.Errorf("confidence = %.2f, want %.2f", m.Confidence, tt.wantConfid)
					}
				}
			}
		})
	}
}
