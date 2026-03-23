import { useEffect, useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/fetch'

function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    return payload.exp * 1000 < Date.now()
  } catch {
    return true
  }
}

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

interface WebSocketMessage {
  type: 'notification' | 'ping' | 'pong'
  payload?: {
    notification?: Notification
  }
}

const notificationsKey = ['notifications'] as const

// API functions
const fetchNotifications = async (unreadOnly = false, limit = 20, offset = 0) => {
  const params = new URLSearchParams()
  if (unreadOnly) params.append('unread_only', 'true')
  params.append('limit', limit.toString())
  params.append('offset', offset.toString())
  return api.get<NotificationListResponse>(`/notifications?${params.toString()}`)
}

const markAsRead = async (ids: number[]) => {
  return api.patch('/notifications/read', { ids })
}

const markAllAsRead = async () => {
  return api.patch('/notifications/read-all', {})
}

const deleteNotifications = async (ids: number[]) => {
  return api.post('/notifications/delete', { ids })
}

const fetchUnreadCount = async () => {
  const res = await api.get<{ count: number }>('/notifications/unread-count')
  return res.count
}

// Hook for WebSocket connection — accepts queryClient to avoid module-level global
export function useWebSocket() {
  const [connected, setConnected] = useState(false)
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const reconnectAttempts = useRef(0)
  const queryClient = useQueryClient()
  // Use ref so the connect closure always sees the latest queryClient
  const queryClientRef = useRef(queryClient)
  useEffect(() => {
    queryClientRef.current = queryClient
  }, [queryClient])

  const MAX_RECONNECT_ATTEMPTS = 5
  const RECONNECT_DELAY = 3000

  useEffect(() => {
    function connect() {
      const token = localStorage.getItem('access_token')
      if (!token || isTokenExpired(token)) return

      const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const wsUrl = `${proto}//${window.location.host}/api/ws?token=${token}`

      try {
        const ws = new WebSocket(wsUrl)
        wsRef.current = ws

        ws.onopen = () => {
          setConnected(true)
          reconnectAttempts.current = 0
        }

        ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data)
            setLastMessage(message)
            if (message.type === 'notification' && message.payload?.notification) {
              queryClientRef.current.invalidateQueries({ queryKey: notificationsKey })
            }
          } catch {
            // Ignore invalid JSON
          }
        }

        ws.onclose = () => {
          setConnected(false)
          wsRef.current = null
          // Only reconnect if token is still valid
          const currentToken = localStorage.getItem('access_token')
          if (
            currentToken &&
            !isTokenExpired(currentToken) &&
            reconnectAttempts.current < MAX_RECONNECT_ATTEMPTS
          ) {
            reconnectAttempts.current++
            const delay = RECONNECT_DELAY * reconnectAttempts.current
            reconnectTimeoutRef.current = setTimeout(connect, delay)
          }
        }

        ws.onerror = () => {
          // Error handled by onclose
        }
      } catch {
        setConnected(false)
      }
    }

    connect()

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
        reconnectTimeoutRef.current = null
      }
      if (wsRef.current) {
        wsRef.current.close()
        wsRef.current = null
      }
    }
  }, [])

  return { connected, lastMessage }
}

export function useNotifications(unreadOnly = false, limit = 20, offset = 0) {
  return useQuery({
    queryKey: [...notificationsKey, 'list', { unreadOnly, limit, offset }],
    queryFn: () => fetchNotifications(unreadOnly, limit, offset),
    staleTime: 30000,
  })
}

export function useUnreadCount() {
  return useQuery({
    queryKey: [...notificationsKey, 'unread-count'],
    queryFn: fetchUnreadCount,
    staleTime: 10000,
  })
}

export function useMarkNotificationsAsRead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: markAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}

export function useMarkAllNotificationsAsRead() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: markAllAsRead,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}

export function useDeleteNotifications() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: deleteNotifications,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: notificationsKey })
    },
  })
}

// Combined hook for notification bell component
export function useNotificationBell() {
  const { data: unreadData } = useUnreadCount()
  const { data: notificationsData } = useNotifications(false, 10, 0)
  const { connected } = useWebSocket()

  return {
    unreadCount: unreadData ?? 0,
    recentNotifications: notificationsData?.notifications ?? [],
    connected,
  }
}
