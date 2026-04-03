/**
 * Library management hooks (admin)
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api'
import type { FsBrowseResponse, Library, CreateLibraryRequest } from '../../types'

// Filesystem browser API (admin only)
export function useFsBrowse(path: string) {
  return useQuery({
    queryKey: ['fs-browse', path],
    queryFn: () => api.get<FsBrowseResponse>(`/admin/fs/browse?path=${encodeURIComponent(path)}`),
    staleTime: 0,
  })
}

// Library API Functions
const libraryApi = {
  list: () => api.get<Library[]>('/libraries'),
  create: (data: CreateLibraryRequest) => api.post<Library>('/libraries', data),
  delete: (id: number) => api.delete(`/libraries/${id}`),
  scan: (id: number, force = false) =>
    api.post<void>(`/libraries/${id}/scan${force ? '?force=true' : ''}`, {}),
}

// Query Keys
export const libraryKeys = {
  all: ['libraries'] as const,
  list: () => [...libraryKeys.all, 'list'] as const,
  detail: (id: number) => [...libraryKeys.all, 'detail', id] as const,
}

// React Query Hooks
export function useLibraries() {
  return useQuery({
    queryKey: libraryKeys.list(),
    queryFn: libraryApi.list,
    staleTime: 5 * 60 * 1000,
  })
}

export function useCreateLibrary() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: libraryApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: libraryKeys.list() })
    },
  })
}

export function useDeleteLibrary() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: libraryApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: libraryKeys.list() })
    },
  })
}

export function useScanLibrary() {
  return useMutation({
    mutationFn: ({ id, force = false }: { id: number; force?: boolean }) =>
      libraryApi.scan(id, force),
  })
}
