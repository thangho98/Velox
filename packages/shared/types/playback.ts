// Playback, Streaming, Subtitle types

export interface StreamUrls {
  direct: string
  hls?: string
  primary_file_id?: number
  stream_session_id?: string
}

export interface SubtitleTrack {
  id: number
  media_id: number
  language: string
  label: string
  file_path: string
  is_default: boolean
}

export interface AudioTrack {
  id: number
  media_id: number
  language: string
  label: string
  codec: string
  channels: number
  is_default: boolean
}

export interface PlaybackInfoRequest {
  video_codecs?: string[]
  audio_codecs?: string[]
  containers?: string[]
  max_height?: number
  media_file_id?: number
  selected_audio_track?: number
  selected_subtitle?: string
  selected_subtitle_id?: number
}

export interface SkipSegment {
  type: 'intro' | 'credits'
  start: number
  end: number
  source: string
  confidence: number
}

export interface PlaybackSubtitleTrack {
  id: number
  language: string
  label: string
  format: string
  is_default: boolean
  is_image: boolean
}

export interface PlaybackAudioTrack {
  id: number
  language: string
  label: string
  codec: string
  channels: number
  bitrate: number
  sample_rate: number
  is_default: boolean
  selected: boolean
}

export interface SubtitleSearchResult {
  provider: 'opensubtitles' | 'subdl'
  external_id: string
  title: string
  language: string
  format: string
  downloads: number
  rating: number
  forced: boolean
  hearing_impaired: boolean
  ai_translated: boolean
}

export interface SubtitleDownloadRequest {
  provider: string
  external_id: string
  language: string
}

export interface PlaybackInfo {
  media_id: number
  primary_file_id: number
  stream_session_id?: string
  method: string
  stream_url: string
  direct_url?: string
  abr_url?: string
  video_codec: string
  video_profile: string
  video_level: number
  video_fps: number
  audio_codec: string
  container: string
  file_size: number
  bitrate: number
  duration: number
  width: number
  height: number
  audio_tracks: PlaybackAudioTrack[]
  subtitle_tracks: PlaybackSubtitleTrack[]
  decision_reason: string
  estimated_bitrate: number
  pt_video_codec?: string
  pt_audio_codec?: string
  pt_height?: number
  pt_video_bitrate?: number
  pt_audio_bitrate?: number
  position: number
  skip_segments?: SkipSegment[]
  available_qualities?: QualityOption[]
}

export interface QualityOption {
  height: number
  label: string
  instant: boolean
  source: 'original' | 'pretranscode' | 'transcode'
  bitrate_kbps?: number
}

// User progress & continue watching

export interface UserData {
  user_id: number
  media_id: number
  position: number
  completed: boolean
  is_favorite: boolean
  rating?: number
  play_count: number
  last_played_at?: string
  updated_at: string
  // JOIN fields
  media_title?: string
  media_poster?: string
  media_duration?: number
}

export interface UpdateProgressRequest {
  position: number
  completed: boolean
}

export interface ToggleFavoriteResponse {
  is_favorite: boolean
}

export interface FavoritesListParams {
  limit?: number
  offset?: number
}

export interface RecentlyWatchedParams {
  limit?: number
}

export interface ContinueWatchingItem {
  media_id: number
  series_id?: number
  position: number
  completed: boolean
  last_played_at?: string
  title: string
  poster_path: string
  backdrop_path: string
  media_type: 'movie' | 'episode'
  duration: number
  series_title?: string
  season_number?: number
  episode_number?: number
}

export interface NextUpItem {
  media_id: number
  series_id: number
  title: string
  episode_title: string
  media_type: 'episode'
  still_path: string
  backdrop_path: string
  duration: number
  season_number: number
  episode_number: number
  series_title: string
  series_poster: string
  last_watched_at?: string
}

export interface ContinueWatchingParams {
  limit?: number
}

export interface NextUpParams {
  limit?: number
}
