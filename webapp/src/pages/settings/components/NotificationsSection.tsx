import { useState } from 'react'
import { useNavigate } from 'react-router'
import { LuCheck, LuTrash2, LuBell } from 'react-icons/lu'
import { Select } from '@/components/ui/Select'
import {
  useNotifications,
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

export function NotificationsSection() {
  const { t } = useTranslation('settings')
  const { t: tNav } = useTranslation('navigation')
  const navigate = useNavigate()
  const [filter, setFilter] = useState<string>('all')
  const [limit, setLimit] = useState(50)
  const unreadOnly = filter === 'unread'
  const { data, isLoading } = useNotifications(unreadOnly, limit, 0)
  const { mutate: markAsRead } = useMarkNotificationsAsRead()
  const { mutate: markAllAsRead } = useMarkAllNotificationsAsRead()
  const { mutate: deleteNotifications } = useDeleteNotifications()

  const notifications = data?.notifications ?? []
  const unreadCount = data?.unread_count ?? 0

  const handleClick = (n: Notification) => {
    if (!n.read) markAsRead([n.id])
    if (n.data.media_id) navigate(`/movies/${n.data.media_id}`)
    else if (n.data.series_id) navigate(`/series/${n.data.series_id}`)
    else if (n.data.library_id) navigate(`/browse?library=${n.data.library_id}`)
  }

  return (
    <div>
      <h2 className="mb-1 text-2xl font-bold">{t('sections.notifications.title')}</h2>
      <p className="mb-6 text-sm text-gray-400">{t('sections.notifications.description')}</p>

      {/* Toolbar */}
      <div className="mb-4 flex items-center gap-4">
        <Select
          value={filter}
          onChange={(e) => setFilter(e.target.value)}
          className="rounded bg-netflix-gray px-3 py-2 text-sm text-white"
        >
          <option value="all">{t('sections.notifications.filterAll')}</option>
          <option value="unread">
            {t('sections.notifications.filterUnread')} ({unreadCount})
          </option>
        </Select>
        <Select
          value={limit}
          onChange={(e) => setLimit(Number(e.target.value))}
          className="rounded bg-netflix-gray px-3 py-2 text-sm text-white"
        >
          <option value={25}>25</option>
          <option value={50}>50</option>
          <option value={100}>100</option>
        </Select>
        {unreadCount > 0 && (
          <button
            onClick={() => markAllAsRead()}
            className="ml-auto flex items-center gap-1.5 rounded bg-netflix-gray px-3 py-2 text-sm text-white transition-colors hover:bg-netflix-red"
          >
            <LuCheck size={14} />
            {tNav('notifications.markAllRead')}
          </button>
        )}
      </div>

      {/* List */}
      {isLoading ? (
        <p className="py-8 text-center text-sm text-gray-400">Loading...</p>
      ) : notifications.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-lg bg-netflix-gray/30 py-12">
          <LuBell size={36} className="mb-3 text-gray-500" />
          <p className="text-sm text-gray-400">{tNav('notifications.empty')}</p>
        </div>
      ) : (
        <div className="divide-y divide-white/5 rounded-lg bg-netflix-gray/30">
          {notifications.map((n) => (
            <div
              key={n.id}
              className={`group flex items-start gap-3 px-4 py-3 transition-colors hover:bg-white/5 ${
                !n.read ? 'bg-white/[0.03]' : ''
              }`}
            >
              {/* Icon */}
              <div className="mt-0.5 flex-shrink-0 text-lg">
                {NOTIFICATION_ICONS[n.type] || '🔔'}
              </div>

              {/* Content */}
              <div className="flex-1 cursor-pointer" onClick={() => handleClick(n)}>
                <p className={`text-sm ${n.read ? 'text-gray-300' : 'font-medium text-white'}`}>
                  {n.title}
                </p>
                <p className="mt-0.5 text-xs text-gray-400">{n.message}</p>
                <p className="mt-1 text-xs text-gray-500">
                  {new Date(n.created_at).toLocaleString()}
                </p>
              </div>

              {/* Actions */}
              <div className="flex items-center gap-1">
                {!n.read && (
                  <button
                    onClick={() => markAsRead([n.id])}
                    className="rounded p-1.5 text-gray-500 opacity-0 transition-all hover:bg-white/10 hover:text-white group-hover:opacity-100"
                    title={tNav('notifications.markRead')}
                  >
                    <LuCheck size={14} />
                  </button>
                )}
                <button
                  onClick={() => deleteNotifications([n.id])}
                  className="rounded p-1.5 text-gray-500 opacity-0 transition-all hover:bg-red-500/10 hover:text-red-400 group-hover:opacity-100"
                  title={tNav('notifications.delete')}
                >
                  <LuTrash2 size={14} />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
