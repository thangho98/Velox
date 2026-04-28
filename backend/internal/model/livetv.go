package model

import "time"

// LivePlaylist represents a source of IPTV channels (usually an m3u URL or file)
type LivePlaylist struct {
	ID                int64      `json:"id"`
	Name              string     `json:"name"`
	URL               string     `json:"url"` // External URL or file:/// path
	EpgURL            string     `json:"epgUrl,omitempty"`
	LastSyncedAt      *time.Time `json:"lastSyncedAt,omitempty"`
	SyncIntervalHours int        `json:"syncIntervalHours"`
	IsActive          bool       `json:"isActive"`
	CreatedAt         time.Time  `json:"createdAt"`
	UpdatedAt         time.Time  `json:"updatedAt"`
}

// LiveChannel represents a single streaming channel extracted from a playlist.
// UserAgent/Referer/Origin come from per-channel #EXTVLCOPT hints and are
// forwarded by the proxy when fetching upstream (some CDNs block unknown UAs).
type LiveChannel struct {
	ID            int64      `json:"id"`
	PlaylistID    int64      `json:"playlistId"`
	ChannelID     string     `json:"channelId"` // tvg-id if available
	Name          string     `json:"name"`
	Logo          string     `json:"logo"`
	GroupTitle    string     `json:"groupTitle"`
	StreamURL     string     `json:"streamUrl"`
	Country       string     `json:"country"`
	UserAgent     string     `json:"userAgent,omitempty"`
	Referer       string     `json:"referer,omitempty"`
	Origin        string     `json:"origin,omitempty"`
	IsActive      bool       `json:"isActive"`
	IsHidden      bool       `json:"isHidden"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	LastWatchedAt *time.Time `json:"lastWatchedAt,omitempty"`
}

// LiveProgram represents an EPG entry for a LiveChannel
type LiveProgram struct {
	ID          int64     `json:"id"`
	ChannelID   int64     `json:"channel_id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}
