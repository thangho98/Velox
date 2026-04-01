# Plan R: Refactor Frontend

Created: 2026-04-01
Status: ⬜ Pending

## Overview

Refactor large React/TypeScript files for readability and maintainability.
Pure structural refactor — no behavior changes, no new features.

## Rules
1. **No behavior changes** — only move code between files
2. **Re-export from original locations** where needed to avoid breaking imports
3. **Each file under 400 lines** (target, not hard limit)
4. **Build check** (`npm run build`) after each phase
5. **No new dependencies**

---

## Phase 1: SettingsPage.tsx (3,620 → ~16 section files + layout)

The biggest win. 16 independent sections rendered by `activeSection` state.

### Step 1: Create settings section directory + shared utilities

```
src/pages/settings/
├── SettingsPage.tsx          # Layout + sidebar nav + section router (~200 lines)
├── ProfileSection.tsx        # Display name, avatar (~60 lines)
├── PreferencesSection.tsx    # Language, theme, quality defaults (~100 lines)
├── SecuritySection.tsx       # Password change (~80 lines)
├── SessionsSection.tsx       # Active sessions (~60 lines)
├── NotificationsSection.tsx  # Push notification prefs (~120 lines)
├── MetadataSection.tsx       # TMDb, OMDb, TVDB, Fanart API keys (~430 lines)
├── SubtitlesSection.tsx      # OpenSubs, Subdl, DeepL configs (~450 lines)
├── PlaybackSection.tsx       # Playback settings (~80 lines)
├── GeneralSection.tsx        # Server info, version, DB stats (~160 lines)
├── LibrariesSection.tsx      # Library CRUD, scan, browse (~280 lines)
├── UsersSection.tsx          # User CRUD with modals (~310 lines)
├── ActivitySection.tsx       # Activity log (~120 lines)
├── TasksSection.tsx          # Scheduled tasks (~120 lines)
├── WebhooksSection.tsx       # Webhook CRUD (~190 lines)
├── CinemaSection.tsx         # Cinema intro upload (~100 lines)
├── MarkersSection.tsx        # Intro/credits markers (~230 lines)
├── PretranscodeSection.tsx   # Transcoding profiles, queues (~280 lines)
└── shared.tsx                # SectionHeader, Field, SaveButton, Modal, Spinner (~70 lines)
```

### Step 2: Update SettingsPage.tsx
- Keep sidebar navigation + `activeSection` state
- Import and render section components
- Move utility components to `shared.tsx`

### Step 3: Update route (if SettingsPage path changes)
- Keep `/settings` route pointing to new `settings/SettingsPage.tsx`

---

## Phase 2: useMedia.ts (726 lines → 6 domain files)

41 exports grouped into domain-specific hook files.

```
src/hooks/stores/
├── useMedia.ts               # Re-exports all (backward compat) (~30 lines)
├── media/
│   ├── useLibrary.ts         # useLibraries, useCreateLibrary, useDeleteLibrary, useScanLibrary, useFsBrowse (~80 lines)
│   ├── useMediaQuery.ts      # useMediaList, useMedia, useMediaWithFiles, useSearch, useFolderBrowse (~100 lines)
│   ├── useProgress.ts        # useProgress, useUpdateProgress, useDismissProgress, useFavorites, useToggleFavorite (~80 lines)
│   ├── useContinueWatching.ts # useContinueWatching, useNextUp, useRecentlyWatched (~60 lines)
│   ├── useSeries.ts          # useSeriesList, useSeriesDetail, useSeriesSearch, useSeasons, useEpisodes (~100 lines)
│   ├── usePlayback.ts        # useStreamUrls, useSubtitles, useAudioTracks, usePlaybackInfo, useStreamUrl (~80 lines)
│   ├── useSubtitleOps.ts     # useSubtitleSearch, useDownloadSubtitle, useTranslateSubtitle (~50 lines)
│   ├── useMetadataOps.ts     # useRefreshMetadata, useEditMedia/Series/EpisodeMetadata, useUploadImage, useUnlockMetadata (~100 lines)
│   └── useGenres.ts          # useAllGenres, useMediaGenres, useSeriesGenres, useMediaCredits, useSeriesCredits (~60 lines)
```

**Key:** `useMedia.ts` becomes a barrel re-export file for backward compatibility.

---

## Phase 3: useSettings.ts (448 lines → factory + domain files)

### Step 1: Create settings hook factory

```typescript
// src/hooks/stores/settings/factory.ts (~40 lines)
export function createSettingsHook<T>(endpoint: string, queryKey: string[]) {
  function useGet() {
    return useQuery({ queryKey, queryFn: () => fetchWithAuth<T>(`/api/admin/settings/${endpoint}`) })
  }
  function useUpdate() {
    const qc = useQueryClient()
    return useMutation({
      mutationFn: (data: Partial<T>) => fetchWithAuth<T>(`/api/admin/settings/${endpoint}`, { method: 'PUT', body: data }),
      onSuccess: (data) => qc.setQueryData(queryKey, data),
    })
  }
  return { useGet, useUpdate }
}
```

### Step 2: Split into domain files

```
src/hooks/stores/
├── useSettings.ts            # Re-exports all (backward compat) (~30 lines)
├── settings/
│   ├── factory.ts            # createSettingsHook factory (~40 lines)
│   ├── useMetadataSettings.ts # TMDb, OMDb, TVDB, Fanart (~40 lines using factory)
│   ├── useSubtitleSettings.ts # OpenSubs, Subdl, DeepL, AutoSub (~40 lines)
│   ├── usePlaybackSettings.ts # Playback, Cinema (~30 lines)
│   └── usePretranscodeSettings.ts # Pretranscode settings + control hooks (~100 lines)
```

---

## Phase 4: types/api.ts (643 lines → 6 domain files)

```
src/types/
├── api.ts                    # Re-exports all (backward compat) (~20 lines)
├── auth.ts                   # LoginRequest, LoginResponse, User, Session, UserPreferences (~80 lines)
├── media.ts                  # Media, MediaFile, MediaWithFiles, MediaListItem, Genre, BrowseResult, SearchResult (~120 lines)
├── series.ts                 # Series, SeriesListItem, Season, Episode, metadata edit types (~150 lines)
├── playback.ts               # StreamUrls, PlaybackInfo, QualityOption, SubtitleTrack, AudioTrack, SkipSegment (~130 lines)
├── settings.ts               # All settings interfaces (OpenSubs, TMDb, etc.) (~100 lines)
└── admin.ts                  # ServerInfo, LibraryStats, Webhook, ScheduledTask, ActivityLog (~70 lines)
```

---

## Phase 5: WatchPage.tsx (2,017 lines → page + sub-components + hooks)

Most complex refactor. Extract concerns incrementally.

### Step 1: Extract hooks

```
src/hooks/
├── usePlayerControls.ts      # togglePlay, toggleMute, seekTo, handleVolumeChange (~80 lines)
├── useKeyboardShortcuts.ts   # onKeyDown handler with seek/volume/fullscreen (~60 lines)
└── useFullscreen.ts          # toggleFullscreen, fullscreen event listeners (~50 lines)
```

### Step 2: Extract sub-components

```
src/components/watch/
├── PlayerControls.tsx        # Bottom bar: play/pause, volume, timeline, time display (~200 lines)
├── TimelineBar.tsx           # Seekable progress bar with trickplay preview (~150 lines)
├── EpisodeDrawer.tsx         # Season/episode browser panel (~150 lines)
├── WatchTopBar.tsx            # (already exists — 83 lines)
├── WatchDetailSheet.tsx       # (already exists — 290 lines)
└── WatchPlaybackStatsOverlay.tsx # (already exists)
```

### Step 3: Slim down WatchPage.tsx
- Keep: video element, HLS setup, state coordination
- Target: ~500 lines (down from 2,017)

---

## Phase 6: SubtitlePicker.tsx (458 lines → component + utils)

```
src/lib/languages.ts          # LANG_NAMES, normalizeLanguageCode, languageMatches (~70 lines)
src/components/
├── SubtitlePicker.tsx        # Main picker, slimmed (~200 lines)
├── SubtitleRow.tsx            # Single subtitle option row (~60 lines)
└── SubtitleSourceSelector.tsx # Source selection modal (~80 lines)
src/hooks/useSubtitleTranslation.ts  # Translation state + API call (~50 lines)
```

---

## Verification

After each phase:
```sh
cd webapp && npm run build && npm run lint
```

## Summary

| Phase | File | Before | After | Impact |
|-------|------|--------|-------|--------|
| 1 | SettingsPage.tsx | 3,620 | ~18 files (60-450 each) | Critical |
| 2 | useMedia.ts | 726 | 9 files (50-100 each) | High |
| 3 | useSettings.ts | 448 | 5 files (30-100 each) | High |
| 4 | types/api.ts | 643 | 7 files (20-150 each) | Medium |
| 5 | WatchPage.tsx | 2,017 | ~500 + 5 extracted | High |
| 6 | SubtitlePicker.tsx | 458 | ~200 + 4 extracted | Medium |
| **Total** | | **7,912** | **~45 files** | |
