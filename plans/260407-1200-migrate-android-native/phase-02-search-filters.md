# Phase 02: Search Type Filter + Layout Fix
Status: ⬜ Pending
Dependencies: None

## Objective
Fix SearchScreen Android Native: thêm type filter chips, đổi layout từ LazyRow → Grid.

## Features to Implement

### 1. Type Filter Chips
```
- Chips: All, Movies, Series
- Active chip có background color (NetflixRed)
- Tap chip → filter results theo type
```

### 2. Results Count Header
```
- Text: "Found X results for 'query'"
- Hiện khi có query và có results
```

### 3. Layout Fix
```
- Đổi từ LazyRow (horizontal scroll)
- Sang Grid layout (2-3 columns)
- Giống MoviesScreen grid
```

## Files to Modify

### Modify Files:
- `presentation/viewmodel/SearchViewModel.kt` - Add typeFilter state, filter logic
- `presentation/ui/screens/search/SearchScreen.kt` - Add chips, fix layout

## Implementation Steps

1. **Update SearchViewModel**
   - [ ] Add `typeFilter: String` state ("", "movie", "series")
   - [ ] Add `setTypeFilter()` method
   - [ ] Filter results client-side hoặc call API với type param

2. **Update SearchScreen UI**
   - [ ] Add Row of FilterChips (All, Movies, Series)
   - [ ] Add results count Text above results
   - [ ] Change LazyRow → LazyVerticalGrid (2-3 columns)
   - [ ] Update cards to use MediaCard instead of SearchMovieCard/SearchSeriesCard

## Test Criteria
- [ ] Tap "Movies" chip → only movies shown
- [ ] Tap "Series" chip → only series shown
- [ ] Tap "All" chip → all results shown
- [ ] Results count updates correctly
- [ ] Grid layout displays properly

## Reference
Mobile implementation: `mobile/src/screens/SearchScreen.tsx` lines 250-320
