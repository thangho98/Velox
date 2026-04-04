# Phase 02: Project Foundation + Build System
Status: ⬜ Pending
Dependencies: Phase 01

## Objective
Dựng project Android native mới, compile được từ đầu, có CI/lint/test baseline, DI/navigation/theme skeleton, và module structure đúng cho scale-up.

## Proposed Stack
- Android Gradle Plugin latest stable compatible with Kotlin target
- Kotlin 2.x
- Compose BOM
- Hilt
- Navigation Compose
- Coil
- Detekt + Ktlint
- JUnit + Turbine + MockWebServer

## Build Decisions to Lock
- App lives in `android-native/`
- Gradle Kotlin DSL everywhere
- version catalog for dependency ownership
- separate beta/release configuration possible
- package/app ID strategy allows internal coexistence with RN build if needed

## Implementation Steps

### 1. Create Gradle Root
- [ ] Add `settings.gradle.kts`
- [ ] Add root `build.gradle.kts`
- [ ] Add `gradle/libs.versions.toml`
- [ ] Add wrapper and shared JVM/toolchain settings

### 2. Create Modules
- [ ] `app`
- [ ] `core:common`
- [ ] `core:model`
- [ ] `core:network`
- [ ] `core:datastore`
- [ ] `core:designsystem`
- [ ] `core:player`
- [ ] `feature:auth`
- [ ] `feature:home`
- [ ] `feature:browse`
- [ ] `feature:detail`
- [ ] `feature:search`
- [ ] `feature:player`
- [ ] `feature:favorites`
- [ ] `feature:settings`
- [ ] `feature:profile`
- [ ] `feature:cast`
- [ ] `feature:tv`

### 3. Create Baseline App Shell
- [ ] `VeloxApplication` with Hilt
- [ ] `MainActivity`
- [ ] `VeloxApp()` root composable
- [ ] placeholder nav graph for auth/main/player routes

### 4. Create Design System Baseline
- [ ] colors
- [ ] typography
- [ ] spacing
- [ ] focus/focused-card tokens for future TV use
- [ ] loading and empty-state placeholders

### 5. Add Quality Gates
- [ ] Detekt
- [ ] Ktlint
- [ ] unit test config
- [ ] Compose UI test config
- [ ] basic CI command list in docs

## Tasks
1. [ ] Create `android-native/` Gradle root with Kotlin DSL and version catalog.
2. [ ] Add modules:
   - `app`
   - `core:model`
   - `core:common`
   - `core:network`
   - `core:datastore`
   - `core:designsystem`
   - `core:player`
   - `feature:auth`
   - `feature:home`
   - `feature:browse`
   - `feature:detail`
   - `feature:search`
   - `feature:player`
   - `feature:settings`
   - `feature:profile`
   - `feature:favorites`
   - `feature:cast`
   - `feature:tv`
3. [ ] Configure Hilt across modules and app entrypoint.
4. [ ] Configure Compose theme with Velox branding, dark-first but not TV-hostile contrast.
5. [ ] Create a root `VeloxApp()` composable with nav graph placeholder screens.
6. [ ] Add build variants:
   - `debug`
   - `release`
   - optional `staging`
7. [ ] Add runtime config for base API URL and environment handling.
8. [ ] Add baseline tooling:
   - Detekt
   - Ktlint
   - unit test dependencies
   - Compose UI test setup
9. [ ] Add a sample CI command set:
   - `./gradlew assembleDebug`
   - `./gradlew testDebugUnitTest`
   - `./gradlew lint`
10. [ ] Add a minimal home placeholder so app launches through real navigation.
11. [ ] Document module dependency rules to avoid feature cycles.

## Suggested Files / Directories
- [ ] `android-native/settings.gradle.kts`
- [ ] `android-native/build.gradle.kts`
- [ ] `android-native/gradle/libs.versions.toml`
- [ ] `android-native/app/src/main/java/.../VeloxApplication.kt`
- [ ] `android-native/app/src/main/java/.../MainActivity.kt`
- [ ] `android-native/app/src/main/java/.../VeloxApp.kt`
- [ ] baseline source sets for each `core/*` and `feature/*` module

## Done When
- [ ] `./gradlew assembleDebug` passes from `android-native/`
- [ ] app boots to a native shell screen
- [ ] DI is wired
- [ ] code quality tools run locally

## Verification
- [ ] `./gradlew assembleDebug` passes.
- [ ] App launches on emulator and shows native Compose shell.
- [ ] Hilt injection works in at least one sample screen.
- [ ] Detekt/Ktlint run successfully.

## Exit Criteria
- Native app can compile/run independently of Expo/RN.
- All later phases have a clean module base to land in.
