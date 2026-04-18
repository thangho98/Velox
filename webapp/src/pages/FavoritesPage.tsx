import { useInfiniteFavorites, useFavoritesAlphabet } from '@/hooks/stores/useMedia'
import { useIntersectionObserver } from '@/hooks/useIntersectionObserver'
import { useEffect, useState } from 'react'
import { MediaCard } from '@/components/MediaCard'
import { LuHeart } from 'react-icons/lu'
import { AlphaIndex } from '@/components/AlphaIndex'

export default function FavoritesPage() {
  const [startChar, setStartChar] = useState<string>('')

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteFavorites({
    limit: 100,
    start_char: startChar || undefined,
  })

  const { data: alphabetCounts } = useFavoritesAlphabet()

  const favorites = data?.pages.flat() || []

  const { targetRef, isIntersecting } = useIntersectionObserver({
    rootMargin: '200px',
  })

  useEffect(() => {
    if (isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage()
    }
  }, [isIntersecting, hasNextPage, isFetchingNextPage, fetchNextPage])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Favorites</h1>
          <p className="text-gray-400">
            {favorites?.length || 0} {favorites?.length === 1 ? 'item' : 'items'} in your favorites
          </p>
        </div>
      </div>

      <div className="flex gap-6">
        <div className="flex-1">
          {/* Favorites Grid */}
          {isLoading ? (
            <div className="flex h-64 items-center justify-center">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-pink-500 border-t-transparent" />
            </div>
          ) : favorites?.length === 0 ? (
            <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-netflix-dark">
              <LuHeart size={48} className="mb-4 text-gray-600" />
              <p className="text-gray-400">No favorites yet</p>
              <p className="text-sm text-gray-500">
                Click the heart icon on any movie or series to add it here
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
              {favorites?.map((item, index) => {
                const isTarget = index === Math.max(0, favorites.length - 30)
                return (
                  <div key={item.media_id} ref={isTarget ? targetRef : undefined}>
                    <MediaCard
                      id={item.media_id}
                      title={item.media_title || 'Unknown'}
                      poster={item.poster}
                      progress={{
                        position: item.position,
                        duration: item.media_duration || 1,
                        completed: item.completed,
                        is_favorite: item.is_favorite,
                      }}
                    />
                  </div>
                )
              })}

              {hasNextPage && (
                <div className="col-span-full flex h-24 items-center justify-center">
                  <div className="h-8 w-8 animate-spin rounded-full border-2 border-pink-500 border-t-transparent" />
                </div>
              )}
            </div>
          )}
        </div>

        {/* Alpha Index Sidebar */}
        {!isLoading && (favorites.length > 0 || alphabetCounts?.length) && (
          <div className="w-10 flex-shrink-0">
            <div className="sticky top-24">
              <AlphaIndex counts={alphabetCounts} activeChar={startChar} onChange={setStartChar} />
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
