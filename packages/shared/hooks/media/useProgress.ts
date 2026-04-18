/**
 * User progress hooks — position, favorites, recently watched
 */

import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type {
  UserData,
  UpdateProgressRequest,
  ToggleFavoriteResponse,
  FavoritesListParams,
  RecentlyWatchedParams,
} from '../../types'
import { continueWatchingKeys, nextUpKeys, userDataKeys } from './queryKeys'

// User Data API Functions (Progress, Favorites)
const userDataApi = {
  getProgress: (mediaId: number) => api.get<UserData | null>(`/profile/progress/${mediaId}`),
  updateProgress: (mediaId: number, data: UpdateProgressRequest) =>
    api.put<void>(`/profile/progress/${mediaId}`, data),
  listFavorites: (params: FavoritesListParams) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    if (params.offset) searchParams.append('offset', String(params.offset))
    if (params.start_char) searchParams.append('start_char', params.start_char)
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/favorites${query ? `?${query}` : ''}`)
  },
  getFavoritesAlphabet: () => api.get<any[]>('/profile/favorites/alphabet'),
  toggleFavorite: (mediaId: number) =>
    api.post<ToggleFavoriteResponse>(`/profile/favorites/${mediaId}`, {}),
  listRecentlyWatched: (params: RecentlyWatchedParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<UserData[]>(`/profile/recently-watched${query ? `?${query}` : ''}`)
  },
}

// React Query Hooks
export function useProgress(mediaId: number) {
  return useQuery({
    queryKey: userDataKeys.progress(mediaId),
    queryFn: () => userDataApi.getProgress(mediaId),
    staleTime: 30 * 1000,
    enabled: mediaId > 0,
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

export function useInfiniteFavorites(params: FavoritesListParams = {}) {
  return useInfiniteQuery({
    queryKey: [...userDataKeys.favorites(params), 'infinite'],
    queryFn: ({ pageParam = 0 }) => userDataApi.listFavorites({ ...params, offset: pageParam as number, limit: params.limit || 100 }),
    initialPageParam: 0,
    getNextPageParam: (lastPage, allPages) => {
      const limit = params.limit || 100;
      if (lastPage.length < limit) return undefined;
      return allPages.length * limit;
    },
    staleTime: 60 * 1000,
  })
}

export function useFavoritesAlphabet() {
  return useQuery({
    queryKey: [...userDataKeys.all, 'favorites', 'alphabet'],
    queryFn: () => userDataApi.getFavoritesAlphabet(),
    staleTime: 60 * 1000,
  })
}

export function useToggleFavorite() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: userDataApi.toggleFavorite,
    onSuccess: () => {
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
