package tvmaze

import "errors"

// ErrNotFound is returned when TVmaze has no data for the given ID/query.
var ErrNotFound = errors.New("tvmaze: not found")

// Show represents a TV show from TVmaze.
type Show struct {
	ID             int       `json:"id"`
	URL            string    `json:"url"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`           // "Scripted", "Animation", etc.
	Language       string    `json:"language"`       // "English", "Japanese", etc.
	Genres         []string  `json:"genres"`         // ["Drama", "Thriller"]
	Status         string    `json:"status"`         // "Running", "Ended", "To Be Determined", "In Development"
	Runtime        int       `json:"runtime"`        // minutes
	AverageRuntime int       `json:"averageRuntime"` // minutes
	Premiered      string    `json:"premiered"`      // "2013-06-24"
	Ended          string    `json:"ended"`          // "2015-09-10" or empty
	OfficialSite   string    `json:"officialSite"`   // URL
	Schedule       Schedule  `json:"schedule"`
	Rating         Rating    `json:"rating"`
	Weight         int       `json:"weight"`
	Network        *Network  `json:"network"`
	WebChannel     *Network  `json:"webChannel"`
	Externals      Externals `json:"externals"`
	Image          *Image    `json:"image"`
	Summary        string    `json:"summary"` // HTML
	Updated        int64     `json:"updated"` // Unix timestamp
	Embedded       *Embedded `json:"_embedded,omitempty"`
}

// Schedule contains the show's airing schedule.
type Schedule struct {
	Time string   `json:"time"` // "22:00"
	Days []string `json:"days"` // ["Thursday"]
}

// Rating contains the average rating.
type Rating struct {
	Average *float64 `json:"average"` // nullable
}

// Network represents a TV network or streaming service.
type Network struct {
	ID           int      `json:"id"`
	Name         string   `json:"name"` // "CBS", "Netflix"
	Country      *Country `json:"country"`
	OfficialSite string   `json:"officialSite"`
}

// Country represents a country.
type Country struct {
	Name     string `json:"name"`     // "United States"
	Code     string `json:"code"`     // "US"
	Timezone string `json:"timezone"` // "America/New_York"
}

// Externals contains external provider IDs.
type Externals struct {
	TVRage  *int   `json:"tvrage"`
	TheTVDB *int   `json:"thetvdb"`
	IMDb    string `json:"imdb"` // "tt1553656"
}

// Image contains image URLs.
type Image struct {
	Medium   string `json:"medium"`   // ~210x295
	Original string `json:"original"` // full resolution
}

// SearchResult wraps a show with a relevance score.
type SearchResult struct {
	Score float64 `json:"score"`
	Show  Show    `json:"show"`
}

// Episode represents a TV episode.
type Episode struct {
	ID       int    `json:"id"`
	URL      string `json:"url"`
	Name     string `json:"name"`
	Season   int    `json:"season"`
	Number   int    `json:"number"`
	Type     string `json:"type"`     // "regular", "significant_special"
	Airdate  string `json:"airdate"`  // "2013-06-24"
	Airtime  string `json:"airtime"`  // "22:00"
	Airstamp string `json:"airstamp"` // ISO 8601: "2013-06-25T02:00:00+00:00"
	Runtime  int    `json:"runtime"`
	Rating   Rating `json:"rating"`
	Image    *Image `json:"image"`
	Summary  string `json:"summary"` // HTML
}

// Season represents a TV season.
type Season struct {
	ID           int    `json:"id"`
	URL          string `json:"url"`
	Number       int    `json:"number"`
	Name         string `json:"name"`
	EpisodeOrder int    `json:"episodeOrder"` // total episodes in season
	PremiereDate string `json:"premiereDate"` // "2013-06-24"
	EndDate      string `json:"endDate"`      // "2013-09-16"
	Image        *Image `json:"image"`
	Summary      string `json:"summary"` // HTML
}

// Embedded contains embedded resources (episodes, seasons).
type Embedded struct {
	Episodes []Episode `json:"episodes,omitempty"`
	Seasons  []Season  `json:"seasons,omitempty"`
}

// NetworkName returns the network or web channel name, or empty string.
func (s *Show) NetworkName() string {
	if s.Network != nil {
		return s.Network.Name
	}
	if s.WebChannel != nil {
		return s.WebChannel.Name
	}
	return ""
}
