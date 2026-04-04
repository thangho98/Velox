# Phase 01: Architecture + Scope Freeze
Status: ⬜ Pending
Dependencies: none

## Objective
Chốt scope V1/V1.5, repo layout, module boundaries, API contracts, và feature cut-lines trước khi dựng codebase native.

## Deliverables
- `android-native/` direction approved
- feature matrix Android phone vs Android TV
- API contract checklist frozen
- list parity features: now / later / never
- technical ADR cho stack native
- planning artifacts committed:
  - `DESIGN.md`
  - `parity-matrix.md`
  - `api-contract-checklist.md`

## Current References
- `mobile/App.tsx` — current route graph and screen inventory
- `mobile/src/screens/*` — current surface area
- `mobile/src/screens/VideoPlayerScreen.tsx` — highest-risk port
- `mobile/modules/velox-player/android/*` — current native playback logic
- `packages/shared/hooks/*` and `packages/shared/types/*` — contract source of truth

## Implementation Steps

### 1. Freeze Product Scope
- [ ] Review all current mobile surfaces and classify them:
  - must-have for first beta
  - important but can land after playback MVP
  - admin-only / defer to web
- [ ] Confirm that Android and Android TV share one codebase, with TV-specific flows where needed.
- [ ] Confirm iOS is out of scope for this plan.

### 2. Freeze Contract Scope
- [ ] Enumerate auth/session/profile endpoints from `packages/shared/hooks/auth.ts`.
- [ ] Enumerate browse/detail/search endpoints from current media/series hooks.
- [ ] Enumerate progress/favorites/playback endpoints.
- [ ] Decide which admin settings remain web-only.

### 3. Freeze Technical Direction
- [ ] Confirm Compose, Hilt, Retrofit, OkHttp, Kotlinx Serialization, DataStore, Media3.
- [ ] Reject early KMP unless there is a proven near-term need.
- [ ] Confirm repository + ViewModel + StateFlow pattern.

### 4. Freeze TV Strategy
- [ ] Decide whether TV has a dedicated nav graph.
- [ ] Decide focus-management strategy and restoration requirements.
- [ ] Decide minimum acceptable TV search experience for beta.

### 5. Freeze Cut Lines
- [ ] Write explicit “ship now / defer / web-admin only” decisions.
- [ ] Ensure admin tools cannot silently expand scope.

## Tasks
1. [ ] Audit RN app screens/components/features và map chúng vào buckets: `must-have`, `tv-only`, `later`.
2. [ ] Freeze V1 scope:
   - Login
   - Home
   - Browse
   - Media detail / series detail
   - Search
   - Player
   - Progress
   - Subtitle/audio/quality
3. [ ] Freeze V1.5 scope:
   - Favorites
   - Profile
   - Settings
   - Chromecast
4. [ ] Decide explicit out-of-scope for first beta:
   - metadata editor
   - subtitle translate/search UI if needed
   - deep admin operations
5. [ ] Convert backend contract references from TS into a Kotlin model checklist:
   - auth payloads
   - media DTOs
   - playback DTOs
   - progress/favorites/settings DTOs
6. [ ] Decide module strategy for `android-native/`:
   - `app`
   - `core/*`
   - `feature/*`
7. [ ] Decide state pattern:
   - repository + use-case style where needed
   - `ViewModel + StateFlow`
   - one immutable `UiState` per screen
8. [ ] Decide TV strategy:
   - same app with runtime TV mode
   - dedicated TV composables for focus-heavy screens
9. [ ] Define code migration rules:
   - RN app is reference only
   - no destructive deletes in `mobile/` during early phases
10. [ ] Write a short ADR summary inside the plan folder before Phase 02 starts.

## Files to Produce in This Phase
- [ ] `DESIGN.md`
- [ ] `parity-matrix.md`
- [ ] `api-contract-checklist.md`
- [ ] optional ADR notes inside plan folder if more decisions are needed

## Done When
- [ ] Native target architecture is frozen enough that Phase 02 can scaffold without churn.
- [ ] Every “should this feature be in beta?” question has an explicit answer.
- [ ] API scope is concrete enough to start porting DTOs and repositories.

## Verification
- [ ] Team can answer “what is in V1?” without ambiguity.
- [ ] Team can answer “what will stay in RN/web temporarily?” without ambiguity.
- [ ] All required backend endpoints for Ship 1/2 are enumerated.
- [ ] Module boundaries are concrete enough to start scaffolding.

## Exit Criteria
- No major scope questions remain open for foundation work.
- Phase 02 can start without re-litigating architecture.
