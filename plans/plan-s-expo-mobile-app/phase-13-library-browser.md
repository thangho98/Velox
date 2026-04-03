# Phase 13: Library Browser + Media Grid
Status: ⬜ Pending
Dependencies: Phase 12

## Objective
Library tab với library selector, media grid, basic filter/sort, search.

## Implementation Steps

### 1. Library Selector
- [ ] Tạo `mobile/src/components/LibrarySelector.tsx`:
  - `useLibraries()` from `@velox/shared/hooks/media`
  - Horizontal ScrollView of pill/chip buttons
  - "All" chip + one per library (name from API)
  - Selected state: highlighted chip (red background)
  - Pass selected `library_id` to parent

### 2. Media Grid
- [ ] `mobile/app/(tabs)/library.tsx`:
  ```typescript
  import { useLibraries } from '@velox/shared/hooks/media'
  // Use media list + series list queries from shared hooks

  export default function LibraryScreen() {
    const [libraryId, setLibraryId] = useState<number | undefined>()
    const [sort, setSort] = useState('date_added')
    // ... query with libraryId filter

    return (
      <SafeAreaView>
        <LibrarySelector selected={libraryId} onSelect={setLibraryId} />
        <SortHeader sort={sort} onSortChange={setSort} />
        <FlatList
          data={media}
          numColumns={3}
          renderItem={({ item }) => <MediaCard media={item} />}
          onEndReached={loadMore}  // infinite scroll
          onEndReachedThreshold={0.5}
          ListFooterComponent={isLoading ? <ActivityIndicator /> : null}
        />
      </SafeAreaView>
    )
  }
  ```
  - FlatList with `numColumns={3}` (phone) / `numColumns={5}` (tablet)
  - Infinite scroll: pagination with `limit=30` + `offset`
  - Loading skeleton placeholders (first load)
  - ActivityIndicator footer (loading more)

### 3. Sort/Filter
- [ ] Tạo `mobile/src/components/SortHeader.tsx`:
  - Sort by: Title, Date Added, Year, Rating
  - Genre filter (useGenres from shared → dropdown/bottom sheet)
  - Simple horizontal row of filter chips / dropdown
  - Year range filter (optional, can be Phase 18 polish)

### 4. Search Tab
- [ ] `mobile/app/(tabs)/search.tsx`:
  - Search text input with debounce (300ms)
  - Uses **unified `useSearch(query)`** from `@velox/shared/hooks/media` (in `useGenres.ts`)
    - Endpoint: `GET /api/search?q=...&limit=20`
    - Returns mixed results (movies + series in one response)
  - Results: FlatList with MediaCard for each result
  - Empty state (no query): "Search for movies and shows"
  - Empty state (no results): "No results for '{query}'"
  - Recent searches list (MMKV persist, optional)

## Files to Create
- `mobile/src/components/LibrarySelector.tsx`
- `mobile/src/components/SortHeader.tsx`
- `mobile/app/(tabs)/library.tsx`
- `mobile/app/(tabs)/search.tsx`

## Test Criteria
- [ ] Library chips switch between libraries
- [ ] Grid displays poster cards (3 columns on phone)
- [ ] Infinite scroll loads more items
- [ ] Sort changes order (title, date, year, rating)
- [ ] Search finds movies + series with debounced input
- [ ] Empty states display correctly

---
Next Phase: [phase-14-media-detail.md](phase-14-media-detail.md)
