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

	// Year pattern
	yearPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)

	// Quality patterns
	qualityPattern = regexp.MustCompile(`\b(4K|2160p|1080p|720p|480p|360p|240p)\b`)

	// Codec patterns
	codecPattern = regexp.MustCompile(`\b(x264|x265|HEVC|AVC|h\.?264|h\.?265|VP9|AV1|MPEG[24])\b`)

	// Release group pattern (usually at end in brackets/parens)
	releaseGroupPattern = regexp.MustCompile(`[-\s]*[\[\(]([^\]\)]+)[\]\)]$`)

	// Separators to split title
	separatorPattern = regexp.MustCompile(`[._\-]+`)

	// Extra whitespace
	multiSpacePattern = regexp.MustCompile(`\s+`)
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

	// Try to extract release group first (from end)
	if match := releaseGroupPattern.FindStringSubmatch(base); match != nil {
		result.ReleaseGroup = match[1]
		base = releaseGroupPattern.ReplaceAllString(base, "")
	}

	// Look for season/episode patterns
	if match := seasonEpisodePattern.FindStringSubmatch(base); match != nil {
		result.MediaType = "episode"
		result.Season, _ = strconv.Atoi(match[1])
		result.Episode, _ = strconv.Atoi(match[2])
		if match[3] != "" {
			result.EndEpisode, _ = strconv.Atoi(match[3])
		}
		base = seasonEpisodePattern.ReplaceAllString(base, "")
	} else if match := seasonEpisodePattern2.FindStringSubmatch(base); match != nil {
		result.MediaType = "episode"
		result.Season, _ = strconv.Atoi(match[1])
		result.Episode, _ = strconv.Atoi(match[2])
		if match[3] != "" {
			result.EndEpisode, _ = strconv.Atoi(match[3])
		}
		base = seasonEpisodePattern2.ReplaceAllString(base, "")
	}

	// Extract quality
	if match := qualityPattern.FindStringSubmatch(base); match != nil {
		result.Quality = match[1]
		base = qualityPattern.ReplaceAllString(base, "")
	}

	// Extract codec
	if match := codecPattern.FindStringSubmatch(base); match != nil {
		result.Codec = strings.ToUpper(match[1])
		base = codecPattern.ReplaceAllString(base, "")
	}

	// Extract year
	if match := yearPattern.FindStringSubmatch(base); match != nil {
		result.Year, _ = strconv.Atoi(match[1])
		base = yearPattern.ReplaceAllString(base, "")
	}

	// Clean up remaining base as title
	title := cleanTitle(base)
	result.Title = title

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

// cleanTitle normalizes a title string
func cleanTitle(s string) string {
	// Replace separators with spaces
	s = separatorPattern.ReplaceAllString(s, " ")

	// Trim whitespace
	s = strings.TrimSpace(s)

	// Remove extra spaces
	s = multiSpacePattern.ReplaceAllString(s, " ")

	// Title case (optional - keeping original for now)
	return s
}

// IsMultiEpisode returns true if this is a multi-episode file
func (p ParsedMedia) IsMultiEpisode() bool {
	return p.EndEpisode > p.Episode && p.EndEpisode > 0
}
