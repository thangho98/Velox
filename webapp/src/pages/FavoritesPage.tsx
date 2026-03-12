import { useFavorites } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'

export function FavoritesPage() {
  const { data: favorites, isLoading } = useFavorites({ limit: 100 })

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-white">Favorites</h1>
        <p className="text-gray-400">
          {favorites?.length || 0} {favorites?.length === 1 ? 'item' : 'items'} in your favorites
        </p>
      </div>

      {/* Favorites Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-pink-500 border-t-transparent" />
        </div>
      ) : favorites?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
          <svg
            className="mb-4 h-12 w-12 text-gray-600"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z"
            />
          </svg>
          <p className="text-gray-400">No favorites yet</p>
          <p className="text-sm text-gray-500">
            Click the heart icon on any movie or series to add it here
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
          {favorites?.map((item) => (
            <MediaCard
              key={item.media_id}
              id={item.media_id}
              title={item.media_title || 'Unknown'}
              posterPath={item.media_poster}
              progress={{
                position: item.position,
                duration: item.media_duration || 1,
                completed: item.completed,
                is_favorite: item.is_favorite,
              }}
            />
          ))}
        </div>
      )}
    </div>
  )
}
