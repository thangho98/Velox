package model

// MediaMarker represents a skip segment (intro or credits) for a media file
type MediaMarker struct {
	ID          int64   `json:"id"`
	MediaFileID int64   `json:"media_file_id"`
	MarkerType  string  `json:"marker_type"` // "intro" | "credits"
	StartSec    float64 `json:"start_sec"`
	EndSec      float64 `json:"end_sec"`
	Source      string  `json:"source"`     // "chapter" | "fingerprint" | "manual"
	Confidence  float64 `json:"confidence"` // 1.0 = chapter/manual, <1.0 = fingerprint
	Label       string  `json:"label"`      // Original chapter title or user label
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// SkipSegment is the API DTO for skip markers
type SkipSegment struct {
	Type       string  `json:"type"`       // "intro" | "credits"
	Start      float64 `json:"start"`      // seconds
	End        float64 `json:"end"`        // seconds
	Source     string  `json:"source"`     // "chapter" | "fingerprint" | "manual"
	Confidence float64 `json:"confidence"` // 1.0 = high confidence
}
