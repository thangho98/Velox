import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'

// --- Types ---

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

// --- Query Keys ---

const pretranscodeKeys = {
  all: ['pretranscode'] as const,
  settings: () => [...pretranscodeKeys.all, 'settings'] as const,
  status: () => [...pretranscodeKeys.all, 'status'] as const,
  profiles: () => [...pretranscodeKeys.all, 'profiles'] as const,
  estimate: (libraryId: number) => [...pretranscodeKeys.all, 'estimate', libraryId] as const,
}

// --- Hooks ---

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
