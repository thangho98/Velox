/**
 * Batch fetch watch progress for multiple episodes
 */

import { useQuery } from '@tanstack/react-query'
import { api } from '../../api'
import type { UserData } from '../../types'

const episodesProgressApi = {
  // Batch fetch progress for multiple media IDs (returns null for unwatched)
  batchGetProgress: (mediaIds: number[]) =>
    Promise.all(mediaIds.map((id) =>
      api.get<UserData | null>(`/profile/progress/${id}`).catch(() => null)
    )),
}

export const episodesProgressKeys = {
  all: ['episodesProgress'] as const,
  batch: (mediaIds: number[]) => [...episodesProgressKeys.all, 'batch', ...mediaIds.sort()] as const,
}

export function useEpisodesProgress(mediaIds: number[]) {
  const validIds = mediaIds.filter((id) => id > 0)
  return useQuery({
    queryKey: episodesProgressKeys.batch(validIds),
    queryFn: () => episodesProgressApi.batchGetProgress(validIds),
    staleTime: 30 * 1000,
    // Don't refetch if we already have data
    enabled: validIds.length > 0,
  })
}
