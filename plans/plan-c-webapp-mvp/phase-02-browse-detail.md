# Phase 02: Browse & Detail Pages
Status: ⬜ Pending
Plan: C - Web App MVP
Dependencies: Phase 01

## Tasks

### 1. Media Card Component
- [ ] `src/components/MediaCard.tsx` - poster thumbnail + title + year + rating badge
- [ ] Hover: scale up, show play button overlay
- [ ] Watched badge (checkmark) if completed
- [ ] Progress bar at bottom if partially watched
- [ ] Lazy load images with placeholder/skeleton

### 2. Movies Page
- [ ] Route: `/movies`
- [ ] Grid layout of MediaCards (responsive: 2/3/4/5/6 columns)
- [ ] Sort dropdown: Title, Rating, Year, Date Added
- [ ] Genre filter chips (horizontal scroll)
- [ ] Infinite scroll or pagination
- [ ] Empty state: "No movies found. Add a library to get started."

### 3. Series Page
- [ ] Route: `/series`
- [ ] Same grid layout as Movies
- [ ] Show series poster, title, year range, episode count
- [ ] Sort + filter same as movies

### 4. Movie Detail Page
- [ ] Route: `/movie/:id`
- [ ] Hero section: backdrop image, poster, title, year, rating, runtime, genres
- [ ] Overview/synopsis
- [ ] Cast row (horizontal scroll): headshot, name, character
- [ ] Play button (big, prominent)
- [ ] Resume button nếu có progress ("Resume from 1:23:45")
- [ ] Media info: resolution, codec, file size, audio channels

### 5. Series Detail Page
- [ ] Route: `/series/:id`
- [ ] Hero section same as movie
- [ ] Season tabs/dropdown selector
- [ ] Episode list for selected season:
  - Episode number, title, still image, overview (truncated), duration
  - Watched badge, progress bar
  - Click → play episode

### 6. Cast/Person Display
- [ ] Horizontal scroll row of cast members
- [ ] Photo, name, character name
- [ ] Click → filter/search by person (future enhancement, placeholder for now)

### 7. Genre Page
- [ ] Route: `/genre/:id` hoặc `/movies?genre=action`
- [ ] Same grid layout, filtered by genre
- [ ] Genre name as page title

### 8. Search Page
- [ ] Route: `/search?q=keyword`
- [ ] Search input in navbar (debounced 300ms)
- [ ] Results: mixed movies + series, grouped or combined
- [ ] Show media type badge (Movie/Series)
- [ ] Empty state: "No results for 'keyword'"

## Files to Create
- `src/components/MediaCard.tsx`
- `src/components/CastRow.tsx`
- `src/components/GenreChips.tsx`
- `src/pages/MoviesPage.tsx`
- `src/pages/SeriesPage.tsx`
- `src/pages/MovieDetailPage.tsx`
- `src/pages/SeriesDetailPage.tsx`
- `src/pages/SearchPage.tsx`

---
Next: phase-03-video-player.md
