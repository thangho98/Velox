# Phase 08: Settings / Profile / Feature Parity Review
Status: ⬜ Pending
Dependencies: Phase 03, Phase 04, Phase 06, Phase 07

## Objective
Fill the most important everyday gaps after the browsing and playback spine is stable, and explicitly decide what stays out of the native client for now.

## Scope
- profile
- settings
- favorites
- parity review for admin and advanced subtitle tools

## Important Scope Rule
This phase is for finishing the everyday user experience, not for reopening admin scope or overloading the native app with power-user tooling.

## Implementation Steps

### 1. Favorites
- [ ] favorites screen
- [ ] favorite toggles from detail/listing contexts
- [ ] empty/error/loading states

### 2. Profile
- [ ] user info
- [ ] logout
- [ ] optional device/session summary if low effort

### 3. Settings
- [ ] playback defaults
- [ ] subtitle defaults
- [ ] quality defaults
- [ ] any user-facing server/app preferences that belong locally

### 4. Local Preference Persistence
- [ ] DataStore-backed player preferences
- [ ] restore on app startup
- [ ] ensure TV settings UI remains usable by remote

### 5. Parity Review
- [ ] explicitly decide what stays web-admin only
- [ ] explicitly decide which subtitle advanced flows are deferred
- [ ] update `parity-matrix.md` with final status notes

## Tasks
1. [ ] Implement favorites list + toggle actions if not already completed in earlier phases.
2. [ ] Implement settings screens for:
   - server URL/env if needed
   - playback preferences
   - subtitle defaults
   - quality defaults
3. [ ] Implement profile/account screen:
   - user info
   - logout
   - session/device info if useful
4. [ ] Persist player preferences in DataStore.
5. [ ] Review whether notification/admin surfaces belong in native V1.
6. [ ] Review whether metadata editor belongs in native or should remain web-admin only.
7. [ ] Review subtitle translate/search parity and decide:
   - ship now
   - defer
   - keep web/mobile legacy only
8. [ ] Add phone and TV UX passes for settings/profile flows.
9. [ ] Update the parity matrix with final “ship / defer / drop” decisions.

## Files / Modules to Create
- [ ] `feature:favorites/...`
- [ ] `feature:profile/...`
- [ ] `feature:settings/...`
- [ ] DataStore preference models in `core:datastore`

## Test Criteria
- [ ] favorite toggles are reflected across app surfaces
- [ ] settings survive app restart
- [ ] logout clears all relevant local state
- [ ] profile/settings are usable on both phone and TV where exposed

## Done When
- [ ] Native client covers the core day-to-day user flows without requiring RN fallback.

## Verification
- [ ] User can manage core preferences without touching RN app.
- [ ] Favorites and profile flows work on real backend data.
- [ ] Deferred features are clearly documented, not accidental omissions.

## Exit Criteria
- Native app covers the day-to-day user-facing feature set needed for beta rollout.
