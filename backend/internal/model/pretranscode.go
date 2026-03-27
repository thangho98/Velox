package model

// PretranscodeProfile defines a quality preset for offline encoding.
type PretranscodeProfile struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	Height       int    `json:"height"`
	VideoBitrate int    `json:"video_bitrate"` // kbps
	AudioBitrate int    `json:"audio_bitrate"` // kbps
	VideoCodec   string `json:"video_codec"`
	AudioCodec   string `json:"audio_codec"`
	Enabled      bool   `json:"enabled"`
	CreatedAt    string `json:"created_at"`
}

// PretranscodeFile represents a single pre-transcoded output file.
type PretranscodeFile struct {
	ID           int64   `json:"id"`
	MediaFileID  int64   `json:"media_file_id"`
	ProfileID    int64   `json:"profile_id"`
	FilePath     string  `json:"file_path"`
	FileSize     int64   `json:"file_size"`
	DurationSecs float64 `json:"duration_secs"`
	Status       string  `json:"status"` // pending, encoding, ready, failed
	ErrorMessage string  `json:"error_message,omitempty"`
	StartedAt    string  `json:"started_at,omitempty"`
	CompletedAt  string  `json:"completed_at,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

// PretranscodeQueueItem represents a pending encode job.
type PretranscodeQueueItem struct {
	ID          int64  `json:"id"`
	MediaFileID int64  `json:"media_file_id"`
	ProfileID   int64  `json:"profile_id"`
	Priority    int    `json:"priority"`
	Status      string `json:"status"` // queued, encoding, done, failed, cancelled
	CreatedAt   string `json:"created_at"`
}

// PretranscodeStatus is the aggregated progress response.
type PretranscodeStatus struct {
	Enabled     bool   `json:"enabled"`
	Schedule    string `json:"schedule"`
	Concurrency int    `json:"concurrency"`
	Paused      bool   `json:"paused"`
	Total       int    `json:"total"`
	Done        int    `json:"done"`
	Encoding    int    `json:"encoding"`
	Failed      int    `json:"failed"`
	Queued      int    `json:"queued"`
	DiskUsed    int64  `json:"disk_used"`    // bytes used by pre-transcode files
	CurrentFile string `json:"current_file"` // title of currently encoding media
	Speed       string `json:"speed"`        // e.g. "2.5x"
}

// StorageEstimate is the result of estimating pre-transcode disk usage.
type StorageEstimate struct {
	Profiles      []ProfileEstimate `json:"profiles"`
	TotalBytes    int64             `json:"total_bytes"`
	DiskFreeBytes int64             `json:"disk_free_bytes"`
	FileCount     int               `json:"file_count"`
}

// ProfileEstimate is per-profile disk usage estimate.
type ProfileEstimate struct {
	ProfileID   int64   `json:"profile_id"`
	ProfileName string  `json:"profile_name"`
	Height      int     `json:"height"`
	EstimatedGB float64 `json:"estimated_gb"`
	FileCount   int     `json:"file_count"`
}
