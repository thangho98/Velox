import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@velox/shared/api'

export interface LiveChannel {
  id: number
  playlistId: number
  channelId: string
  name: string
  logo: string
  groupTitle: string
  streamUrl: string
  country: string
  isActive: boolean
  isHidden: boolean
  createdAt: string
  updatedAt: string
  lastWatchedAt?: string
}

export interface LivePlaylist {
  id: number
  name: string
  url: string
  lastSyncedAt?: string
  syncIntervalHours: number
  isActive: boolean
  createdAt: string
  updatedAt: string
}

export interface LiveProgram {
  id: number
  channel_id: number
  title: string
  description?: string
  start_time: string
  end_time: string
}

export function useLiveTVGroups() {
  return useQuery({
    queryKey: ['livetv', 'groups'],
    queryFn: async () => {
      const data = await api.get<string[]>('/livetv/groups')
      return data || []
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useLiveTVChannels(group?: string | null) {
  return useQuery({
    queryKey: ['livetv', 'channels', group],
    queryFn: async () => {
      const params = new URLSearchParams()
      if (group) params.set('group', group)
      const data = await api.get<LiveChannel[]>(`/livetv/channels?${params.toString()}`)
      return data || []
    },
    staleTime: 5 * 60 * 1000,
  })
}

export function useLiveTVRecentChannels(limit: number = 20) {
  return useQuery({
    queryKey: ['livetv', 'recent', limit],
    queryFn: async () => {
      const data = await api.get<LiveChannel[]>(`/livetv/channels/recent?limit=${limit}`)
      return data || []
    },
    staleTime: 60 * 1000,
  })
}

export function useLiveTVChannel(id?: number | null) {
  return useQuery({
    queryKey: ['livetv', 'channel', id],
    queryFn: async () => {
      if (!id) return null
      const data = await api.get<LiveChannel>(`/livetv/channels/${id}`)
      return data || null
    },
    enabled: !!id,
    staleTime: 5 * 60 * 1000,
  })
}

export function useSearchLiveTVChannels(query: string) {
  return useQuery({
    queryKey: ['livetv', 'search', query],
    queryFn: async () => {
      if (!query.trim()) return []
      const data = await api.get<LiveChannel[]>(
        `/livetv/channels/search?q=${encodeURIComponent(query)}`,
      )
      return data || []
    },
    enabled: query.length > 2,
    staleTime: 5 * 60 * 1000,
  })
}

// ── Admin Hooks ──────────────────────────────────────────────────────────────

export function useLiveTVPlaylists() {
  return useQuery({
    queryKey: ['livetv', 'playlists'],
    queryFn: async () => {
      const data = await api.get<LivePlaylist[]>('/livetv/playlists')
      return data || []
    },
    staleTime: 5 * 60 * 1000,
  })
}
export function useLiveTVEpg(channelId?: number | null) {
  return useQuery({
    queryKey: ['livetv', 'epg', channelId],
    queryFn: async () => {
      if (!channelId) return []
      const data = await api.get<LiveProgram[]>(`/livetv/channels/${channelId}/epg`)
      return data || []
    },
    enabled: !!channelId,
    staleTime: 5 * 60 * 1000,
  })
}

export function useToggleChannelHidden() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: async (id: number) => {
      return api.post<{ success: boolean; is_hidden: boolean }>(
        `/livetv/channels/${id}/toggle-hidden`,
        {},
      )
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['livetv', 'channels'] })
      queryClient.invalidateQueries({ queryKey: ['livetv', 'search'] })
    },
  })
}
