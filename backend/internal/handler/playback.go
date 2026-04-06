package handler

import (
	"github.com/thawng/velox/internal/auth"
	"github.com/thawng/velox/internal/model"
	"github.com/thawng/velox/internal/service"
	"github.com/thawng/velox/internal/transcoder"
)

var supportsSubtitleBurnIn = transcoder.SupportsSubtitleBurnIn

// PlaybackHandler provides playback decision and info endpoints
type PlaybackHandler struct {
	mediaSvc      *service.MediaService
	streamSvc     *service.StreamService
	userDataSvc   *service.UserDataService
	subtitleSvc   *service.SubtitleService
	audioTrackSvc *service.AudioTrackService
	markerSvc     *service.MarkerService
	configSvc     *service.PlaybackConfigService
	apiKeyStore   *auth.APIKeyStore
	activitySvc   *service.ActivityService
}

// NewPlaybackHandler creates a new playback handler
func NewPlaybackHandler(
	mediaSvc *service.MediaService,
	streamSvc *service.StreamService,
	userDataSvc *service.UserDataService,
	subtitleSvc *service.SubtitleService,
	audioTrackSvc *service.AudioTrackService,
	markerSvc *service.MarkerService,
	configSvc *service.PlaybackConfigService,
	apiKeyStore *auth.APIKeyStore,
) *PlaybackHandler {
	return &PlaybackHandler{
		mediaSvc:      mediaSvc,
		streamSvc:     streamSvc,
		userDataSvc:   userDataSvc,
		subtitleSvc:   subtitleSvc,
		audioTrackSvc: audioTrackSvc,
		markerSvc:     markerSvc,
		configSvc:     configSvc,
		apiKeyStore:   apiKeyStore,
	}
}

func (h *PlaybackHandler) SetActivityService(svc *service.ActivityService) {
	h.activitySvc = svc
}

// PlaybackInfoRequest represents client-sent capabilities
type PlaybackInfoRequest struct {
	VideoCodecs        []string `json:"video_codecs,omitempty"`
	AudioCodecs        []string `json:"audio_codecs,omitempty"`
	Containers         []string `json:"containers,omitempty"`
	MaxHeight          int      `json:"max_height,omitempty"`
	PreferDirectPlay   bool     `json:"prefer_direct_play,omitempty"`
	MediaFileID        int64    `json:"media_file_id,omitempty"`
	SelectedAudioTrack int      `json:"selected_audio_track,omitempty"`
	SelectedSubtitle   string   `json:"selected_subtitle,omitempty"`
	SelectedSubtitleID int      `json:"selected_subtitle_id,omitempty"`
}

// PlaybackInfoResponse represents playback decision response
type PlaybackInfoResponse struct {
	MediaID            int                 `json:"media_id"`
	PrimaryFileID      int64               `json:"primary_file_id,omitempty"`
	StreamSessionID    string              `json:"stream_session_id,omitempty"`
	Method             string              `json:"method"`
	StreamURL          string              `json:"stream_url"`
	DirectURL          string              `json:"direct_url,omitempty"`
	AbrURL             string              `json:"abr_url,omitempty"`
	VideoCodec         string              `json:"video_codec"`
	VideoProfile       string              `json:"video_profile,omitempty"`
	VideoLevel         int                 `json:"video_level,omitempty"`
	VideoFPS           float64             `json:"video_fps,omitempty"`
	AudioCodec         string              `json:"audio_codec"`
	Container          string              `json:"container"`
	FileSize           int64               `json:"file_size,omitempty"`
	Bitrate            int                 `json:"bitrate,omitempty"`
	Duration           int                 `json:"duration,omitempty"`
	Width              int                 `json:"width,omitempty"`
	Height             int                 `json:"height,omitempty"`
	AudioTracks        []AudioTrackInfo    `json:"audio_tracks,omitempty"`
	SubtitleTracks     []SubtitleTrackInfo `json:"subtitle_tracks,omitempty"`
	DecisionReason     string              `json:"decision_reason"`
	EstimatedBitrate   int                 `json:"estimated_bitrate,omitempty"`
	PtVideoCodec       string              `json:"pt_video_codec,omitempty"`
	PtAudioCodec       string              `json:"pt_audio_codec,omitempty"`
	PtHeight           int                 `json:"pt_height,omitempty"`
	PtVideoBitrate     int                 `json:"pt_video_bitrate,omitempty"`
	PtAudioBitrate     int                 `json:"pt_audio_bitrate,omitempty"`
	Position           float64             `json:"position,omitempty"`
	SkipSegments       []model.SkipSegment `json:"skip_segments,omitempty"`
	AvailableQualities []QualityOption     `json:"available_qualities,omitempty"`
}

// AudioTrackInfo represents an audio track
type AudioTrackInfo struct {
	ID         int    `json:"id"`
	Language   string `json:"language"`
	Label      string `json:"label"`
	Codec      string `json:"codec,omitempty"`
	Channels   int    `json:"channels,omitempty"`
	Bitrate    int    `json:"bitrate,omitempty"`
	SampleRate int    `json:"sample_rate,omitempty"`
	IsDefault  bool   `json:"is_default"`
	Selected   bool   `json:"selected"`
}

// SubtitleTrackInfo represents a subtitle track
type SubtitleTrackInfo struct {
	ID        int    `json:"id"`
	Language  string `json:"language"`
	Label     string `json:"label"`
	Format    string `json:"format"`
	IsDefault bool   `json:"is_default"`
	IsImage   bool   `json:"is_image"`
}
