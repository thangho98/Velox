package model

// ActivityLog represents a single activity entry
type ActivityLog struct {
	ID          int64  `json:"id"`
	UserID      *int64 `json:"user_id,omitempty"`
	Username    string `json:"username,omitempty"` // joined from users table
	Action      string `json:"action"`
	MediaID     *int64 `json:"media_id,omitempty"`
	MediaTitle  string `json:"media_title,omitempty"` // joined from media table
	DetailsJSON string `json:"details"`
	IPAddress   string `json:"ip_address"`
	CreatedAt   string `json:"created_at"`
}

// ActivityFilter specifies filters for listing activity logs
type ActivityFilter struct {
	UserID *int64
	Action string
	From   string // ISO 8601 datetime
	To     string // ISO 8601 datetime
	Limit  int
	Offset int
}

// PlaybackStatsResult holds aggregated playback statistics
type PlaybackStatsResult struct {
	MostWatched  []MostWatchedItem `json:"most_watched"`
	MostActive   []MostActiveUser  `json:"most_active_users"`
	TotalPlays   int               `json:"total_plays"`
	UniqueTitles int               `json:"unique_titles"`
}

// MostWatchedItem represents a frequently watched media item
type MostWatchedItem struct {
	MediaID    int64  `json:"media_id"`
	Title      string `json:"title"`
	PlayCount  int    `json:"play_count"`
	PosterPath string `json:"poster_path"`
}

// MostActiveUser represents a user ranked by activity
type MostActiveUser struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	ActionCount int    `json:"action_count"`
}

// ServerInfo holds server status information for the admin dashboard
type ServerInfo struct {
	Version     string `json:"version"`
	Uptime      string `json:"uptime"`
	GoVersion   string `json:"go_version"`
	OS          string `json:"os"`
	Arch        string `json:"arch"`
	FFmpegVer   string `json:"ffmpeg_version"`
	Database    string `json:"database"`
	HWAccel     string `json:"hw_accel"`
	MediaCount  int    `json:"media_count"`
	SeriesCount int    `json:"series_count"`
	UserCount   int    `json:"user_count"`
	TotalSize   int64  `json:"total_size_bytes"`
}

// LibraryStats holds per-library statistics
type LibraryStats struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	ItemCount   int    `json:"item_count"`
	FileCount   int    `json:"file_count"`
	TotalSize   int64  `json:"total_size_bytes"`
	LastScanned string `json:"last_scanned,omitempty"`
}

// Webhook represents a webhook subscription
type Webhook struct {
	ID        int64  `json:"id"`
	URL       string `json:"url"`
	Events    string `json:"events"` // JSON array
	Secret    string `json:"-"`      // never expose
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// TaskInfo holds status information for a scheduled task
type TaskInfo struct {
	Name     string `json:"name"`
	Interval string `json:"interval"`
	LastRun  string `json:"last_run,omitempty"`
	NextRun  string `json:"next_run"`
	Running  bool   `json:"running"`
}
