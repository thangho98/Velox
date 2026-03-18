import { useState } from 'react'
import { useSearchParams } from 'react-router'
import { useSearch } from '@/hooks/stores/useMedia'
import { MediaCard } from '@/components/MediaCard'
import { LuSearch, LuFilm, LuTv } from 'react-icons/lu'

export function SearchPage() {
  const [searchParams] = useSearchParams()
  const query = searchParams.get('q') || ''
  const [typeFilter, setTypeFilter] = useState(searchParams.get('type') || '')

  const { data: searchResult, isLoading } = useSearch(query, 50)

  const movies = searchResult?.movies ?? []
  const series = searchResult?.series ?? []
  const totalCount =
    (typeFilter === 'series' ? 0 : movies.length) + (typeFilter === 'movie' ? 0 : series.length)

  return (
    <div className="space-y-6">
      {/* Header + Type Filter */}
      <div className="space-y-4">
        {query && (
          <p className="text-gray-400">
            Found {totalCount} {totalCount === 1 ? 'result' : 'results'} for &quot;
            <span className="text-white">{query}</span>&quot;
          </p>
        )}

        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => setTypeFilter('')}
            className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm transition-all ${
              !typeFilter
                ? 'bg-[#e50914] text-white'
                : 'bg-[#1a1a1a] text-gray-300 hover:text-white'
            }`}
          >
            All
          </button>
          <button
            onClick={() => setTypeFilter('movie')}
            className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm transition-all ${
              typeFilter === 'movie'
                ? 'bg-[#e50914] text-white'
                : 'bg-[#1a1a1a] text-gray-300 hover:text-white'
            }`}
          >
            <LuFilm size={16} />
            Movies
          </button>
          <button
            onClick={() => setTypeFilter('series')}
            className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm transition-all ${
              typeFilter === 'series'
                ? 'bg-[#e50914] text-white'
                : 'bg-[#1a1a1a] text-gray-300 hover:text-white'
            }`}
          >
            <LuTv size={16} />
            Series
          </button>
        </div>
      </div>

      {/* Results */}
      {!query ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-[#1a1a1a]">
          <LuSearch size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">Use the search bar above to find movies and series</p>
        </div>
      ) : isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#e50914] border-t-transparent" />
        </div>
      ) : totalCount === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-[#1a1a1a]">
          <LuSearch size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">No results for &quot;{query}&quot;</p>
        </div>
      ) : (
        <div className="space-y-8">
          {/* Movies */}
          {(!typeFilter || typeFilter === 'movie') && movies.length > 0 && (
            <section>
              <h2 className="mb-4 text-lg font-semibold text-white flex items-center gap-2">
                <LuFilm size={20} />
                Movies ({movies.length})
              </h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {movies.map((item) => (
                  <MediaCard
                    key={`media-${item.id}`}
                    id={item.id}
                    title={item.title}
                    posterPath={item.poster_path}
                    type="movie"
                    year={item.release_date ? new Date(item.release_date).getFullYear() : undefined}
                    rating={item.rating}
                  />
                ))}
              </div>
            </section>
          )}

          {/* Series */}
          {(!typeFilter || typeFilter === 'series') && series.length > 0 && (
            <section>
              <h2 className="mb-4 text-lg font-semibold text-white flex items-center gap-2">
                <LuTv size={20} />
                Series ({series.length})
              </h2>
              <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6">
                {series.map((item) => (
                  <MediaCard
                    key={`series-${item.id}`}
                    id={item.id}
                    title={item.title}
                    posterPath={item.poster_path}
                    type="series"
                    seriesId={item.id}
                    year={
                      item.first_air_date ? new Date(item.first_air_date).getFullYear() : undefined
                    }
                  />
                ))}
              </div>
            </section>
          )}
        </div>
      )}
    </div>
  )
}
