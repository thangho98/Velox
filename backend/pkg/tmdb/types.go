package tmdb

import "time"

// Search results structures

type MovieSearchResults struct {
	Page         int            `json:"page"`
	Results      []MovieSummary `json:"results"`
	TotalResults int            `json:"total_results"`
	TotalPages   int            `json:"total_pages"`
}

type TVSearchResults struct {
	Page         int         `json:"page"`
	Results      []TVSummary `json:"results"`
	TotalResults int         `json:"total_results"`
	TotalPages   int         `json:"total_pages"`
}

// Summary structures (from search results)

type MovieSummary struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	OriginalTitle string  `json:"original_title"`
	Overview      string  `json:"overview"`
	ReleaseDate   string  `json:"release_date"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	VoteAverage   float64 `json:"vote_average"`
	VoteCount     int     `json:"vote_count"`
	Popularity    float64 `json:"popularity"`
	Adult         bool    `json:"adult"`
	Video         bool    `json:"video"`
	GenreIDs      []int   `json:"genre_ids"`
}

type TVSummary struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	FirstAirDate     string   `json:"first_air_date"`
	PosterPath       string   `json:"poster_path"`
	BackdropPath     string   `json:"backdrop_path"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	GenreIDs         []int    `json:"genre_ids"`
	OriginCountry    []string `json:"origin_country"`
	OriginalLanguage string   `json:"original_language"`
}

// Full detail structures

type MovieDetails struct {
	ID                  int         `json:"id"`
	IMDbID              string      `json:"imdb_id"`
	Title               string      `json:"title"`
	OriginalTitle       string      `json:"original_title"`
	Overview            string      `json:"overview"`
	ReleaseDate         string      `json:"release_date"`
	PosterPath          string      `json:"poster_path"`
	BackdropPath        string      `json:"backdrop_path"`
	VoteAverage         float64     `json:"vote_average"`
	VoteCount           int         `json:"vote_count"`
	Popularity          float64     `json:"popularity"`
	Runtime             int         `json:"runtime"`
	Status              string      `json:"status"`
	Tagline             string      `json:"tagline"`
	Budget              int64       `json:"budget"`
	Revenue             int64       `json:"revenue"`
	Homepage            string      `json:"homepage"`
	Adult               bool        `json:"adult"`
	Video               bool        `json:"video"`
	Genres              []Genre     `json:"genres"`
	Credits             *Credits    `json:"credits"`
	Keywords            *Keywords   `json:"keywords"`
	BelongsToCollection *Collection `json:"belongs_to_collection"`
	ProductionCompanies []Company   `json:"production_companies"`
	ProductionCountries []Country   `json:"production_countries"`
	SpokenLanguages     []Language  `json:"spoken_languages"`
}

type TVDetails struct {
	ID               int             `json:"id"`
	Name             string          `json:"name"`
	OriginalName     string          `json:"original_name"`
	Overview         string          `json:"overview"`
	FirstAirDate     string          `json:"first_air_date"`
	LastAirDate      string          `json:"last_air_date"`
	PosterPath       string          `json:"poster_path"`
	BackdropPath     string          `json:"backdrop_path"`
	VoteAverage      float64         `json:"vote_average"`
	VoteCount        int             `json:"vote_count"`
	Popularity       float64         `json:"popularity"`
	NumberOfEpisodes int             `json:"number_of_episodes"`
	NumberOfSeasons  int             `json:"number_of_seasons"`
	EpisodeRunTime   []int           `json:"episode_run_time"`
	Status           string          `json:"status"`
	Tagline          string          `json:"tagline"`
	Homepage         string          `json:"homepage"`
	Type             string          `json:"type"`
	InProduction     bool            `json:"in_production"`
	Genres           []Genre         `json:"genres"`
	Credits          *Credits        `json:"credits"`
	Keywords         *Keywords       `json:"keywords"`
	ExternalIDs      *ExternalIDs    `json:"external_ids"`
	Seasons          []SeasonSummary `json:"seasons"`
	CreatedBy        []PersonSummary `json:"created_by"`
	Networks         []Network       `json:"networks"`
	OriginCountry    []string        `json:"origin_country"`
	OriginalLanguage string          `json:"original_language"`
	Languages        []string        `json:"languages"`
}

// Season and Episode structures

type SeasonSummary struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Overview     string `json:"overview"`
	SeasonNumber int    `json:"season_number"`
	EpisodeCount int    `json:"episode_count"`
	PosterPath   string `json:"poster_path"`
	AirDate      string `json:"air_date"`
}

type SeasonDetails struct {
	ID           string           `json:"_id"`
	AirDate      string           `json:"air_date"`
	Name         string           `json:"name"`
	Overview     string           `json:"overview"`
	PosterPath   string           `json:"poster_path"`
	SeasonNumber int              `json:"season_number"`
	Episodes     []EpisodeDetails `json:"episodes"`
}

type EpisodeDetails struct {
	ID             int          `json:"id"`
	Name           string       `json:"name"`
	Overview       string       `json:"overview"`
	AirDate        string       `json:"air_date"`
	EpisodeNumber  int          `json:"episode_number"`
	SeasonNumber   int          `json:"season_number"`
	StillPath      string       `json:"still_path"`
	VoteAverage    float64      `json:"vote_average"`
	VoteCount      int          `json:"vote_count"`
	Runtime        int          `json:"runtime"`
	ProductionCode string       `json:"production_code"`
	ShowID         int          `json:"show_id"`
	Crew           []CrewMember `json:"crew"`
	GuestStars     []CastMember `json:"guest_stars"`
}

// Supporting structures

type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Credits struct {
	Cast []CastMember `json:"cast"`
	Crew []CrewMember `json:"crew"`
}

type CastMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Character   string `json:"character"`
	ProfilePath string `json:"profile_path"`
	Order       int    `json:"order"`
	CastID      int    `json:"cast_id"`
	CreditID    string `json:"credit_id"`
}

type CrewMember struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Job         string `json:"job"`
	Department  string `json:"department"`
	ProfilePath string `json:"profile_path"`
	CreditID    string `json:"credit_id"`
}

type Keywords struct {
	Keywords []Keyword `json:"keywords"`
}

type Keyword struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExternalIDs struct {
	IMDbID      string `json:"imdb_id"`
	FreebaseMID string `json:"freebase_mid"`
	FreebaseID  string `json:"freebase_id"`
	TVDBID      int    `json:"tvdb_id"`
	TVRageID    int    `json:"tvrage_id"`
	Facebook    string `json:"facebook_id"`
	Instagram   string `json:"instagram_id"`
	Twitter     string `json:"twitter_id"`
}

type Collection struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	PosterPath   string `json:"poster_path"`
	BackdropPath string `json:"backdrop_path"`
}

type Company struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

type Country struct {
	ISO3166_1 string `json:"iso_3166_1"`
	Name      string `json:"name"`
}

type Language struct {
	ISO639_1    string `json:"iso_639_1"`
	Name        string `json:"name"`
	EnglishName string `json:"english_name"`
}

type PersonSummary struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ProfilePath string `json:"profile_path"`
}

type Network struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	LogoPath      string `json:"logo_path"`
	OriginCountry string `json:"origin_country"`
}

// Find results (for external ID lookup)

type FindResults struct {
	MovieResults  []MovieSummary `json:"movie_results"`
	TVResults     []TVSummary    `json:"tv_results"`
	PersonResults []struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		ProfilePath string `json:"profile_path"`
	} `json:"person_results"`
}

// ReleaseDate structures

type ReleaseDates struct {
	ID      int              `json:"id"`
	Results []CountryRelease `json:"results"`
}

type CountryRelease struct {
	ISO3166_1    string        `json:"iso_3166_1"`
	ReleaseDates []ReleaseInfo `json:"release_dates"`
}

type ReleaseInfo struct {
	Certification string `json:"certification"`
	ISO639_1      string `json:"iso_639_1"`
	Note          string `json:"note"`
	ReleaseDate   string `json:"release_date"`
	Type          int    `json:"type"`
}

// Helper methods

// GetYear extracts the year from a release date string
func GetYear(dateStr string) int {
	if dateStr == "" {
		return 0
	}
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return 0
	}
	return t.Year()
}
