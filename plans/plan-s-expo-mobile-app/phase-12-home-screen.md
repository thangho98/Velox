# Phase 12: Home Screen (Continue Watching + Next Up)
Status: ⬜ Pending
Dependencies: Phase 11

## Objective
Home screen với Continue Watching, Next Up rows.

## Context — Available APIs
| Row | Hook | Endpoint | Data |
|-----|------|----------|------|
| Continue Watching | `useContinueWatching()` | GET /profile/continue-watching | Items with position + duration |
| Next Up | `useNextUp()` | GET /profile/next-up | Next episodes for in-progress series |

⚠️ **"Recently Added" CHƯA CÓ API riêng.** Có thể dùng media list sorted by date, nhưng đó là scope Plan M (search/filter). Home screen MVP chỉ cần Continue Watching + Next Up.

## Image URL Resolution (verified from code)
Backend image paths có 3 dạng:
1. **TMDb:** `/hashpath.jpg` → proxy qua `/api/images/tmdb/{size}/{hashpath.jpg}`
2. **Local uploaded:** `local://` prefix → `/api/images/local/media/{path}` hoặc `/api/images/local/series/{path}`
3. **Full URL:** `http://...` → pass through

Webapp dùng relative URL vì Vite proxy. Mobile cần absolute URL.

## Implementation Steps

### 1. Image URL helper
- [ ] Tạo `mobile/src/utils/image.ts`:
  ```typescript
  import { getPlatform } from '@velox/shared/platform'

  /**
   * Convert server-relative image paths to absolute URLs for mobile.
   * Webapp uses relative paths (proxied by Vite), mobile needs absolute.
   *
   * Matches logic in webapp/src/lib/image.ts but with absolute URLs.
   */
  export function mediaImage(
    path: string | null | undefined,
    size: string = 'w500',
  ): string | undefined {
    if (!path) return undefined
    const apiBase = getPlatform().getApiBaseUrl()  // e.g. "http://192.168.1.100:8098/api"

    // Local uploaded image — served by backend
    if (path.startsWith('local://')) {
      return `${apiBase}/images/local/media/${path.slice(8)}`
    }
    // Already a full URL
    if (path.startsWith('http')) return path
    // Already an API path
    if (path.startsWith('/api/')) {
      const serverBase = apiBase.replace(/\/api$/, '')
      return `${serverBase}${path}`
    }
    // TMDb path — proxy through backend
    const cleaned = path.startsWith('/') ? path.slice(1) : path
    return `${apiBase}/images/tmdb/${size}/${cleaned}`
  }

  export function seriesImage(
    path: string | null | undefined,
    size: string = 'w500',
  ): string | undefined {
    if (!path) return undefined
    const apiBase = getPlatform().getApiBaseUrl()

    if (path.startsWith('local://')) {
      return `${apiBase}/images/local/series/${path.slice(8)}`
    }
    if (path.startsWith('http')) return path
    if (path.startsWith('/api/')) {
      const serverBase = apiBase.replace(/\/api$/, '')
      return `${serverBase}${path}`
    }
    const cleaned = path.startsWith('/') ? path.slice(1) : path
    return `${apiBase}/images/tmdb/${size}/${cleaned}`
  }
  ```

### 2. MediaCard component
- [ ] Tạo `mobile/src/components/MediaCard.tsx`:
  - `expo-image` for poster (cached automatically by expo-image)
  - Title overlay at bottom (text-white, dark gradient)
  - Progress bar overlay if `item.position > 0`:
    ```typescript
    const percent = item.duration > 0 ? (item.position / item.duration) * 100 : 0
    ```
  - Pressable → navigate:
    - Movie: `router.push(\`/media/${item.media_id}\`)`
    - Episode: `router.push(\`/series/${item.series_id}\`)` (go to series detail)
  - Aspect ratio: 2:3 (poster)
  - Responsive width: `(screenWidth - padding) / 3` for 3 columns

### 3. MediaRow component
- [ ] Tạo `mobile/src/components/MediaRow.tsx`:
  - Section title text (bold, left-aligned)
  - Horizontal FlatList of MediaCards
  - Card width: ~130px for phone, ~160px for tablet
  - Gap between cards: 8px
  - Empty state: hide entire row (don't show title)
  - Loading state: 3 skeleton cards (placeholder shimmer)

### 4. Home Screen
- [ ] `mobile/app/(tabs)/index.tsx`:
  ```typescript
  import { useContinueWatching, useNextUp } from '@velox/shared/hooks/media'
  import { useQueryClient } from '@tanstack/react-query'

  export default function HomeScreen() {
    const { data: continueWatching, isLoading: cwLoading } = useContinueWatching({ limit: 20 })
    const { data: nextUp, isLoading: nuLoading } = useNextUp({ limit: 20 })
    const queryClient = useQueryClient()

    const onRefresh = () => {
      queryClient.invalidateQueries()
    }

    return (
      <SafeAreaView className="flex-1 bg-black">
        <ScrollView
          refreshControl={
            <RefreshControl refreshing={cwLoading || nuLoading} onRefresh={onRefresh} />
          }
        >
          {/* Header: "Velox" text or logo */}
          <Text className="text-2xl font-bold text-white px-4 pt-4 pb-2">Velox</Text>

          {continueWatching && continueWatching.length > 0 && (
            <MediaRow title="Continue Watching" items={continueWatching} />
          )}
          {nextUp && nextUp.length > 0 && (
            <MediaRow title="Next Up" items={nextUp} />
          )}

          {/* Future: Recently Added (needs media list API with sort=date_added) */}
        </ScrollView>
      </SafeAreaView>
    )
  }
  ```
  - SafeAreaView for status bar
  - Pull-to-refresh invalidates all queries

## Files to Create
- `mobile/src/utils/image.ts`
- `mobile/src/components/MediaCard.tsx`
- `mobile/src/components/MediaRow.tsx`
- `mobile/app/(tabs)/index.tsx` — home screen

## Test Criteria
- [ ] Continue Watching row shows items with progress bars
- [ ] Next Up row shows next episodes
- [ ] TMDb images load via proxy: `{server}/api/images/tmdb/w500/{path}`
- [ ] Local uploaded images load: `{server}/api/images/local/media/{path}`
- [ ] Pull-to-refresh reloads data
- [ ] Tap card navigates to detail (route can be placeholder)
- [ ] Rows hidden when no data (no empty title bar)

---
Next Phase: [phase-13-library-browser.md](phase-13-library-browser.md)
