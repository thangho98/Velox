// API Response Types - matches backend JSON structure

// Filesystem browser (used by Add Library UI)
export interface FsDirEntry {
  name: string
  path: string
}

export interface FsBrowseResponse {
  current: string
  parent?: string
  dirs: FsDirEntry[]
}

// Generic API responses
export interface ApiResponse<T> {
  data: T
}

export interface ApiErrorResponse {
  error: string
}

// Auth Types
export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface RefreshRequest {
  refresh_token: string
}

export interface RefreshResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface ChangePasswordRequest {
  old_password: string
  new_password: string
}

export interface UserInfo {
  id: number
  username: string
  display_name: string
  is_admin: boolean
}

// User Types
export interface User {
  id: number
  username: string
  display_name: string
  is_admin: boolean
  avatar_path: string
  created_at: string
  updated_at: string
}

export interface UserPreferences {
  user_id: number
  subtitle_language: string
  audio_language: string
  max_streaming_quality: string
  theme: string
}

export interface UpdateProfileRequest {
  display_name: string
}

// Session Types
export interface Session {
  id: number
  user_id: number
  refresh_token_id?: number
  device_name: string
  ip_address: string
  user_agent: string
  expires_at: string
  last_active_at: string
  created_at: string
}

// Library Types
export interface Library {
  id: number
  name: string
  type: string
  paths: string[]
  created_at: string
}

export interface CreateLibraryRequest {
  name: string
  type: string
  paths: string[]
}

// Media Types
export interface Media {
  id: number
  library_id: number
  media_type: 'movie' | 'episode'
  title: string
  sort_title: string
  tmdb_id?: number
  imdb_id?: string
  tvdb_id?: number
  overview: string
  release_date: string
  rating: number
  imdb_rating: number
  rt_score: number
  metacritic_score: number
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  duration?: number // from media_files join
  series_id?: number // for episodes
  season_id?: number // for episodes
  created_at: string
  updated_at: string
}

export interface MediaFile {
  id: number
  media_id: number
  file_path: string
  file_size: number
  duration: number
  width: number
  height: number
  video_codec: string
  audio_codec: string
  container: string
  bitrate: number
  fingerprint: string
  is_primary: boolean
  added_at: string
  last_verified_at?: string
}

export interface MediaWithFiles {
  media: Media
  files: MediaFile[]
}

export interface MediaListItem {
  id: number
  title: string
  sort_title: string
  poster_path: string
  media_type: 'movie' | 'episode'
  type?: 'movie' | 'series' // frontend-friendly alias
  genres: string[]
  release_date?: string
  rating?: number
  overview?: string
}

// Season Types
export interface Season {
  id: number
  series_id: number
  season_number: number
  title: string
  overview: string
  poster_path: string
  air_date?: string
  episode_count: number
  created_at: string
  updated_at: string
}

// Episode Types
export interface Episode {
  id: number
  series_id: number
  season_id: number
  episode_number: number
  title: string
  overview: string
  still_path?: string
  air_date?: string
  duration?: number
  media_files?: MediaFile[]
  created_at: string
  updated_at: string
}

// UserData Types
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

// List Query Parameters
export interface MediaListParams {
  library_id?: number
  type?: 'movie' | 'episode'
  limit?: number
  offset?: number
}

export interface FavoritesListParams {
  limit?: number
  offset?: number
}

export interface RecentlyWatchedParams {
  limit?: number
}

// Streaming Types
export interface StreamUrls {
  direct: string
  hls?: string
  abr?: string // adaptive bitrate HLS (multi-quality variants)
  primary_file_id?: number
}

// Subtitle Types
export interface SubtitleTrack {
  id: number
  media_id: number
  language: string
  label: string
  file_path: string
  is_default: boolean
}

// Audio Track Types
export interface AudioTrack {
  id: number
  media_id: number
  language: string
  label: string
  codec: string
  channels: number
  is_default: boolean
}

// Playback Info Request (body for POST /api/playback/{id}/info)
export interface PlaybackInfoRequest {
  video_codecs?: string[]
  audio_codecs?: string[]
  containers?: string[]
  max_height?: number
  media_file_id?: number
  selected_audio_track?: number // 0 = default
  selected_subtitle?: string // language code or 'off'
}

// Playback Info Types (from POST /api/playback/{id}/info)
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
  is_default: boolean
  selected: boolean
}

// External Subtitle Search Types
export interface SubtitleSearchResult {
  provider: 'opensubtitles' | 'podnapisi'
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
  method: string // DirectPlay, DirectStream, TranscodeAudio, FullTranscode
  stream_url: string
  abr_url?: string // adaptive bitrate HLS master playlist (populated when transcoding)
  video_codec: string
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
  position: number
}

// Activity Log
export interface ActivityLog {
  id: number
  user_id?: number
  username?: string
  action: string
  media_id?: number
  media_title?: string
  details: string
  ip_address: string
  created_at: string
}

export interface PlaybackStats {
  most_watched: Array<{ media_id: number; title: string; play_count: number }>
  most_active_users: Array<{ user_id: number; username: string; play_count: number }>
  plays_today: number
  plays_this_week: number
  plays_this_month: number
  total_plays: number
}

// Server Info
export interface ServerInfo {
  version: string
  uptime: string
  go_version: string
  os: string
  arch: string
  ffmpeg_version: string
  database: string
  hw_accel: string
  media_count: number
  series_count: number
  user_count: number
  total_size_bytes: number
}

// Library Stats
export interface LibraryStatsItem {
  id: number
  name: string
  type: string
  item_count: number
  file_count: number
  total_size_bytes: number
  last_scanned?: string
}

// Webhook
export interface Webhook {
  id: number
  url: string
  events: string
  active: boolean
  created_at: string
  updated_at: string
}

// Scheduled Tasks
export interface ScheduledTask {
  name: string
  interval: string
  last_run?: string
  next_run: string
  running: boolean
}
