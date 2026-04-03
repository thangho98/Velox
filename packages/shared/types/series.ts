// Series, Season, Episode types

import type { MediaFile, CreditInput } from './media'

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

export interface SeriesWithSeasons {
  series: Series
  seasons: Season[]
}

export interface SeriesListParams {
  library_id?: number
  search?: string
  genre?: string
  year?: string
  sort?: 'newest' | 'oldest' | 'rating' | 'title'
  limit?: number
  offset?: number
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

export interface EpisodeMetadataEditRequest {
  title?: string
  overview?: string
  air_date?: string
  episode_number?: number
  metadata_locked?: boolean
}

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
