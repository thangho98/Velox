package model

// AudioFingerprint stores cached chromaprint fingerprint data for a media file region.
type AudioFingerprint struct {
	ID          int64   `json:"id"`
	MediaFileID int64   `json:"media_file_id"`
	Region      string  `json:"region"`       // "intro_region" | "credits_region"
	Fingerprint []byte  `json:"-"`            // raw uint32 array, little-endian
	DurationSec float64 `json:"duration_sec"` // duration of analyzed region
	SampleCount int     `json:"sample_count"` // number of uint32 samples
	CreatedAt   string  `json:"created_at"`
}
