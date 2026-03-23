package model

import (
	"encoding/json"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationScanComplete       NotificationType = "scan_complete"
	NotificationMediaAdded         NotificationType = "media_added"
	NotificationTranscodeComplete  NotificationType = "transcode_complete"
	NotificationTranscodeFailed    NotificationType = "transcode_failed"
	NotificationSubtitleDownloaded NotificationType = "subtitle_downloaded"
	NotificationIdentifyComplete   NotificationType = "identify_complete"
	NotificationLibraryWatcher     NotificationType = "library_watcher"
)

// Notification represents a user notification (persistent inbox + real-time)
type Notification struct {
	ID        int64            `json:"id"`
	UserID    *int64           `json:"user_id,omitempty"` // nil = broadcast to all
	Type      NotificationType `json:"type"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Data      json.RawMessage  `json:"data"`
	Read      bool             `json:"read"`
	CreatedAt time.Time        `json:"created_at"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
}

// NotificationData holds structured data for different notification types
type NotificationData struct {
	// Common fields
	LibraryID *int64 `json:"library_id,omitempty"`
	MediaID   *int64 `json:"media_id,omitempty"`
	SeriesID  *int64 `json:"series_id,omitempty"`
	EpisodeID *int64 `json:"episode_id,omitempty"`

	// Scan-specific
	ScannedCount int `json:"scanned_count,omitempty"`
	NewCount     int `json:"new_count,omitempty"`
	ErrorCount   int `json:"error_count,omitempty"`

	// Transcode-specific
	Quality  string `json:"quality,omitempty"`
	Duration int    `json:"duration_seconds,omitempty"`

	// Subtitle-specific
	Language string `json:"language,omitempty"`
	Provider string `json:"provider,omitempty"`

	// Media-specific
	MediaTitle string `json:"media_title,omitempty"`
	MediaType  string `json:"media_type,omitempty"` // "movie" | "episode"
}

// ToJSON serializes NotificationData to json.RawMessage
func (d NotificationData) ToJSON() json.RawMessage {
	b, _ := json.Marshal(d)
	return b
}

// NotificationFilter specifies filters for listing notifications
type NotificationFilter struct {
	UserID     *int64
	UnreadOnly bool
	Limit      int
	Offset     int
}

// UnreadCountResult holds the unread count for a user
type UnreadCountResult struct {
	Count int64 `json:"count"`
}

// WebSocketMessage is the envelope for WebSocket messages
type WebSocketMessage struct {
	Type    string          `json:"type"` // "notification", "ping", "pong"
	Payload json.RawMessage `json:"payload"`
}

// WSNotificationPayload is the payload for notification messages
type WSNotificationPayload struct {
	Notification *Notification `json:"notification"`
}

// NotificationPreferences for user notification settings
type NotificationPreferences struct {
	ScanComplete       bool `json:"scan_complete"`
	MediaAdded         bool `json:"media_added"`
	TranscodeComplete  bool `json:"transcode_complete"`
	SubtitleDownloaded bool `json:"subtitle_downloaded"`
	IdentifyComplete   bool `json:"identify_complete"`
	LibraryWatcher     bool `json:"library_watcher"`
	BrowserPush        bool `json:"browser_push"` // Web Push API (optional)
}
