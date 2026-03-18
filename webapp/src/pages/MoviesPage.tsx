import { useMediaList, useGenres } from '@/hooks/stores/useMedia'
import { useFilterParams } from '@/hooks/useFilterParams'
import { MediaCard } from '@/components/MediaCard'
import { FilterBar } from '@/components/FilterBar'
import { AlphaIndex, useAlphaScroll } from '@/components/AlphaIndex'
import { LuFilm } from 'react-icons/lu'

export function MoviesPage() {
  const { filters, setGenre, setYear, setSort, clearFilters, hasActiveFilters } = useFilterParams()

  const { data: movies, isLoading } = useMediaList({
    type: 'movie',
    genre: filters.genre || undefined,
    year: filters.year || undefined,
    sort: filters.sort,
    limit: 500,
  })

  const { data: genreList } = useGenres('movie')
  const genres = genreList?.map((g) => g.name) ?? []

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: currentYear - 1900 + 1 }, (_, i) => String(currentYear - i))

  const { activeLetters, currentLetter, scrollToLetter, getLetterForTitle } = useAlphaScroll(movies)

  const showAlphaIndex = filters.sort === 'title' && (movies?.length ?? 0) > 0

  // Track which letters have been seen to mark only the first item per letter
  const seenLetters = new Set<string>()

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-3xl font-bold text-white">Movies</h1>
          <p className="text-gray-400">
            {movies?.length || 0} {movies?.length === 1 ? 'movie' : 'movies'}
          </p>
        </div>
      </div>

      {/* Filter Bar */}
      <FilterBar
        genre={filters.genre}
        year={filters.year}
        sort={filters.sort}
        genres={genres}
        years={years}
        onGenreChange={setGenre}
        onYearChange={setYear}
        onSortChange={setSort}
        onClearFilters={clearFilters}
        hasActiveFilters={hasActiveFilters}
      />

      {showAlphaIndex && (
        <AlphaIndex
          activeLetters={activeLetters}
          currentLetter={currentLetter}
          onSelect={scrollToLetter}
        />
      )}

      {/* Movies Grid */}
      {isLoading ? (
        <div className="flex h-64 items-center justify-center">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#e50914] border-t-transparent" />
        </div>
      ) : movies?.length === 0 ? (
        <div className="flex h-64 flex-col items-center justify-center rounded-lg bg-[#1a1a1a]">
          <LuFilm size={48} className="mb-4 text-gray-600" />
          <p className="text-gray-400">
            {hasActiveFilters ? 'No movies match your filters.' : 'No movies in your library yet.'}
          </p>
        </div>
      ) : (
        <div
          className={`grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 ${showAlphaIndex ? 'pr-8' : ''}`}
        >
          {movies?.map((movie) => {
            const letter = getLetterForTitle(movie.sort_title || movie.title)
            const isFirstOfLetter = showAlphaIndex && !seenLetters.has(letter)
            if (isFirstOfLetter) seenLetters.add(letter)
            return (
              <div key={movie.id} {...(isFirstOfLetter ? { 'data-alpha-letter': letter } : {})}>
                <MediaCard
                  id={movie.id}
                  title={movie.title}
                  posterPath={movie.poster_path}
                  type={movie.media_type === 'episode' ? 'series' : 'movie'}
                  year={movie.release_date ? new Date(movie.release_date).getFullYear() : undefined}
                  rating={movie.rating}
                />
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
