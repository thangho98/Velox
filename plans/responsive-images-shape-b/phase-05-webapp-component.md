# Phase 05: Webapp `<ResponsiveImage>` Component + Migration

Status: ✅ Complete
Dependencies: Phase 02 (new fields in responses) — blurhash optional (Phase 04 can still be in progress)

## Objective

Create a single `<ResponsiveImage>` component that consumes `ImageResource` and renders `<picture>` with srcset + blurhash placeholder fade-in. Migrate all `<img src={mediaImage(...)} />` call sites to use it.

## Changes

### 1. Dependencies

`webapp/package.json`:
```json
"dependencies": {
  "react-blurhash": "^0.3.0"
}
```

### 2. Types (shared package — source of truth)

Webapp re-exports API types from `@velox/shared/types` via [webapp/src/types/api.ts](../../webapp/src/types/api.ts). **All type additions MUST land in the shared package** — the webapp `types/` folder is a pass-through. Adding types directly under `webapp/src/types/` would desync shared hooks from webapp consumers and break Android (which also consumes shared types in its own way).

Add to [packages/shared/types/common.ts](../../packages/shared/types/common.ts) (common file since ImageResource is used across media/series/user):

```ts
export interface ImageResource {
  url: string
  srcset: Record<string, string>  // { "185": "...", "342": "...", "500": "...", "original": "..." }
  type: 'poster' | 'backdrop' | 'still' | 'logo'
  aspect: string                   // "2:3" | "16:9" | "variable"
  width?: number
  height?: number
  blurhash?: string | null
}
```

Re-export via [packages/shared/types/index.ts](../../packages/shared/types/index.ts) if not auto-exported by barrel.

Extend existing types in their respective files:

- [packages/shared/types/media.ts](../../packages/shared/types/media.ts) — Media, MediaListItem, ContinueWatchingItem, NextUpItem, UserData, MostWatchedItem
- [packages/shared/types/series.ts](../../packages/shared/types/series.ts) — Series, Season, Episode, SeriesListItem
- [packages/shared/types/auth.ts](../../packages/shared/types/auth.ts) — User (only if Profile scope is expanded; per plan.md, Profile is deferred so User stays on legacy string path here)

Shape per type (keep legacy fields during transition — removed in Phase 07):
```ts
export interface Media {
  // ... existing legacy fields stay ...
  poster?: ImageResource | null
  backdrop?: ImageResource | null
  logo?: ImageResource | null
  thumb?: ImageResource | null
}
```

Verify Android isn't broken: Android has its own Kotlin DTO layer ([data/model/PlaybackModels.kt](../../android/app/src/main/java/com/velox/app/data/model/) etc.) so adding optional fields in shared TS doesn't directly affect it. Phase 06 handles Android DTO updates.

### 3. Component

[webapp/src/components/ResponsiveImage.tsx](../../webapp/src/components/ResponsiveImage.tsx) — new:

```tsx
import { Blurhash } from 'react-blurhash'
import { useState } from 'react'
import type { ImageResource } from '@/types/image'

interface Props {
  data: ImageResource | null | undefined
  sizes: string           // CSS `sizes` attribute, e.g. "(max-width: 768px) 185px, 500px"
  alt: string
  className?: string
  loading?: 'lazy' | 'eager'
}

export function ResponsiveImage({ data, sizes, alt, className = '', loading = 'lazy' }: Props) {
  const [loaded, setLoaded] = useState(false)
  if (!data) return <div className={`bg-zinc-900 ${className}`} style={{ aspectRatio: '2/3' }} />

  const srcsetAttr = Object.entries(data.srcset)
    .filter(([k]) => k !== 'original')
    .map(([width, url]) => `${url} ${width}w`)
    .join(', ')

  const aspectStyle = aspectToCss(data.aspect)

  return (
    <div className={`relative overflow-hidden ${className}`} style={aspectStyle}>
      {data.blurhash && !loaded && (
        <Blurhash
          hash={data.blurhash}
          width="100%"
          height="100%"
          resolutionX={32}
          resolutionY={32}
          punch={1}
        />
      )}
      <picture>
        <img
          src={data.url}
          srcSet={srcsetAttr}
          sizes={sizes}
          alt={alt}
          loading={loading}
          onLoad={() => setLoaded(true)}
          className={`absolute inset-0 h-full w-full object-cover transition-opacity duration-300 ${loaded ? 'opacity-100' : 'opacity-0'}`}
        />
      </picture>
    </div>
  )
}

function aspectToCss(aspect: string): React.CSSProperties {
  if (aspect === '2:3') return { aspectRatio: '2/3' }
  if (aspect === '16:9') return { aspectRatio: '16/9' }
  return {}
}
```

### 4. Migrate call sites

Replace `<img src={mediaImage(media.poster_path)} />` → `<ResponsiveImage data={media.poster} sizes="..." alt="..." />`.

**Authoritative discovery** (run before starting migration, re-run after):
```bash
grep -rn "tmdbImage\|mediaImage\|seriesImage\|poster_path\|backdrop_path\|logo_path\|still_path" webapp/src | grep -v "\.test\."
```
Every hit must either be migrated or explicitly justified as "stays on legacy string" (e.g. passed into an external API that expects a URL string). Phase 07 will require this grep to return zero results, so any leftover here becomes a blocker there.

Affected (~20 files at current count; verify with the grep above):
- Card/row components: [MediaCard.tsx](../../webapp/src/components/MediaCard.tsx), [MediaRow.tsx](../../webapp/src/components/MediaRow.tsx), [EpisodeCard.tsx](../../webapp/src/components/EpisodeCard.tsx), [NextUpCard.tsx](../../webapp/src/components/NextUpCard.tsx), [ContinueWatchingCard.tsx](../../webapp/src/components/ContinueWatchingCard.tsx), [FolderCard.tsx](../../webapp/src/components/FolderCard.tsx)
- Detail pages: [MediaDetailPage.tsx](../../webapp/src/pages/MediaDetailPage.tsx), [SeriesDetailPage.tsx](../../webapp/src/pages/SeriesDetailPage.tsx)
- Listing pages: [BrowsePage.tsx](../../webapp/src/pages/BrowsePage.tsx), [HomePage.tsx](../../webapp/src/pages/HomePage.tsx), [SearchPage.tsx](../../webapp/src/pages/SearchPage.tsx), [MoviesPage.tsx](../../webapp/src/pages/MoviesPage.tsx), [SeriesPage.tsx](../../webapp/src/pages/SeriesPage.tsx)
- **[WatchPage.tsx](../../webapp/src/pages/WatchPage.tsx)** — 3 current `tmdbImage` call sites (MediaSession metadata artwork at line 385, logo overlay at 456, up-next thumbnail at 1849). Player page = high-impact; do NOT skip.
- [WatchDetailSheet.tsx](../../webapp/src/components/watch/WatchDetailSheet.tsx), [library/LibraryContent.tsx](../../webapp/src/components/library/LibraryContent.tsx)
- [MetadataEditor.tsx](../../webapp/src/components/metadata/MetadataEditor.tsx) (already uses `tmdbImage` — swap to `data={entity.poster}`)

**Note on non-component usage of tmdbImage:** some call sites pass the URL into browser APIs (WatchPage line 385 sets `navigator.mediaSession.metadata` artwork). Those can't use `<ResponsiveImage>` — instead read `ImageResource.url` directly:
```ts
const posterUrl = media.poster?.url ?? ''
navigator.mediaSession.metadata = new MediaMetadata({ ..., artwork: [{ src: posterUrl }] })
```

Each call site picks appropriate `sizes` attribute:
- Card grid: `"(max-width: 768px) 185px, 342px"`
- Detail backdrop: `"100vw"`
- Detail poster: `"(max-width: 768px) 342px, 500px"`
- Notification thumbnail: `"185px"`

### 5. Legacy helper stays temporarily

[packages/shared/lib/image.ts](../../packages/shared/lib/image.ts) — helpers `tmdbImage/mediaImage/seriesImage` continue working for any missed call site or non-critical usage. Remove in Phase 07.

## Tests

- Basic render test with Blurhash placeholder showing when `loaded=false`.
- Fallback `<div>` rendering when `data=null`.
- Snapshot: srcset attribute format.

## Acceptance

- `npx tsc --noEmit` green.
- `npm run lint` — no new errors from changes (pre-existing errors allowed).
- Manual: open browse page → cards render with blurhash flash then fade to image. DevTools Network: confirm correct srcset variant selected per viewport.
- No CLS (cumulative layout shift) — aspect ratio locks container size.

## Notes

- `react-blurhash` uses canvas; for SSR (if ever added) wrap in `suppressHydrationWarning` or use `react-blurhash/server`.
- `sizes` attribute is mandatory for `<img srcset>` — browsers need it to pick. Empty string = use default (not great).
- `loading="lazy"` default, eager only for above-the-fold (e.g. detail page hero).

---
Next: [Phase 06 — Android Coil blurhash + migration](phase-06-android-component.md)
