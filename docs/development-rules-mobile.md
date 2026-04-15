# Velox Mobile Development Rules (Android / Kotlin)

> Living document — updated as patterns evolve. For project overview and build commands, see [CLAUDE.md](../CLAUDE.md). See also: [backend rules](development-rules-backend.md), [webapp rules](development-rules-webapp.md).

> Stack: **Kotlin + Jetpack Compose + Dagger Hilt + Retrofit + Media3 ExoPlayer + Coroutines/Flow**. Clean Architecture (data / domain / presentation). No React Native, no XML layouts.

---

## General Rules

### File Size
- **Composable file:** max ~300 lines. Break into smaller Composables when larger.
- **ViewModel:** max ~300 lines. Split by feature/use-case when larger.
- **Repository / UseCase:** max ~250 lines.
- **Exception:** generated code (Hilt, Room).

### Naming
- Code comments and variable names in **English**.
- Plan/spec files may contain **Vietnamese**.
- Commit messages in **English**, prefixed: `Add(scope)`, `Fix(scope)`, `Enhance(scope)`, `Refactor(scope)`, `Chore:`.

### No Premature Abstractions
- Don't wrap one-call Retrofit services behind extra interfaces "just in case".
- Three similar lines > one premature abstraction.
- Only abstract when the pattern repeats 3+ times.

---

## Project Structure

```
android/app/src/main/java/com/velox/app/
├── VeloxApp.kt                    # @HiltAndroidApp Application
├── MainActivity.kt                # Single activity, hosts NavHost
├── data/
│   ├── api/                       # Retrofit services, AuthManager, TokenAuthenticator
│   ├── local/                     # DataStore / Room DAOs
│   ├── model/                     # DTOs (API wire format)
│   ├── repository/                # Repository impls (data sources → domain)
│   └── util/                      # Mappers, network helpers
├── domain/
│   ├── model/                     # Pure Kotlin domain models (no Android deps)
│   ├── repository/                # Repository interfaces
│   └── usecase/                   # Business logic, one class per operation
├── presentation/
│   ├── navigation/                # NavGraph, routes, typed args
│   ├── viewmodel/                 # @HiltViewModel classes
│   ├── ui/
│   │   ├── screens/               # One subfolder per feature (home/, detail/, player/, ...)
│   │   └── components/            # Shared Composables (cards, buttons, dialogs)
│   └── cast/                      # Chromecast session management
├── di/                            # Hilt modules (NetworkModule, RepositoryModule, ...)
├── ui/theme/                      # Color.kt, Theme.kt, Type.kt, Shape.kt
└── utils/                         # Extension functions, helpers
```

**Feature folder pattern** (for a screen with multiple Composables):
```
presentation/ui/screens/detail/
├── MediaDetailScreen.kt           # @Composable entry, state collection, navigation
├── components/
│   ├── DetailBackdrop.kt
│   ├── DetailActions.kt
│   ├── CastRow.kt
│   └── EpisodeList.kt
└── MediaDetailUiState.kt          # sealed interface for UI state
```

---

## Layer Architecture (Clean Architecture + MVVM)

```
┌──────────────────────────────────────────────────────┐
│  Presentation (Compose UI + ViewModel)               │
│    ↓ collects StateFlow                              │
│  Domain (UseCase + Repository interfaces + Model)    │
│    ↓ suspend functions                               │
│  Data (Repository impl + Retrofit + DataStore)       │
└──────────────────────────────────────────────────────┘
```

| Layer | Responsibility | Allowed Dependencies |
|-------|---------------|----------------------|
| **UI / Composable** | Render state, emit events upward | `domain/model`, `ui/theme`, other Composables. **No** ViewModels imported by nested Composables — pass state + lambdas as params. |
| **ViewModel** | Hold UI state (`StateFlow`), invoke UseCases, transform domain → UI model | `domain/usecase`, `domain/model`, `javax.inject`, `kotlinx.coroutines` |
| **UseCase** | One business operation (`operator fun invoke`) | `domain/repository`, `domain/model` |
| **Repository (interface)** | Domain-facing contract | `domain/model` only |
| **Repository (impl)** | Coordinate remote + local sources, map DTO → domain | `data/api`, `data/local`, `data/model`, `domain/*` |
| **DTO** (`data/model`) | API wire format, `@Serializable` or Moshi annotated | `kotlinx.serialization` / `moshi` |
| **Domain Model** | Pure Kotlin data class | Nothing Android-specific. No `@Parcelize` unless needed for nav args — use a separate nav model then. |

**Hard rules:**
- `domain/` **MUST NOT** import anything from `android.*`, `androidx.*`, Retrofit, Room, or `data/`.
- Composables **MUST NOT** inject ViewModels directly via Hilt except at the **screen root**. Nested Composables receive state + event lambdas as parameters (stateless, testable, previewable).
- Repository implementations return domain models, not DTOs. Map at the repository boundary.

---

## Compose UI Rules

### Composable Design
- **Stateless by default.** State hoisting: child takes `state: T` + `onEvent: (Event) -> Unit`, parent owns the state.
- **Screen root is stateful**, injects ViewModel via `hiltViewModel()`, collects state with `collectAsStateWithLifecycle()`.
- **Previews mandatory** for components > 50 lines or with multiple states. Use `@PreviewParameter` for state variations.
- **One Composable per responsibility.** If a Composable has > 3 parameters OR > 80 lines, extract subcomposables.
- **`@Stable` / `@Immutable` annotations** on data classes used as Composable parameters to help skipping.

### State Management
- **`StateFlow<UiState>`** as the single source of UI state per screen. Avoid multiple flows for one screen.
- **Sealed interface** for UI state: `Loading`, `Success(data)`, `Error(throwable)`, `Empty`.
- **One-shot events** (navigation, snackbars) via `Channel<Event>` or `SharedFlow` — **never** in `StateFlow` (causes replay on config change).
- **`remember { ... }`** for derived values within a Composable. **`rememberSaveable`** for state that must survive process death (text fields, scroll position).
- **`LaunchedEffect(key)`** for side effects. `DisposableEffect` for resources needing cleanup.

### Recomposition Performance
- Pass lambdas as `remember`-ed references when they capture unstable types; otherwise let the compiler handle it.
- Use `key()` composable around list items that change identity.
- `LazyColumn` / `LazyRow` with stable `key` parameter: `items(list, key = { it.id })`.
- Avoid reading `state.value` in Composables — use `state` delegate (`by state`) for proper recomposition scoping.
- **Never** create `Modifier` chains inside loops — hoist to constants or remember.

### Navigation (Compose Navigation)
- Single `NavHost` in `MainActivity`. Routes defined as **typed sealed class** (Kotlin 2.0 type-safe navigation preferred).
- Pass IDs only via nav args — look up full domain model via ViewModel + UseCase.
- Back stack hygiene: use `popUpTo` with `inclusive` for login → home (clears auth stack).

```kotlin
@Serializable data class MediaDetail(val mediaId: Long, val mediaType: String)

composable<MediaDetail> { backStackEntry ->
    val args = backStackEntry.toRoute<MediaDetail>()
    MediaDetailScreen(args.mediaId, args.mediaType)
}
```

---

## Theming

- **All tokens** live in `ui/theme/`. Never hardcode colors, dimensions, or typography inside Composables.
- `MaterialTheme.colorScheme.*` for colors, `MaterialTheme.typography.*` for text styles, `MaterialTheme.shapes.*` for shapes.
- Custom tokens (outside Material spec) go in a `CompositionLocal`:

```kotlin
// ui/theme/Color.kt — Netflix-inspired palette
val NetflixRed = Color(0xFFE50914)
val SurfaceDark = Color(0xFF1F1F1F)
val BackgroundDeep = Color(0xFF0A0A0A)

// ui/theme/Theme.kt
@Composable fun VeloxTheme(content: @Composable () -> Unit) {
    MaterialTheme(colorScheme = darkColorScheme(...), typography = VeloxTypography, content = content)
}
```

- Dark theme is default and primary. Light theme optional; if added, verify all screens.
- Dimension scale: use multiples of `4.dp` (4, 8, 12, 16, 20, 24, 32). Declare common ones in `Dimens.kt`.

---

## Dependency Injection (Hilt)

- **`@HiltAndroidApp`** on `VeloxApp`. **`@AndroidEntryPoint`** on `MainActivity`.
- **`@HiltViewModel`** for all ViewModels. Inject via `hiltViewModel()` only at screen root.
- Modules organized by concern in `di/`:
  - `NetworkModule` — Retrofit, OkHttp, interceptors, Moshi/kotlinx.serialization.
  - `RepositoryModule` — `@Binds` repository interface → impl.
  - `DataStoreModule` — DataStore instances.
  - Feature modules when a domain grows its own DI surface.
- **Prefer `@Binds` over `@Provides`** for interface→impl mapping (cheaper to generate).
- **Scoping:** `@Singleton` for repositories, API clients, DataStore. `@ViewModelScoped` for per-screen coordinators.
- **Never** create Hilt qualifiers for a single binding — use them only when ≥ 2 impls of the same type exist.

---

## Networking & Auth

- **Retrofit + OkHttp + kotlinx.serialization** (or Moshi if legacy). Configured in `NetworkModule`.
- **Base URL** resolved at runtime from `ServerPrefsManager` (user-entered). Never hardcode `http://...` constants.
- **Auth flow:**
  - `AuthManager` holds tokens in encrypted DataStore (not SharedPreferences).
  - `TokenAuthenticator` (OkHttp `Authenticator`) handles 401 → refresh → retry. Single-flight refresh via `Mutex`.
  - Access token: 15min. Refresh token: 7d. Both stored encrypted.
- **Stream URLs** use Jellyfin-style `api_key` (32-char hex, 2h): obtain via `POST /api/stream/{id}/url`, append as query param. Required for ExoPlayer HTTP data source (cannot attach headers per segment reliably).
- **Cleartext LAN traffic:** declare `android:usesCleartextTraffic="true"` in `AndroidManifest.xml` OR a `network_security_config.xml` scoped to LAN IPs (preferred).
- **Error mapping:** `Response<T>` in repository layer → map to sealed `Result<Success, Error>` domain type. Never leak `HttpException` / `IOException` to ViewModels.

---

## Coroutines & Flow

- **Structured concurrency:** always launch from a scope (`viewModelScope`, `lifecycleScope`). No `GlobalScope`.
- **Dispatchers:** inject via constructor (`@Dispatcher(IO) dispatcher: CoroutineDispatcher`). Never hardcode `Dispatchers.IO` in repositories — blocks testing.
- **Repository functions are `suspend`**, marked `withContext(ioDispatcher)` at the boundary.
- **ViewModel exposes `StateFlow`**, not `LiveData`. Collect in UI with `collectAsStateWithLifecycle()` (lifecycle-aware, stops in background).
- **`SharedFlow` for events**, `StateFlow` for state. Don't mix.
- **`flow { emit(...) }.stateIn(scope, SharingStarted.WhileSubscribed(5000), initial)`** — standard pattern for upstream → UI state.
- **Cancellation-safe:** check `ensureActive()` before heavy work inside long-running suspend functions.

---

## Video Player (Media3 / ExoPlayer)

- **Wrap ExoPlayer in a Service** (`MediaSessionService` or foreground service) — don't keep `ExoPlayer` inside a Composable or ViewModel. Player survives config changes and supports MediaSession (lock screen, Auto, Cast).
- **Single `Player` instance per app.** ViewModels observe via a repository/holder; Composables get a stable `Player` reference to render with `PlayerView` (via `AndroidView`).
- **Resume position:** read `playbackInfo.position` from backend (single source of truth per Plan S). Seek on prepare.
- **HLS vs Direct Play:** detect from `playbackInfo.direct_url` vs `hls_url`. Prefer direct when browser/device compatible.
- **Track selection:** Media3 `TrackSelectionParameters` — set preferred audio language, subtitle language, max height.
- **Subtitles:** server burns in when `subtitle_burned_in` flag set; otherwise pass sideloaded SRT/VTT tracks.
- **Progress reporting:** coroutine in service polls `player.currentPosition` every 10s → `POST /api/playback/{id}/progress { position, completed }`. Pause reporting when buffering or paused.
- **Keep screen on:** `PlayerView.keepScreenOn = player.isPlaying` observed via listener. Don't `FLAG_KEEP_SCREEN_ON` on whole activity.
- **PiP:** declare `android:supportsPictureInPicture="true"`; handle `onPictureInPictureModeChanged` to hide controls.

---

## Local Storage

- **DataStore Preferences** for simple key/value (settings, flags). **Proto DataStore** for typed structured data.
- **EncryptedSharedPreferences** (or encrypted DataStore) for tokens, server URL, API keys.
- **Room** only when there's meaningful offline data (watchlist cache, continue-watching mirror). Don't use Room for one-off prefs.
- **Never** use raw `SharedPreferences`.

---

## Error Handling

- **UI-facing:** sealed `UiError` → map to user-friendly Vietnamese strings in `strings.xml` (+ `values-vi/`).
- **Snackbars** for transient errors, **Dialog** for blocking errors that need user action.
- **Repository returns `Result<T>`** (Kotlin stdlib) or a sealed `DataResult<T>` — never throws to ViewModel.
- **Timber** for logging in debug builds. Release builds: strip via ProGuard or use a no-op tree. **No `Log.d`** direct calls.
- **Crashlytics** (optional) for release-build crash reporting.

---

## Accessibility

- **`contentDescription`** on every `Icon`, `Image`, `IconButton` — use `null` only for purely decorative elements next to labeled text.
- **Touch targets ≥ 48dp.** Use `Modifier.minimumInteractiveComponentSize()` or explicit `size(48.dp)`.
- **Semantics** for custom interactive Composables: `Modifier.semantics { role = Role.Button; contentDescription = "..." }`.
- **Font scaling:** test at 130% system font size. Avoid fixed `height` on text containers — use `heightIn(min = ...)`.
- **TalkBack** test login, home, detail, player before merging a feature.

---

## Testing

- **Unit tests** (JUnit 4 + MockK + Turbine): UseCases, ViewModels, mappers. Place in `test/`.
- **Instrumented UI tests** (Compose `createAndroidComposeRule`): critical flows only. Place in `androidTest/`.
- **Fakes over mocks** for repositories — write a `FakeMediaRepository` in `test/` and reuse across ViewModel tests.
- **Flow testing:** `Turbine` — `repository.flow.test { awaitItem() }`.
- **Run:** `./gradlew test` (unit), `./gradlew connectedAndroidTest` (instrumented).

---

## Build & Run

```sh
cd android
./gradlew build              # Compile + lint + unit tests
./gradlew assembleDebug      # Build debug APK
./gradlew installDebug       # Install debug APK on connected device/emulator
./gradlew test               # Unit tests
./gradlew lint               # Android Lint
./gradlew ktlintCheck        # Kotlin style check (if ktlint configured)
./gradlew ktlintFormat       # Auto-format Kotlin files
```

**Verification before commit:**
```sh
cd android && ./gradlew ktlintCheck lint testDebug
```

**Requirements:** JDK 17, Android SDK 34+, Android Studio Ladybug+ (or newer), Gradle 8.7+.

---

## Git & CI

### Commit Convention
```
Add(scope): new feature
Fix(scope): bug fix
Enhance(scope): improve existing feature
Refactor(scope): structural change, no behavior change
Chore: tooling, deps, config
```

### Pre-commit Hooks (Husky + lint-staged)
- `.kt` → ktlint auto-format (if configured at repo root)
- Runs automatically on `git commit`

---

## Anti-Patterns

| ❌ Don't | ✅ Do |
|----------|-------|
| XML layouts for new screens | Jetpack Compose |
| Import Android types in `domain/` | Pure Kotlin domain models |
| `hiltViewModel()` inside nested Composables | Inject only at screen root, hoist state down |
| `GlobalScope.launch { ... }` | `viewModelScope` / `lifecycleScope` |
| `Dispatchers.IO` hardcoded in repository | Inject dispatcher via constructor |
| `LiveData` for new code | `StateFlow` + `collectAsStateWithLifecycle` |
| Events in `StateFlow` (replay on rotate) | `SharedFlow` or `Channel` for one-shot events |
| `Log.d(...)` in production | Timber, stripped in release |
| Raw `SharedPreferences` | DataStore (Proto or Preferences) |
| `SharedPreferences` for tokens | EncryptedSharedPreferences / encrypted DataStore |
| Return DTOs from repository | Map DTO → domain model at repository boundary |
| Throwing exceptions to ViewModel | Sealed `Result` / `DataResult` |
| Hardcoded colors / dp in Composables | `MaterialTheme` tokens + `Dimens.kt` |
| Stateful Composables everywhere | Hoist state — stateless children, stateful root |
| ExoPlayer inside Composable / ViewModel | Wrap in Service, observe via holder |
| `FLAG_KEEP_SCREEN_ON` on Activity | `PlayerView.keepScreenOn` toggled on play/pause |
| Hardcoded base URL | `ServerPrefsManager` runtime config |
| 500+ line Composable | Extract subcomposables, preview each |
| `@Composable` with 10+ params | Group params into a state data class |
| Missing `contentDescription` on icons | Always label or explicit `null` for decorative |
| Skipping `key` in `LazyColumn` items | `items(list, key = { it.id })` |
