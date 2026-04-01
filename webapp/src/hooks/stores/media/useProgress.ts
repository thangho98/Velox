import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type {
  UserData,
  UpdateProgressRequest,
  ToggleFavoriteResponse,
  FavoritesListParams,
  RecentlyWatchedParams,
} from '@/types/api'
import { continueWatchingKeys, nextUpKeys } from './useContinueWatching'

// User Data API Functions (Progress, Favorites)
const userDataApi = {
  getProgress: (mediaId: number) => api.get<UserData | null>(`/profile/progress/${mediaId}`),
  updateProgress: (mediaId: number, data: UpdateProgressRequest) =>
    api.put<void>(`/profile/progress/${mediaId}`, data),
  listFavorites: (params: FavoritesListParams) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/favorites${query ? `?${query}` : ''}`)
  },
  toggleFavorite: (mediaId: number) =>
    api.post<ToggleFavoriteResponse>(`/profile/favorites/${mediaId}`, {}),
  listRecentlyWatched: (params: RecentlyWatchedParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/recently-watched${query ? `?${query}` : ''}`)
  },
}

// Query Keys
export const userDataKeys = {
  all: ['userData'] as const,
  progress: (mediaId: number) => [...userDataKeys.all, 'progress', mediaId] as const,
  favorites: (params: FavoritesListParams) => [...userDataKeys.all, 'favorites', params] as const,
  recentlyWatched: (params: RecentlyWatchedParams) =>
    [...userDataKeys.all, 'recentlyWatched', params] as const,
}

// React Query Hooks
export function useProgress(mediaId: number) {
  return useQuery({
    queryKey: userDataKeys.progress(mediaId),
    queryFn: () => userDataApi.getProgress(mediaId),
    staleTime: 30 * 1000, // 30s — progress is invalidated by useUpdateProgress on save
  })
}

export function useUpdateProgress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ mediaId, data }: { mediaId: number; data: UpdateProgressRequest }) =>
      userDataApi.updateProgress(mediaId, data),
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: userDataKeys.progress(variables.mediaId),
      })
      queryClient.invalidateQueries({
        queryKey: userDataKeys.recentlyWatched({}),
      })
      // Invalidate continue-watching and next-up when progress updates
      queryClient.invalidateQueries({
        queryKey: continueWatchingKeys.all,
      })
      queryClient.invalidateQueries({
        queryKey: nextUpKeys.all,
      })
    },
  })
}

export function useFavorites(params: FavoritesListParams = {}) {
  return useQuery({
    queryKey: userDataKeys.favorites(params),
    queryFn: () => userDataApi.listFavorites(params),
    staleTime: 60 * 1000,
  })
}

export function useToggleFavorite() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: userDataApi.toggleFavorite,
    onSuccess: () => {
      // Invalidate favorites list and any media item that might show favorite status
      queryClient.invalidateQueries({ queryKey: userDataKeys.all })
    },
  })
}

export function useRecentlyWatched(params: RecentlyWatchedParams = { limit: 20 }) {
  return useQuery({
    queryKey: userDataKeys.recentlyWatched(params),
    queryFn: () => userDataApi.listRecentlyWatched(params),
    staleTime: 60 * 1000,
  })
}
