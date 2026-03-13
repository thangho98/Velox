package model

// Media represents a logical media item (movie or episode)
type Media struct {
	ID              int64   `json:"id"`
	LibraryID       int64   `json:"library_id"`
	MediaType       string  `json:"media_type"` // "movie" | "episode"
	Title           string  `json:"title"`
	SortTitle       string  `json:"sort_title"`
	TmdbID          *int64  `json:"tmdb_id,omitempty"`
	ImdbID          *string `json:"imdb_id,omitempty"`
	TvdbID          *int64  `json:"tvdb_id,omitempty"`
	Overview        string  `json:"overview"`
	ReleaseDate     string  `json:"release_date"` // YYYY-MM-DD
	Rating          float64 `json:"rating"`
	IMDbRating      float64 `json:"imdb_rating"`
	RTScore         int     `json:"rt_score"`
	MetacriticScore int     `json:"metacritic_score"`
	PosterPath      string  `json:"poster_path"`
	BackdropPath    string  `json:"backdrop_path"`
	LogoPath        string  `json:"logo_path"`
	ThumbPath       string  `json:"thumb_path"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
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
	AudioCodec     string  `json:"audio_codec"`
	Container      string  `json:"container"`
	Bitrate        int     `json:"bitrate"`
	Fingerprint    string  `json:"fingerprint"` // "{file_size}:{xxhash64_first_64KB}"
	IsPrimary      bool    `json:"is_primary"`
	AddedAt        string  `json:"added_at"`
	LastVerifiedAt *string `json:"last_verified_at,omitempty"`
}

// MediaWithFiles combines media with its files
type MediaWithFiles struct {
	Media Media       `json:"media"`
	Files []MediaFile `json:"files"`
}

// MediaListItem represents a media item for list views with genres
type MediaListItem struct {
	ID         int64    `json:"id"`
	Title      string   `json:"title"`
	SortTitle  string   `json:"sort_title"`
	PosterPath string   `json:"poster_path"`
	MediaType  string   `json:"media_type"`
	Genres     []string `json:"genres"`
}
