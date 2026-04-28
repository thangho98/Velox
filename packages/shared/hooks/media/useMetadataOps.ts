/**
 * Metadata refresh, edit, image upload hooks
 */

import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api'
import type {
  Media,
  MetadataEditRequest,
  SeriesMetadataEditRequest,
  EpisodeMetadataEditRequest,
} from '../../types'
import { mediaKeys } from './useMediaQuery'
import { seriesKeys } from './useSeries'

export function useRefreshMetadata(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<Media>(`/media/${mediaId}/refresh`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      queryClient.invalidateQueries({ queryKey: mediaKeys.detail(mediaId) })
    },
  })
}

export function useCloudProbe(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<unknown>(`/media/${mediaId}/cloud-probe`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      queryClient.invalidateQueries({ queryKey: ['streaming'] })
    },
  })
}

export function useDownloadToNas(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<unknown>(`/media/${mediaId}/download`, {}),
    onSuccess: () => {
      // Just invalidate download tasks or media keys if necessary
      queryClient.invalidateQueries({ queryKey: ['downloads'] })
    },
  })
}

export function useRemoveDownloadFromNas(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.delete<unknown>(`/media/${mediaId}/download`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['downloads'] })
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      // If it's part of a series, this will eventually trigger a re-fetch of episodes if configured
    },
  })
}


export function useDownloadSeriesToNas(seriesId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<unknown>(`/series/${seriesId}/download`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['downloads'] })
    },
  })
}

export function useRemoveSeriesDownloadFromNas(seriesId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.delete<unknown>(`/series/${seriesId}/download`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['downloads'] })
      queryClient.invalidateQueries({ queryKey: mediaKeys.all() }) // series and episodes might be updated
    },
  })
}

// --- Metadata Editor hooks ---

export function useEditMediaMetadata(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: MetadataEditRequest) => api.patch<Media>(`/media/${mediaId}/metadata`, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      queryClient.invalidateQueries({ queryKey: mediaKeys.detail(mediaId) })
    },
  })
}

export function useEditSeriesMetadata(seriesId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (req: SeriesMetadataEditRequest) =>
      api.patch<unknown>(`/series/${seriesId}/metadata`, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: seriesKeys.detail(seriesId) })
    },
  })
}

export function useUploadMediaImage(mediaId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ imageType, file }: { imageType: string; file: File }) => {
      const form = new FormData()
      form.append('image_type', imageType)
      form.append('file', file)
      return api.uploadFormData<{ path: string }>(`/media/${mediaId}/images`, form)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(mediaId) })
      queryClient.invalidateQueries({ queryKey: mediaKeys.detail(mediaId) })
    },
  })
}

export function useUploadSeriesImage(seriesId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ imageType, file }: { imageType: string; file: File }) => {
      const form = new FormData()
      form.append('image_type', imageType)
      form.append('file', file)
      return api.uploadFormData<{ path: string }>(`/series/${seriesId}/images`, form)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: seriesKeys.detail(seriesId) })
    },
  })
}

export function useUnlockMetadata(type: 'media' | 'series', id: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.delete(`/${type === 'media' ? 'media' : 'series'}/${id}/metadata/lock`),
    onSuccess: () => {
      if (type === 'media') {
        queryClient.invalidateQueries({ queryKey: mediaKeys.withFiles(id) })
      } else {
        queryClient.invalidateQueries({ queryKey: seriesKeys.detail(id) })
      }
    },
  })
}

export function useEditEpisodeMetadata(seriesId: number, seasonId: number) {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ episodeId, req }: { episodeId: number; req: EpisodeMetadataEditRequest }) =>
      api.patch<unknown>(`/episodes/${episodeId}/metadata`, req),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: seriesKeys.episodes(seriesId, seasonId) })
    },
  })
}
