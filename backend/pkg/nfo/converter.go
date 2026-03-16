package nfo

import (
	"fmt"
	"strconv"
	"strings"
)

// MediaData holds the data needed to build a movie NFO from DB models.
type MediaData struct {
	Title        string
	SortTitle    string
	Overview     string
	Tagline      string
	ReleaseDate  string // YYYY-MM-DD
	Rating       float64
	TmdbID       *int64
	ImdbID       *string
	TvdbID       *int64
	PosterPath   string
	BackdropPath string
	Genres       []string
	Cast         []CastData
	Directors    []string
	Writers      []string
}

// CastData represents a cast member for NFO generation.
type CastData struct {
	Name        string
	Character   string
	Order       int
	ProfilePath string
}

// SeriesData holds the data needed to build a tvshow NFO from DB models.
type SeriesData struct {
	Title        string
	SortTitle    string
	Overview     string
	Status       string
	Network      string
	FirstAirDate string
	TmdbID       *int64
	ImdbID       *string
	TvdbID       *int64
	PosterPath   string
	BackdropPath string
	Genres       []string
	Cast         []CastData
	Directors    []string
	Writers      []string
}

// EpisodeData holds the data needed to build an episode NFO.
type EpisodeData struct {
	Title         string
	ShowTitle     string
	SeasonNumber  int
	EpisodeNumber int
	Overview      string
	ReleaseDate   string
	Rating        float64
	TmdbID        *int64
	ImdbID        *string
}

// MovieFromData converts MediaData to a Movie NFO struct ready for writing.
func MovieFromData(d MediaData) *Movie {
	m := &Movie{
		Title:     d.Title,
		SortTitle: d.SortTitle,
		Plot:      d.Overview,
		Tagline:   d.Tagline,
		Rating:    d.Rating,
		Premiered: d.ReleaseDate,
		Genres:    d.Genres,
		Directors: d.Directors,
		Credits:   d.Writers,
		Poster:    d.PosterPath,
	}

	// Year from release date
	if len(d.ReleaseDate) >= 4 {
		if y, err := strconv.Atoi(d.ReleaseDate[:4]); err == nil {
			m.Year = y
		}
	}

	// IDs
	if d.TmdbID != nil {
		m.TMDbID = fmt.Sprintf("%d", *d.TmdbID)
		m.ID = m.TMDbID
	}
	if d.ImdbID != nil && *d.ImdbID != "" {
		m.IMDbID = *d.ImdbID
	}
	if d.TvdbID != nil {
		m.TVDBID = fmt.Sprintf("%d", *d.TvdbID)
	}

	// Fanart (backdrop)
	if d.BackdropPath != "" && !strings.HasPrefix(d.BackdropPath, "local://") {
		m.Fanart = []Fanart{{URL: d.BackdropPath}}
	}

	// Cast
	for _, c := range d.Cast {
		m.Actors = append(m.Actors, Actor{
			Name:  c.Name,
			Role:  c.Character,
			Order: c.Order,
			Thumb: c.ProfilePath,
		})
	}

	return m
}

// TVShowFromData converts SeriesData to a TVShow NFO struct ready for writing.
func TVShowFromData(d SeriesData) *TVShow {
	s := &TVShow{
		Title:     d.Title,
		SortTitle: d.SortTitle,
		Plot:      d.Overview,
		Status:    d.Status,
		Premiered: d.FirstAirDate,
		Genres:    d.Genres,
		Poster:    d.PosterPath,
	}

	if len(d.FirstAirDate) >= 4 {
		if y, err := strconv.Atoi(d.FirstAirDate[:4]); err == nil {
			s.Year = y
		}
	}

	if d.Network != "" {
		s.Studios = []string{d.Network}
	}

	if d.TmdbID != nil {
		s.TMDbID = fmt.Sprintf("%d", *d.TmdbID)
		s.ID = s.TMDbID
	}
	if d.ImdbID != nil && *d.ImdbID != "" {
		s.IMDbID = *d.ImdbID
	}
	if d.TvdbID != nil {
		s.TVDBID = fmt.Sprintf("%d", *d.TvdbID)
	}

	if d.BackdropPath != "" && !strings.HasPrefix(d.BackdropPath, "local://") {
		s.Fanart = []Fanart{{URL: d.BackdropPath}}
	}

	for _, c := range d.Cast {
		s.Actors = append(s.Actors, Actor{
			Name:  c.Name,
			Role:  c.Character,
			Order: c.Order,
			Thumb: c.ProfilePath,
		})
	}

	return s
}

// EpisodeFromData converts EpisodeData to an EpisodeNFO struct ready for writing.
func EpisodeFromData(d EpisodeData) *EpisodeNFO {
	e := &EpisodeNFO{
		Title:     d.Title,
		ShowTitle: d.ShowTitle,
		Season:    d.SeasonNumber,
		Episode:   d.EpisodeNumber,
		Plot:      d.Overview,
		Rating:    d.Rating,
		Aired:     d.ReleaseDate,
	}

	if d.TmdbID != nil {
		e.TMDbID = fmt.Sprintf("%d", *d.TmdbID)
	}
	if d.ImdbID != nil && *d.ImdbID != "" {
		e.IMDbID = *d.ImdbID
	}

	return e
}
