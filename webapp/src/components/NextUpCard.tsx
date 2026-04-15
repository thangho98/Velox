import { Link } from 'react-router'
import type { NextUpItem } from '@/types/api'
import { ResponsiveImage } from '@/components/ResponsiveImage'

interface NextUpCardProps {
  item: NextUpItem
}

export function NextUpCard({ item }: NextUpCardProps) {
  // Format: "S04E24 · Friends"
  const displayTitle = `S${item.season_number}E${item.episode_number} · ${item.series_title}`

  // Prefer episode still → episode backdrop → series poster.
  const imgData = item.still || item.backdrop || item.poster

  return (
    <Link
      to={`/watch/${item.media_id}`}
      className="group relative block w-full shrink-0 overflow-hidden rounded-lg bg-netflix-dark transition-transform hover:scale-105"
    >
      {/* Thumbnail */}
      <div className="aspect-video w-full overflow-hidden">
        {imgData ? (
          <ResponsiveImage
            data={imgData}
            sizes="780px"
            alt={item.episode_title}
            className="h-full w-full"
          />
        ) : (
          <div className="h-full w-full bg-netflix-gray" />
        )}
      </div>

      {/* Info overlay */}
      <div className="absolute inset-0 flex flex-col items-center justify-center bg-black/30 p-3 text-center transition-colors group-hover:bg-black/40">
        <p className="text-sm font-semibold text-white drop-shadow-lg">{displayTitle}</p>
        <p className="mt-1 text-xs text-gray-200 drop-shadow-md">{item.episode_title}</p>
      </div>
    </Link>
  )
}
