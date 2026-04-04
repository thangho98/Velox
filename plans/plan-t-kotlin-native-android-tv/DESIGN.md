# DESIGN: Plan T - Kotlin Native Android + Android TV

Created: 2026-04-04
Based on: [plan.md](plan.md)

---

## 1. Product Intent

Velox không chỉ cần một Android app "chạy được", mà cần một native client khai thác đúng các lợi thế mà backend/playback engine đã có:

- Direct Play / Direct Stream / Transcode decision từ backend
- audio-only incompatibility không kéo tụt video quality nếu không cần
- pretranscode để instant playback
- multi-quality HLS thay vì single-quality restart-heavy flow
- Android TV remote-first UX

Native rewrite chỉ đáng làm nếu client mới tôn trọng và tận dụng được toàn bộ các điểm này.

---

## 2. Current Client Inventory

### 2.1. Current RN App Surfaces

| Current Surface | Source | Native Destination | Ship |
|---|---|---|---|
| Login | `mobile/src/screens/LoginScreen.tsx` | `feature:auth` | Ship 1 |
| Home | `mobile/src/screens/HomeScreen.tsx` | `feature:home` | Ship 1 |
| Movies / Series listing | `mobile/src/screens/MoviesScreen.tsx`, `SeriesScreen.tsx` | `feature:home`, `feature:browse` | Ship 1 |
| Library browse | `mobile/src/screens/BrowseScreen.tsx`, `LibraryBrowseScreen.tsx` | `feature:browse` | Ship 1 |
| Search | `mobile/src/screens/SearchScreen.tsx` | `feature:search` | Ship 1 |
| Media detail | `mobile/src/screens/MediaDetailScreen.tsx` | `feature:detail` | Ship 1 |
| Series detail / episodes | `mobile/src/screens/SeriesDetailScreen.tsx` | `feature:detail` | Ship 1 |
| Player | `mobile/src/screens/VideoPlayerScreen.tsx` | `feature:player` + `core:player` | Ship 1 |
| Chromecast | `mobile/src/hooks/useChromecast.ts` | `feature:cast` | Ship 2/3 |
| Favorites | `mobile/src/screens/FavoritesScreen.tsx` | `feature:favorites` | Ship 3 |
| Profile | `mobile/src/screens/ProfileScreen.tsx` | `feature:profile` | Ship 3 |
| Settings | `mobile/src/screens/SettingsScreen.tsx` | `feature:settings` | Ship 3 |
| Admin | `mobile/src/screens/AdminScreen.tsx` | defer / web-admin | Later |
| TV utilities | `mobile/src/lib/tv.ts`, `mobile/src/components/FocusableCard.tsx` | `feature:tv` | Ship 2 |

### 2.2. Current Reusable Knowledge

| Area | Source | Reuse Mode |
|---|---|---|
| Auth endpoints/contracts | `packages/shared/hooks/auth.ts`, `packages/shared/types/*` | port to Kotlin |
| Media/detail/search endpoints | `packages/shared/hooks/media/useMediaQuery.ts`, `useGenres.ts`, `useSeries.ts` | port to Kotlin |
| Progress/favorites | `packages/shared/hooks/media/useProgress.ts` | port to Kotlin |
| Playback contracts | `packages/shared/hooks/media/usePlayback.ts`, `packages/shared/types/playback.ts` | port to Kotlin |
| Existing Android player logic | `mobile/modules/velox-player/android/*` | refactor/rewrite into pure native module |
| TV behavior expectations | `mobile/src/lib/tv.ts` | translate into Compose TV UX rules |

---

## 3. Native Repo Layout

```text
android-native/
├── app/
├── core/
│   ├── common/
│   ├── model/
│   ├── network/
│   ├── datastore/
│   ├── designsystem/
│   └── player/
├── feature/
│   ├── auth/
│   ├── home/
│   ├── browse/
│   ├── detail/
│   ├── search/
│   ├── player/
│   ├── favorites/
│   ├── settings/
│   ├── profile/
│   ├── cast/
│   └── tv/
└── benchmark/
```

### 3.1. Module Dependency Rules

```text
app
 -> feature:*
 -> core:designsystem
 -> core:common

feature:*
 -> core:model
 -> core:network
 -> core:datastore
 -> core:designsystem
 -> core:player (player-related features only)
 -> core:common

core:player
 -> core:model
 -> core:common

core:network
 -> core:model
 -> core:common
```

Rules:
- feature modules never depend on each other directly unless absolutely necessary
- `core:model` owns DTO/domain model definitions
- `core:network` owns Retrofit services + repositories
- `core:player` owns Media3 integration and must remain UI-framework-light

---

## 4. Navigation Design

### 4.1. Phone / Tablet Graph

```text
RootGraph
├── Bootstrap
├── AuthGraph
│   └── Login
└── MainGraph
    ├── Home
    ├── Browse
    ├── Search
    ├── Favorites
    ├── Profile
    ├── Settings
    ├── MediaDetail/{mediaId}
    ├── SeriesDetail/{seriesId}
    └── Player/{mediaId}
```

### 4.2. TV Graph

```text
TvRootGraph
├── TvBootstrap
├── TvAuth
└── TvMain
    ├── TvHome
    ├── TvBrowse
    ├── TvSearch
    ├── TvMediaDetail/{mediaId}
    ├── TvSeriesDetail/{seriesId}
    └── TvPlayer/{mediaId}
```

TV should not just be a larger version of the phone graph. Focus behavior and list/detail hierarchy must be explicit.

---

## 5. Data Layer Design

### 5.1. Contract Porting Strategy

Source of truth for API contracts:
- `packages/shared/types/*`
- `packages/shared/hooks/auth.ts`
- `packages/shared/hooks/media/useMediaQuery.ts`
- `packages/shared/hooks/media/useGenres.ts`
- `packages/shared/hooks/media/useSeries.ts`
- `packages/shared/hooks/media/useProgress.ts`
- `packages/shared/hooks/media/usePlayback.ts`

Porting rule:
- port DTO names/fields faithfully first
- only introduce Android-specific domain wrappers after DTO parity is verified

### 5.2. Repository Boundary

Repositories:
- `AuthRepository`
- `ProfileRepository`
- `LibraryRepository`
- `MediaRepository`
- `SeriesRepository`
- `SearchRepository`
- `PlaybackRepository`
- `FavoritesRepository`
- `SettingsRepository`
- `CastRepository`

Each repository returns:
- success data
- typed failure category
- never raw Retrofit exceptions into UI

### 5.3. Auth Refresh Flow

```text
Request -> OkHttp interceptor adds bearer token
       -> 401
       -> Authenticator attempts single refresh
       -> DataStore tokens updated
       -> Original request retried
       -> If refresh fails: session cleared + UI redirected to auth
```

Need parity with the current TS token-refresh queue behavior, but implemented natively with OkHttp.

---

## 6. State Management Design

### 6.1. Screen State

Every screen should expose:

```kotlin
data class ScreenUiState(
  val isLoading: Boolean = false,
  val error: UiError? = null,
  val data: ...
)
```

Pattern:
- `ViewModel` owns `MutableStateFlow`
- Composables render stateless content from immutable `UiState`
- one-off events use `SharedFlow` or event channel, not mutable flags sprinkled through UI

### 6.2. Persistent Local State

Use DataStore for:
- auth/session tokens
- server selection if needed
- player defaults
- subtitle defaults
- playback preferences
- last used quality preference

Do not port Zustand semantics literally.

---

## 7. Player Architecture

### 7.1. Core Requirement

The native player must preserve backend-driven playback behavior:

- Direct Play when possible
- Direct Stream/remux when needed
- HLS/transcode when required
- quality switching through `available_qualities`
- progress sync through `/profile/progress/{id}`
- subtitle/audio selection through playback request parameters

### 7.2. Native Player Layers

```text
feature:player
 -> PlayerViewModel
 -> PlayerUiState
 -> PlayerScreen / Controls / Pickers

core:player
 -> VeloxPlayerEngine
 -> PlayerSessionState
 -> TrackMapper
 -> QualityMapper
 -> ProgressSyncCoordinator
 -> Media3 adapter
```

### 7.3. Playback Request Builder

The client should build a request profile from device capabilities and user selections:

- supported video codecs
- supported audio codecs
- supported containers
- `max_height`
- selected audio track
- selected subtitle track

This builder is a first-class native component, not ad hoc logic inside the player screen.

### 7.4. Error and Fallback Rules

If local playback fails:
- inspect whether backend response also includes/permits HLS fallback
- retry with fallback path where safe
- avoid silent black-screen failures

If audio-only incompatibility exists:
- do not force an unnecessary video downgrade
- respect backend playback response and expose the resulting stream method clearly in debug state

### 7.5. Quality UX

Quality picker should expose:
- Auto
- Original
- Pretranscode qualities
- Live transcode qualities

Suggested badge labels:
- `Instant` for pretranscode
- `Live` for realtime transcode
- `Original` for direct

This matters because Velox is stronger than Jellyfin/Emby in this area, and the client should make that advantage visible.

---

## 8. Android TV UX Rules

### 8.1. Remote-First Principles

- Every primary action must be reachable with DPAD only
- Focus state must always be visible
- No hidden hover-only affordances
- No tiny tap targets reused blindly from phone UI

### 8.2. TV Component Strategy

Use dedicated TV composables for:
- content rows
- posters/cards
- action rails
- focused buttons
- player controls

Do not rely on mobile `LazyRow`/`LazyVerticalGrid` behavior without explicit focus testing.

### 8.3. Focus Rules

- entering detail from a row should restore focus to the same item on back
- player exit should restore focus to the invoking CTA if possible
- modal sheets on TV must trap and restore focus predictably

### 8.4. TV Search

V1 can ship with a basic text-entry screen if:
- it is actually usable with remote
- keyboard/IME behavior is verified

If not, search should be simplified rather than shipping a broken text field experience.

---

## 9. What Stays Out of V1

Keep out unless they are cheap and already stable:
- metadata editor
- admin ops
- deep subtitle translate/search management
- setup wizard management
- power-user diagnostic screens

These can remain in web admin or legacy RN while Android native catches up.

---

## 10. Testing Strategy

### 10.1. Unit Tests
- auth refresh logic
- request builders
- repositories
- player state reducers / mappers

### 10.2. Integration Tests
- Retrofit + MockWebServer for auth/media/playback
- progress sync flows
- track selection mapping

### 10.3. Compose UI Tests
- login
- main navigation
- search
- detail to player routing
- settings/profile basics

### 10.4. Real Device Matrix
- Android phone
- Android tablet if officially supported
- Android TV / Google TV
- Chromecast receiver path

### 10.5. Media Playback Matrix
- H.264 MP4
- HEVC MKV
- files needing audio-only transcode
- HLS fallback
- pretranscode quality
- subtitle formats used in the real library

---

## 11. Rollout Strategy

### Stage A
- internal builds only
- compare native browse/playback against RN on same sample library

### Stage B
- native becomes recommended for Android testers
- RN remains fallback path

### Stage C
- native becomes default Android recommendation
- RN Android feature development freezes

### Stage D
- evaluate whether RN Android is archived or kept as emergency fallback

---

## 12. Acceptance Gates

### Gate 1: Foundation
- app builds
- login works
- session restore works

### Gate 2: Browse
- home/browse/search/detail usable

### Gate 3: Playback
- movie and episode playback stable
- progress sync stable

### Gate 4: Advanced Playback
- subtitle/audio/quality/cast stable

### Gate 5: TV
- remote-only navigation passes manual checklist

### Gate 6: Cutover
- real-device beta feedback acceptable
- no P0/P1 blockers in daily usage flows
