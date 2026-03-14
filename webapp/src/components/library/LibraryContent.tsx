import { memo } from 'react'
import { Link } from 'react-router'
import { LuChevronRight, LuFilm, LuLayoutGrid, LuList } from 'react-icons/lu'
import { MediaCard } from '@/components/MediaCard'
import { useMediaList } from '@/hooks/stores/useMedia'
import { useUIStore } from '@/stores/ui'

interface LibraryContentProps {
  libraryId: number
  libraryName: string
}

export const LibraryContent = memo(function LibraryContent({
  libraryId,
  libraryName,
}: LibraryContentProps) {
  const { libraryViewMode, setLibraryViewMode } = useUIStore()
  const { data: media, isLoading } = useMediaList({ library_id: libraryId, limit: 50 })

  return (
    <div>
      <div className="mb-4 flex items-center justify-between">
        <h2 className="text-lg font-semibold text-white">{libraryName}</h2>
        <div className="flex gap-2">
          <button
            onClick={() => setLibraryViewMode('grid')}
            className={`flex items-center gap-2 rounded px-3 py-1.5 text-sm transition-colors ${
              libraryViewMode === 'grid'
                ? 'bg-netflix-red text-white'
                : 'bg-netflix-gray text-gray-300 hover:bg-gray-700'
            }`}
          >
            <LuLayoutGrid size={16} />
            Grid
          </button>
          <button
            onClick={() => setLibraryViewMode('list')}
            className={`flex items-center gap-2 rounded px-3 py-1.5 text-sm transition-colors ${
              libraryViewMode === 'list'
                ? 'bg-netflix-red text-white'
                : 'bg-netflix-gray text-gray-300 hover:bg-gray-700'
            }`}
          >
            <LuList size={16} />
            List
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-netflix-red border-t-transparent" />
        </div>
      ) : media?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <LuFilm size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">No media found in this library.</p>
        </div>
      ) : libraryViewMode === 'grid' ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
          {media?.map((item) => (
            <MediaCard
              key={item.id}
              id={item.id}
              title={item.title}
              posterPath={item.poster_path}
              type={item.media_type === 'episode' ? 'series' : 'movie'}
              seriesId={item.series_id}
              year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
              rating={item.rating}
            />
          ))}
        </div>
      ) : (
        <div className="space-y-2">
          {media?.map((item) => (
            <Link
              key={item.id}
              to={
                item.media_type === 'episode'
                  ? `/series/${item.series_id || item.id}`
                  : `/movies/${item.id}`
              }
              className="flex items-center gap-4 rounded-lg bg-netflix-dark p-3 transition-colors hover:bg-netflix-gray"
            >
              <div className="h-16 w-12 flex-shrink-0 overflow-hidden rounded bg-netflix-gray">
                {item.poster_path ? (
                  <img
                    src={item.poster_path}
                    alt={item.title}
                    className="h-full w-full object-cover"
                  />
                ) : (
                  <div className="flex h-full w-full items-center justify-center">
                    <LuFilm size={24} className="text-gray-600" />
                  </div>
                )}
              </div>
              <div className="min-w-0 flex-1">
                <h3 className="truncate font-medium text-white">{item.title}</h3>
                <p className="truncate text-sm text-gray-400">{item.genres.join(', ')}</p>
              </div>
              <LuChevronRight size={20} className="text-gray-500" />
            </Link>
          ))}
        </div>
      )}
    </div>
  )
})
