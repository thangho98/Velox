# Phase 01: Movies/Series Filters + A-Z Index
Status: ⬜ Pending
Dependencies: None

## Objective
Thêm Genre filter, Year filter, A-Z sidebar index, QuickActionsMenu vào MoviesScreen và SeriesScreen của Android Native.

## Features to Implement

### 1. FilterBottomSheet Component (NEW)
```
UI: Bottom sheet với danh sách options để chọn
- Genre filter (lấy từ useGenres hook)
- Year filter (generation từ current year về 1970)
```

### 2. Genre Filter
```
- Gọi API get genres theo type (movie/series)
- Hiện BottomSheet khi tap Genre chip
- Selected genre hiển thị trên chip
- Gọi lại API movies với genre param
```

### 3. Year Filter
```
- Generate list years từ currentYear → 1970
- Hiện BottomSheet khi tap Year chip
- Selected year hiển thị trên chip
- Gọi lại API movies với year param
```

### 4. A-Z Sidebar Index
```
- Chỉ hiện khi sort = "az" hoặc "title"
- Vertical alphabet list ở right side
- Tap letter → scroll FlatList/LazyColumn to position
- Highlight active letter khi scroll
```

### 5. QuickActionsMenu
```
- Long press on movie/series card → show menu
- Options: Play, View Details
- Menu hiện ở center màn hình
```

## Files to Create/Modify

### New Files:
- `presentation/ui/components/FilterBottomSheet.kt` - Reusable bottom sheet
- `presentation/ui/components/QuickActionsMenu.kt` - Action menu modal

### Modify Files:
- `presentation/viewmodel/MoviesViewModel.kt` - Add genre/year state, filter methods
- `presentation/viewmodel/SeriesViewModel.kt` - Add genre/year state, filter methods
- `presentation/ui/screens/movies/MoviesScreen.kt` - Add filter UI, A-Z sidebar
- `presentation/ui/screens/series/SeriesScreen.kt` - Add filter UI, A-Z sidebar
- `domain/model/MediaModels.kt` - Add year, genre fields if needed

## Implementation Steps

1. **Create FilterBottomSheet component**
   - [ ] Generic bottom sheet nhận list of options
   - [ ] Single select behavior
   - [ ] Done/Clear buttons

2. **Update MoviesViewModel**
   - [ ] Add `selectedGenre`, `selectedYear` state
   - [ ] Add `genres` list from API
   - [ ] Add `setGenre()`, `setYear()`, `clearFilters()` methods
   - [ ] Filter movies by genre/year in repository call

3. **Update MoviesScreen UI**
   - [ ] Connect Genre chip → show FilterBottomSheet
   - [ ] Connect Year chip → show FilterBottomSheet
   - [ ] Add A-Z sidebar (only when sort=az)
   - [ ] Add LongPress → QuickActionsMenu

4. **Apply same to SeriesScreen**
   - [ ] Same pattern as MoviesScreen

5. **Create QuickActionsMenu**
   - [ ] AlertDialog hoặc ModalBottomSheet
   - [ ] Actions: Play, View Details
   - [ ] Callback khi chọn action

## Test Criteria
- [ ] Tap Genre → show list → select → chip shows selected + filter works
- [ ] Tap Year → show list → select → chip shows selected + filter works
- [ ] Sort A-Z → A-Z sidebar appears → tap letter → scroll to position
- [ ] Long press card → menu appears → tap Play → navigates to Player
- [ ] Pull to refresh → movies reload

## Reference
Mobile implementation: `mobile/src/screens/MoviesScreen.tsx` lines 30-230
