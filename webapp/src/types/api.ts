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
  language: string
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
  tagline: string
  release_date: string
  rating: number
  imdb_rating: number
  rt_score: number
  metacritic_score: number
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  metadata_locked: boolean
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
  // Episode-only fields
  series_id?: number
  season_id?: number
  episode_number?: number
  season_number?: number
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
  series_id?: number
}

// SeriesListItem is a superset of Series — all Series fields + genres, season/episode counts.
// Used by GET /api/series list endpoint. No fields removed vs Series.
export interface SeriesListItem {
  id: number
  library_id: number
  title: string
  sort_title: string
  tmdb_id?: number
  imdb_id?: string
  tvdb_id?: number
  overview: string
  status: string
  network: string
  first_air_date: string
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  metadata_locked: boolean
  created_at: string
  updated_at: string
  // Extra fields beyond Series
  genres: string[]
  season_count?: number
  episode_count?: number
}

// Series Types (from GET /api/series)
export interface Series {
  id: number
  library_id: number
  title: string
  sort_title: string
  tmdb_id?: number
  imdb_id?: string
  tvdb_id?: number
  overview: string
  status: string // "Returning Series" | "Ended" | "Canceled"
  network: string
  first_air_date: string
  poster_path: string
  backdrop_path: string
  logo_path: string
  thumb_path: string
  metadata_locked: boolean
  created_at: string
  updated_at: string
}

// Metadata editing types
export interface MetadataEditRequest {
  title?: string
  sort_title?: string
  overview?: string
  tagline?: string
  release_date?: string
  rating?: number
  genres?: string[]
  credits?: CreditInput[]
  save_nfo?: boolean
  metadata_locked?: boolean
}

export interface SeriesMetadataEditRequest {
  title?: string
  sort_title?: string
  overview?: string
  status?: string
  network?: string
  first_air_date?: string
  genres?: string[]
  credits?: CreditInput[]
  save_nfo?: boolean
  metadata_locked?: boolean
}

export interface CreditInput {
  person_name: string
  character?: string
  role: 'cast' | 'director' | 'writer'
  order: number
}

export interface EpisodeMetadataEditRequest {
  title?: string
  overview?: string
  air_date?: string
  episode_number?: number
  metadata_locked?: boolean
}

export interface CreditWithPerson {
  credit: {
    id: number
    person_id: number
    character: string
    role: string
    display_order: number
  }
  person: {
    id: number
    name: string
    profile_path: string
  }
}

export interface SeriesWithSeasons {
  series: Series
  seasons: Season[]
}

export interface SeriesListParams {
  library_id?: number
  // Filter params (Plan M)
  search?: string
  genre?: string
  year?: string
  sort?: 'newest' | 'oldest' | 'rating' | 'title'
  limit?: number
  offset?: number
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
  media_id: number
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
  // Filter params (Plan M)
  search?: string
  genre?: string
  year?: string
  sort?: 'newest' | 'oldest' | 'rating' | 'title'
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
  selected_subtitle_id?: number // exact subtitle track id
}

// Skip Segment Types (for intro/credits skip)
export interface SkipSegment {
  type: 'intro' | 'credits'
  start: number
  end: number
  source: string
  confidence: number
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
  bitrate: number
  sample_rate: number
  is_default: boolean
  selected: boolean
}

// External Subtitle Search Types
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
  method: string // DirectPlay, DirectStream, TranscodeAudio, FullTranscode
  stream_url: string
  abr_url?: string // adaptive bitrate HLS master playlist (populated when transcoding)
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
  position: number
  skip_segments?: SkipSegment[]
}

// Continue Watching / Next Up Types
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

// Genre Types
export interface Genre {
  id: number
  name: string
  tmdb_id?: number
}

// Plan M: Browse Folder Types — matches backend BrowseResult/BrowseFolderItem
export interface BrowseFolderItem {
  name: string
  path: string
  media_count?: number
  poster?: string
}

export interface BrowseResult {
  library_id: number
  path: string
  parent: string
  folders: BrowseFolderItem[]
  media: MediaListItem[]
}

// Unified Search Result (from GET /api/search)
export interface SearchResult {
  movies: MediaListItem[]
  series: SeriesListItem[]
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
