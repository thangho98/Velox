import { useState } from 'react'
import { Link } from 'react-router'
import {
  useDismissProgress,
  useProgress,
  useUpdateProgress,
  useStreamUrl,
} from '@/hooks/stores/useMedia'
import { LuFilm, LuPlay, LuPencil, LuLink, LuCheck, LuCircle } from 'react-icons/lu'
import { tmdbImage } from '@/lib/image'
import { ActionMenu } from '@/components/ActionMenu'
import type { ActionMenuItem } from '@/components/ActionMenu'
import type { Episode } from '@/types/api'

interface EpisodeCardProps {
  episode: Episode
  isAdmin?: boolean
  onEdit?: (episode: Episode) => void
}

export function EpisodeCard({ episode, isAdmin, onEdit }: EpisodeCardProps) {
  const { data: progress } = useProgress(episode.media_id)
  const { mutate: updateProgress } = useUpdateProgress()
  const { mutate: dismissProgress } = useDismissProgress()
  const { mutate: getStreamUrl } = useStreamUrl(episode.media_id)
  const [copied, setCopied] = useState(false)
  const duration = episode.duration || 0
  const hasProgress = !!progress && progress.position > 0 && !progress.completed && duration > 0
  const progressPercent = hasProgress ? Math.min(100, (progress.position / duration) * 100) : 0

  const menuItems: ActionMenuItem[] = [
    {
      label: progress?.completed ? 'Mark as unwatched' : 'Mark as watched',
      icon: progress?.completed ? <LuCircle size={16} /> : <LuCheck size={16} />,
      onClick: () => {
        if (progress?.completed) {
          dismissProgress(episode.media_id)
        } else {
          updateProgress({
            mediaId: episode.media_id,
            data: { position: duration > 0 ? duration : 0, completed: true },
          })
        }
      },
    },
    {
      label: copied ? 'Copied!' : 'Copy stream URL',
      icon: <LuLink size={16} className={copied ? 'text-green-400' : ''} />,
      onClick: () => {
        getStreamUrl(undefined, {
          onSuccess: (data) => {
            navigator.clipboard.writeText(data.direct_url)
            setCopied(true)
            setTimeout(() => setCopied(false), 2000)
          },
        })
      },
    },
    {
      label: 'Edit metadata',
      icon: <LuPencil size={16} />,
      onClick: () => onEdit?.(episode),
      adminOnly: true,
      separator: true,
    },
  ]

  return (
    <div className="group flex items-center gap-4 rounded-lg bg-netflix-dark/80 p-4 backdrop-blur-sm transition-colors hover:bg-netflix-gray">
      {/* Thumbnail */}
      <div className="relative flex h-20 w-32 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-netflix-black">
        {episode.still_path ? (
          <img
            src={tmdbImage(episode.still_path, 'w300')!}
            alt={episode.title}
            className="h-full w-full object-cover"
          />
        ) : (
          <div className="flex h-full w-full items-center justify-center">
            <LuFilm size={32} className="text-gray-600" />
          </div>
        )}
        <Link
          to={`/watch/${episode.media_id}`}
          className="absolute inset-0 flex items-center justify-center bg-black/60 opacity-0 transition-opacity group-hover:opacity-100"
        >
          <div className="rounded-full bg-netflix-red p-2">
            <LuPlay size={20} className="text-white" />
          </div>
        </Link>
      </div>

      {/* Info */}
      <div className="flex-1">
        <div className="flex items-center gap-3">
          <span className="text-lg font-bold text-gray-500">{episode.episode_number}</span>
          <h3 className="font-semibold text-white">{episode.title}</h3>
          {progress?.completed && <LuCheck size={16} className="text-green-500" />}
        </div>
        {episode.overview && (
          <p className="mt-1 line-clamp-2 text-sm text-gray-400">{episode.overview}</p>
        )}
        {hasProgress && (
          <div className="mt-2 max-w-sm">
            <div className="h-1 rounded-full bg-gray-700">
              <div
                className="h-1 rounded-full bg-netflix-red"
                style={{ width: `${progressPercent}%` }}
              />
            </div>
            <p className="mt-1 text-xs text-gray-500">{Math.round(progressPercent)}% watched</p>
          </div>
        )}
        {episode.duration && (
          <p className="mt-1 text-xs text-gray-500">{formatDuration(episode.duration)}</p>
        )}
      </div>

      {/* Actions */}
      <div className="flex items-center gap-2">
        <Link
          to={`/watch/${episode.media_id}`}
          className="flex items-center gap-2 rounded-full bg-white/10 px-4 py-2 text-sm font-medium text-white opacity-0 transition-all group-hover:opacity-100 hover:bg-netflix-red"
        >
          <LuPlay size={16} />
          {hasProgress ? 'Resume' : 'Play'}
        </Link>
        <ActionMenu items={menuItems} isAdmin={isAdmin} size={18} />
      </div>
    </div>
  )
}

function formatDuration(seconds: number): string {
  const hours = Math.floor(seconds / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (hours > 0) {
    return `${hours}h ${mins}m`
  }
  return `${mins}m`
}
