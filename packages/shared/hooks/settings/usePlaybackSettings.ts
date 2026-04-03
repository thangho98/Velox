/**
 * Playback & cinema settings hooks
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api'
import { createSettingsHooks } from './factory'
import { settingsKeys } from './factory'

// --- Types ---

interface PlaybackSettings {
  playback_mode: 'auto' | 'direct_play'
}

export interface CinemaSettings {
  enabled: boolean
  max_trailers: string
  has_intro: boolean
}

// --- Hooks ---

export const [usePlaybackSettings, useUpdatePlaybackSettings] =
  createSettingsHooks<PlaybackSettings>('playback')

// Cinema has a custom key since it also has upload
const cinemaKey = settingsKeys.key('cinema')

export function useCinemaSettings() {
  return useQuery({
    queryKey: cinemaKey,
    queryFn: () => api.get<CinemaSettings>('/admin/settings/cinema'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateCinemaSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { enabled?: boolean; max_trailers?: string }) =>
      api.put<unknown>('/admin/settings/cinema', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cinemaKey })
    },
  })
}

export function useUploadCinemaIntro() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData()
      form.append('file', file)
      return api.uploadFormData<{ path: string }>('/admin/cinema/intro', form)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: cinemaKey })
    },
  })
}
