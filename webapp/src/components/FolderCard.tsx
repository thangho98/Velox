import { LuFolder } from 'react-icons/lu'
import { resolveImageUrl } from '@/lib/image'

interface FolderCardProps {
  name: string
  path: string
  // Backend sends a pre-normalized URL string for folder previews; no
  // ImageResource available (folders derive poster from the first contained
  // media's path, not an image_metadata row). Blurhash not applicable here.
  poster?: string
  mediaCount?: number
  onClick: () => void
}

export function FolderCard({ name, poster, mediaCount, onClick }: FolderCardProps) {
  return (
    <button
      onClick={onClick}
      className="group relative flex flex-col gap-2 text-left transition-transform hover:scale-105"
    >
      <div className="aspect-[2/3] overflow-hidden rounded-lg bg-[#1a1a1a] border border-gray-800 group-hover:border-[#e50914] transition-colors">
        {poster ? (
          <img
            src={resolveImageUrl(poster)}
            alt={name}
            className="h-full w-full object-cover transition-opacity group-hover:opacity-80"
            loading="lazy"
          />
        ) : (
          <div className="flex h-full flex-col items-center justify-center gap-2">
            <LuFolder className="h-16 w-16 text-gray-600 group-hover:text-[#e50914] transition-colors" />
          </div>
        )}
      </div>
      <div>
        <p className="line-clamp-2 text-sm font-medium text-white group-hover:text-[#e50914]">
          {name}
        </p>
        {mediaCount != null && mediaCount > 0 && (
          <p className="text-xs text-gray-500">
            {mediaCount} {mediaCount === 1 ? 'item' : 'items'}
          </p>
        )}
      </div>
    </button>
  )
}
