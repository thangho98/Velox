package scanner

import (
	"regexp"
	"strings"

	"github.com/thawng/velox/pkg/ffprobe"
)

// Compiled regex patterns for chapter classification (Jellyfin-compatible)
var (
	// Intro patterns: "Intro", "Opening", "OP", "OP1", "Theme", "Title Sequence",
	// "Previously On", "Recap", "Cold Open". Avoids "Introduction to..." false positives.
	introRe = regexp.MustCompile(`^(intro|opening|opening credits|op\s?\d*|theme|title sequence|previously on|recap|cold open)$`)

	// Credits patterns: "Credits", "End Credits", "Closing Credits", "End Titles",
	// "Closing Titles", "Outro", "ED", "ED1", "Ending"
	creditsRe = regexp.MustCompile(`^(credits|end credits|closing credits|end titles|closing titles|outro|ed\s?\d*|ending)$`)

	// Generic chapter name pattern — "Chapter 1", "Chapter 01", etc.
	genericChapterRe = regexp.MustCompile(`^chapter\s*\d+$`)

	nonWordRe    = regexp.MustCompile(`[^\w\s]`)
	multiSpaceRe = regexp.MustCompile(`\s+`)
)

// Timing constraints per Jellyfin intro-skipper
const (
	minIntroDuration   = 15.0  // Minimum 15 seconds for intro
	maxIntroDuration   = 120.0 // Maximum 2 minutes for intro
	minCreditsDuration = 15.0  // Minimum 15 seconds for credits
	maxCreditsDuration = 450.0 // Maximum 7.5 minutes for credits
	maxIntroStart      = 900.0 // Intro must start within first 15 minutes
)

// ExtractChapterMarkers identifies intro/credits markers from ffprobe chapters.
// First tries named chapter matching, then falls back to unnamed chapter heuristics.
func ExtractChapterMarkers(chapters []ffprobe.ChapterInfo) []DetectedMarker {
	if len(chapters) == 0 {
		return nil
	}

	// Try named chapter matching first
	markers := extractNamedChapterMarkers(chapters)
	if len(markers) > 0 {
		return markers
	}

	// Fallback: infer from unnamed/generic chapters (BluRay pattern)
	if len(chapters) >= 2 {
		totalDuration := chapters[len(chapters)-1].EndTime
		return inferFromUnnamedChapters(chapters, totalDuration)
	}

	return nil
}

// extractNamedChapterMarkers matches chapter titles against known patterns.
func extractNamedChapterMarkers(chapters []ffprobe.ChapterInfo) []DetectedMarker {
	var markers []DetectedMarker
	for _, ch := range chapters {
		markerType := classifyChapter(ch.Title)
		if markerType == "" {
			continue
		}
		if !isValidSegment(ch.StartTime, ch.EndTime, markerType) {
			continue
		}
		markers = append(markers, DetectedMarker{
			Type:       markerType,
			StartSec:   ch.StartTime,
			EndSec:     ch.EndTime,
			Source:     "chapter",
			Confidence: 1.0,
			Label:      ch.Title,
		})
	}
	return markers
}

// inferFromUnnamedChapters detects intro/credits from unnamed chapter patterns.
// BluRay discs typically have numbered chapters without meaningful titles.
// Heuristic: first short chapter starting at 0 = intro, last short chapter = credits.
func inferFromUnnamedChapters(chapters []ffprobe.ChapterInfo, totalDuration float64) []DetectedMarker {
	if totalDuration <= 0 {
		return nil
	}

	// Check all chapters are unnamed/generic
	hasNamedChapter := false
	for _, ch := range chapters {
		normalized := normalizeChapterTitle(ch.Title)
		if normalized != "" && !genericChapterRe.MatchString(normalized) {
			hasNamedChapter = true
			break
		}
	}
	if hasNamedChapter {
		return nil // Has meaningful chapter names but none matched — don't guess
	}

	var markers []DetectedMarker

	// First chapter: starts at 0 (±1s tolerance), duration 15-120s → likely intro.
	// Uses source="fingerprint" with low confidence so chromaprint can override
	// with more precise boundaries (chapter includes cold open + theme).
	first := chapters[0]
	firstDur := first.EndTime - first.StartTime
	if first.StartTime <= 1.0 && firstDur >= minIntroDuration && firstDur <= maxIntroDuration {
		markers = append(markers, DetectedMarker{
			Type:       "intro",
			StartSec:   first.StartTime,
			EndSec:     first.EndTime,
			Source:     "fingerprint",
			Confidence: 0.60,
			Label:      "unnamed chapter heuristic",
		})
	}

	// Last chapter: starts > 85% of total duration, duration 15-120s → likely credits.
	// Credits heuristic is more reliable than intro (no "cold ending" ambiguity).
	last := chapters[len(chapters)-1]
	lastDur := last.EndTime - last.StartTime
	if last.StartTime > totalDuration*0.85 && lastDur >= minCreditsDuration && lastDur <= maxIntroDuration {
		markers = append(markers, DetectedMarker{
			Type:       "credits",
			StartSec:   last.StartTime,
			EndSec:     last.EndTime,
			Source:     "fingerprint",
			Confidence: 0.70,
			Label:      "unnamed chapter heuristic",
		})
	}

	return markers
}

// classifyChapter determines if a chapter title indicates intro or credits
func classifyChapter(title string) string {
	normalized := normalizeChapterTitle(title)
	if normalized == "" {
		return ""
	}

	if introRe.MatchString(normalized) {
		return "intro"
	}
	if creditsRe.MatchString(normalized) {
		return "credits"
	}

	return ""
}

// normalizeChapterTitle normalizes a chapter title for comparison
func normalizeChapterTitle(title string) string {
	s := strings.ToLower(title)
	s = nonWordRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	s = multiSpaceRe.ReplaceAllString(s, " ")
	return s
}

// isValidSegment validates that a segment meets timing criteria
func isValidSegment(start, end float64, markerType string) bool {
	if end <= start {
		return false
	}

	duration := end - start

	switch markerType {
	case "intro":
		if duration < minIntroDuration || duration > maxIntroDuration {
			return false
		}
		if start > maxIntroStart {
			return false
		}
	case "credits":
		if duration < minCreditsDuration || duration > maxCreditsDuration {
			return false
		}
	}

	return true
}
