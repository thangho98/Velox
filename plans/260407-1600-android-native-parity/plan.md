# Plan: Android Native - Mobile Feature Parity
Created: 2026-04-07
Status: ✅ Complete
Parent: plans/260407-1200-migrate-android-native/

## Overview
Migrate all missing features from React Native mobile app to Android Native (Jetpack Compose) to achieve full feature parity.

## Missing Features

### MediaDetailScreen
- [ ] Poster (centered with shadow)
- [ ] Tech specs (Video · Audio · Container · Size)
- [ ] "Ends at" time calculation
- [ ] Watched/Check button
- [ ] Inline subtitle selector
- [ ] Tablet side-by-side layout

### SeriesDetailScreen
- [ ] Status badge, Network info
- [ ] Edit badge, Lock badge
- [ ] "Read more" overview toggle
- [ ] Episode duration
- [ ] EpisodeEditDialog
- [ ] Episode metadata editing

## Tech Stack
- Android: Kotlin + Jetpack Compose + Hilt + Coil
- Backend: Already implemented (reuse existing API)

## Phases

| Phase | Name | Status | Progress |
|-------|------|--------|----------|
| 01 | MediaDetailScreen Parity | ✅ Complete | 100% |
| 02 | SeriesDetailScreen Parity | ✅ Complete | 100% |
| 03 | Build & Verify | ✅ Complete | 100% |

## Quick Commands
- Start Phase 1: `/code phase-01`
- Check progress: `/next`

## Notes
- Follow mobile app logic exactly ("migrate sao logic vẫn giống nhau là được")
- Reuse existing domain models where possible
- Test on both phone and tablet layouts
