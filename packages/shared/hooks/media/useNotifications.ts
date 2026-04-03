/**
 * Notification hooks for mobile app
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../../api'

export interface NotificationData {
  library_id?: number
  media_id?: number
  series_id?: number
  episode_id?: number
  scanned_count?: number
  new_count?: number
  error_count?: number
  quality?: string
  duration_seconds?: number
  language?: string
  provider?: string
  media_title?: string
  media_type?: string
}

export interface Notification {
  id: number
  user_id?: number
  type: NotificationType
  title: string
  message: string
  data: NotificationData
  read: boolean
  created_at: string
  read_at?: string
}

export type NotificationType =
  | 'scan_complete'
  | 'media_added'
  | 'transcode_complete'
  | 'transcode_failed'
  | 'subtitle_downloaded'
  | 'identify_complete'
  | 'library_watcher'

interface NotificationListResponse {
  notifications: Notification[]
  unread_count: number
}

const notificationsKey = ['notifications'] as const

// API functions
const notificationsApi = {
  list: (unreadOnly = false, limit = 20, offset = 0) => {
    const params = new URLSearchParams()
    if (unreadOnly) params.append('unread_only', 'true')
    params.append('limit', String(limit))
    params.append('offset', String(offset))
    return api.get<NotificationListResponse>(`/notifications?${params.toString()}`)
  },

  markAsRead: (ids: number[]) =>
    api.patch('/notifications/read', { ids }),

  markAllAsRead: () =>
    api.patch('/notifications/read-all', {}),

  delete: (ids: number[]) =>
    api.post('/notifications/delete', { ids }),

  unreadCount: () =>
    api.get<{ count: number }>('/notifications/unread-count'),
}

export function useNotifications(unreadOnly = false, limit = 20, offset = 0) {
  return useQuery({
    queryKey: [...notificationsKey, 'list', { unreadOnly, limit, offset }],
    queryFn: () => notificationsApi.list(unreadOnly, limit, offset),
    staleTime: 30000,
  })
}

export function useUnreadCount() {
  return useQuery({
    queryKey: [...notificationsKey, 'unread-count'],
    queryFn: notificationsApi.unreadCount,
    staleTime: 10000,
  })
}

export function useMarkNotificationsAsRead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: notificationsApi.markAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}

export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: notificationsApi.markAllAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}

export function useDeleteNotifications() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: notificationsApi.delete,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}
