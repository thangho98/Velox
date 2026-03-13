import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'

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
}

interface OMDbSettings {
  api_key: string
}

interface TVDBSettings {
  api_key: string
}

interface FanartSettings {
  api_key: string
}

const settingsKeys = {
  all: ['settings'] as const,
  openSubs: () => [...settingsKeys.all, 'opensubtitles'] as const,
  tmdb: () => [...settingsKeys.all, 'tmdb'] as const,
  omdb: () => [...settingsKeys.all, 'omdb'] as const,
  tvdb: () => [...settingsKeys.all, 'tvdb'] as const,
  fanart: () => [...settingsKeys.all, 'fanart'] as const,
}

export function useOpenSubsSettings() {
  return useQuery({
    queryKey: settingsKeys.openSubs(),
    queryFn: () => api.get<OpenSubsSettings>('/admin/settings/opensubtitles'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateOpenSubsSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: OpenSubsUpdateRequest) =>
      api.put<OpenSubsSettings>('/admin/settings/opensubtitles', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.openSubs() })
    },
  })
}

export function useTMDbSettings() {
  return useQuery({
    queryKey: settingsKeys.tmdb(),
    queryFn: () => api.get<TMDbSettings>('/admin/settings/tmdb'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateTMDbSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: TMDbSettings) => api.put<TMDbSettings>('/admin/settings/tmdb', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.tmdb() })
    },
  })
}

export function useOMDbSettings() {
  return useQuery({
    queryKey: settingsKeys.omdb(),
    queryFn: () => api.get<OMDbSettings>('/admin/settings/omdb'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateOMDbSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: OMDbSettings) => api.put<OMDbSettings>('/admin/settings/omdb', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.omdb() })
    },
  })
}

export function useTVDBSettings() {
  return useQuery({
    queryKey: settingsKeys.tvdb(),
    queryFn: () => api.get<TVDBSettings>('/admin/settings/tvdb'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateTVDBSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: TVDBSettings) => api.put<TVDBSettings>('/admin/settings/tvdb', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.tvdb() })
    },
  })
}

export function useFanartSettings() {
  return useQuery({
    queryKey: settingsKeys.fanart(),
    queryFn: () => api.get<FanartSettings>('/admin/settings/fanart'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateFanartSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: FanartSettings) => api.put<FanartSettings>('/admin/settings/fanart', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.fanart() })
    },
  })
}

export function useBulkRefreshRatings() {
  return useMutation({
    mutationFn: () => api.post<{ updated: number }>('/admin/metadata/refresh-ratings', {}),
  })
}
