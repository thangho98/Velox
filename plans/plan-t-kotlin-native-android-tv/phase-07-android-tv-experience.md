# Phase 07: Android TV Experience
Status: ⬜ Pending
Dependencies: Phase 04, Phase 05

## Objective
Adapt the native Android app into a true Android TV client instead of a phone UI stretched onto a large screen.

## Design Principles
- Every critical path is remote-first.
- Focus state is always visible.
- Navigation depth is shallow and predictable.
- Text entry/search is explicitly considered for TV.

## Current Reference Inputs
- `mobile/src/lib/tv.ts`
- `mobile/src/components/FocusableCard.tsx`
- current browse/detail/player information architecture in RN app

## Implementation Steps

### 1. Detect TV Mode
- [ ] runtime detection
- [ ] dedicated TV nav shell selection
- [ ] TV-specific theme tweaks if needed

### 2. Build TV Home / Browse
- [ ] hero row
- [ ] horizontal rows
- [ ] focus-aware poster cards
- [ ] clear active section indication

### 3. Build TV Detail Screens
- [ ] large-screen hero
- [ ] focused CTA row
- [ ] season/episode traversal with remote

### 4. Build TV Player Controls
- [ ] center to play/pause
- [ ] left/right seek
- [ ] up/down to reveal or move between control layers
- [ ] visible focus ring on each actionable control

### 5. Focus Restoration
- [ ] returning from detail restores invoking card
- [ ] returning from player restores play CTA or originating row
- [ ] modal exit restores prior focus

### 6. TV Search
- [ ] basic usable keyboard/IME flow
- [ ] explicit loading/results states tuned for TV
- [ ] no mobile-text-field UX copied blindly

## Tasks
1. [ ] Detect TV runtime and route into TV-optimized navigation shell.
2. [ ] Build TV home layout:
   - left rail or top tabs
   - hero/featured row
   - horizontal content rows
3. [ ] Build TV media cards with:
   - visible focus ring
   - scale/elevation feedback
   - safe D-pad traversal
4. [ ] Build TV browse and detail screens using larger spacing and reduced clutter.
5. [ ] Build TV player controls:
   - DPAD left/right seek
   - center to pause/play
   - up/down or menu to open control layers
6. [ ] Implement focus restoration when returning from detail/player to list.
7. [ ] Add TV-friendly search strategy:
   - basic text entry
   - optional keyboard-first fallback
   - avoid shipping unusable phone search on TV
8. [ ] Review subtitle, quality, audio sheets for TV usability.
9. [ ] Review cast entry points on TV and hide flows that do not make sense.
10. [ ] Test on real Android TV / Google TV hardware, not emulator only.
11. [ ] Tune performance and image loading for 10-foot browsing.
12. [ ] Add Compose UI tests or scripted manual test matrix for remote navigation.

## Files / Modules to Create
- [ ] `feature:tv/.../TvRootGraph.kt`
- [ ] `feature:tv/.../TvHomeScreen.kt`
- [ ] `feature:tv/.../TvBrowseScreen.kt`
- [ ] `feature:tv/.../TvMediaDetailScreen.kt`
- [ ] `feature:tv/.../TvSeriesDetailScreen.kt`
- [ ] TV card/focus components in `feature:tv` or `core:designsystem`

## Test Criteria
- [ ] app is fully usable without touch
- [ ] focus never disappears
- [ ] back stack is predictable
- [ ] long lists remain performant on TV hardware
- [ ] player controls are operable from a real remote

## Done When
- [ ] Android TV build feels intentional, not just phone UI scaled up.

## Verification
- [ ] App is fully usable without touch.
- [ ] Focus never gets lost in home, browse, detail, or player flows.
- [ ] Player controls are comfortable from a remote.
- [ ] Back navigation is predictable across screens.

## Exit Criteria
- Android TV build feels intentionally native, not merely “works”.
