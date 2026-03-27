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
  has_builtin: boolean
}

interface TMDbUpdateRequest {
  api_key: string
}

interface OMDbSettings {
  api_key: string
  has_builtin: boolean
}

interface OMDbUpdateRequest {
  api_key: string
}

interface TVDBSettings {
  api_key: string
  has_builtin: boolean
}

interface TVDBUpdateRequest {
  api_key: string
}

interface FanartSettings {
  api_key: string
  has_builtin: boolean
}

interface FanartUpdateRequest {
  api_key: string
}

interface SubdlSettings {
  api_key: string
  has_builtin: boolean
}

interface SubdlUpdateRequest {
  api_key: string
}

interface DeepLSettings {
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
  deepl: () => [...settingsKeys.all, 'deepl'] as const,
  playback: () => [...settingsKeys.all, 'playback'] as const,
  autoSub: () => [...settingsKeys.all, 'auto-subtitles'] as const,
  cinema: () => [...settingsKeys.all, 'cinema'] as const,
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
    mutationFn: (data: TMDbUpdateRequest) => api.put<TMDbSettings>('/admin/settings/tmdb', data),
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
    mutationFn: (data: OMDbUpdateRequest) => api.put<OMDbSettings>('/admin/settings/omdb', data),
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
    mutationFn: (data: TVDBUpdateRequest) => api.put<TVDBSettings>('/admin/settings/tvdb', data),
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
    mutationFn: (data: FanartUpdateRequest) =>
      api.put<FanartSettings>('/admin/settings/fanart', data),
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
    mutationFn: (data: SubdlUpdateRequest) => api.put<SubdlSettings>('/admin/settings/subdl', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.subdl() })
    },
  })
}

export function useDeepLSettings() {
  return useQuery({
    queryKey: settingsKeys.deepl(),
    queryFn: () => api.get<DeepLSettings>('/admin/settings/deepl'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdateDeepLSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: DeepLSettings) => api.put<DeepLSettings>('/admin/settings/deepl', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: settingsKeys.deepl() })
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

export interface CinemaSettings {
  enabled: boolean
  max_trailers: string
  has_intro: boolean
}

export function useCinemaSettings() {
  return useQuery({
    queryKey: settingsKeys.cinema(),
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
      queryClient.invalidateQueries({ queryKey: settingsKeys.cinema() })
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
      queryClient.invalidateQueries({ queryKey: settingsKeys.cinema() })
    },
  })
}

export function useBulkRefreshRatings() {
  return useMutation({
    mutationFn: () => api.post<{ updated: number }>('/admin/metadata/refresh-ratings', {}),
  })
}

// --- Pre-transcode (Plan P) ---

export interface PretranscodeSettings {
  enabled: boolean
  schedule: string
  concurrency: string
}

export interface PretranscodeProfile {
  id: number
  name: string
  height: number
  video_bitrate: number
  audio_bitrate: number
  video_codec: string
  audio_codec: string
  enabled: boolean
}

export interface PretranscodeStatus {
  enabled: boolean
  schedule: string
  concurrency: number
  paused: boolean
  total: number
  done: number
  encoding: number
  failed: number
  queued: number
  disk_used: number
  current_file: string
  speed: string
}

export interface StorageEstimate {
  profiles: {
    profile_id: number
    profile_name: string
    height: number
    estimated_gb: number
    file_count: number
  }[]
  total_bytes: number
  disk_free_bytes: number
  file_count: number
}

const pretranscodeKeys = {
  all: ['pretranscode'] as const,
  settings: () => [...pretranscodeKeys.all, 'settings'] as const,
  status: () => [...pretranscodeKeys.all, 'status'] as const,
  profiles: () => [...pretranscodeKeys.all, 'profiles'] as const,
  estimate: (libraryId: number) => [...pretranscodeKeys.all, 'estimate', libraryId] as const,
}

export function usePretranscodeSettings() {
  return useQuery({
    queryKey: pretranscodeKeys.settings(),
    queryFn: () => api.get<PretranscodeSettings>('/admin/settings/pretranscode'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useUpdatePretranscodeSettings() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: Partial<PretranscodeSettings>) =>
      api.put<PretranscodeSettings>('/admin/settings/pretranscode', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.settings() })
    },
  })
}

export function usePretranscodeStatus() {
  return useQuery({
    queryKey: pretranscodeKeys.status(),
    queryFn: () => api.get<PretranscodeStatus>('/admin/pretranscode/status'),
    refetchInterval: 5000,
  })
}

export function usePretranscodeProfiles() {
  return useQuery({
    queryKey: pretranscodeKeys.profiles(),
    queryFn: () => api.get<PretranscodeProfile[]>('/admin/pretranscode/profiles'),
    staleTime: 5 * 60 * 1000,
  })
}

export function useTogglePretranscodeProfile() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, enabled }: { id: number; enabled: boolean }) =>
      api.put<unknown>(`/admin/pretranscode/profiles/${id}`, { enabled }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.profiles() })
    },
  })
}

export function usePretranscodeEstimate(libraryId: number) {
  return useQuery({
    queryKey: pretranscodeKeys.estimate(libraryId),
    queryFn: () => api.get<StorageEstimate>(`/admin/pretranscode/estimate?library_id=${libraryId}`),
    enabled: libraryId > 0,
    staleTime: 60 * 1000,
  })
}

export function useStartPretranscode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ enqueued: number }>('/admin/pretranscode/start', {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.status() })
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.settings() })
    },
  })
}

export function useStopPretranscode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ cancelled: number }>('/admin/pretranscode/stop', {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.status() })
    },
  })
}

export function useResumePretranscode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<unknown>('/admin/pretranscode/resume', {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.status() })
    },
  })
}

export function useCleanupPretranscode() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: () => api.post<{ removed: number }>('/admin/pretranscode/cleanup-files', {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.status() })
      queryClient.invalidateQueries({ queryKey: pretranscodeKeys.settings() })
    },
  })
}
