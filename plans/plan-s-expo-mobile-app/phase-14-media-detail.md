# Phase 14: Media Detail + Series Detail
Status: ⬜ Pending
Dependencies: Phase 12

## Objective
Movie detail page + series/season/episode browser.

## Implementation Steps

### 1. Movie Detail Screen
- [ ] `mobile/app/media/[id].tsx`:
  ```typescript
  import { useMediaWithFiles, useMediaGenres, useMediaCredits } from '@velox/shared/hooks/media'
  import { useToggleFavorite, useProgress } from '@velox/shared/hooks/media'

  export default function MediaDetailScreen() {
    const { id } = useLocalSearchParams<{ id: string }>()
    const mediaId = Number(id)
    const { data: media } = useMediaWithFiles(mediaId)
    const { data: genres } = useMediaGenres(mediaId)
    const { data: progress } = useProgress(mediaId)
    const toggleFav = useToggleFavorite()

    return (
      <ScrollView>
        {/* Backdrop image (full width, 16:9) */}
        {/* Back button overlay (top-left) */}
        {/* Title + year + runtime + rating */}
        {/* Play button (large, prominent, red) */}
        {/* Resume info if has progress: "Resume from 1:23:45" */}
        {/* Overview/description text */}
        {/* Genre chips (horizontal scroll) */}
        {/* Favorite heart button */}
        {/* Cast/crew horizontal scroll (avatar + name) */}
        {/* File info: resolution, codecs, size */}
      </ScrollView>
    )
  }
  ```
  - Play button → `router.push(\`/watch/${mediaId}\`)`
  - Backdrop: `expo-image` with `getImageUrl(media.backdrop_path, 'w1280')`
  - Favorite toggle: heart icon, calls `useToggleFavorite`

### 2. Series Detail Screen
- [ ] `mobile/app/series/[id].tsx`:
  ```typescript
  import { useSeriesDetail, useSeasons, useEpisodes } from '@velox/shared/hooks/media'

  export default function SeriesDetailScreen() {
    const { id } = useLocalSearchParams<{ id: string }>()
    const seriesId = Number(id)
    const { data: series } = useSeriesDetail(seriesId)
    const { data: seasons } = useSeasons(seriesId)
    const [selectedSeason, setSelectedSeason] = useState<number>(0)
    const { data: episodes } = useEpisodes(seriesId, selectedSeason)

    return (
      <ScrollView>
        {/* Backdrop + series info (same layout as movie) */}
        {/* Season selector: horizontal scroll of "Season 1", "Season 2", ... */}
        {/* Episode list */}
        <FlatList
          data={episodes}
          renderItem={({ item }) => <EpisodeCard episode={item} />}
          scrollEnabled={false}  // nested in ScrollView
        />
      </ScrollView>
    )
  }
  ```
  - Season selector: pill buttons or dropdown
  - Auto-select first season on load

### 3. Episode Card
- [ ] Tạo `mobile/src/components/EpisodeCard.tsx`:
  - Thumbnail/still image (16:9, left side)
  - Episode number + title (right side)
  - Air date + runtime
  - Progress bar (if partially watched)
  - Brief description (2 lines, ellipsis)
  - Pressable → `router.push(\`/watch/${episode.media_id}\`)`

## Files to Create
- `mobile/app/media/[id].tsx` — movie detail
- `mobile/app/series/[id].tsx` — series detail
- `mobile/src/components/EpisodeCard.tsx`

## Test Criteria
- [ ] Movie detail shows backdrop, metadata, genres, cast
- [ ] Play button navigates to watch screen
- [ ] Resume text shows for partially watched items
- [ ] Favorite toggle works (heart fills/unfills)
- [ ] Series shows season tabs + episode list
- [ ] Episode card shows thumbnail + info
- [ ] Tap episode → navigate to watch

---
Next Phase: [phase-15-video-player.md](phase-15-video-player.md)
