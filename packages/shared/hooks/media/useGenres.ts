/**
 * Genre, search, folder browse hooks (Plan M)
 */

import { useQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type { CreditWithPerson, SearchResult, Genre, BrowseResult } from '../../types'

export function useAllGenres(type?: 'movie' | 'series') {
  return useGenres(type)
}

export function useMediaGenres(mediaId: number) {
  return useQuery({
    queryKey: ['media', mediaId, 'genres'],
    queryFn: () => api.get<{ id: number; name: string }[]>(`/media/${mediaId}/genres`),
    enabled: mediaId > 0,
    select: (data) => data ?? [],
  })
}

export function useMediaCredits(mediaId: number) {
  return useQuery({
    queryKey: ['media', mediaId, 'credits'],
    queryFn: () => api.get<CreditWithPerson[]>(`/media/${mediaId}/credits`),
    enabled: mediaId > 0,
    select: (data) => data ?? [],
  })
}

export function useSeriesGenres(seriesId: number) {
  return useQuery({
    queryKey: ['series', seriesId, 'genres'],
    queryFn: () => api.get<{ id: number; name: string }[]>(`/series/${seriesId}/genres`),
    enabled: seriesId > 0,
    select: (data) => data ?? [],
  })
}

export function useSeriesCredits(seriesId: number) {
  return useQuery({
    queryKey: ['series', seriesId, 'credits'],
    queryFn: () => api.get<CreditWithPerson[]>(`/series/${seriesId}/credits`),
    enabled: seriesId > 0,
    select: (data) => data ?? [],
  })
}

// Unified Search Hook
const searchApi = {
  search: (query: string, limit = 20) => {
    const searchParams = new URLSearchParams()
    searchParams.append('q', query)
    if (limit) searchParams.append('limit', String(limit))
    return api.get<SearchResult>(`/search?${searchParams.toString()}`)
  },
}

export const searchKeys = {
  all: ['search'] as const,
  query: (query: string) => [...searchKeys.all, query] as const,
}

export function useSearch(query: string, limit = 20) {
  return useQuery({
    queryKey: searchKeys.query(query),
    queryFn: () => searchApi.search(query, limit),
    staleTime: 60 * 1000,
    enabled: query.length > 0,
  })
}

// Enhanced Genres Hook with type filter
export function useGenres(type?: 'movie' | 'series') {
  const searchParams = new URLSearchParams()
  if (type) searchParams.append('type', type)
  const query = searchParams.toString()

  return useQuery({
    queryKey: ['genres', type ?? 'all'],
    queryFn: () => api.get<Genre[]>(`/genres${query ? `?${query}` : ''}`),
    staleTime: 5 * 60 * 1000,
  })
}

// Folder Browse Hook
const browseApi = {
  browse: (libraryId?: number, path: string = '') => {
    const searchParams = new URLSearchParams()
    if (libraryId) searchParams.append('library_id', String(libraryId))
    if (path) searchParams.append('path', path)
    return api.get<BrowseResult>(`/browse?${searchParams.toString()}`)
  },
}

export const browseKeys = {
  all: ['browse'] as const,
  folder: (libraryId: number | undefined, path: string) =>
    [...browseKeys.all, libraryId ?? 0, path] as const,
}

export function useFolderBrowse(libraryId?: number, path: string = '') {
  return useQuery({
    queryKey: browseKeys.folder(libraryId, path),
    queryFn: () => browseApi.browse(libraryId, path),
    staleTime: 60 * 1000,
  })
}
