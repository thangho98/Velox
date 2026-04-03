/**
 * Continue watching & next up hooks
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api'
import type {
  ContinueWatchingItem,
  NextUpItem,
  ContinueWatchingParams,
  NextUpParams,
} from '../../types'
import { userDataKeys, continueWatchingKeys, nextUpKeys } from './queryKeys'

// Continue Watching / Next Up API Functions
const continueWatchingApi = {
  list: (params: ContinueWatchingParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<ContinueWatchingItem[]>(`/profile/continue-watching${query ? `?${query}` : ''}`)
  },
  dismiss: (mediaId: number) => api.delete(`/profile/progress/${mediaId}/dismiss`),
}

const nextUpApi = {
  list: (params: NextUpParams = {}) => {
    const searchParams = new URLSearchParams()
    if (params.limit) searchParams.append('limit', String(params.limit))
    const query = searchParams.toString()
    return api.get<NextUpItem[]>(`/profile/next-up${query ? `?${query}` : ''}`)
  },
}

// React Query Hooks
export function useContinueWatching(params: ContinueWatchingParams = { limit: 20 }) {
  return useQuery({
    queryKey: continueWatchingKeys.list(params),
    queryFn: () => continueWatchingApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useNextUp(params: NextUpParams = { limit: 20 }) {
  return useQuery({
    queryKey: nextUpKeys.list(params),
    queryFn: () => nextUpApi.list(params),
    staleTime: 60 * 1000,
  })
}

export function useDismissProgress() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (mediaId: number) => continueWatchingApi.dismiss(mediaId),
    onSuccess: (_, mediaId) => {
      queryClient.invalidateQueries({ queryKey: userDataKeys.progress(mediaId) })
      queryClient.invalidateQueries({ queryKey: userDataKeys.recentlyWatched({}) })
      queryClient.invalidateQueries({ queryKey: continueWatchingKeys.all })
      queryClient.invalidateQueries({ queryKey: nextUpKeys.all })
    },
  })
}
