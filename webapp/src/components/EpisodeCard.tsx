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
import { useTranslation } from '@/hooks/useTranslation'

interface EpisodeCardProps {
  episode: Episode
  isAdmin?: boolean
  onEdit?: (episode: Episode) => void
}

export function EpisodeCard({ episode, isAdmin, onEdit }: EpisodeCardProps) {
  const { t } = useTranslation('media')
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
      label: progress?.completed ? t('actions.markAsUnwatched') : t('actions.markAsWatched'),
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
      label: copied ? t('actions.copied') : t('actions.copyStreamUrl'),
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
      label: t('actions.editMetadata'),
      icon: <LuPencil size={16} />,
      onClick: () => onEdit?.(episode),
      adminOnly: true,
      separator: true,
    },
  ]

  return (
    <div className="group flex items-start gap-3 rounded-lg bg-netflix-dark/80 p-3 backdrop-blur-sm transition-colors hover:bg-netflix-gray sm:items-center sm:gap-4 sm:p-4">
      {/* Thumbnail */}
      <div className="relative flex h-16 w-24 flex-shrink-0 items-center justify-center overflow-hidden rounded bg-netflix-black sm:h-20 sm:w-32">
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
      <div className="min-w-0 flex-1">
        <div className="flex min-w-0 items-center gap-2 sm:gap-3">
          <span className="flex-shrink-0 text-base font-bold text-gray-500 sm:text-lg">
            {episode.episode_number}
          </span>
          <h3 className="truncate font-semibold text-white">{episode.title}</h3>
          {progress?.completed && <LuCheck size={16} className="flex-shrink-0 text-green-500" />}
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
            <p className="mt-1 text-xs text-gray-500">
              {t('detail.percentWatched', { percent: Math.round(progressPercent) })}
            </p>
          </div>
        )}
        {episode.duration && (
          <p className="mt-1 text-xs text-gray-500">{formatDuration(episode.duration)}</p>
        )}
      </div>

      {/* Actions */}
      <div className="flex flex-shrink-0 items-center gap-1 self-center sm:gap-2">
        <Link
          to={`/watch/${episode.media_id}`}
          className="flex h-9 w-9 items-center justify-center rounded-full bg-white/10 text-white hover:bg-netflix-red sm:hidden"
          aria-label={hasProgress ? t('actions.resume') : t('actions.play')}
        >
          <LuPlay size={16} />
        </Link>
        <Link
          to={`/watch/${episode.media_id}`}
          className="hidden items-center gap-2 rounded-full bg-white/10 px-4 py-2 text-sm font-medium text-white opacity-0 transition-all hover:bg-netflix-red group-hover:opacity-100 sm:flex"
        >
          <LuPlay size={16} />
          {hasProgress ? t('actions.resume') : t('actions.play')}
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
