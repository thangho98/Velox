package model

import "time"

type DownloadStatus string

const (
	DownloadStatusPending     DownloadStatus = "pending"
	DownloadStatusDownloading DownloadStatus = "downloading"
	DownloadStatusCompleted   DownloadStatus = "completed"
	DownloadStatusFailed      DownloadStatus = "failed"
)

type DownloadTask struct {
	ID          string         `json:"id"`
	MediaID     int64          `json:"media_id"`
	MediaFileID int64          `json:"media_file_id"`
	Title       string         `json:"title"`
	Progress    float64        `json:"progress"`
	Status      DownloadStatus `json:"status"`
	Error       string         `json:"error,omitempty"`
	Speed       string         `json:"speed,omitempty"`
	ETA         string         `json:"eta,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
}
