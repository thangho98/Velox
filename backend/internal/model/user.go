package model

// User represents a system user
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	DisplayName  string `json:"display_name"`
	PasswordHash string `json:"-"` // never expose
	IsAdmin      bool   `json:"is_admin"`
	AvatarPath   string `json:"avatar_path"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// UserPreferences stores user-specific settings
type UserPreferences struct {
	UserID              int64  `json:"user_id"`
	SubtitleLanguage    string `json:"subtitle_language"`
	AudioLanguage       string `json:"audio_language"`
	MaxStreamingQuality string `json:"max_streaming_quality"`
	Theme               string `json:"theme"`
	Language            string `json:"language"`
}

// UserLibraryAccess links users to libraries they can access
type UserLibraryAccess struct {
	UserID    int64 `json:"user_id"`
	LibraryID int64 `json:"library_id"`
}

// UserData represents unified per-user-per-media state (Emby pattern)
type UserData struct {
	UserID       int64    `json:"user_id"`
	MediaID      int64    `json:"media_id"`
	Position     float64  `json:"position"`
	Completed    bool     `json:"completed"`
	IsFavorite   bool     `json:"is_favorite"`
	Rating       *float64 `json:"rating"` // nil = not rated, 1.0-10.0
	PlayCount    int      `json:"play_count"`
	LastPlayedAt *string  `json:"last_played_at"` // nil = never played
	UpdatedAt    string   `json:"updated_at"`

	// JOIN fields (populated by queries with media)
	MediaTitle    string  `json:"media_title,omitempty"`
	MediaPoster   string  `json:"media_poster,omitempty"`
	MediaDuration float64 `json:"media_duration,omitempty"`
}

// UserSeriesData represents series-level favorite/rating
type UserSeriesData struct {
	UserID     int64    `json:"user_id"`
	SeriesID   int64    `json:"series_id"`
	IsFavorite bool     `json:"is_favorite"`
	Rating     *float64 `json:"rating"`
	UpdatedAt  string   `json:"updated_at"`
}

// ContinueWatchingItem represents an in-progress media item (movie or episode)
type ContinueWatchingItem struct {
	// UserData fields
	MediaID      int64   `json:"media_id"`
	SeriesID     int64   `json:"series_id,omitempty"`
	Position     float64 `json:"position"`
	Completed    bool    `json:"completed"`
	LastPlayedAt *string `json:"last_played_at"`

	// Media fields
	Title         string  `json:"title"`
	PosterPath    string  `json:"poster_path"`
	BackdropPath  string  `json:"backdrop_path"`
	MediaType     string  `json:"media_type"`
	MediaDuration float64 `json:"duration"`

	// Episode context (nullable for movies)
	SeriesTitle   string `json:"series_title,omitempty"`
	SeasonNumber  int    `json:"season_number,omitempty"`
	EpisodeNumber int    `json:"episode_number,omitempty"`
}

// NextUpItem represents the next unwatched episode for a series
type NextUpItem struct {
	MediaID       int64   `json:"media_id"`
	SeriesID      int64   `json:"series_id"`
	Title         string  `json:"title"`
	EpisodeTitle  string  `json:"episode_title"`
	MediaType     string  `json:"media_type"`
	StillPath     string  `json:"still_path"`
	BackdropPath  string  `json:"backdrop_path"`
	Duration      float64 `json:"duration"`
	SeasonNumber  int     `json:"season_number"`
	EpisodeNumber int     `json:"episode_number"`
	SeriesTitle   string  `json:"series_title"`
	SeriesPoster  string  `json:"series_poster"`
	LastWatchedAt *string `json:"last_watched_at"`
}
