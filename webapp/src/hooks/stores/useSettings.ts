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

interface SubdlSettings {
  api_key: string
}

interface PlaybackSettings {
  playback_mode: 'auto' | 'direct_play'
}

interface AutoSubSettings {
  languages: string
}

const settingsKeys = {
  all: ['settings'] as const,
  openSubs: () => [...settingsKeys.all, 'opensubtitles'] as const,
  tmdb: () => [...settingsKeys.all, 'tmdb'] as const,
  omdb: () => [...settingsKeys.all, 'omdb'] as const,
  tvdb: () => [...settingsKeys.all, 'tvdb'] as const,
  fanart: () => [...settingsKeys.all, 'fanart'] as const,
  subdl: () => [...settingsKeys.all, 'subdl'] as const,
  playback: () => [...settingsKeys.all, 'playback'] as const,
  autoSub: () => [...settingsKeys.all, 'auto-subtitles'] as const,
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

export function useSubdlSettings() {
  return useQuery({
    queryKey: settingsKeys.subdl(),
    queryFn: () => api.get<SubdlSettings>('/admin/settings/subdl'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateSubdlSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: SubdlSettings) => api.put<SubdlSettings>('/admin/settings/subdl', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.subdl() })
    },
  })
}

export function usePlaybackSettings() {
  return useQuery({
    queryKey: settingsKeys.playback(),
    queryFn: () => api.get<PlaybackSettings>('/admin/settings/playback'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdatePlaybackSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: PlaybackSettings) =>
      api.put<PlaybackSettings>('/admin/settings/playback', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.playback() })
    },
  })
}

export function useAutoSubSettings() {
  return useQuery({
    queryKey: settingsKeys.autoSub(),
    queryFn: () => api.get<AutoSubSettings>('/admin/settings/auto-subtitles'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateAutoSubSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: AutoSubSettings) =>
      api.put<AutoSubSettings>('/admin/settings/auto-subtitles', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.autoSub() })
    },
  })
}

export function useBulkRefreshRatings() {
  return useMutation({
    mutationFn: () => api.post<{ updated: number }>('/admin/metadata/refresh-ratings', {}),
  })
}
