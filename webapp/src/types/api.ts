// API Response Types - matches backend JSON structure

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
  path: string
  type: string
  created_at: string
  updated_at: string
}

export interface CreateLibraryRequest {
  name: string
  path: string
  type: string
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
  overview: string
  release_date: string
  rating: number
  poster_path: string
  backdrop_path: string
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

export interface PlaybackInfo {
  media_id: number
  primary_file_id: number
  method: string // DirectPlay, DirectStream, TranscodeAudio, FullTranscode
  stream_url: string
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
