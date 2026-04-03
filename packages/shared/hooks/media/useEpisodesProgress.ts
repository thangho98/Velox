/**
 * Batch fetch watch progress for multiple episodes
 */

import { useQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type { UserData } from '../../types'

const episodesProgressApi = {
  // Batch fetch progress for multiple media IDs
  batchGetProgress: (mediaIds: number[]) =>
    Promise.all(mediaIds.map((id) => api.get<UserData | null>(`/profile/progress/${id}`))),
}

export const episodesProgressKeys = {
  all: ['episodesProgress'] as const,
  batch: (mediaIds: number[]) => [...episodesProgressKeys.all, 'batch', ...mediaIds.sort()] as const,
}

export function useEpisodesProgress(mediaIds: number[]) {
  return useQuery({
    queryKey: episodesProgressKeys.batch(mediaIds),
    queryFn: () => episodesProgressApi.batchGetProgress(mediaIds),
    staleTime: 30 * 1000,
    // Don't refetch if we already have data
    enabled: mediaIds.length > 0,
  })
}
