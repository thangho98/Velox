# Plan S: Expo Mobile App (Monorepo)

Created: 2026-04-01
Status: ⬜ Pending

## Overview

Android mobile app cho Velox media server. Dùng Expo + expo-video (ExoPlayer) để Direct Play nhiều format hơn browser. Monorepo architecture với shared package để reuse types, hooks, stores giữa web app và mobile app.

Android TV dùng web app hiện tại.

## Architecture

```
velox/
├── backend/                       # Go backend (unchanged)
├── packages/
│   └── shared/                    # @velox/shared — reuse giữa web + mobile
│       ├── types/                 # API types (100% reuse)
│       ├── api/                   # API client + query hooks (90% reuse)
│       ├── stores/                # Zustand store factories (80% reuse)
│       └── lib/                   # Utils: languages, image (100% reuse)
├── webapp/                        # Web app — import from @velox/shared
│   └── src/
│       └── platform/              # Web-specific: localStorage, browser fetch
├── mobile/                        # Expo app — import from @velox/shared
│   └── src/
│       └── platform/              # Mobile-specific: SecureStore, RN fetch
├── package.json                   # pnpm workspace root
└── pnpm-workspace.yaml
```

```
┌──────────────┐     REST API      ┌──────────────┐
│  Velox       │ ◄───────────────► │  Web App     │
│  Backend     │                   │  (React)     │
│  (Go+SQLite) │                   └──────┬───────┘
│              │                          │
│              │     REST API      ┌──────┴───────┐
│              │ ◄───────────────► │  Mobile App  │
└──────────────┘                   │  (Expo)      │
                                   └──────────────┘
                                          │
                                   @velox/shared
                                   (types, hooks,
                                    stores, utils)
```

## Tech Stack

| Layer | Web App | Mobile App | Shared |
|-------|---------|-----------|--------|
| Framework | React 19 + Vite | Expo SDK 53+ | — |
| Language | TypeScript | TypeScript | TypeScript |
| Navigation | React Router | Expo Router | — |
| Video | HTML5 `<video>` + HLS.js | expo-video (ExoPlayer) | — |
| Server State | TanStack React Query | TanStack React Query | Query hooks |
| Client State | Zustand | Zustand | Store factories |
| Styling | TailwindCSS 4 | NativeWind | — |
| Storage | localStorage | SecureStore + MMKV | Storage adapter interface |
| Types | @velox/shared | @velox/shared | ✅ Single source |
| API Client | @velox/shared | @velox/shared | ✅ Platform-agnostic |

## Shared Package Design (`@velox/shared`)

### Platform Adapter Interface

```typescript
// packages/shared/platform.ts
export interface PlatformAdapter {
  storage: {
    getItem(key: string): Promise<string | null>
    setItem(key: string, value: string): Promise<void>
    removeItem(key: string): Promise<void>
  }
  secureStorage: {
    getItem(key: string): Promise<string | null>
    setItem(key: string, value: string): Promise<void>
    removeItem(key: string): Promise<void>
  }
}
```

Web provides: `localStorage` adapter
Mobile provides: `MMKV` + `expo-secure-store` adapter

### Shared Types (~550 lines, 100% reuse)
```
packages/shared/types/
├── index.ts        # barrel re-export
├── common.ts       # ApiResponse, FsBrowseResponse
├── auth.ts         # Login, User, Session, Preferences
├── media.ts        # Media, Library, Genre, Search
├── series.ts       # Series, Season, Episode
├── playback.ts     # StreamUrls, PlaybackInfo, Subtitles
└── admin.ts        # ServerInfo, Webhook, Task
```
Move directly from `webapp/src/types/` — already split by domain.

### Shared API Client (~200 lines, 90% reuse)
```
packages/shared/api/
├── client.ts       # Platform-agnostic fetch wrapper
│                   # Receives token getter/setter via init()
│                   # Handles refresh, error mapping, FormData
└── index.ts
```

### Shared Query Hooks (~700 lines, 90% reuse)
```
packages/shared/hooks/
├── useAuth.ts           # login, logout, refresh, profile, preferences, sessions
├── media/
│   ├── useLibrary.ts
│   ├── useMediaQuery.ts
│   ├── usePlayback.ts
│   ├── useProgress.ts
│   ├── useContinueWatching.ts
│   ├── useSeries.ts
│   ├── useSubtitleOps.ts
│   ├── useMetadataOps.ts
│   └── useGenres.ts
├── settings/
│   ├── factory.ts
│   ├── useMetadataSettings.ts
│   ├── useSubtitleSettings.ts
│   ├── usePlaybackSettings.ts
│   └── usePretranscodeSettings.ts
├── useAdmin.ts
├── useUsers.ts
└── useNotifications.ts
```
Move from `webapp/src/hooks/stores/` — already split by domain (Plan R Phase 2-3).

### Shared Stores (~260 lines, 80% reuse)
```
packages/shared/stores/
├── createAuthStore.ts     # Factory: inject storage adapter
└── createPlayerStore.ts   # Factory: inject storage adapter
```

### Shared Utils (~150 lines, 100% reuse)
```
packages/shared/lib/
├── languages.ts     # LANG_NAMES, normalize, parseSubtitleLabel
└── image.ts         # tmdbImage URL builder
```

## Reuse Summary

| Category | Lines | Reuse % | Shared |
|----------|-------|---------|--------|
| Types | ~550 | 100% | ✅ |
| API client | ~200 | 90% | ✅ |
| Query hooks | ~700 | 90% | ✅ |
| Zustand stores | ~260 | 80% | ✅ |
| Lib utils | ~150 | 100% | ✅ |
| **Total shared** | **~1,860** | | |
| UI components | ~5,000 | 0% | ❌ Platform-specific |

---

## Phases

| Phase | Name | Status |
|-------|------|--------|
| 01 | Monorepo Setup + Extract Shared Package | ⬜ Pending |
| 02 | Migrate Web App to Use @velox/shared | ⬜ Pending |
| 03 | Expo Project Setup + Auth | ⬜ Pending |
| 04 | Home + Browse Screens | ⬜ Pending |
| 05 | Media Detail + Series | ⬜ Pending |
| 06 | Video Player (ExoPlayer Direct Play) | ⬜ Pending |
| 07 | Player Controls + Subtitles | ⬜ Pending |
| 08 | Profile + Settings | ⬜ Pending |
| 09 | Polish + Build APK | ⬜ Pending |

---

## Phase 1: Monorepo Setup + Extract Shared Package

### Tasks
- [ ] Install pnpm (if not already): `npm i -g pnpm`
- [ ] Create `pnpm-workspace.yaml` at root:
  ```yaml
  packages:
    - 'packages/*'
    - 'webapp'
    - 'mobile'
  ```
- [ ] Create `packages/shared/package.json`:
  ```json
  { "name": "@velox/shared", "private": true, "main": "index.ts" }
  ```
- [ ] Create `packages/shared/tsconfig.json`
- [ ] Move `webapp/src/types/*.ts` → `packages/shared/types/` (keep barrel)
- [ ] Create `packages/shared/api/client.ts` — platform-agnostic API client
  - Extract core logic from `webapp/src/lib/fetch.ts`
  - Accept token getter/setter via `initApiClient(adapter)` function
  - No direct `localStorage` reference — use adapter
- [ ] Create `packages/shared/stores/createAuthStore.ts` — factory
- [ ] Create `packages/shared/stores/createPlayerStore.ts` — factory
- [ ] Move `webapp/src/lib/languages.ts` → `packages/shared/lib/languages.ts`
- [ ] Move `webapp/src/lib/image.ts` → `packages/shared/lib/image.ts`
- [ ] Create `packages/shared/platform.ts` — PlatformAdapter interface
- [ ] Verify: `pnpm install` + all packages resolve

---

## Phase 2: Migrate Web App to Use @velox/shared

### Tasks
- [ ] Update `webapp/package.json`: add `"@velox/shared": "workspace:*"`
- [ ] Update `webapp/tsconfig.app.json`: add path alias `@velox/shared`
- [ ] Create `webapp/src/platform/storage.ts` — localStorage adapter
- [ ] Update all imports in webapp:
  - `@/types/api` → `@velox/shared/types`
  - `@/lib/languages` → `@velox/shared/lib/languages`
  - `@/lib/image` → `@velox/shared/lib/image`
- [ ] Update `webapp/src/lib/fetch.ts` → wrap shared client with web adapter
- [ ] Update `webapp/src/stores/auth.ts` → use `createAuthStore(webStorage)`
- [ ] Update `webapp/src/stores/player.ts` → use `createPlayerStore(webStorage)`
- [ ] Move query hooks to shared (one by one, verify after each):
  - `hooks/stores/media/*.ts` → `@velox/shared/hooks/media/`
  - `hooks/stores/settings/*.ts` → `@velox/shared/hooks/settings/`
  - `hooks/stores/useAuth.ts` → `@velox/shared/hooks/useAuth.ts`
  - `hooks/stores/useAdmin.ts` → `@velox/shared/hooks/useAdmin.ts`
  - `hooks/stores/useUsers.ts` → `@velox/shared/hooks/useUsers.ts`
- [ ] Verify: `cd webapp && npm run build && npm run lint` — zero errors
- [ ] Webapp must work identically after migration

---

## Phase 3: Expo Project Setup + Auth

### Tasks
- [ ] Create Expo project: `cd mobile && npx create-expo-app@latest .`
- [ ] Update `mobile/package.json`: add `"@velox/shared": "workspace:*"`
- [ ] Install deps: expo-router, expo-video, expo-secure-store, nativewind, zustand, @tanstack/react-query, react-native-mmkv
- [ ] Setup NativeWind (TailwindCSS for RN)
- [ ] Create `mobile/src/platform/storage.ts` — MMKV + SecureStore adapters
- [ ] Create `mobile/src/stores/auth.ts`:
  ```typescript
  import { createAuthStore } from '@velox/shared/stores/createAuthStore'
  import { mobileSecureStorage } from '../platform/storage'
  export const useAuthStore = createAuthStore(mobileSecureStorage)
  ```
- [ ] Create `app/login.tsx` — Server URL input + login form
- [ ] Create `app/_layout.tsx` — Auth guard + QueryClientProvider
- [ ] Server config: URL input saved to MMKV, passed to API client
- [ ] Verify: Login + token refresh works on Android emulator

---

## Phase 4: Home + Browse Screens

### Tasks
- [ ] `app/(tabs)/index.tsx` — Home screen
  - Continue Watching row (using `useContinueWatching` from shared)
  - Next Up row (using `useNextUp` from shared)
  - Recently Added
- [ ] `app/(tabs)/library.tsx` — Library browser
  - Library selector (using `useLibraries` from shared)
  - Media grid with poster cards
  - Filter/sort by genre, year, rating
- [ ] `components/MediaCard.tsx` — Poster card (React Native)
- [ ] `components/MediaRow.tsx` — Horizontal FlatList
- [ ] Pull-to-refresh, expo-image for caching
- [ ] Bottom tab navigation: Home, Library, Search, Profile

---

## Phase 5: Media Detail + Series

### Tasks
- [ ] `app/media/[id].tsx` — Movie detail
  - Backdrop + metadata (using `useMediaWithFiles` from shared)
  - Play button → navigate to watch
  - Genres, credits (using `useMediaGenres`, `useMediaCredits` from shared)
- [ ] `app/series/[id].tsx` — Series detail
  - Season tabs + episode list
  - Using `useSeriesDetail`, `useSeasons`, `useEpisodes` from shared
- [ ] `components/EpisodeList.tsx` — Season/episode browser

---

## Phase 6: Video Player (ExoPlayer Direct Play)

### Tasks
- [ ] `app/watch/[id].tsx` — Fullscreen video player
- [ ] Fetch playback info using `usePlaybackInfo` from shared
  - Send ExoPlayer capabilities: h264, hevc, vp9, av1, aac, ac3, eac3, dts, mkv, mp4
  - Backend returns Direct Play for 90%+ files
- [ ] expo-video setup:
  - Direct Play: pass URL directly to VideoView
  - HLS fallback: expo-video handles HLS natively
- [ ] Auth header injection for stream URLs
- [ ] Resume from saved position (usePlayerStore from shared)
- [ ] Progress saving every 10s (useUpdateProgress from shared)
- [ ] Force landscape orientation during playback
- [ ] Background audio continuation

---

## Phase 7: Player Controls + Subtitles

### Tasks
- [ ] `components/PlayerControls.tsx` — Custom controls overlay
  - Play/pause, seek ±10s, progress scrubbing
  - Auto-hide after 3s (same as web)
- [ ] `components/SubtitlePicker.tsx` — Bottom sheet
  - Using shared language utils from `@velox/shared/lib/languages`
  - ExoPlayer renders SRT/ASS/PGS natively
- [ ] `components/QualityPicker.tsx` — Quality selection bottom sheet
- [ ] Audio track picker
- [ ] Skip intro/credits (skip_segments from playback info)
- [ ] Next episode auto-play
- [ ] Gesture controls: horizontal swipe = seek, vertical = volume/brightness
- [ ] PiP (Picture-in-Picture) support

---

## Phase 8: Profile + Settings

### Tasks
- [ ] `app/(tabs)/profile.tsx` — Profile + settings
  - Display name (using `useProfile` from shared)
  - Favorites list (using `useFavorites` from shared)
  - Watch history
- [ ] Settings:
  - Server URL management
  - Subtitle defaults (language, size)
  - Audio defaults
  - Quality defaults
  - Player preferences (all from shared player store)
- [ ] Notification preferences

---

## Phase 9: Polish + Build APK

### Tasks
- [ ] App icon + splash screen
- [ ] Loading skeletons + error states
- [ ] Network error handling (server offline, timeout)
- [ ] EAS Build configuration
- [ ] Generate APK for sideloading
- [ ] Test on physical Android devices
- [ ] Performance: image caching, list virtualization, memory
- [ ] Deep linking: `velox://watch/123`

---

## Backend Changes Needed

Minimal — API already mobile-ready:

| Change | Why | Effort |
|--------|-----|--------|
| Accept `mkv` in container capabilities | ExoPlayer handles MKV natively | 1 line |
| Extended codec list in playback decision | More Direct Play for capable clients | Already works |

---

## Future (Post-MVP)

- **Offline download** — expo-file-system + background download
- **Chromecast** — Cast SDK integration
- **Home screen widget** — Continue Watching
- **iOS** — Same Expo codebase, minimal changes (AVPlayer via expo-video)
- **Android TV** — Evaluate after mobile launch (Leanback or enhanced web app)
