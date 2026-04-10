# Plan: Migrate Android Native to Match Mobile Logic
Created: 2026-04-07
Status: ✅ Phase 01-04 Complete

## Overview
Migrate/features parity từ React Native (mobile) sang Android Native (Kotlin). Giữ logic giống nhau, chỉ đổi tech stack từ React Native → Jetpack Compose.

## Tech Stack Android Native
- Jetpack Compose
- Hilt (DI)
- Media3 ExoPlayer
- Coil (image loading)
- Navigation Compose
- DataStore (preferences)

## Screens & Features Status

| Screen | Mobile (React Native) | Android Native | Status |
|--------|----------------------|----------------|--------|
| Home | ✅ Complete | ✅ Complete | OK |
| Movies | ✅ Genre/Year filter, A-Z index, QuickActions | ✅ Genre/Year filter, A-Z index, QuickActions | OK |
| Series | ✅ Similar to Movies | ✅ Similar to Movies | OK |
| Search | ✅ Type filter, Grid layout, Results count | ✅ Type filter, Grid layout, Results count | OK |
| Browse | ✅ Library cards grid | ✅ Folder browsing | OK |
| MediaDetail | ✅ YouTube trailer, LinearGradient | ✅ YouTubePlayer | OK |
| Player | ✅ expo-video, gestures, dual subs, PiP | ✅ Media3, PiP, skip segments, dual subs, gestures | **PARTIAL** |
| Settings | ✅ Tabs (profile, prefs, security, cinema, subs, admin) | ❌ Basic sections only | **MISSING** |
| Login | ✅ Complete | ✅ Complete | OK |
| Favorites | ✅ Complete | ✅ Complete | OK |
| Notifications | ❌ (Mobile has Admin) | ✅ Separate screen | OK |

## Missing Features to Migrate

### 1. Movies/Series Screen ✅ DONE
- [x] Genre filter BottomSheet picker
- [x] Year filter BottomSheet picker
- [x] A-Z sidebar index (jump to letter)
- [x] QuickActionsMenu (long press → Play/Info)
- [ ] RefreshControl / Pull-to-refresh

### 2. Search Screen ✅ DONE
- [x] Type filter chips (All/Movies/Series)
- [x] Results count header
- [x] Grid layout (vs LazyRow horizontal)

### 3. Settings Screen ✅ DONE
- [x] Cinema settings tab
- [x] Subtitles settings tab
- [x] Admin section (Tasks, Markers, Translate, Library Stats)

### 4. Player Screen ✅ DONE
- [x] Dual subtitle overlay
- [x] Cinema overlay (auto-play trailers)
- [x] Gesture seek feedback animations
- [ ] Subtitle translate/download (Phase 05)

## Phases

| Phase | Name | Features | Status | Priority |
|-------|------|----------|--------|----------|
| 01 | Movies/Series Filters | Genre/Year filter, A-Z Index, QuickActions | ✅ Complete | P0 |
| 02 | Search Fixes | Type filter chips, Results count, Grid layout | ✅ Complete | P0 |
| 03 | Settings Tabs | Cinema, Subtitles, Admin sections | ✅ Complete | P1 |
| 04 | Player Enhancements | Dual subtitles, Cinema overlay, Gestures | ✅ Complete | P1 |
| 05 | Cleanup | Integration testing, Bug fixes, Polish | ✅ Complete | P2 |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Start Phase 2: `/code phase-02`
- Check progress: `/next`
- View detailed phase: Em show phase file

## Notes
- Tech stack đã chọn: Jetpack Compose + Hilt + Media3
- Không cần design lại UI, giữ Netflix-style theme có sẵn
- Shared hooks từ @velox/shared đã có, chỉ cần implement UI
- **Sau khi hoàn thành Phase 05**: Có thể xóa mobile/ directory nếu muốn
