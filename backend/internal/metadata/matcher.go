package metadata

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/thawng/velox/pkg/nameparser"
	"github.com/thawng/velox/pkg/nfo"
	"github.com/thawng/velox/pkg/tmdb"
)

// Matcher handles metadata matching between local files and TMDb
type Matcher struct {
	tmdbClient *tmdb.Client
}

// NewMatcher creates a new metadata matcher
func NewMatcher(tmdbClient *tmdb.Client) *Matcher {
	return &Matcher{tmdbClient: tmdbClient}
}

// MatchResult contains the matched metadata
type MatchResult struct {
	Found         bool
	TMDbID        int
	IMDbID        string
	Title         string
	OriginalTitle string
	Overview      string
	ReleaseDate   string
	Year          int
	PosterPath    string
	BackdropPath  string
	StillPath     string // For TV episodes
	Rating        float64
	Genres        []GenreInfo
	Cast          []CastInfo
	Crew          []CrewInfo
	Confidence    float64 // 0.0 - 1.0 matching confidence
	Source        string  // "nfo", "tmdb_exact", "tmdb_search"
}

// GenreInfo represents genre information
type GenreInfo struct {
	ID   int
	Name string
}

// CastInfo represents cast member
type CastInfo struct {
	ID          int
	Name        string
	Character   string
	ProfilePath string
	Order       int
}

// CrewInfo represents crew member
type CrewInfo struct {
	ID          int
	Name        string
	Job         string
	Department  string
	ProfilePath string
}

// TVMatchResult contains TV show specific matching
type TVMatchResult struct {
	MatchResult
	SeriesID      int
	SeasonNumber  int
	EpisodeNumber int
	EpisodeTitle  string
}

// MatchMovie tries to match a movie file with TMDb metadata
func (m *Matcher) MatchMovie(ctx context.Context, parsed nameparser.ParsedMedia, filePath string) (*MatchResult, error) {
	// Step 1: Check for NFO file
	nfoPath, hasNFO := nfo.FindMovieNFO(filePath)
	if hasNFO {
		movieNFO, err := nfo.ParseMovie(nfoPath)
		if err == nil && movieNFO != nil {
			tmdbID := movieNFO.GetTMDBID()
			if tmdbID > 0 {
				// Use NFO's TMDb ID to fetch fresh data
				details, err := m.tmdbClient.GetMovieDetails(ctx, tmdbID)
				if err == nil {
					return m.convertMovieDetails(details, 1.0, "nfo"), nil
				}
			}

			// Return NFO data directly if TMDb fetch fails
			return m.convertMovieNFO(movieNFO), nil
		}
	}

	// Step 2: Search TMDb
	results, err := m.tmdbClient.SearchMovies(ctx, parsed.Title, parsed.Year, 1)
	if err != nil {
		return nil, fmt.Errorf("tmdb search: %w", err)
	}

	if len(results.Results) == 0 {
		return &MatchResult{Found: false}, nil
	}

	// Step 3: Find best match
	bestMatch := m.findBestMovieMatch(results.Results, parsed)
	if bestMatch == nil {
		return &MatchResult{Found: false}, nil
	}

	// Step 4: Get full details
	details, err := m.tmdbClient.GetMovieDetails(ctx, bestMatch.ID)
	if err != nil {
		return nil, fmt.Errorf("tmdb details: %w", err)
	}

	confidence := m.calculateMovieConfidence(bestMatch, parsed)
	return m.convertMovieDetails(details, confidence, "tmdb_search"), nil
}

// MatchTVShow tries to match a TV episode with TMDb metadata
func (m *Matcher) MatchTVShow(ctx context.Context, parsed nameparser.ParsedMedia, filePath string) (*TVMatchResult, error) {
	// Step 1: Check for tvshow.nfo in parent folder or grandparent
	// Standard layout: Show/Season 01/Episode.mkv → tvshow.nfo at Show/
	parentDir := filepath.Dir(filePath)
	tvshowNFOPath, hasTVShowNFO := nfo.FindTVShowNFO(parentDir)
	if !hasTVShowNFO {
		// Check one level up (e.g., Show/Season 01/ → Show/)
		grandparentDir := filepath.Dir(parentDir)
		tvshowNFOPath, hasTVShowNFO = nfo.FindTVShowNFO(grandparentDir)
	}
	if hasTVShowNFO {
		tvshowNFO, err := nfo.ParseTVShow(tvshowNFOPath)
		if err == nil && tvshowNFO != nil {
			tmdbID := tvshowNFO.GetTMDBID()
			if tmdbID > 0 {
				return m.matchTVEpisodeBySeriesID(ctx, tmdbID, parsed, filePath, tvshowNFO)
			}
			return m.convertTVShowNFO(tvshowNFO, parsed), nil
		}
	}

	// Step 2: Check for episode NFO
	episodeNFOPath, hasEpisodeNFO := nfo.FindEpisodeNFO(filePath)
	if hasEpisodeNFO {
		episodeNFO, err := nfo.ParseEpisode(episodeNFOPath)
		if err == nil && episodeNFO != nil {
			tmdbID := episodeNFO.GetTMDBID()
			if tmdbID > 0 {
				// Episode NFO might have series TMDb ID
				// Try to get series details from episode
			}
		}
	}

	// Step 3: Search TMDb for series
	results, err := m.tmdbClient.SearchTV(ctx, parsed.Title, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("tmdb search: %w", err)
	}

	if len(results.Results) == 0 {
		return &TVMatchResult{MatchResult: MatchResult{Found: false}}, nil
	}

	// Step 4: Find best series match
	bestMatch := m.findBestTVMatch(results.Results, parsed)
	if bestMatch == nil {
		return &TVMatchResult{MatchResult: MatchResult{Found: false}}, nil
	}

	// Step 5: Get series details and match episode
	return m.matchTVEpisodeBySeriesID(ctx, bestMatch.ID, parsed, filePath, nil)
}

// matchTVEpisodeBySeriesID matches episode using series TMDb ID
func (m *Matcher) matchTVEpisodeBySeriesID(ctx context.Context, seriesID int, parsed nameparser.ParsedMedia, filePath string, nfoShow *nfo.TVShow) (*TVMatchResult, error) {
	// Get series details
	seriesDetails, err := m.tmdbClient.GetTVDetails(ctx, seriesID)
	if err != nil {
		return nil, fmt.Errorf("tmdb series details: %w", err)
	}

	// Get season details
	seasonNum := parsed.Season
	if seasonNum == 0 && nfoShow != nil {
		// Try to infer from parent directory name
		seasonNum = inferSeasonFromPath(filepath.Dir(filePath))
	}

	if seasonNum == 0 {
		seasonNum = 1 // Default to season 1
	}

	seasonDetails, err := m.tmdbClient.GetTVSeason(ctx, seriesID, seasonNum)
	if err != nil {
		log.Printf("Failed to get season %d for series %d: %v", seasonNum, seriesID, err)
		// Return series info without episode details
		return m.convertSeriesToTVResult(seriesDetails, seasonNum, parsed.Episode), nil
	}

	// Find episode
	episodeNum := parsed.Episode
	if episodeNum == 0 {
		return m.convertSeriesToTVResult(seriesDetails, seasonNum, 0), nil
	}

	var episode *tmdb.EpisodeDetails
	for _, ep := range seasonDetails.Episodes {
		if ep.EpisodeNumber == episodeNum {
			episode = &ep
			break
		}
	}

	if episode == nil {
		return m.convertSeriesToTVResult(seriesDetails, seasonNum, episodeNum), nil
	}

	return m.convertEpisodeDetails(seriesDetails, seasonDetails, episode, nfoShow), nil
}

// findBestMovieMatch finds the best matching movie from search results
func (m *Matcher) findBestMovieMatch(results []tmdb.MovieSummary, parsed nameparser.ParsedMedia) *tmdb.MovieSummary {
	if len(results) == 0 {
		return nil
	}

	var best *tmdb.MovieSummary
	bestScore := -1.0

	for i := range results {
		result := &results[i]
		score := m.calculateMovieMatchScore(result, parsed)

		if score > bestScore {
			bestScore = score
			best = result
		}
	}

	// Require minimum confidence
	if bestScore < 0.5 {
		return nil
	}

	return best
}

// findBestTVMatch finds the best matching TV show from search results
func (m *Matcher) findBestTVMatch(results []tmdb.TVSummary, parsed nameparser.ParsedMedia) *tmdb.TVSummary {
	if len(results) == 0 {
		return nil
	}

	var best *tmdb.TVSummary
	bestScore := -1.0

	for i := range results {
		result := &results[i]
		score := m.calculateTVMatchScore(result, parsed)

		if score > bestScore {
			bestScore = score
			best = result
		}
	}

	if bestScore < 0.5 {
		return nil
	}

	return best
}

// calculateMovieMatchScore calculates matching score for a movie
func (m *Matcher) calculateMovieMatchScore(result *tmdb.MovieSummary, parsed nameparser.ParsedMedia) float64 {
	score := 0.0

	// Title similarity (0.6 weight)
	titleScore := stringSimilarity(strings.ToLower(result.Title), strings.ToLower(parsed.Title))
	score += titleScore * 0.6

	// Year match (0.4 weight)
	if parsed.Year > 0 {
		resultYear := tmdb.GetYear(result.ReleaseDate)
		if resultYear == parsed.Year {
			score += 0.4
		} else if abs(resultYear-parsed.Year) <= 1 {
			score += 0.2 // Off by one year (common with different release dates)
		}
	} else {
		score += 0.2 // No year to compare, partial credit
	}

	return score
}

// calculateTVMatchScore calculates matching score for a TV show
func (m *Matcher) calculateTVMatchScore(result *tmdb.TVSummary, parsed nameparser.ParsedMedia) float64 {
	score := 0.0

	// Title similarity (0.8 weight)
	titleScore := stringSimilarity(strings.ToLower(result.Name), strings.ToLower(parsed.Title))
	score += titleScore * 0.8

	// Year match (0.2 weight)
	if parsed.Year > 0 {
		resultYear := tmdb.GetYear(result.FirstAirDate)
		if resultYear == parsed.Year {
			score += 0.2
		}
	} else {
		score += 0.1
	}

	return score
}

// calculateMovieConfidence calculates final confidence score
func (m *Matcher) calculateMovieConfidence(match *tmdb.MovieSummary, parsed nameparser.ParsedMedia) float64 {
	return m.calculateMovieMatchScore(match, parsed)
}

// Helper functions

func (m *Matcher) convertMovieDetails(details *tmdb.MovieDetails, confidence float64, source string) *MatchResult {
	result := &MatchResult{
		Found:         true,
		TMDbID:        details.ID,
		IMDbID:        details.IMDbID,
		Title:         details.Title,
		OriginalTitle: details.OriginalTitle,
		Overview:      details.Overview,
		ReleaseDate:   details.ReleaseDate,
		Year:          tmdb.GetYear(details.ReleaseDate),
		PosterPath:    details.PosterPath,
		BackdropPath:  details.BackdropPath,
		Rating:        details.VoteAverage,
		Confidence:    confidence,
		Source:        source,
	}

	// Convert genres
	for _, g := range details.Genres {
		result.Genres = append(result.Genres, GenreInfo{ID: g.ID, Name: g.Name})
	}

	// Convert cast and crew
	if details.Credits != nil {
		for _, c := range details.Credits.Cast {
			result.Cast = append(result.Cast, CastInfo{
				ID:          c.ID,
				Name:        c.Name,
				Character:   c.Character,
				ProfilePath: c.ProfilePath,
				Order:       c.Order,
			})
		}

		for _, c := range details.Credits.Crew {
			result.Crew = append(result.Crew, CrewInfo{
				ID:          c.ID,
				Name:        c.Name,
				Job:         c.Job,
				Department:  c.Department,
				ProfilePath: c.ProfilePath,
			})
		}
	}

	return result
}

func (m *Matcher) convertMovieNFO(nfoMovie *nfo.Movie) *MatchResult {
	result := &MatchResult{
		Found:         true,
		Title:         nfoMovie.Title,
		OriginalTitle: nfoMovie.OriginalTitle,
		Overview:      nfoMovie.Plot,
		Year:          nfoMovie.Year,
		PosterPath:    nfoMovie.Poster,
		Rating:        nfoMovie.Rating,
		Confidence:    1.0,
		Source:        "nfo",
	}

	// Parse genres
	for _, g := range nfoMovie.Genres {
		result.Genres = append(result.Genres, GenreInfo{Name: g})
	}

	// Parse cast
	for _, a := range nfoMovie.Actors {
		result.Cast = append(result.Cast, CastInfo{
			Name:      a.Name,
			Character: a.Role,
		})
	}

	return result
}

func (m *Matcher) convertTVShowNFO(nfoShow *nfo.TVShow, parsed nameparser.ParsedMedia) *TVMatchResult {
	result := &TVMatchResult{
		MatchResult: MatchResult{
			Found:         true,
			Title:         nfoShow.Title,
			OriginalTitle: nfoShow.OriginalTitle,
			Overview:      nfoShow.Plot,
			Year:          nfoShow.Year,
			PosterPath:    nfoShow.Poster,
			Rating:        nfoShow.Rating,
			Confidence:    1.0,
			Source:        "nfo",
		},
		SeasonNumber:  parsed.Season,
		EpisodeNumber: parsed.Episode,
	}

	for _, g := range nfoShow.Genres {
		result.Genres = append(result.Genres, GenreInfo{Name: g})
	}

	return result
}

func (m *Matcher) convertSeriesToTVResult(series *tmdb.TVDetails, seasonNum, episodeNum int) *TVMatchResult {
	result := &TVMatchResult{
		MatchResult: MatchResult{
			Found:         true,
			TMDbID:        series.ID,
			Title:         series.Name,
			OriginalTitle: series.OriginalName,
			Overview:      series.Overview,
			ReleaseDate:   series.FirstAirDate,
			Year:          tmdb.GetYear(series.FirstAirDate),
			PosterPath:    series.PosterPath,
			BackdropPath:  series.BackdropPath,
			Rating:        series.VoteAverage,
			Confidence:    0.7,
			Source:        "tmdb_partial",
		},
		SeriesID:      series.ID,
		SeasonNumber:  seasonNum,
		EpisodeNumber: episodeNum,
	}

	for _, g := range series.Genres {
		result.Genres = append(result.Genres, GenreInfo{ID: g.ID, Name: g.Name})
	}

	return result
}

func (m *Matcher) convertEpisodeDetails(series *tmdb.TVDetails, season *tmdb.SeasonDetails, episode *tmdb.EpisodeDetails, nfoShow *nfo.TVShow) *TVMatchResult {
	confidence := 0.9
	if nfoShow != nil {
		confidence = 1.0
	}

	result := &TVMatchResult{
		MatchResult: MatchResult{
			Found:       true,
			TMDbID:      episode.ID,
			Title:       episode.Name,
			Overview:    episode.Overview,
			ReleaseDate: episode.AirDate,
			Year:        tmdb.GetYear(episode.AirDate),
			StillPath:   episode.StillPath,
			Rating:      episode.VoteAverage,
			Confidence:  confidence,
			Source:      "tmdb_search",
		},
		SeriesID:      series.ID,
		SeasonNumber:  episode.SeasonNumber,
		EpisodeNumber: episode.EpisodeNumber,
		EpisodeTitle:  episode.Name,
	}

	// Use series poster if episode has no still
	if result.StillPath == "" {
		result.PosterPath = series.PosterPath
	}

	// Use series genres
	for _, g := range series.Genres {
		result.Genres = append(result.Genres, GenreInfo{ID: g.ID, Name: g.Name})
	}

	// Convert guest stars to cast
	for _, star := range episode.GuestStars {
		result.Cast = append(result.Cast, CastInfo{
			ID:          star.ID,
			Name:        star.Name,
			Character:   star.Character,
			ProfilePath: star.ProfilePath,
		})
	}

	return result
}

// stringSimilarity calculates similarity between two strings (0.0 - 1.0)
func stringSimilarity(a, b string) float64 {
	if a == b {
		return 1.0
	}

	// Normalize: strip punctuation, collapse spaces, lowercase
	a = normalizeString(a)
	b = normalizeString(b)

	if a == b {
		return 0.95
	}

	if strings.Contains(a, b) || strings.Contains(b, a) {
		return 0.85
	}

	// Levenshtein-based similarity
	maxLen := len(a)
	if len(b) > maxLen {
		maxLen = len(b)
	}
	if maxLen == 0 {
		return 1.0
	}

	dist := levenshtein(a, b)
	return 1.0 - float64(dist)/float64(maxLen)
}

// levenshtein computes the edit distance between two strings.
func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Single-row DP
	prev := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(a); i++ {
		curr := make([]int, len(b)+1)
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			ins := curr[j-1] + 1
			del := prev[j] + 1
			sub := prev[j-1] + cost
			curr[j] = min(ins, min(del, sub))
		}
		prev = curr
	}

	return prev[len(b)]
}

func normalizeString(s string) string {
	var result strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result.WriteRune(unicode.ToLower(r))
		} else if unicode.IsSpace(r) {
			result.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(result.String()), " ")
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func inferSeasonFromPath(dirPath string) int {
	// Extract season number from the directory name (not full path)
	// e.g., "Season 1", "Season 01", "S01"
	dir := strings.ToLower(filepath.Base(dirPath))

	if strings.Contains(dir, "season") {
		var season int
		// Find "season" and parse the number after it
		idx := strings.Index(dir, "season")
		fmt.Sscanf(dir[idx:], "season %d", &season)
		return season
	}

	// Try "S01" pattern
	if strings.HasPrefix(dir, "s") && len(dir) >= 2 {
		var season int
		fmt.Sscanf(dir, "s%d", &season)
		if season > 0 {
			return season
		}
	}

	return 0
}
