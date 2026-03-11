package nfo

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Movie represents movie metadata from NFO file
type Movie struct {
	XMLName       xml.Name  `xml:"movie"`
	Title         string    `xml:"title"`
	OriginalTitle string    `xml:"originaltitle"`
	SortTitle     string    `xml:"sorttitle"`
	Year          int       `xml:"year"`
	Rating        float64   `xml:"rating"`
	Votes         int       `xml:"votes"`
	Plot          string    `xml:"plot"`
	Tagline       string    `xml:"tagline"`
	Runtime       int       `xml:"runtime"`
	MPAA          string    `xml:"mpaa"`
	PlayCount     int       `xml:"playcount"`
	LastPlayed    string    `xml:"lastplayed"`
	ID            string    `xml:"id"` // TMDb ID
	IMDbID        string    `xml:"imdbid"`
	TMDbID        string    `xml:"tmdbid"`
	TVDBID        string    `xml:"tvdbid"`
	Genres        []string  `xml:"genre"`
	Tags          []string  `xml:"tag"`
	Countries     []string  `xml:"country"`
	Studios       []string  `xml:"studio"`
	Directors     []string  `xml:"director"`
	Credits       []string  `xml:"credits"`   // Writers
	Premiered     string    `xml:"premiered"` // YYYY-MM-DD
	Status        string    `xml:"status"`
	Poster        string    `xml:"poster"`
	Fanart        []Fanart  `xml:"fanart>thumb"`
	Actors        []Actor   `xml:"actor"`
	FileInfo      *FileInfo `xml:"fileinfo"`
	Trailer       string    `xml:"trailer"`
	Watched       bool      `xml:"watched"`
}

// TVShow represents TV show metadata from tvshow.nfo
type TVShow struct {
	XMLName       xml.Name `xml:"tvshow"`
	Title         string   `xml:"title"`
	OriginalTitle string   `xml:"showtitle"`
	SortTitle     string   `xml:"sorttitle"`
	Year          int      `xml:"year"`
	Rating        float64  `xml:"rating"`
	Votes         int      `xml:"votes"`
	Plot          string   `xml:"plot"`
	Tagline       string   `xml:"tagline"`
	Runtime       int      `xml:"runtime"`
	MPAA          string   `xml:"mpaa"`
	EpisodeGuide  string   `xml:"episodeguide"`
	ID            string   `xml:"id"` // TMDb ID
	IMDbID        string   `xml:"imdbid"`
	TMDbID        string   `xml:"tmdbid"`
	TVDBID        string   `xml:"tvdbid"`
	Genres        []string `xml:"genre"`
	Tags          []string `xml:"tag"`
	Countries     []string `xml:"country"`
	Studios       []string `xml:"studio"`
	Premiered     string   `xml:"premiered"`
	Status        string   `xml:"status"`
	Poster        string   `xml:"poster"`
	Fanart        []Fanart `xml:"fanart>thumb"`
	Actors        []Actor  `xml:"actor"`
	Seasons       []Season `xml:"namedseason"`
}

// Episode represents episode metadata from episode.nfo
type Episode struct {
	XMLName        xml.Name  `xml:"episodedetails"`
	Title          string    `xml:"title"`
	ShowTitle      string    `xml:"showtitle"`
	Rating         float64   `xml:"rating"`
	Votes          int       `xml:"votes"`
	Plot           string    `xml:"plot"`
	Runtime        int       `xml:"runtime"`
	MPAA           string    `xml:"mpaa"`
	PlayCount      int       `xml:"playcount"`
	LastPlayed     string    `xml:"lastplayed"`
	ID             string    `xml:"id"`
	IMDbID         string    `xml:"imdbid"`
	TMDbID         string    `xml:"tmdbid"`
	TVDBID         string    `xml:"tvdbid"`
	Season         int       `xml:"season"`
	Episode        int       `xml:"episode"`
	DisplaySeason  int       `xml:"displayseason"`
	DisplayEpisode int       `xml:"displayepisode"`
	Aired          string    `xml:"aired"` // YYYY-MM-DD
	Studios        []string  `xml:"studio"`
	Directors      []string  `xml:"director"`
	Credits        []string  `xml:"credits"`
	Actors         []Actor   `xml:"actor"`
	FileInfo       *FileInfo `xml:"fileinfo"`
	Watched        bool      `xml:"watched"`
}

// Supporting structures

type Actor struct {
	Name  string `xml:"name"`
	Role  string `xml:"role"`
	Order int    `xml:"order"`
	Thumb string `xml:"thumb"`
}

type Fanart struct {
	URL     string `xml:",chardata"`
	Colors  string `xml:"colors,attr"`
	Preview string `xml:"preview,attr"`
}

type Season struct {
	Number int    `xml:"number,attr"`
	Name   string `xml:",chardata"`
}

type FileInfo struct {
	StreamDetails *StreamDetails `xml:"streamdetails"`
}

type StreamDetails struct {
	Video    *VideoStream     `xml:"video"`
	Audio    []AudioStream    `xml:"audio"`
	Subtitle []SubtitleStream `xml:"subtitle"`
}

type VideoStream struct {
	Codec    string  `xml:"codec"`
	Aspect   float64 `xml:"aspect"`
	Width    int     `xml:"width"`
	Height   int     `xml:"height"`
	Duration int     `xml:"durationinseconds"`
	Language string  `xml:"language"`
}

type AudioStream struct {
	Codec    string `xml:"codec"`
	Language string `xml:"language"`
	Channels int    `xml:"channels"`
}

type SubtitleStream struct {
	Language string `xml:"language"`
}

// ParseMovie parses a movie.nfo file
func ParseMovie(path string) (*Movie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading nfo file: %w", err)
	}

	var movie Movie
	if err := xml.Unmarshal(data, &movie); err != nil {
		return nil, fmt.Errorf("parsing nfo xml: %w", err)
	}

	return &movie, nil
}

// ParseTVShow parses a tvshow.nfo file
func ParseTVShow(path string) (*TVShow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading nfo file: %w", err)
	}

	var tvshow TVShow
	if err := xml.Unmarshal(data, &tvshow); err != nil {
		return nil, fmt.Errorf("parsing nfo xml: %w", err)
	}

	return &tvshow, nil
}

// ParseEpisode parses an episode.nfo file
func ParseEpisode(path string) (*Episode, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading nfo file: %w", err)
	}

	var episode Episode
	if err := xml.Unmarshal(data, &episode); err != nil {
		return nil, fmt.Errorf("parsing nfo xml: %w", err)
	}

	return &episode, nil
}

// FindMovieNFO searches for a movie NFO file.
// Checks: movie.nfo, then <video_basename>.nfo in the same directory.
func FindMovieNFO(videoPath string) (string, bool) {
	dir := filepath.Dir(videoPath)

	// Standard Kodi name
	movieNFO := filepath.Join(dir, "movie.nfo")
	if _, err := os.Stat(movieNFO); err == nil {
		return movieNFO, true
	}

	// Same name as video file: "Movie Title (2024).nfo"
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
	videoNFO := filepath.Join(dir, base+".nfo")
	if _, err := os.Stat(videoNFO); err == nil {
		return videoNFO, true
	}

	return "", false
}

// FindTVShowNFO searches for tvshow.nfo in the given directory
func FindTVShowNFO(dir string) (string, bool) {
	path := filepath.Join(dir, "tvshow.nfo")
	if _, err := os.Stat(path); err == nil {
		return path, true
	}
	return "", false
}

// FindEpisodeNFO searches for an episode's nfo file
func FindEpisodeNFO(videoPath string) (string, bool) {
	base := strings.TrimSuffix(videoPath, filepath.Ext(videoPath))
	nfoPath := base + ".nfo"
	if _, err := os.Stat(nfoPath); err == nil {
		return nfoPath, true
	}
	return "", false
}

// GetTMDBID extracts the TMDb ID from NFO data
func (m *Movie) GetTMDBID() int {
	// Try explicit tmdbid field first
	if m.TMDbID != "" {
		var id int
		fmt.Sscanf(m.TMDbID, "%d", &id)
		return id
	}

	// Try id field (some scrapers put tmdb id here)
	if m.ID != "" {
		var id int
		fmt.Sscanf(m.ID, "%d", &id)
		return id
	}

	return 0
}

// GetTMDBID extracts the TMDb ID from TV show NFO
func (t *TVShow) GetTMDBID() int {
	if t.TMDbID != "" {
		var id int
		fmt.Sscanf(t.TMDbID, "%d", &id)
		return id
	}

	if t.ID != "" {
		var id int
		fmt.Sscanf(t.ID, "%d", &id)
		return id
	}

	return 0
}

// GetTMDBID extracts the TMDb ID from episode NFO
func (e *Episode) GetTMDBID() int {
	if e.TMDbID != "" {
		var id int
		fmt.Sscanf(e.TMDbID, "%d", &id)
		return id
	}

	if e.ID != "" {
		var id int
		fmt.Sscanf(e.ID, "%d", &id)
		return id
	}

	return 0
}
