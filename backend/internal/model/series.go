package model

// Series represents a TV show
type Series struct {
	ID             int64   `json:"id"`
	LibraryID      int64   `json:"library_id"`
	Title          string  `json:"title"`
	SortTitle      string  `json:"sort_title"`
	TmdbID         *int64  `json:"tmdb_id,omitempty"`
	ImdbID         *string `json:"imdb_id,omitempty"`
	TvdbID         *int64  `json:"tvdb_id,omitempty"`
	Overview       string  `json:"overview"`
	Status         string  `json:"status"`         // "Returning Series" | "Ended" | "Canceled"
	Network        string  `json:"network"`        // "CBS", "Netflix", etc.
	FirstAirDate   string  `json:"first_air_date"` // YYYY-MM-DD
	PosterPath     string  `json:"poster_path"`
	BackdropPath   string  `json:"backdrop_path"`
	LogoPath       string  `json:"logo_path"`
	ThumbPath      string  `json:"thumb_path"`
	MetadataLocked bool    `json:"metadata_locked"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      string  `json:"updated_at"`
}

// SeriesMetadataEditRequest represents a partial metadata edit for series.
type SeriesMetadataEditRequest struct {
	Title          *string       `json:"title"`
	SortTitle      *string       `json:"sort_title"`
	Overview       *string       `json:"overview"`
	Status         *string       `json:"status"`
	Network        *string       `json:"network"`
	FirstAirDate   *string       `json:"first_air_date"`
	Genres         []string      `json:"genres"`
	Credits        []CreditInput `json:"credits"`
	SaveNFO        bool          `json:"save_nfo"`
	MetadataLocked *bool         `json:"metadata_locked"`
}

// Season represents a season of a series
type Season struct {
	ID           int64  `json:"id"`
	SeriesID     int64  `json:"series_id"`
	SeasonNumber int    `json:"season_number"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	PosterPath   string `json:"poster_path"`
	EpisodeCount int    `json:"episode_count"`
	CreatedAt    string `json:"created_at"`
}

// Episode represents a single episode linking to a media item
type Episode struct {
	ID            int64  `json:"id"`
	SeriesID      int64  `json:"series_id"`
	SeasonID      int64  `json:"season_id"`
	MediaID       int64  `json:"media_id"`
	EpisodeNumber int    `json:"episode_number"`
	Title         string `json:"title"`
	Overview      string `json:"overview"`
	StillPath     string `json:"still_path"`
	AirDate       string `json:"air_date"` // YYYY-MM-DD
	CreatedAt     string `json:"created_at"`
}

// EpisodeWithMedia combines episode with its media and media file
type EpisodeWithMedia struct {
	Episode     Episode    `json:"episode"`
	Media       Media      `json:"media"`
	PrimaryFile *MediaFile `json:"primary_file,omitempty"`
}

// SeriesWithSeasons combines series with its seasons
type SeriesWithSeasons struct {
	Series  Series   `json:"series"`
	Seasons []Season `json:"seasons"`
}
