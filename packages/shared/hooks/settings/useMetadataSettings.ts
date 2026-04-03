/**
 * Metadata provider settings hooks
 */

import { useMutation } from '@tanstack/react-query'
import { api } from '../../api'
import { createSettingsHooks } from './factory'

// --- Types ---

interface OpenSubsSettings {
  api_key: string
  username: string
  password_set: boolean
}

interface OpenSubsUpdateRequest {
  api_key: string
  username: string
  password: string
}

interface TMDbSettings {
  api_key: string
  has_builtin: boolean
}

interface OMDbSettings {
  api_key: string
  has_builtin: boolean
}

interface TVDBSettings {
  api_key: string
  has_builtin: boolean
}

interface FanartSettings {
  api_key: string
  has_builtin: boolean
}

// --- Hooks (factory-generated) ---

export const [useTMDbSettings, useUpdateTMDbSettings] = createSettingsHooks<
  TMDbSettings,
  { api_key: string }
>('tmdb')

export const [useOMDbSettings, useUpdateOMDbSettings] = createSettingsHooks<
  OMDbSettings,
  { api_key: string }
>('omdb')

export const [useTVDBSettings, useUpdateTVDBSettings] = createSettingsHooks<
  TVDBSettings,
  { api_key: string }
>('tvdb')

export const [useFanartSettings, useUpdateFanartSettings] = createSettingsHooks<
  FanartSettings,
  { api_key: string }
>('fanart')

export const [useOpenSubsSettings, useUpdateOpenSubsSettings] = createSettingsHooks<
  OpenSubsSettings,
  OpenSubsUpdateRequest
>('opensubtitles')

// --- Non-factory hooks ---

export function useBulkRefreshRatings() {
  return useMutation({
    mutationFn: () => api.post<{ updated: number }>('/admin/metadata/refresh-ratings', {}),
  })
}
