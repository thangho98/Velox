// Media, Library, Genre, Browse, Search types

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

export interface CreditInput {
  person_name: string
  character?: string
  role: 'cast' | 'director' | 'writer'
  order: number
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

export interface Genre {
  id: number
  name: string
  tmdb_id?: number
}

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

export interface SearchResult {
  movies: MediaListItem[]
  series: SeriesListItem[]
}

// SeriesListItem is defined in series.ts — import for SearchResult
import type { SeriesListItem } from './series'
export type { SeriesListItem } from './series'
