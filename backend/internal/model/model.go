package model

import "time"

type Library struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

type Media struct {
	ID          int64     `json:"id"`
	LibraryID   int64     `json:"library_id"`
	Title       string    `json:"title"`
	FilePath    string    `json:"file_path"`
	Duration    float64   `json:"duration"` // seconds
	Size        int64     `json:"size"`     // bytes
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	VideoCodec  string    `json:"video_codec"`
	AudioCodec  string    `json:"audio_codec"`
	Container   string    `json:"container"`
	Bitrate     int64     `json:"bitrate"`
	HasSubtitle bool      `json:"has_subtitle"`
	PosterPath  string    `json:"poster_path,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Progress struct {
	MediaID   int64     `json:"media_id"`
	Position  float64   `json:"position"` // seconds
	Completed bool      `json:"completed"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CanDirectPlay returns true if the browser can play this format natively.
func (m *Media) CanDirectPlay() bool {
	switch m.VideoCodec {
	case "h264", "vp8", "vp9", "av1":
		return true
	}
	return false
}
