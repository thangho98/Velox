package service

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/thawng/velox/pkg/subprovider"
)

var subtitleSearchTokenSplitter = regexp.MustCompile(`[^a-z0-9]+`)

func filterAndRankSubtitleSearchResults(
	results []subprovider.Result,
	epInfo *episodeInfo,
	query string,
) []subprovider.Result {
	if len(results) == 0 || epInfo == nil {
		return results
	}

	filtered := make([]subprovider.Result, 0, len(results))
	for _, result := range results {
		if !isExactEpisodeSubtitleMatch(result.Title, epInfo.seasonNumber, epInfo.episodeNumber) {
			continue
		}
		filtered = append(filtered, result)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		left := subtitleSearchScore(filtered[i], query)
		right := subtitleSearchScore(filtered[j], query)
		if left != right {
			return left > right
		}
		return filtered[i].Title < filtered[j].Title
	})

	return filtered
}

func isExactEpisodeSubtitleMatch(title string, seasonNumber, episodeNumber int) bool {
	normalized := normalizeSubtitleSearchText(title)
	if normalized == "" {
		return false
	}

	season := strconv.Itoa(seasonNumber)
	episode := strconv.Itoa(episodeNumber)
	seasonPadded := fmt.Sprintf("%02d", seasonNumber)
	episodePadded := fmt.Sprintf("%02d", episodeNumber)

	exactTokens := []string{
		"s" + season + "e" + episode,
		"s" + seasonPadded + "e" + episodePadded,
		season + "x" + episodePadded,
		seasonPadded + "x" + episodePadded,
		"season " + season + " episode " + episode,
		"season " + seasonPadded + " episode " + episodePadded,
	}
	for _, token := range exactTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}
	return false
}

func subtitleSearchScore(result subprovider.Result, query string) int {
	score := subtitleSearchOverlapScore(query, result.Title) * 100
	score += result.Downloads
	if !result.HearingImpaired {
		score += 25
	}
	switch result.Provider {
	case "subdl":
		score += 10
	case "podnapisi":
		score += 5
	case "bsplayer":
		score += 3
	}
	return score
}

func subtitleSearchOverlapScore(query, title string) int {
	queryTokens := subtitleSearchTokens(query)
	if len(queryTokens) == 0 {
		return 0
	}
	titleTokens := subtitleSearchTokens(title)
	overlap := 0
	for token := range titleTokens {
		if _, ok := queryTokens[token]; ok {
			overlap++
		}
	}
	return overlap
}

func subtitleSearchTokens(value string) map[string]struct{} {
	normalized := normalizeSubtitleSearchText(value)
	parts := subtitleSearchTokenSplitter.Split(normalized, -1)
	tokens := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		if len(part) < 2 {
			continue
		}
		tokens[part] = struct{}{}
	}
	return tokens
}

func normalizeSubtitleSearchText(value string) string {
	replacer := strings.NewReplacer(".", " ", "_", " ", "-", " ")
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}
