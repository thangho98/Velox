import { useCallback } from 'react'
import { useSearchParams } from 'react-router'

export type SortOption = 'newest' | 'oldest' | 'rating' | 'title'

export interface FilterState {
  genre: string
  year: string
  sort: SortOption
}

/**
 * Hook for managing filter state with URL synchronization.
 * Persists genre/year/sort in URL query params for bookmarking.
 * Search is handled separately by Navbar global search → SearchPage.
 */
export function useFilterParams() {
  const [searchParams, setSearchParams] = useSearchParams()

  const filters: FilterState = {
    genre: searchParams.get('genre') ?? '',
    year: searchParams.get('year') ?? '',
    sort: (searchParams.get('sort') as SortOption) ?? 'title',
  }

  const hasActiveFilters = filters.genre !== '' || filters.year !== ''

  const updateParam = useCallback(
    (key: keyof FilterState, value: string) => {
      const newParams = new URLSearchParams(searchParams)
      if (value && value !== '') {
        newParams.set(key, value)
      } else {
        newParams.delete(key)
      }
      newParams.delete('offset')
      setSearchParams(newParams, { replace: true })
    },
    [searchParams, setSearchParams],
  )

  const setGenre = useCallback((value: string) => updateParam('genre', value), [updateParam])
  const setYear = useCallback((value: string) => updateParam('year', value), [updateParam])
  const setSort = useCallback((value: SortOption) => updateParam('sort', value), [updateParam])

  const clearFilters = useCallback(() => {
    const newParams = new URLSearchParams(searchParams)
    newParams.delete('genre')
    newParams.delete('year')
    setSearchParams(newParams, { replace: true })
  }, [searchParams, setSearchParams])

  return { filters, setGenre, setYear, setSort, clearFilters, hasActiveFilters }
}
