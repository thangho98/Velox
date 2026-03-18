import { Link } from 'react-router'
import { useState } from 'react'
import { useDismissProgress } from '@/hooks/stores/useMedia'
import type { ContinueWatchingItem } from '@/types/api'
import { tmdbImage } from '@/lib/image'
import { LuX } from 'react-icons/lu'

interface ContinueWatchingCardProps {
  item: ContinueWatchingItem
}

export function ContinueWatchingCard({ item }: ContinueWatchingCardProps) {
  const [isDismissed, setIsDismissed] = useState(false)
  const dismissMutation = useDismissProgress()

  const handleDismiss = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsDismissed(true)
    dismissMutation.mutate(item.media_id)
  }

  if (isDismissed) return null

  // Calculate progress percentage
  const progressPercent = item.duration > 0 ? (item.position / item.duration) * 100 : 0

  // Calculate remaining time in minutes
  const remainingMinutes = Math.ceil((item.duration - item.position) / 60)

  // Format title: for episodes show "S04E23 · Friends" format
  const displayTitle =
    item.media_type === 'episode' && item.series_title
      ? `S${item.season_number}E${item.episode_number} · ${item.series_title}`
      : item.title

  // Use backdrop if available, fallback to poster
  const imageUrl =
    tmdbImage(item.backdrop_path, 'w780') ||
    tmdbImage(item.poster_path, 'w500') ||
    '/placeholder-backdrop.png'

  return (
    <Link
      to={`/watch/${item.media_id}`}
      className="group relative block w-full shrink-0 overflow-hidden rounded-lg bg-netflix-dark transition-transform hover:scale-105"
    >
      {/* Thumbnail */}
      <div className="aspect-video w-full overflow-hidden">
        <img
          src={imageUrl}
          alt={item.title}
          className="h-full w-full object-cover transition-transform duration-300 group-hover:scale-110"
          loading="lazy"
        />
      </div>

      {/* Progress bar */}
      <div className="absolute bottom-0 left-0 right-0 h-1 bg-gray-700">
        <div
          className="h-full bg-netflix-red transition-all duration-300"
          style={{ width: `${progressPercent}%` }}
        />
      </div>

      {/* Info */}
      <div className="absolute bottom-1 left-0 right-0 bg-gradient-to-t from-black/80 to-transparent p-3 pt-6">
        <p className="truncate text-sm font-medium text-white">{displayTitle}</p>
        <p className="text-xs text-gray-400">{remainingMinutes}m remaining</p>
      </div>

      {/* Dismiss button */}
      <button
        onClick={handleDismiss}
        className="absolute right-2 top-2 rounded-full bg-black/60 p-1.5 text-white opacity-0 transition-opacity hover:bg-netflix-red group-hover:opacity-100"
        aria-label="Dismiss"
      >
        <LuX size={16} />
      </button>
    </Link>
  )
}
