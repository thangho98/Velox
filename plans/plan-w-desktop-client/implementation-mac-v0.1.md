# Plan W — Desktop Client v0.1 (Mac-only)

**Goal:** Ship a Mac desktop app that has full feature parity with the webapp + libmpv-based playback (HDR/DV correct, max codec coverage, GPU HW decode).

**Scope:** Mac-only. Linux + Windows deferred to Plan W.2.

**Decisions locked:**
- Build Mac first, optimize, then Linux + Windows
- Reuse webapp UI 100% (no custom desktop chrome)
- Custom onboarding: enter server URL + login in 1 step
- Live TV / Channels included in v0.1 (full parity)
- Repo: monorepo `Velox/desktop/` (sibling of `webapp/`, `android/`)
- Dual-subtitle: shipped in v0.1 (not deferred)

**Stack:** Tauri 2 + libmpv2 (custom build w/ libdovi + libplacebo) + Vite/React webapp embedded
**PoC validated:** [poc-01..07] + NSOpenGLView embed (working in `~/Desktop/source/mpv-poc/velox-desktop/`)

---

## Phase 1 — Infrastructure (1 week)

**Goal:** Webapp loads inside Tauri window, user can log into a NAS server.

### Tasks
1. **Repo decision** ([open question](#open-questions)) — relocate PoC into Velox monorepo as `desktop/` (sibling of `webapp/`, `android/`).
2. **Webapp build pipeline:** `npm run build` in `webapp/` outputs to `desktop/dist/`. Tauri config points `frontendDist` to that path.
3. **Dev mode:** Tauri devUrl → Vite dev server on `:3000`. Vite proxy `/api` rewritten at runtime to `http://NAS_IP:8080` (configurable).
4. **DesktopAdapter** (new file `webapp/src/platform/desktop-adapter.ts`):
   - `storage` → `tauri-plugin-store`
   - `secureStorage` → `keyring` crate (Mac Keychain) via Tauri command `secure_get`/`secure_set`/`secure_remove`
   - `getApiBaseUrl` → reads `server_url` from store
   - `getDeviceName` → `tauri::os::hostname`
5. **Adapter selection:** `webapp/src/platform/index.ts` exports active adapter. Detect `window.__TAURI__` at runtime.
6. **Onboarding screen** (Tauri-only, gated by missing `server_url` or token):
   - Inputs: server URL (e.g. `http://192.168.98.98:8098`), username, password
   - On submit: ping `/api/health`, then `POST /api/auth/login`, save token to keychain + URL to store
   - On success: navigate to `/`
7. **Auto-login on launch:** read token + server URL → if both present, attach to API client + skip login.
8. **Logout flow:** clear keychain + store, return to onboarding.

### Acceptance
- App boots → onboarding (first time) or home (returning user)
- Login → home loads, can browse Movies/Series (no playback yet)
- Logout returns to onboarding
- Token survives app restart (keychain)

### Files touched
- `desktop/src-tauri/Cargo.toml` — add `tauri-plugin-store`, `keyring`
- `desktop/src-tauri/src/secure_storage.rs` (new) — keyring commands
- `desktop/src-tauri/src/lib.rs` — register store plugin + secure commands
- `webapp/src/platform/desktop-adapter.ts` (new)
- `webapp/src/platform/index.ts` (new) — adapter selector
- `webapp/src/pages/OnboardingPage.tsx` (new, desktop-only)
- `webapp/src/providers/Router.tsx` — add onboarding route, gate login

---

## Phase 2 — Player adapter (2 weeks) — heaviest phase

**Goal:** WatchPage + WatchLivePage delegate to libmpv on desktop, HTMLVideoElement on web. Same UX both sides.

### Strategy
Introduce `IPlayer` interface. Two implementations (`WebPlayer`, `DesktopPlayer`). WatchPage uses `usePlayer()` hook returning the active impl. **Don't refactor all 2526 lines at once** — wrap existing video logic into `WebPlayer`, build `DesktopPlayer` against same interface, swap at top.

### IPlayer surface (proposal)
```ts
interface IPlayer {
  load(opts: { url, headers?, startTime? }): Promise<void>
  play(): void
  pause(): void
  seek(seconds: number): void
  setVolume(v: number): void
  setRate(r: number): void
  setSubtitle(opts: { url, label, lang } | null): void
  setSubtitleDelay(ms: number): void
  setAudioTrack(idx: number): void
  setQuality(opts: { url, position }): void  // reload at position
  position(): number
  duration(): number
  on(event: 'ready'|'play'|'pause'|'ended'|'error'|'progress', cb): void
}
```

### Tasks
1. **Define `IPlayer` + `usePlayer()` hook** in `webapp/src/lib/player/`.
2. **WebPlayer** — wraps existing HTMLVideoElement + hls.js logic. Goal: zero behavior change vs today.
3. **DesktopPlayer** — calls Tauri commands. Each method = 1 mpv property/command.
4. **Tauri player commands** (extend `desktop/src-tauri/src/player.rs`):
   - `player_load(url, headers, start_time)`, `player_seek(s)`, `player_set_volume(v)`, `player_set_rate(r)`
   - `player_sub_add(url, label, lang)`, `player_sub_delay(ms)`, `player_sub_remove()`
   - `player_audio_set(idx)`, `player_audio_list()` → JSON
   - `player_position()`, `player_duration()`
   - Stream events `time-pos`, `pause`, `playback-restart`, `end-file` → emit `player-event`
5. **Subtitle loading:** mpv `sub-add` with HTTP URL + token header for primary track. **Dual-sub:** disable mpv built-in render (`sub-visibility=no`) → fetch SRT/VTT cues via existing webapp logic, render through `DualSubtitleOverlay` (which sits on the webview, above the mpv layer). PGS/image subtitles render via mpv `sid` (single track, no dual). Track changes feed both renders.
6. **Trickplay preview:** sprite fetch unchanged; overlay positions over webview controls (mpv layer is below webview, so overlay just sits on the webview as today).
7. **Resume position:** `time-pos` tick every 5s → existing progress API.
8. **Quality switching:** unload + reload at saved position. Emit synthetic `loadstart`/`loadeddata` events for WatchPage spinner.
9. **WatchLivePage:** same adapter, just different URL source. mpv handles live HLS with `cache=yes` already.
10. **Cleanup:** stop mpv on route leave + on window close.

### Acceptance
- `WatchPage` works identically on web (regression-test all features)
- `WatchPage` on desktop: play/pause/seek/sub/audio/quality/PiP all work
- `WatchLivePage` plays live channels with channel switch
- Resume position persists, progress tracking syncs to backend
- HDR/DV titles render correctly (validated via `color_info`)

### Files touched
- `webapp/src/lib/player/IPlayer.ts` (new)
- `webapp/src/lib/player/WebPlayer.ts` (new) — extracted from WatchPage
- `webapp/src/lib/player/DesktopPlayer.ts` (new)
- `webapp/src/lib/player/usePlayer.ts` (new)
- `webapp/src/pages/WatchPage.tsx` — replace direct video API with `IPlayer`
- `webapp/src/pages/WatchLivePage.tsx` — same
- `desktop/src-tauri/src/player.rs` — extend command surface
- `desktop/src-tauri/src/lib.rs` — register new commands

---

## Phase 3 — Window UX (1 week)

**Goal:** Native-feeling Mac window with proper fullscreen, shortcuts, file association.

### Tasks
1. **Window state persist** — `tauri-plugin-window-state` (size + position).
2. **Fullscreen sync** — Tauri `set_fullscreen` ↔ webapp `useFullscreen` hook + mpv `fs` property.
3. **Keyboard shortcuts** (when WatchPage focused): space, ←/→ (5s), shift+←/→ (30s), ↑/↓ (vol), F (fs), M (mute), Esc (exit fs), `[` `]` (rate).
4. **Drag-drop local file** — Tauri file-drop event → load into player (skip stream URL flow).
5. **`velox://` URL handler** — open `velox://watch/{id}` from outside app (shareable links).
6. **Mac native menu** — File / View / Playback / Help with shortcuts.
7. **Window close handler** — stop mpv cleanly before exit (avoid render-context-leak crash).

### Acceptance
- Fullscreen toggle smooth (no ugly transition)
- Drag MP4 onto app icon → plays directly
- All keyboard shortcuts behave like a native player

---

## Phase 4 — Settings & polish (3–4 days)

### Tasks
1. **Desktop-only Settings tab** in `SettingsPage`:
   - HW decode toggle (`videotoolbox` / off)
   - Tone-mapping algorithm picker (`bt.2390` / `hable` / `mobius` / `reinhard`)
   - Network timeout + cache size sliders (mpv properties)
   - Server URL change (forces re-login)
   - "Open log folder", "Open config folder" buttons
2. **About dialog** — Velox version + libmpv/ffmpeg/libplacebo versions (already exposed via `get_versions`).
3. **App branding** — Velox icon (1024×1024 source → all sizes), DMG background.
4. **First-run splash** — quick "checking server…" while pinging NAS.

### Acceptance
- Settings persist across restarts
- Log folder accessible from app

---

## Phase 5 — Mac ship (1 week)

### Tasks
1. **Bundle native libs into `.app`:**
   - Copy `libmpv.dylib` + `libavcodec.dylib`/etc + `libdovi.dylib` + `libplacebo.dylib` to `Contents/Frameworks/`
   - `install_name_tool` rewrite paths to `@rpath` / `@executable_path/../Frameworks/`
   - Tauri `bundle.macOS.frameworks` config + `before_bundle_command` script
2. **Codesign + notarize** with Apple Developer ID (sign each dylib + the app).
3. **DMG installer** via `tauri-bundler`.
4. **Tauri auto-updater** — JSON manifest hosted in GitHub Releases.
5. **Smoke test on a clean Mac** (no DYLD env vars, no manually-installed mpv).

### Acceptance
- Drag-and-drop install on a fresh Mac
- App launches, plays HDR/DV title with correct colors
- "Check for updates" prompts and applies update

---

## Out of scope (Plan W.2)

- Linux build (libmpv from package manager + AppImage)
- Windows port (PoC #4 libmpv build, swap NSOpenGLView → ANGLE/D3D11 view, MSI installer)
- Per-server multi-account
- Offline downloads
- Picture-in-Picture mini window (needs separate NSWindow + render context)

---

## Open questions

1. **mpv config:** ship a default `mpv.conf` baked into Settings, or expose every property as a UI toggle? **Recommend:** ship sensible defaults, expose only the 3-4 most-changed (HW decode, tone-mapping, cache).

---

## Timeline

| Phase | Effort | Cumulative |
|-------|--------|------------|
| 1 — Infrastructure  | 1 wk   | 1 wk |
| 2 — Player adapter  | 2 wk   | 3 wk |
| 3 — Window UX       | 1 wk   | 4 wk |
| 4 — Settings polish | 3-4 d  | ~5 wk |
| 5 — Mac ship        | 1 wk   | ~6 wk |

**Target ship date:** ~6 weeks from kickoff.
