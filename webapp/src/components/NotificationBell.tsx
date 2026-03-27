import { useState, useRef, useEffect, useLayoutEffect } from 'react'
import { createPortal } from 'react-dom'
import { LuBell, LuCheck, LuTrash2, LuWifi, LuWifiOff } from 'react-icons/lu'
import { useNavigate } from 'react-router'
import {
  useNotificationBell,
  useMarkNotificationsAsRead,
  useMarkAllNotificationsAsRead,
  useDeleteNotifications,
  type Notification,
} from '@/hooks/useNotifications'
import { useTranslation } from '@/hooks/useTranslation'

const NOTIFICATION_ICONS: Record<string, string> = {
  scan_complete: '🔍',
  media_added: '🎬',
  transcode_complete: '✅',
  transcode_failed: '❌',
  subtitle_downloaded: '📝',
  identify_complete: '🆔',
  library_watcher: '👁️',
}

export function NotificationBell() {
  const navigate = useNavigate()
  const [open, setOpen] = useState(false)
  const buttonRef = useRef<HTMLButtonElement>(null)
  const menuRef = useRef<HTMLDivElement>(null)
  const [pos, setPos] = useState({ top: 0, right: 0 })
  const { unreadCount, recentNotifications, connected } = useNotificationBell()
  const { mutate: markAsRead } = useMarkNotificationsAsRead()
  const { mutate: markAllAsRead } = useMarkAllNotificationsAsRead()
  const { mutate: deleteNotifications } = useDeleteNotifications()
  const { t } = useTranslation('navigation')

  useLayoutEffect(() => {
    if (!open || !buttonRef.current) return
    const rect = buttonRef.current.getBoundingClientRect()
    setPos({
      top: rect.bottom + 8,
      right: window.innerWidth - rect.right,
    })
  }, [open])

  useEffect(() => {
    if (!open) return
    function handleClick(e: MouseEvent) {
      if (
        menuRef.current &&
        !menuRef.current.contains(e.target as Node) &&
        buttonRef.current &&
        !buttonRef.current.contains(e.target as Node)
      ) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', handleClick)
    return () => document.removeEventListener('mousedown', handleClick)
  }, [open])

  const handleNotificationClick = (n: Notification) => {
    if (!n.read) {
      markAsRead([n.id])
    }

    if (n.data.media_id) {
      navigate(`/movies/${n.data.media_id}`)
    } else if (n.data.series_id) {
      navigate(`/series/${n.data.series_id}`)
    } else if (n.data.library_id) {
      navigate(`/browse?library=${n.data.library_id}`)
    }

    setOpen(false)
  }

  const handleMarkAllAsRead = () => {
    markAllAsRead()
  }

  const handleDelete = (id: number) => {
    deleteNotifications([id])
  }

  return (
    <>
      <button
        ref={buttonRef}
        onClick={() => setOpen(!open)}
        className="relative p-2 text-gray-300 transition-colors hover:text-white"
        aria-label="Notifications"
      >
        <LuBell size={20} />
        {unreadCount > 0 ? (
          <span className="absolute -right-0.5 -top-0.5 flex h-5 w-5 items-center justify-center rounded-full bg-netflix-red text-xs font-bold text-white">
            {unreadCount > 9 ? '9+' : unreadCount}
          </span>
        ) : !connected ? (
          <span className="absolute -right-0.5 -top-0.5 h-2 w-2 rounded-full bg-yellow-500" />
        ) : null}
      </button>

      {open &&
        createPortal(
          <div
            ref={menuRef}
            className="fixed z-[9999] overflow-hidden rounded-lg bg-[#1a1a1a] shadow-2xl ring-1 ring-white/10 max-sm:inset-x-2 max-sm:top-14 sm:w-[360px]"
            style={window.innerWidth >= 640 ? { top: pos.top, right: pos.right } : undefined}
          >
            {/* Header */}
            <div className="flex items-center justify-between border-b border-white/10 px-4 py-3">
              <div className="flex items-center gap-2">
                <h3 className="font-semibold text-white">{t('notifications.title')}</h3>
                {connected ? (
                  <LuWifi size={14} className="text-green-500" />
                ) : (
                  <LuWifiOff size={14} className="text-yellow-500" />
                )}
              </div>
              {unreadCount > 0 && (
                <button
                  onClick={handleMarkAllAsRead}
                  className="text-xs text-netflix-red hover:text-red-400"
                >
                  {t('notifications.markAllRead')}
                </button>
              )}
            </div>

            {/* Notification List */}
            <div className="max-h-[400px] overflow-y-auto">
              {recentNotifications.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-gray-400">
                  <LuBell size={32} className="mb-2 opacity-50" />
                  <p className="text-sm">{t('notifications.empty')}</p>
                </div>
              ) : (
                recentNotifications.map((n) => (
                  <div
                    key={n.id}
                    className={`group relative flex gap-3 border-b border-white/5 px-4 py-3 transition-colors last:border-0 hover:bg-white/5 ${
                      !n.read ? 'bg-white/[0.03]' : ''
                    }`}
                  >
                    {/* Icon */}
                    <div className="flex-shrink-0 text-lg">
                      {NOTIFICATION_ICONS[n.type] || '🔔'}
                    </div>

                    {/* Content */}
                    <div
                      className="flex-1 cursor-pointer"
                      onClick={() => handleNotificationClick(n)}
                    >
                      <p
                        className={`text-sm ${n.read ? 'text-gray-300' : 'font-medium text-white'}`}
                      >
                        {n.title}
                      </p>
                      <p className="mt-0.5 line-clamp-2 text-xs text-gray-400">{n.message}</p>
                      <p className="mt-1 text-xs text-gray-500">{formatTime(n.created_at, t)}</p>
                    </div>

                    {/* Actions */}
                    <div className="flex flex-col items-end gap-1">
                      {!n.read && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation()
                            markAsRead([n.id])
                          }}
                          className="rounded p-1 text-gray-500 opacity-0 transition-all hover:bg-white/10 hover:text-white group-hover:opacity-100"
                          title={t('notifications.markRead')}
                        >
                          <LuCheck size={14} />
                        </button>
                      )}
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDelete(n.id)
                        }}
                        className="rounded p-1 text-gray-500 opacity-0 transition-all hover:bg-red-500/10 hover:text-red-400 group-hover:opacity-100"
                        title={t('notifications.delete')}
                      >
                        <LuTrash2 size={14} />
                      </button>
                    </div>
                  </div>
                ))
              )}
            </div>

            {/* Footer */}
            {recentNotifications.length > 0 && (
              <div className="border-t border-white/10 px-4 py-2">
                <button
                  onClick={() => {
                    navigate('/settings?section=notifications')
                    setOpen(false)
                  }}
                  className="w-full text-center text-xs text-gray-400 hover:text-white"
                >
                  {t('notifications.viewAll')}
                </button>
              </div>
            )}
          </div>,
          document.body,
        )}
    </>
  )
}

function formatTime(
  isoString: string,
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  const date = new Date(isoString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) return t('notifications.time.justNow')
  if (diff < 3600000) return t('notifications.time.minutesAgo', { count: Math.floor(diff / 60000) })
  if (diff < 86400000)
    return t('notifications.time.hoursAgo', { count: Math.floor(diff / 3600000) })
  if (diff < 604800000)
    return t('notifications.time.daysAgo', { count: Math.floor(diff / 86400000) })
  return date.toLocaleDateString()
}
