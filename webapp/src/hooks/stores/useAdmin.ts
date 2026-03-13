import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'
import type {
  ServerInfo,
  LibraryStatsItem,
  ActivityLog,
  PlaybackStats,
  Webhook,
  ScheduledTask,
} from '@/types/api'

const adminKeys = {
  all: ['admin'] as const,
  server: () => [...adminKeys.all, 'server'] as const,
  libraryStats: () => [...adminKeys.all, 'library-stats'] as const,
  activity: (filters?: Record<string, string>) => [...adminKeys.all, 'activity', filters] as const,
  playbackStats: () => [...adminKeys.all, 'playback-stats'] as const,
  webhooks: () => [...adminKeys.all, 'webhooks'] as const,
  tasks: () => [...adminKeys.all, 'tasks'] as const,
}

export function useServerInfo() {
  return useQuery({
    queryKey: adminKeys.server(),
    queryFn: () => api.get<ServerInfo>('/admin/server'),
    staleTime: 30 * 1000,
  })
}

export function useLibraryStats() {
  return useQuery({
    queryKey: adminKeys.libraryStats(),
    queryFn: () => api.get<LibraryStatsItem[]>('/admin/stats/libraries'),
    staleTime: 60 * 1000,
  })
}

export function useActivity(filters?: Record<string, string>) {
  const params = filters ? '?' + new URLSearchParams(filters).toString() : ''
  return useQuery({
    queryKey: adminKeys.activity(filters),
    queryFn: () => api.get<ActivityLog[]>(`/admin/activity${params}`),
    staleTime: 10 * 1000,
    refetchInterval: 10 * 1000,
  })
}

export function usePlaybackStats() {
  return useQuery({
    queryKey: adminKeys.playbackStats(),
    queryFn: () => api.get<PlaybackStats>('/admin/stats/playback'),
    staleTime: 60 * 1000,
  })
}

export function useWebhooks() {
  return useQuery({
    queryKey: adminKeys.webhooks(),
    queryFn: () => api.get<Webhook[]>('/admin/webhooks'),
  })
}

export function useCreateWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (data: { url: string; events: string; active: boolean }) =>
      api.post<Webhook>('/admin/webhooks', data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.webhooks() })
    },
  })
}

export function useUpdateWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Webhook> }) =>
      api.put<Webhook>(`/admin/webhooks/${id}`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.webhooks() })
    },
  })
}

export function useDeleteWebhook() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => api.delete(`/admin/webhooks/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.webhooks() })
    },
  })
}

export function useScheduledTasks() {
  return useQuery({
    queryKey: adminKeys.tasks(),
    queryFn: () => api.get<ScheduledTask[]>('/admin/tasks'),
    staleTime: 10 * 1000,
  })
}

export function useRunTask() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (name: string) => api.post(`/admin/tasks/${name}/run`, {}),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: adminKeys.tasks() })
    },
  })
}
