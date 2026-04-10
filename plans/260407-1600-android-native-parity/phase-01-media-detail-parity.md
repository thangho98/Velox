# Phase 01: MediaDetailScreen Parity
Status: ⬜ Pending
Dependencies: None

## Objective
Add all missing features to MediaDetailScreen to match mobile app exactly.

## Tasks

### 1. Add Poster
- [ ] Add poster image (centered) with shadow
- [ ] Show poster placeholder (first char of title) if no poster

### 2. Add Tech Specs Row
- [ ] Show video codec and resolution (e.g., "1080p HEVC")
- [ ] Show audio codec
- [ ] Show container format
- [ ] Show file size (formatted: GB/MB)

### 3. Add "Ends at" Time
- [ ] Calculate remaining time from progress
- [ ] Display "Ends at HH:MM" format

### 4. Add Watched/Check Button
- [ ] Add check icon button next to favorite
- [ ] Toggle watched state

### 5. Add Inline Subtitle Selector
- [ ] Show subtitle label and picker
- [ ] Open subtitle selection modal
- [ ] Support Off option

### 6. Add Tablet Side-by-Side Layout
- [ ] Detect tablet (screen width >= 600dp)
- [ ] Show poster on left, info on right
- [ ] Side-by-side row layout

## Files to Modify
- `MediaDetailScreen.kt` - Add missing UI components
- `MediaDetailUiState.kt` - May need additional state fields

## Test Criteria
- [ ] Poster displays correctly on phone and tablet
- [ ] Tech specs show for media with file info
- [ ] "Ends at" time calculates correctly
- [ ] Watched button toggles state
- [ ] Subtitle selector opens and selects subtitles
- [ ] Tablet shows side-by-side layout

## Reference
Mobile implementation: `mobile/src/screens/MediaDetailScreen.tsx`
