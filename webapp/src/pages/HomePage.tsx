import { Link } from 'react-router'
import { useLibraries, useRecentlyWatched, useMediaList } from '@/hooks/stores/useMedia'
import { MediaRow } from '@/components/MediaRow'
import { useAuthStore } from '@/stores/auth'
import { LuPlay, LuFilm, LuLibrary } from 'react-icons/lu'

export function HomePage() {
  const { user } = useAuthStore()
  const { data: libraries, isLoading: libsLoading } = useLibraries()
  const { data: recentlyWatched, isLoading: recentLoading } = useRecentlyWatched({ limit: 20 })
  const { data: recentMovies, isLoading: moviesLoading } = useMediaList({
    type: 'movie',
    limit: 20,
  })
  // TODO: replace with GET /api/series once backend endpoint is available.
  // Currently fetches episode records as a proxy for series content.
  const { data: recentSeries, isLoading: seriesLoading } = useMediaList({
    type: 'episode',
    limit: 20,
  })

  const hasLibraries = !libsLoading && libraries && libraries.length > 0
  const hasRecentlyWatched = recentlyWatched && recentlyWatched.length > 0

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
      {(recentLoading || hasRecentlyWatched) && (
        <MediaRow
          title="Continue Watching"
          seeAllLink="/recently-watched"
          items={recentlyWatched}
          isLoading={recentLoading}
          showProgress
        />
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
                  className="text-sm text-netflix-light-gray hover:text-white transition-colors"
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
