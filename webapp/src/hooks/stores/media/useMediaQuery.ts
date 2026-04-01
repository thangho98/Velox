import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type { Media, MediaWithFiles, MediaListItem, MediaListParams } from '@/types/api'

// Media API Functions
const mediaApi = {
  list: (params: MediaListParams) => {
    const searchParams = new URLSearchParams()
    if (params.library_id) searchParams.append('library_id', String(params.library_id))
    if (params.type) searchParams.append('type', params.type)
    // Plan M: filter params
    if (params.search) searchParams.append('search', params.search)
    if (params.genre) searchParams.append('genre', params.genre)
    if (params.year) searchParams.append('year', params.year)
    if (params.sort) searchParams.append('sort', params.sort)
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))

    const query = searchParams.toString()
    return api.get<MediaListItem[]>(`/media${query ? `?${query}` : ''}`)
  },
  get: (id: number) => api.get<Media>(`/media/${id}`),
  getWithFiles: (id: number) => api.get<MediaWithFiles>(`/media/${id}/files`),
}

// Query Keys
export const mediaKeys = {
  all: ['media'] as const,
  list: (params: MediaListParams) => [...mediaKeys.all, 'list', params] as const,
  detail: (id: number) => [...mediaKeys.all, 'detail', id] as const,
  withFiles: (id: number) => [...mediaKeys.all, 'withFiles', id] as const,
}

// React Query Hooks
export function useMediaList(params: MediaListParams = {}) {
  return useQuery({
    queryKey: mediaKeys.list(params),
    queryFn: () => mediaApi.list(params),
    staleTime: 60 * 1000, // 1 minute
  })
}

export function useMedia(id: number) {
  return useQuery({
    queryKey: mediaKeys.detail(id),
    queryFn: () => mediaApi.get(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: id > 0,
  })
}

export function useMediaWithFiles(id: number) {
  return useQuery({
    queryKey: mediaKeys.withFiles(id),
    queryFn: () => mediaApi.getWithFiles(id),
    staleTime: 5 * 60 * 1000, // 5 minutes
    enabled: id > 0,
  })
}
