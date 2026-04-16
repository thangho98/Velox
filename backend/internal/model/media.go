package model

// Media represents a logical media item (movie or episode)
type Media struct {
	ID              int64        `json:"id"`
	LibraryID       int64        `json:"library_id"`
	MediaType       string       `json:"media_type"` // "movie" | "episode"
	Title           string       `json:"title"`
	SortTitle       string       `json:"sort_title"`
	TmdbID          *int64       `json:"tmdb_id,omitempty"`
	ImdbID          *string      `json:"imdb_id,omitempty"`
	TvdbID          *int64       `json:"tvdb_id,omitempty"`
	Overview        string       `json:"overview"`
	Tagline         string       `json:"tagline"`
	ReleaseDate     string       `json:"release_date"` // YYYY-MM-DD
	Rating          float64      `json:"rating"`
	IMDbRating      float64      `json:"imdb_rating"`
	RTScore         int          `json:"rt_score"`
	MetacriticScore int          `json:"metacritic_score"`
	PosterPath      PosterPath   `json:"-"`
	BackdropPath    BackdropPath `json:"-"`
	LogoPath        LogoPath     `json:"-"`
	ThumbPath       BackdropPath `json:"-"`

	Poster   *ImageResource `json:"poster,omitempty"`
	Backdrop *ImageResource `json:"backdrop,omitempty"`
	Logo     *ImageResource `json:"logo,omitempty"`
	Thumb    *ImageResource `json:"thumb,omitempty"`

	MetadataLocked bool   `json:"metadata_locked"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

// MetadataEditRequest represents a partial metadata edit for media.
// Pointer fields: nil = don't change. Slice fields: nil = don't change, empty = clear all.
type MetadataEditRequest struct {
	Title          *string       `json:"title"`
	SortTitle      *string       `json:"sort_title"`
	Overview       *string       `json:"overview"`
	Tagline        *string       `json:"tagline"`
	ReleaseDate    *string       `json:"release_date"`
	Rating         *float64      `json:"rating"`
	Genres         []string      `json:"genres"`
	Credits        []CreditInput `json:"credits"`
	SaveNFO        bool          `json:"save_nfo"`
	MetadataLocked *bool         `json:"metadata_locked"`
}

// CreditInput represents a credit entry in a metadata edit request.
type CreditInput struct {
	PersonName string `json:"person_name"`
	Character  string `json:"character,omitempty"`
	Role       string `json:"role"` // "cast" | "director" | "writer"
	Order      int    `json:"order"`
}

// MediaFile represents a physical video file on disk
type MediaFile struct {
	ID             int64   `json:"id"`
	MediaID        int64   `json:"media_id"`
	FilePath       string  `json:"file_path"`
	FileSize       int64   `json:"file_size"`
	Duration       float64 `json:"duration"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	VideoCodec     string  `json:"video_codec"`
	VideoProfile   string  `json:"video_profile,omitempty"`
	VideoLevel     int     `json:"video_level,omitempty"`
	VideoFPS       float64 `json:"video_fps,omitempty"`
	AudioCodec     string  `json:"audio_codec"`
	Container      string  `json:"container"`
	Bitrate        int     `json:"bitrate"`
	IsHDR          bool    `json:"is_hdr,omitempty"`
	DVProfile      int     `json:"dv_profile,omitempty"`
	ColorTransfer  string  `json:"color_transfer,omitempty"`
	ColorPrimaries string  `json:"color_primaries,omitempty"`
	Fingerprint    string  `json:"fingerprint"` // "{file_size}:{xxhash64_first_64KB}"
	IsPrimary      bool    `json:"is_primary"`
	AddedAt        string  `json:"added_at"`
	LastVerifiedAt *string `json:"last_verified_at,omitempty"`
}

// MediaWithFiles combines media with its files
type MediaWithFiles struct {
	Media Media       `json:"media"`
	Files []MediaFile `json:"files"`

	// Episode-only fields (populated when media_type == "episode")
	SeriesID      int64 `json:"series_id,omitempty"`
	SeasonID      int64 `json:"season_id,omitempty"`
	EpisodeNumber int   `json:"episode_number,omitempty"`
	SeasonNumber  int   `json:"season_number,omitempty"`
}

// MediaListItem represents a media item for list views with genres
type MediaListItem struct {
	ID          int64          `json:"id"`
	Title       string         `json:"title"`
	SortTitle   string         `json:"sort_title"`
	PosterPath  PosterPath     `json:"-"`
	Poster      *ImageResource `json:"poster,omitempty"`
	MediaType   string         `json:"media_type"`
	Genres      []string       `json:"genres"`
	SeriesID    int64          `json:"series_id,omitempty"`
	ReleaseDate string         `json:"release_date,omitempty"`
	Rating      float64        `json:"rating,omitempty"`
	Overview    string         `json:"overview,omitempty"`
	// Progress (from user_data, populated when UserID filter is set)
	Position  *float64 `json:"position,omitempty"`
	Duration  *float64 `json:"duration,omitempty"`
	Completed *bool    `json:"completed,omitempty"`
}

// MediaListFilter represents filter parameters for media list queries
type MediaListFilter struct {
	UserID    int64 // populated from auth context for progress join
	LibraryID int64
	MediaType string // "movie" | "episode" | ""
	Search    string // LIKE on title + sort_title
	Genre     string // exact match genre name
	Year      string // 4-digit year string
	Sort      string // "newest" | "oldest" | "rating" | "title"
	Limit     int
	Offset    int
}
