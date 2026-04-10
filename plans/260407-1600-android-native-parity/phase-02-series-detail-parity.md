# Phase 02: SeriesDetailScreen Parity
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Add all missing features to SeriesDetailScreen to match mobile app exactly.

## Tasks

### 1. Add Status Badge
- [ ] Show status badge (e.g., "Returning Series")
- [ ] Purple background styling

### 2. Add Network Info
- [ ] Show TV network name with icon
- [ ] Gray text styling

### 3. Add Edit & Lock Badges
- [ ] Edit badge with pencil icon
- [ ] Lock badge (yellow) when metadata_locked
- [ ] Edit opens metadata editor

### 4. Add "Read More" Overview Toggle
- [ ] Show 3 lines initially
- [ ] "Read more" text in red
- [ ] Expand to full on tap

### 5. Add Episode Duration
- [ ] Show duration in "XXm" format
- [ ] Display in episode info row

### 6. Add EpisodeEditDialog
- [ ] Create EpisodeEditDialog composable
- [ ] Edit title, overview fields
- [ ] Save/Cancel buttons

### 7. Add Episode Metadata Editing
- [ ] Open dialog on episode menu press
- [ ] Call API to update episode metadata
- [ ] Refresh episode list after save

## Files to Modify
- `SeriesDetailScreen.kt` - Add missing UI components
- `SeriesDetailViewModel.kt` - Add edit episode method
- `EpisodeEditDialog.kt` - Create new dialog component

## Test Criteria
- [ ] Status badge shows correctly
- [ ] Network info displays
- [ ] Edit/Lock badges work
- [ ] Overview "Read more" toggles
- [ ] Episode duration shows
- [ ] EpisodeEditDialog opens and saves

## Reference
Mobile implementation: `mobile/src/screens/SeriesDetailScreen.tsx`
