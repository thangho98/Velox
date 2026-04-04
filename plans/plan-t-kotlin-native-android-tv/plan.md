# Plan T: Kotlin Native Android + Android TV
Created: 2026-04-04
Status: ⬜ Pending

## Overview
Rewrite Android app của Velox theo hướng native bằng Kotlin, ưu tiên Android phone/tablet và Android TV trước. Mục tiêu là giữ nguyên backend hiện tại, tận dụng contract playback/auth/media đã có, và chuyển phần client sang stack native ổn định hơn cho playback, cast, D-pad focus, và hiệu năng 10-foot UI.

RN app hiện tại sẽ được giữ lại như reference trong suốt migration. Không xóa `mobile/` cho tới khi Android native đạt mức usable thực tế.

## Goal
- Ship một app Android native đủ tốt để thay thế RN app trên Android/Android TV.
- Ưu tiên playback reliability, remote/D-pad UX, và startup/performance.
- Reuse backend/API contracts tối đa, nhưng không ép reuse TypeScript hooks/stores bằng mọi giá.

## Non-Goals (V1)
- Không làm iOS native ở plan này.
- Không làm Kotlin Multiplatform/KMP ở giai đoạn đầu.
- Không cố đạt 100% feature parity RN ngay từ ship đầu.
- Không xóa RN app sớm khi chưa có telemetry, QA, và beta đủ tốt.

## Why Rewrite
- Player hiện tại là điểm nặng nhất và là nơi RN dễ phát sinh friction.
- Android TV / 10-foot UX hợp với native focus system hơn RN.
- Repo đã có native player module Kotlin + Media3 để tái sử dụng về logic.
- Shared layer hiện hữu chủ yếu là TypeScript hooks/stores; giữ backend contract dễ hơn là kéo toàn bộ frontend abstraction sang native.

## Planning Artifacts
- [DESIGN.md](DESIGN.md) — kiến trúc chi tiết, module map, player/TV strategy, rollout gates
- [parity-matrix.md](parity-matrix.md) — mapping RN screen/feature -> native phase/ship
- [api-contract-checklist.md](api-contract-checklist.md) — checklist endpoint/DTO cần port từ TS sang Kotlin

## Proposed Repo Structure

```text
velox/
├── backend/                               # Go backend (unchanged)
├── mobile/                                # Existing Expo/RN app (reference during migration)
├── packages/shared/                       # Existing TS shared contracts (reference only)
├── webapp/                                # Existing web app
└── android-native/                        # New native Android project
    ├── app/                               # launcher app
    ├── core/
    │   ├── model/                         # Kotlin data models for API/domain
    │   ├── network/                       # Retrofit/OkHttp/auth interceptors
    │   ├── datastore/                     # tokens, preferences, player settings
    │   ├── designsystem/                  # theme, colors, typography, shared composables
    │   ├── player/                        # Media3 wrapper, track selection, playback state
    │   └── common/                        # result wrappers, utils, dispatchers
    ├── feature/
    │   ├── auth/
    │   ├── home/
    │   ├── browse/
    │   ├── detail/
    │   ├── search/
    │   ├── favorites/
    │   ├── player/
    │   ├── settings/
    │   ├── profile/
    │   ├── cast/
    │   └── tv/
    └── benchmark/                         # optional macrobenchmark phase
```

## Architecture Decisions
- UI: Jetpack Compose
- TV UI: Compose + `androidx.tv` components where helpful
- Navigation: Navigation Compose
- DI: Hilt
- Networking: Retrofit + OkHttp + Kotlinx Serialization
- Async/state: ViewModel + StateFlow + structured UI state
- Image loading: Coil
- Playback: Media3/ExoPlayer
- Persistence: DataStore
- Error handling: typed repository results + mapped UI error states
- Build: Gradle Kotlin DSL + version catalog

## Contract Strategy
- Backend API là source of truth.
- TypeScript shared package chỉ dùng làm reference để port model/request/response sang Kotlin.
- Ưu tiên freeze các contract sau trước khi code mạnh:
  - auth/login/refresh/logout
  - home/browse/search/detail
  - playback info / stream url / progress
  - favorites / profile / settings
  - subtitle ops nào cần cho V1

## Scope by Ship

### Ship 1: Android Playback MVP
- Login/session refresh
- Home
- Browse/library
- Media detail / series detail
- Search
- Fullscreen player
- Resume progress
- Audio/subtitle/quality switching

### Ship 2: Android TV MVP
- TV home shell
- D-pad focus traversal
- TV player controls
- Lean-back detail/listing flows
- Remote-friendly settings and search entry strategy

### Ship 3: Cutover Readiness
- Favorites/profile/settings parity
- Chromecast parity
- Stability/performance passes
- Internal beta rollout
- Decide whether admin/subtitle advanced tools stay in RN/web only or move later

## RN Surface Inventory (Current Reference App)

| Surface | Current Source | Native Target | Ship |
|--------|----------------|---------------|------|
| Login | `mobile/src/screens/LoginScreen.tsx` | `feature:auth` | Ship 1 |
| Home | `mobile/src/screens/HomeScreen.tsx` | `feature:home` | Ship 1 |
| Movies / Series tabs | `mobile/src/screens/MoviesScreen.tsx`, `SeriesScreen.tsx` | `feature:home`, `feature:browse`, `feature:detail` | Ship 1 |
| Browse / library traversal | `mobile/src/screens/BrowseScreen.tsx`, `LibraryBrowseScreen.tsx` | `feature:browse` | Ship 1 |
| Search | `mobile/src/screens/SearchScreen.tsx` | `feature:search` | Ship 1 |
| Media detail | `mobile/src/screens/MediaDetailScreen.tsx` | `feature:detail` | Ship 1 |
| Series detail / episodes | `mobile/src/screens/SeriesDetailScreen.tsx` | `feature:detail` | Ship 1 |
| Video player | `mobile/src/screens/VideoPlayerScreen.tsx` | `feature:player` + `core:player` | Ship 1 |
| Chromecast | `mobile/src/hooks/useChromecast.ts` | `feature:cast` | Ship 2/3 |
| Favorites | `mobile/src/screens/FavoritesScreen.tsx` | `feature:favorites` | Ship 3 |
| Settings | `mobile/src/screens/SettingsScreen.tsx` | `feature:settings` | Ship 3 |
| Profile | `mobile/src/screens/ProfileScreen.tsx` | `feature:profile` | Ship 3 |
| Admin | `mobile/src/screens/AdminScreen.tsx` | defer / web-admin | Later |
| TV behavior | `mobile/src/lib/tv.ts`, `FocusableCard.tsx` | `feature:tv` + TV-aware composables | Ship 2 |

## Phases

| Phase | Name | Status | Tasks |
|------|------|--------|-------|
| 01 | [Architecture + Scope Freeze](phase-01-architecture-scope.md) | ⬜ Pending | 10 |
| 02 | [Project Foundation + Build System](phase-02-project-foundation.md) | ⬜ Pending | 11 |
| 03 | [Auth + Network + Persistence](phase-03-auth-network-persistence.md) | ⬜ Pending | 10 |
| 04 | [Catalog Shell: Home / Browse / Detail / Search](phase-04-catalog-shell.md) | ⬜ Pending | 12 |
| 05 | [Player Core + Progress Sync](phase-05-player-core.md) | ⬜ Pending | 12 |
| 06 | [Advanced Playback + Cast](phase-06-advanced-playback-cast.md) | ⬜ Pending | 11 |
| 07 | [Android TV Experience](phase-07-android-tv-experience.md) | ⬜ Pending | 12 |
| 08 | [Settings / Profile / Feature Parity Review](phase-08-settings-profile-parity.md) | ⬜ Pending | 9 |
| 09 | [QA + Release + Android Cutover](phase-09-qa-release-cutover.md) | ⬜ Pending | 12 |

**Total:** 99 planned tasks across 9 phases

## Playback-Specific Product Rules
- Client phải khai thác đúng playback decision của backend:
  - Direct Play
  - Direct Stream / remux
  - Transcode
- Native player không được giả định rằng "không direct play" nghĩa là luôn hạ video mạnh.
- Khi chỉ audio incompatible, client phải tôn trọng response backend cho phép giữ video quality tối đa có thể.
- Client phải expose `available_qualities` để tận dụng điểm mạnh hiện có của Velox:
  - pretranscode instant playback
  - multi-quality HLS
- Nếu local direct source fail nhưng backend có HLS fallback hợp lệ, native player cần có retry path rõ ràng.

## Existing Code We Can Reuse Intelligently
- Playback logic concepts from `mobile/modules/velox-player/android/*`
- API contract knowledge from `packages/shared/api/*` and `packages/shared/types/*`
- Screen information architecture from `mobile/src/screens/*`
- Current TV behavior expectations from `mobile/src/lib/tv.ts`
- E2E coverage ideas from `mobile/maestro/*`

## Existing Code We Should Not Copy Blindly
- React Query hook layering from `packages/shared/hooks/*`
- Zustand stores as-is
- RN screen composition / giant player screen structure
- Expo-specific platform shims

## Milestones
- Milestone A: Phase 01-03 xong -> app native login được, bootstrap session ổn, chạy shell trống được
- Milestone B: Phase 04 xong -> app native browse/search/detail usable
- Milestone C: Phase 05 xong -> native playback usable hàng ngày trên Android phone/tablet
- Milestone D: Phase 06-07 xong -> playback advanced + Android TV usable bằng remote
- Milestone E: Phase 08-09 xong -> beta + cutover decision cho Android

## Key Decisions
- New Android app lives in `android-native/`, not inside Expo-generated `mobile/android/`.
- RN app remains the reference client until native reaches beta quality.
- V1 avoids Room/offline catalog sync unless a real bottleneck appears.
- `admin` features are explicitly phase-gated; they should not block playback MVP.
- TV quality matters enough to have a dedicated phase, not just responsive polish.
- Use backend contract as source of truth instead of inventing KMP/shared abstractions too early.
- Keep application ID / signing strategy flexible so internal beta can coexist with RN Android if needed.

## Risks
- Rewriting too much parity too early and delaying the first usable build.
- Porting TS hooks/stores literally instead of redesigning around repositories + flows.
- Mixing TV-specific requirements too late and having to rebuild navigation/focus.
- Migrating player logic without preserving backend-driven playback contract.
- Underestimating cast/session/resume edge cases.

## Exit Criteria
- Android phone build is stable for daily playback.
- Android TV build is navigable entirely by remote.
- Progress, subtitles, audio, quality, and cast work on real devices.
- RN Android is no longer the recommended path for internal usage.
- Deferred features are intentional, documented, and do not block everyday playback use.

## Quick Commands (for later execution)
- Start architecture work: `/code phase-01`
- Start foundation work: `/code phase-02`
- Start player work: `/code phase-05`
- Start TV work: `/code phase-07`
