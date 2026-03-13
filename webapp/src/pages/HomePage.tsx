import { Link } from 'react-router'
import {
  useLibraries,
  useContinueWatching,
  useNextUp,
  useMediaList,
  useSeriesList,
} from '@/hooks/stores/useMedia'
import { MediaRow } from '@/components/MediaRow'
import { ContinueWatchingCard } from '@/components/ContinueWatchingCard'
import { NextUpCard } from '@/components/NextUpCard'
import { useAuthStore } from '@/stores/auth'
import { LuPlay, LuFilm, LuLibrary } from 'react-icons/lu'

export function HomePage() {
  const { user } = useAuthStore()
  const { data: libraries, isLoading: libsLoading } = useLibraries()
  const { data: continueWatching, isLoading: continueLoading } = useContinueWatching({ limit: 20 })
  const { data: nextUp, isLoading: nextUpLoading } = useNextUp({ limit: 20 })
  const { data: recentMovies, isLoading: moviesLoading } = useMediaList({
    type: 'movie',
    limit: 20,
  })
  // Use real series data from GET /api/series
  const { data: rawSeries, isLoading: seriesLoading } = useSeriesList({ limit: 20 })

  // Map Series[] to MediaRow-compatible shape
  const recentSeries = rawSeries?.map((s) => ({
    id: s.id,
    title: s.title,
    sort_title: s.sort_title,
    poster_path: s.poster_path,
    media_type: 'episode' as const, // MediaRow uses this for type detection
    type: 'series' as const,
    genres: [] as string[],
    release_date: s.first_air_date,
    series_id: s.id, // for MediaRow → MediaCard routing
  }))

  const hasLibraries = !libsLoading && libraries && libraries.length > 0
  const hasContinueWatching = continueWatching && continueWatching.length > 0
  const hasNextUp = nextUp && nextUp.length > 0

  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <section className="relative -mt-4 -mx-4 lg:-mx-8">
        <div className="relative h-[45vh] bg-gradient-to-b from-netflix-dark to-netflix-black lg:h-[55vh]">
          <div className="absolute inset-0 bg-gradient-to-r from-netflix-black via-netflix-black/60 to-transparent" />
          <div className="absolute bottom-0 left-0 p-8 lg:p-16">
            <h1 className="mb-3 max-w-2xl text-3xl font-bold text-white lg:text-5xl">
              Welcome back{user?.display_name ? `, ${user.display_name}` : ''}
            </h1>
            <p className="mb-6 max-w-lg text-gray-300">Your personal media server.</p>
            <div className="flex gap-3">
              <Link
                to="/movies"
                className="flex items-center gap-2 rounded bg-netflix-red px-5 py-2.5 font-semibold text-white transition-colors hover:bg-netflix-red-hover"
              >
                <LuPlay size={16} />
                Movies
              </Link>
              <Link
                to="/series"
                className="flex items-center gap-2 rounded bg-white/10 px-5 py-2.5 font-semibold text-white backdrop-blur-sm transition-colors hover:bg-white/20"
              >
                Series
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* Continue Watching */}
      {(continueLoading || hasContinueWatching) && (
        <section className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white lg:text-xl">Continue Watching</h2>
          </div>
          {continueLoading ? (
            <div className="hide-scrollbar flex gap-3 overflow-x-auto pb-2">
              {[...Array(4)].map((_, i) => (
                <div
                  key={i}
                  className="w-48 shrink-0 animate-pulse overflow-hidden rounded-lg bg-netflix-gray"
                >
                  <div className="aspect-video bg-gray-700" />
                </div>
              ))}
            </div>
          ) : (
            <div className="hide-scrollbar flex gap-3 overflow-x-auto pb-2">
              {continueWatching?.map((item) => (
                <div key={item.media_id} className="w-48 shrink-0 lg:w-56">
                  <ContinueWatchingCard item={item} />
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Next Up */}
      {(nextUpLoading || hasNextUp) && (
        <section className="space-y-3">
          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold text-white lg:text-xl">Next Up</h2>
          </div>
          {nextUpLoading ? (
            <div className="hide-scrollbar flex gap-3 overflow-x-auto pb-2">
              {[...Array(4)].map((_, i) => (
                <div
                  key={i}
                  className="w-48 shrink-0 animate-pulse overflow-hidden rounded-lg bg-netflix-gray"
                >
                  <div className="aspect-video bg-gray-700" />
                </div>
              ))}
            </div>
          ) : (
            <div className="hide-scrollbar flex gap-3 overflow-x-auto pb-2">
              {nextUp?.map((item) => (
                <div key={item.media_id} className="w-48 shrink-0 lg:w-56">
                  <NextUpCard item={item} />
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Recently Added Movies */}
      {hasLibraries && (
        <MediaRow
          title="Movies"
          seeAllLink="/movies"
          items={recentMovies}
          isLoading={moviesLoading}
        />
      )}

      {/* Recently Added Series */}
      {hasLibraries && (
        <MediaRow
          title="Series"
          seeAllLink="/series"
          items={recentSeries}
          isLoading={seriesLoading}
        />
      )}

      {/* Empty state — no libraries */}
      {!libsLoading && libraries?.length === 0 && (
        <div className="rounded-xl bg-netflix-dark p-12 text-center">
          <LuLibrary size={56} className="mx-auto mb-4 text-gray-600" />
          <p className="mb-2 text-lg font-medium text-gray-300">No libraries yet</p>
          {user?.is_admin ? (
            <Link
              to="/admin/libraries"
              className="mt-3 inline-block rounded bg-netflix-red px-5 py-2 text-sm font-medium text-white transition-colors hover:bg-netflix-red-hover"
            >
              Add your first library
            </Link>
          ) : (
            <p className="text-sm text-gray-500">Contact your administrator to add libraries</p>
          )}
        </div>
      )}

      {/* Libraries list (only shown when there are libraries and no content rows loaded yet) */}
      {hasLibraries &&
        !moviesLoading &&
        !recentMovies?.length &&
        !seriesLoading &&
        !recentSeries?.length && (
          <section>
            <div className="mb-4 flex items-center justify-between">
              <h2 className="text-lg font-semibold text-white">Your Libraries</h2>
              {user?.is_admin && (
                <Link
                  to="/admin/libraries"
                  className="text-sm text-netflix-light-gray transition-colors hover:text-white"
                >
                  Manage →
                </Link>
              )}
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
              {libraries.map((lib) => (
                <Link
                  key={lib.id}
                  to={`/libraries?library=${lib.id}`}
                  className="group relative overflow-hidden rounded-lg bg-netflix-dark p-6 transition-all hover:bg-netflix-gray"
                >
                  <div className="mb-3 flex h-10 w-10 items-center justify-center rounded-lg bg-netflix-red/20">
                    <LuFilm size={20} className="text-netflix-red" />
                  </div>
                  <h3 className="mb-1 font-semibold text-white transition-colors group-hover:text-netflix-red">
                    {lib.name}
                  </h3>
                  <p className="text-sm capitalize text-gray-400">{lib.type}</p>
                </Link>
              ))}
            </div>
          </section>
        )}
    </div>
  )
}
