/**
 * Series, Season, Episode hooks
 */

import { useQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type {
  Season,
  Episode,
  Series,
  SeriesListItem,
  SeriesListParams,
} from '../../types'

// Season/Episode API Functions
const seriesApi = {
  list: (params: SeriesListParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.search) searchParams.append('search', params.search)
    if (params.genre) searchParams.append('genre', params.genre)
    if (params.year) searchParams.append('year', params.year)
    if (params.sort) searchParams.append('sort', params.sort)
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    const query = searchParams.toString()
    return api.get<SeriesListItem[]>(`/series${query ? `?${query}` : ''}`)
  },
  get: (id: number) => api.get<Series>(`/series/${id}`),
  search: (query: string, limit = 20) =>
    api.get<Series[]>(`/series/search?q=${encodeURIComponent(query)}&limit=${limit}`),
  getSeasons: (seriesId: number) => api.get<Season[]>(`/series/${seriesId}/seasons`),
  getEpisodes: (seriesId: number, seasonId: number) =>
    api.get<Episode[]>(`/series/${seriesId}/seasons/${seasonId}/episodes`),
  getEpisode: (episodeId: number) => api.get<Episode>(`/episodes/${episodeId}`),
}

// Query Keys
export const seriesKeys = {
  all: ['series'] as const,
  list: (params: SeriesListParams) => [...seriesKeys.all, 'list', params] as const,
  detail: (id: number) => [...seriesKeys.all, 'detail', id] as const,
  search: (query: string) => [...seriesKeys.all, 'search', query] as const,
  seasons: (seriesId: number) => [...seriesKeys.all, 'seasons', seriesId] as const,
  episodes: (seriesId: number, seasonId: number) =>
    [...seriesKeys.all, 'episodes', seriesId, seasonId] as const,
  episode: (episodeId: number) => [...seriesKeys.all, 'episode', episodeId] as const,
}

// React Query Hooks
export function useSeriesList(params: SeriesListParams = {}) {
  return useQuery({
    queryKey: seriesKeys.list(params),
    queryFn: () => seriesApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useSeriesDetail(id: number) {
  return useQuery({
    queryKey: seriesKeys.detail(id),
    queryFn: () => seriesApi.get(id),
    staleTime: 5 * 60 * 1000,
    enabled: id > 0,
  })
}

export function useSeriesSearch(query: string, limit = 20) {
  return useQuery({
    queryKey: seriesKeys.search(query),
    queryFn: () => seriesApi.search(query, limit),
    staleTime: 60 * 1000,
    enabled: query.length > 0,
  })
}

export function useSeasons(seriesId: number) {
  return useQuery({
    queryKey: seriesKeys.seasons(seriesId),
    queryFn: () => seriesApi.getSeasons(seriesId),
    staleTime: 5 * 60 * 1000,
    enabled: seriesId > 0,
  })
}

export function useEpisodes(seriesId: number, seasonId: number) {
  return useQuery({
    queryKey: seriesKeys.episodes(seriesId, seasonId),
    queryFn: () => seriesApi.getEpisodes(seriesId, seasonId),
    staleTime: 5 * 60 * 1000,
    enabled: seriesId > 0 && seasonId > 0,
  })
}

export function useEpisode(episodeId: number) {
  return useQuery({
    queryKey: seriesKeys.episode(episodeId),
    queryFn: () => seriesApi.getEpisode(episodeId),
    staleTime: 5 * 60 * 1000,
    enabled: episodeId > 0,
  })
}
