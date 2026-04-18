/**
 * Media query hooks — list, detail, with files
 */

import { useQuery, useInfiniteQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type { Media, MediaWithFiles, MediaListItem, MediaListParams, AlphabetCount } from '../../types'

// Media API Functions
const mediaApi = {
  list: (params: MediaListParams) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.type) searchParams.append('type', params.type)
    if (params.search) searchParams.append('search', params.search)
    if (params.genre) searchParams.append('genre', params.genre)
    if (params.year) searchParams.append('year', params.year)
    if (params.sort) searchParams.append('sort', params.sort)
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    if (params.start_char) searchParams.append('start_char', params.start_char)

    const query = searchParams.toString()
    return api.get<MediaListItem[]>(`/media${query ? `?${query}` : ''}`)
  },
  alphabet: (params: MediaListParams) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.type) searchParams.append('type', params.type)
    if (params.genre) searchParams.append('genre', params.genre)
    if (params.year) searchParams.append('year', params.year)
    
    const query = searchParams.toString()
    return api.get<AlphabetCount[]>(`/media/alphabet${query ? `?${query}` : ''}`)
  },
  get: (id: number) => api.get<Media>(`/media/${id}`),
  getWithFiles: (id: number) => api.get<MediaWithFiles>(`/media/${id}/files`),
}

// Query Keys
export const mediaKeys = {
  all: ['media'] as const,
  list: (params: MediaListParams) => [...mediaKeys.all, 'list', params] as const,
  alphabet: (params: MediaListParams) => [...mediaKeys.all, 'alphabet', params] as const,
  detail: (id: number) => [...mediaKeys.all, 'detail', id] as const,
  withFiles: (id: number) => [...mediaKeys.all, 'withFiles', id] as const,
}

// React Query Hooks
export function useMediaList(params: MediaListParams = {}) {
  return useQuery({
    queryKey: mediaKeys.list(params),
    queryFn: () => mediaApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useMediaAlphabet(params: MediaListParams = {}) {
  // Omit fields that don't affect alphabet counts
  const filteredParams = {
    library_id: params.library_id,
    type: params.type,
    genre: params.genre,
    year: params.year,
  }
  return useQuery({
    queryKey: mediaKeys.alphabet(filteredParams),
    queryFn: () => mediaApi.alphabet(filteredParams),
    staleTime: 5 * 60 * 1000,
  })
}

export function useInfiniteMediaList(params: MediaListParams = {}) {
  return useInfiniteQuery({
    queryKey: [...mediaKeys.list(params), 'infinite'],
    queryFn: ({ pageParam = 0 }) => mediaApi.list({ ...params, offset: pageParam as number, limit: params.limit || 100 }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      const limit = params.limit || 100;
      if (lastPage.length < limit) return undefined;
      return allPages.length * limit;
    },
    staleTime: 60 * 1000,
  })
}

export function useMedia(id: number) {
  return useQuery({
    queryKey: mediaKeys.detail(id),
    queryFn: () => mediaApi.get(id),
    staleTime: 5 * 60 * 1000,
    enabled: id > 0,
  })
}

export function useMediaWithFiles(id: number) {
  return useQuery({
    queryKey: mediaKeys.withFiles(id),
    queryFn: () => mediaApi.getWithFiles(id),
    staleTime: 5 * 60 * 1000,
    enabled: id > 0,
  })
}
