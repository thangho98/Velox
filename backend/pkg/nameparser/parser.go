package nameparser

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ParsedMedia represents the result of parsing a filename
type ParsedMedia struct {
	Title        string `json:"title"`
	EpisodeTitle string `json:"episode_title,omitempty"` // Episode name (part after SxxExx)
	Year         int    `json:"year"`
	Season       int    `json:"season"`      // -1 if not a series
	Episode      int    `json:"episode"`     // -1 if not a series
	EndEpisode   int    `json:"end_episode"` // For multi-episode files, -1 otherwise
	MediaType    string `json:"media_type"`  // "movie" | "episode"
	Quality      string `json:"quality"`     // "1080p", "4K", etc.
	Codec        string `json:"codec"`       // "x264", "x265", "HEVC"
	ReleaseGroup string `json:"release_group"`
}

var (
	// Patterns for series detection
	seasonEpisodePattern  = regexp.MustCompile(`[Ss](\d{1,2})[Ee](\d{1,4})(?:[-Ee]?(\d{1,4}))?`)
	seasonEpisodePattern2 = regexp.MustCompile(`(\d{1,2})[xX](\d{1,4})(?:[-]?(\d{1,4}))?`)
	seasonOnlyPattern     = regexp.MustCompile(`[Ss]eason\s*(\d{1,2})`)
	episodeOnlyPattern    = regexp.MustCompile(`[Ee]pisode\s*(\d{1,4})`)

	// Year pattern — matches (2000) or standalone 2000
	yearPattern       = regexp.MustCompile(`\(?(19\d{2}|20\d{2})\)?`)
	yearStrictPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

	// Quality patterns
	qualityPattern = regexp.MustCompile(`(?i)\b(4K|2160p|1080p|720p|480p|360p|240p)\b`)

	// Codec patterns
	codecPattern = regexp.MustCompile(`(?i)\b(x264|x265|HEVC|AVC|h\.?264|h\.?265|VP9|AV1|MPEG[24])\b`)

	// Release group pattern (usually at end in brackets/parens)
	releaseGroupPattern = regexp.MustCompile(`[-\s]*[\[\(]([^\]\)]+)[\]\)]$`)

	// Junk words — source/encoding tags commonly found in release filenames
	junkPattern = regexp.MustCompile(`(?i)\b(AMZN|NF|DSNP|HMAX|ATVP|PCOK|PMTP|WEB[-.]?DL|WEBRip|BluRay|BDRip|BRRip|HDRip|DVDRip|REMUX|PROPER|REPACK|INTERNAL|DTS|DDP?5\.?1|AAC|EAC3|AC3|FLAC|ATMOS|10bit|8bit|HDR|HDR10|DV|DoVi|SDR|Hybrid)\b`)

	// Separators to split title
	separatorPattern = regexp.MustCompile(`[._]+`)

	// Extra whitespace
	multiSpacePattern = regexp.MustCompile(`\s+`)

	// Empty parentheses/brackets
	emptyBracketsPattern = regexp.MustCompile(`[\(\[]\s*[\)\]]`)
)

// Parse extracts metadata from a filename
func Parse(filename string) ParsedMedia {
	result := ParsedMedia{
		Year:       0,
		Season:     0,
		Episode:    0,
		EndEpisode: 0,
		MediaType:  "movie",
	}

	// Get basename without extension
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, filepath.Ext(base))

	// Extract quality and codec from the full string first (before release group removal)
	if match := qualityPattern.FindStringSubmatch(base); match != nil {
		result.Quality = match[1]
		if strings.EqualFold(result.Quality, "4k") {
			result.Quality = "4K"
		}
	}
	if match := codecPattern.FindStringSubmatch(base); match != nil {
		result.Codec = strings.ToUpper(strings.ReplaceAll(match[1], ".", ""))
	}

	// Try to extract release group (from end, in parens/brackets)
	if match := releaseGroupPattern.FindStringSubmatch(base); match != nil {
		group := match[1]
		// If the "release group" contains quality/codec/junk, it's not a real group name —
		// extract just the last word as group, or skip if it's all junk
		if qualityPattern.MatchString(group) || codecPattern.MatchString(group) || junkPattern.MatchString(group) {
			// Try to find a clean group name (last non-junk word)
			words := strings.Fields(group)
			for i := len(words) - 1; i >= 0; i-- {
				w := words[i]
				if !qualityPattern.MatchString(w) && !codecPattern.MatchString(w) && !junkPattern.MatchString(w) {
					result.ReleaseGroup = w
					break
				}
			}
		} else {
			result.ReleaseGroup = group
		}
		base = releaseGroupPattern.ReplaceAllString(base, "")
	}

	// Look for season/episode patterns — split into before (series title) and after (episode title + junk)
	var beforeSE, afterSE string
	var sePattern *regexp.Regexp

	if loc := seasonEpisodePattern.FindStringIndex(base); loc != nil {
		sePattern = seasonEpisodePattern
		beforeSE = base[:loc[0]]
		afterSE = base[loc[1]:]
		match := seasonEpisodePattern.FindStringSubmatch(base)
		result.MediaType = "episode"
		result.Season, _ = strconv.Atoi(match[1])
		result.Episode, _ = strconv.Atoi(match[2])
		if match[3] != "" {
			result.EndEpisode, _ = strconv.Atoi(match[3])
		}
	} else if loc := seasonEpisodePattern2.FindStringIndex(base); loc != nil {
		sePattern = seasonEpisodePattern2
		beforeSE = base[:loc[0]]
		afterSE = base[loc[1]:]
		match := seasonEpisodePattern2.FindStringSubmatch(base)
		result.MediaType = "episode"
		result.Season, _ = strconv.Atoi(match[1])
		result.Episode, _ = strconv.Atoi(match[2])
		if match[3] != "" {
			result.EndEpisode, _ = strconv.Atoi(match[3])
		}
	}
	_ = sePattern

	if result.MediaType == "episode" {
		// Series title = part before SxxExx
		result.Title = cleanSeriesTitle(beforeSE)

		// Episode title = part after SxxExx, before quality/codec junk
		result.EpisodeTitle = extractEpisodeTitle(afterSE)

		// Extract year from the before part
		if match := yearStrictPattern.FindStringSubmatch(beforeSE); match != nil {
			result.Year, _ = strconv.Atoi(match[1])
		}
	} else {
		// Movie: extract year, then clean the whole thing
		if match := yearStrictPattern.FindStringSubmatch(base); match != nil {
			result.Year, _ = strconv.Atoi(match[1])
		}

		// For movies, title is everything before year or quality/codec markers
		result.Title = cleanMovieTitle(base, result.Year)
	}

	return result
}

// ParseWithParents parses filename with parent folder context (for series)
func ParseWithParents(filePath string) ParsedMedia {
	result := Parse(filePath)

	// If not detected as episode but parent folder suggests series structure
	if result.MediaType == "movie" {
		dir := filepath.Dir(filePath)
		parent := filepath.Base(dir)

		// Check if parent folder has season info
		if match := seasonOnlyPattern.FindStringSubmatch(parent); match != nil {
			result.MediaType = "episode"
			result.Season, _ = strconv.Atoi(match[1])
		}
	}

	return result
}

// cleanSeriesTitle extracts a clean series name from the part before SxxExx.
// E.g. "Malcolm in the Middle (2000) - " → "Malcolm in the Middle"
func cleanSeriesTitle(s string) string {
	// Replace file separators (dots, underscores) with spaces
	s = separatorPattern.ReplaceAllString(s, " ")

	// Remove year with optional parens: (2000) or 2000
	s = yearPattern.ReplaceAllString(s, "")

	// Remove empty brackets
	s = emptyBracketsPattern.ReplaceAllString(s, "")

	// Remove trailing dashes/whitespace
	s = strings.TrimRight(s, " -–—")
	s = strings.TrimSpace(s)
	s = multiSpacePattern.ReplaceAllString(s, " ")
	return s
}

// extractEpisodeTitle extracts episode name from after SxxExx.
// Input like " - Pilot (1080p AMZN WEB-DL x265" → "Pilot"
func extractEpisodeTitle(s string) string {
	// Replace file separators
	s = separatorPattern.ReplaceAllString(s, " ")

	// Strip leading dashes/whitespace
	s = strings.TrimLeft(s, " -–—")

	// Remove everything from the first quality/codec/junk marker onward
	// Find the earliest position of any junk marker
	cutPos := len(s)
	for _, pat := range []*regexp.Regexp{qualityPattern, codecPattern, junkPattern} {
		if loc := pat.FindStringIndex(s); loc != nil && loc[0] < cutPos {
			cutPos = loc[0]
		}
	}
	s = s[:cutPos]

	// Remove any remaining parenthesized content (leftover source tags)
	s = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(s, "")
	s = emptyBracketsPattern.ReplaceAllString(s, "")

	s = strings.TrimRight(s, " -–—")
	s = strings.TrimSpace(s)
	s = multiSpacePattern.ReplaceAllString(s, " ")
	return s
}

// cleanMovieTitle extracts movie title, cutting before year or quality markers.
func cleanMovieTitle(base string, year int) string {
	s := separatorPattern.ReplaceAllString(base, " ")

	// Cut at year position if found
	if year > 0 {
		yearStr := strconv.Itoa(year)
		if idx := strings.Index(s, yearStr); idx > 0 {
			s = s[:idx]
		}
	} else {
		// Cut at first quality/codec/junk marker
		cutPos := len(s)
		for _, pat := range []*regexp.Regexp{qualityPattern, codecPattern, junkPattern} {
			if loc := pat.FindStringIndex(s); loc != nil && loc[0] < cutPos {
				cutPos = loc[0]
			}
		}
		s = s[:cutPos]
	}

	s = emptyBracketsPattern.ReplaceAllString(s, "")
	s = strings.TrimRight(s, " -–—(")
	s = strings.TrimSpace(s)
	s = multiSpacePattern.ReplaceAllString(s, " ")
	return s
}

// IsMultiEpisode returns true if this is a multi-episode file
func (p ParsedMedia) IsMultiEpisode() bool {
	return p.EndEpisode > p.Episode && p.EndEpisode > 0
}
