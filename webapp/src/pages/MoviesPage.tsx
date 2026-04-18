import { useInfiniteMediaList, useGenres } from '@/hooks/stores/useMedia'
import { useFilterParams } from '@/hooks/useFilterParams'
import { useIntersectionObserver } from '@/hooks/useIntersectionObserver'
import { useEffect } from 'react'
import { MediaCard } from '@/components/MediaCard'
import { FilterBar } from '@/components/FilterBar'
import { AlphaIndex, useAlphaScroll } from '@/components/AlphaIndex'
import { useMediaAlphabet } from '@/hooks/stores/useMedia'
import { LuFilm } from 'react-icons/lu'

export default function MoviesPage() {
  const { filters, setGenre, setYear, setSort, setStartChar, clearFilters, hasActiveFilters } =
    useFilterParams()

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteMediaList({
    type: 'movie',
    genre: filters.genre || undefined,
    year: filters.year || undefined,
    sort: filters.sort,
    start_char: filters.start_char || undefined,
    limit: 100,
  })

  const movies = data?.pages.flat() || []

  const { targetRef, isIntersecting } = useIntersectionObserver({
    rootMargin: '200px',
  })

  useEffect(() => {
    if (isIntersecting && hasNextPage && !isFetchingNextPage) {
      fetchNextPage()
    }
  }, [isIntersecting, hasNextPage, isFetchingNextPage, fetchNextPage])

  const { data: genreList } = useGenres('movie')
  const genres = genreList?.map((g) => g.name) ?? []

  const currentYear = new Date().getFullYear()
  const years = Array.from({ length: currentYear - 1900 + 1 }, (_, i) => String(currentYear - i))

  const { currentLetter, getLetterForTitle } = useAlphaScroll(movies)

  // Fetch the active alphabet from the server
  const { data: alphabetCounts } = useMediaAlphabet({
    type: 'movie',
    genre: filters.genre || undefined,
    year: filters.year || undefined,
  })

  // The alphabet index is only useful when sorting by title
  const showAlphaIndex = filters.sort === 'title' && (alphabetCounts?.length ?? 0) > 0

  // Extract active letters from the counts API (ignoring ones with count=0)
  const activeLetters = new Set(
    alphabetCounts?.filter((a) => a.count > 0).map((a) => a.letter) || [],
  )

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
          currentLetter={filters.start_char || currentLetter}
          onSelect={setStartChar}
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
          {movies?.map((movie, index) => {
            const letter = getLetterForTitle(movie.sort_title || movie.title)
            const isFirstOfLetter = showAlphaIndex && !seenLetters.has(letter)
            if (isFirstOfLetter) seenLetters.add(letter)

            const isTarget = index === Math.max(0, movies.length - 30)
            return (
              <div
                key={movie.id}
                ref={isTarget ? targetRef : undefined}
                {...(isFirstOfLetter ? { 'data-alpha-letter': letter } : {})}
              >
                <MediaCard
                  id={movie.id}
                  title={movie.title}
                  poster={movie.poster}
                  type={movie.media_type === 'episode' ? 'series' : 'movie'}
                  year={movie.release_date ? new Date(movie.release_date).getFullYear() : undefined}
                  rating={movie.rating}
                  progress={
                    movie.position !== undefined
                      ? {
                          position: movie.position,
                          duration: movie.duration || 1,
                          completed: !!movie.completed,
                          is_favorite: false,
                        }
                      : undefined
                  }
                />
              </div>
            )
          })}

          {hasNextPage && (
            <div className="col-span-full flex h-24 items-center justify-center">
              <div className="h-8 w-8 animate-spin rounded-full border-2 border-[#e50914] border-t-transparent" />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
