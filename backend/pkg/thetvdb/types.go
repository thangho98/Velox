package thetvdb

// API response wrapper — all TVDB v4 responses use this shape.
type apiResponse[T any] struct {
	Status string `json:"status"`
	Data   T      `json:"data"`
}

// LoginResponse is returned by POST /login.
type LoginResponse struct {
	Token string `json:"token"`
}

// SearchResult represents a single search hit.
type SearchResult struct {
	ObjectID        string   `json:"objectID"` // e.g. "series-81189"
	TVDBID          string   `json:"tvdb_id"`
	Name            string   `json:"name"`
	Type            string   `json:"type"` // "series", "movie", "person"
	Year            string   `json:"year"`
	ImageURL        string   `json:"image_url"`
	Overview        string   `json:"overview"`
	Status          string   `json:"status"`
	Slug            string   `json:"slug"`
	FirstAirTime    string   `json:"first_air_time"`
	PrimaryLanguage string   `json:"primary_language"`
	Aliases         []string `json:"aliases"`
}

// SeriesBase is the base series record from GET /series/{id}.
type SeriesBase struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	Slug             string          `json:"slug"`
	Image            string          `json:"image"`
	FirstAired       string          `json:"firstAired"`
	LastAired        string          `json:"lastAired"`
	NextAired        string          `json:"nextAired"`
	Score            int             `json:"score"`
	Status           *StatusRecord   `json:"status"`
	OriginalCountry  string          `json:"originalCountry"`
	OriginalLanguage string          `json:"originalLanguage"`
	Overview         string          `json:"overview"`
	Year             string          `json:"year"`
	AverageRuntime   int             `json:"averageRuntime"`
	Seasons          []SeasonBase    `json:"seasons"`
	Genres           []GenreRecord   `json:"genres"`
	RemoteIDs        []RemoteID      `json:"remoteIds"`
	Artworks         []ArtworkRecord `json:"artworks"`
}

// SeasonBase is the base season record.
type SeasonBase struct {
	ID          int    `json:"id"`
	SeriesID    int    `json:"seriesId"`
	Number      int    `json:"number"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	ImageType   int    `json:"imageType"`
	LastUpdated string `json:"lastUpdated"`
	SeasonType  *Type  `json:"type"`
}

// SeasonExtended is a season with its episodes.
type SeasonExtended struct {
	SeasonBase
	Year     string        `json:"year"`
	Episodes []EpisodeBase `json:"episodes"`
}

// EpisodeBase is the base episode record.
type EpisodeBase struct {
	ID             int    `json:"id"`
	SeriesID       int    `json:"seriesId"`
	Name           string `json:"name"`
	Aired          string `json:"aired"`
	Runtime        int    `json:"runtime"`
	Overview       string `json:"overview"`
	Image          string `json:"image"`
	ImageType      int    `json:"imageType"`
	Number         int    `json:"number"`
	AbsoluteNumber int    `json:"absoluteNumber"`
	SeasonNumber   int    `json:"seasonNumber"`
	FinaleType     string `json:"finaleType"`
	Year           string `json:"year"`
	LastUpdated    string `json:"lastUpdated"`
}

// EpisodeExtended adds characters, remote IDs, etc.
type EpisodeExtended struct {
	EpisodeBase
	Characters     []Character     `json:"characters"`
	ContentRatings []ContentRating `json:"contentRatings"`
	RemoteIDs      []RemoteID      `json:"remoteIds"`
}

// MovieBase is the base movie record.
type MovieBase struct {
	ID       int           `json:"id"`
	Name     string        `json:"name"`
	Slug     string        `json:"slug"`
	Image    string        `json:"image"`
	Score    int           `json:"score"`
	Runtime  int           `json:"runtime"`
	Status   *StatusRecord `json:"status"`
	Year     string        `json:"year"`
	Genres   []GenreRecord `json:"genres"`
	Overview string        `json:"overview,omitempty"`
}

// MovieExtended adds artworks, remote IDs, etc.
type MovieExtended struct {
	MovieBase
	Artworks  []ArtworkRecord `json:"artworks"`
	RemoteIDs []RemoteID      `json:"remoteIds"`
}

// ArtworkRecord is a single artwork entry.
type ArtworkRecord struct {
	ID        int    `json:"id"`
	Image     string `json:"image"`
	Thumbnail string `json:"thumbnail"`
	Language  string `json:"language"`
	Type      int    `json:"type"`
	Score     int    `json:"score"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// Character represents cast/crew on an episode or series.
type Character struct {
	ID         int    `json:"id"`
	PeopleID   int    `json:"peopleId"`
	PersonName string `json:"personName"`
	PeopleType string `json:"peopleType"` // "Director", "Writer", "Guest Star", "Actor"
	Name       string `json:"name"`       // character name
	Image      string `json:"image"`
	Sort       int    `json:"sort"`
	IsFeatured bool   `json:"isFeatured"`
}

// RemoteID is an external provider ID (e.g. IMDB).
type RemoteID struct {
	ID         string `json:"id"`   // e.g. "tt0959621"
	Type       int    `json:"type"` // 2 = IMDB
	SourceName string `json:"sourceName"`
}

// ContentRating represents a content rating.
type ContentRating struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Country     string `json:"country"`
	Description string `json:"description"`
}

// GenreRecord is a genre from TVDB.
type GenreRecord struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// StatusRecord is a status from TVDB.
type StatusRecord struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	RecordType  string `json:"recordType"`
	KeepUpdated bool   `json:"keepUpdated"`
}

// Type represents a season type.
type Type struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"` // "official", "dvd", "absolute"
}

// TranslationRecord from /series/{id}/translations/{lang}.
type TranslationRecord struct {
	Name      string `json:"name"`
	Overview  string `json:"overview"`
	Language  string `json:"language"`
	IsPrimary bool   `json:"isPrimary"`
}

// SeriesEpisodesResponse is the response from /series/{id}/episodes/{seasonType}.
type SeriesEpisodesResponse struct {
	Series   SeriesBase    `json:"series"`
	Episodes []EpisodeBase `json:"episodes"`
}

// IMDbID returns the IMDB ID from a RemoteIDs slice, if present.
func IMDbID(remoteIDs []RemoteID) string {
	for _, r := range remoteIDs {
		if r.Type == 2 || r.SourceName == "IMDB" {
			return r.ID
		}
	}
	return ""
}
