package model

// Library represents a media library (one or more folders on disk)
type Library struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`  // "movies" | "tvshows" | "mixed"
	Paths     []string `json:"paths"` // all root folders; at least one required
	CreatedAt string   `json:"created_at"`
}

// Genre represents a media genre
type Genre struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	TmdbID *int64 `json:"tmdb_id,omitempty"`
}

// MediaGenre links media or series to genres
type MediaGenre struct {
	MediaID  *int64 `json:"media_id,omitempty"`
	SeriesID *int64 `json:"series_id,omitempty"`
	GenreID  int64  `json:"genre_id"`
}

// Person represents an actor, director, or writer
type Person struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	TmdbID      *int64 `json:"tmdb_id,omitempty"`
	ProfilePath string `json:"profile_path"`
}

// Credit represents a person's role in media or series
type Credit struct {
	ID           int64  `json:"id"`
	MediaID      *int64 `json:"media_id,omitempty"`
	SeriesID     *int64 `json:"series_id,omitempty"`
	PersonID     int64  `json:"person_id"`
	Character    string `json:"character"` // For cast roles
	Role         string `json:"role"`      // "cast" | "director" | "writer"
	DisplayOrder int    `json:"display_order"`
}

// CreditWithPerson includes person details
type CreditWithPerson struct {
	Credit Credit `json:"credit"`
	Person Person `json:"person"`
}

// ScanJob tracks library scan progress
type ScanJob struct {
	ID           int64   `json:"id"`
	LibraryID    int64   `json:"library_id"`
	Status       string  `json:"status"` // "queued" | "scanning" | "completed" | "failed"
	TotalFiles   int     `json:"total_files"`
	ScannedFiles int     `json:"scanned_files"`
	NewFiles     int     `json:"new_files"`
	Errors       int     `json:"errors"`
	ErrorLog     string  `json:"error_log"`
	StartedAt    *string `json:"started_at,omitempty"`
	FinishedAt   *string `json:"finished_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// Subtitle represents a subtitle track (embedded or external)
type Subtitle struct {
	ID          int64  `json:"id"`
	MediaFileID int64  `json:"media_file_id"`
	Language    string `json:"language"` // ISO 639-1: 'en', 'vi'
	Codec       string `json:"codec"`    // subrip, ass, webvtt, etc.
	Title       string `json:"title"`
	IsEmbedded  bool   `json:"is_embedded"`
	StreamIndex int    `json:"stream_index"` // -1 for external
	FilePath    string `json:"file_path"`    // empty for embedded
	IsForced    bool   `json:"is_forced"`
	IsDefault   bool   `json:"is_default"`
	IsSDH       bool   `json:"is_sdh"` // Subtitles for Deaf/Hard of Hearing
}

// AudioTrack represents an audio track in a media file
type AudioTrack struct {
	ID            int64  `json:"id"`
	MediaFileID   int64  `json:"media_file_id"`
	StreamIndex   int    `json:"stream_index"`
	Codec         string `json:"codec"`          // aac, ac3, dts, etc.
	Language      string `json:"language"`       // ISO 639-1
	Channels      int    `json:"channels"`       // 2=stereo, 6=5.1, 8=7.1
	ChannelLayout string `json:"channel_layout"` // "stereo", "5.1"
	Bitrate       int    `json:"bitrate"`
	Title         string `json:"title"`
	IsDefault     bool   `json:"is_default"`
}
