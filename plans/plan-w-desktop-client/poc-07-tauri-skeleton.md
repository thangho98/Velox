# PoC #7 — Tauri 2 skeleton + libmpv embed

**Status:** ✅ Skeleton compiles & launches. Playback verification interactive.
**Date:** 2026-04-28
**Path:** `~/Desktop/source/mpv-poc/velox-desktop/`

## Goal
Scaffold a Tauri 2 app that:
1. Renders the (eventually) Velox webapp UI in the webview (currently a hand-written test page).
2. Owns a libmpv instance that can `play`, `pause`, `resume`, `stop`, `probe HDR/DV` via Tauri commands.
3. Forwards mpv events (StartFile, FileLoaded, PlaybackRestart, EndFile, Shutdown) to the webview.

**Deliberately deferred:** real video into a Tauri webview surface (needs render-context API). For PoC, libmpv's own `cocoa-cb` spawns a separate top-level video window — fine for proving the pipeline.

## Decisions
- **Direct `libmpv2 = "5.0"`** instead of `tauri-plugin-libmpv = "0.3.2"`. Plugin is too opinionated and we want fine-grained control over property setup, event extraction, multi-window strategy.
- **`parking_lot::Mutex<Player>`** for shared state across Tauri commands and the event-pump thread. No async (libmpv API is sync).
- **Event pump on a dedicated thread** that calls `wait_event(0.1)` while holding the lock briefly, extracts an owned `PlayerEvent`, then drops the lock and emits via `AppHandle::emit`.
- **Lib crate-types: `["staticlib", "cdylib", "rlib"]`** (Tauri 2 mobile-friendly default).
- **`bundle.active = false`** during PoC — skip codesign/bundle complexity.

## Layout
```
velox-desktop/
├── dist/
│   └── index.html        ← test UI (versions, file path input, play/pause/stop, probe)
└── src-tauri/
    ├── Cargo.toml         ← tauri = 2, libmpv2 = 5.0, parking_lot, serde
    ├── build.rs           ← links libmpv via MPV_LIB_DIR + tauri_build::build()
    ├── tauri.conf.json    ← productName, 1100×720 window, frontendDist=../dist, bundle.active=false
    ├── icons/icon.png     ← 32×32 RGBA placeholder (Tauri requires RGBA, not RGB or paletted)
    └── src/
        ├── main.rs        ← entry point, calls velox_desktop_lib::run()
        ├── lib.rs         ← AppState, Tauri commands (get_versions/play/pause/.../color_info)
        └── player.rs      ← Player struct wrapping Mpv + run_event_loop()
```

## Key code

### src/player.rs — HDR/DV-friendly defaults
```rust
let mpv = Mpv::new()?;
mpv.set_property("hwdec", "videotoolbox")?;
mpv.set_property("vo", "gpu-next")?;
mpv.set_property("gpu-api", "vulkan")?;
mpv.set_property("target-colorspace-hint", "yes")?;
mpv.set_property("tone-mapping", "bt.2390")?;
mpv.set_property("force-window", "yes")?;
mpv.set_property("keep-open", "yes")?;
```

### src/player.rs — event loop ownership trick
```rust
let (payload, shutdown) = {
    let mut p = player.lock();
    match p.inner.wait_event(0.1) {
        Some(Ok(Event::StartFile)) => (Some(PlayerEvent { kind: "start-file", detail: None }), false),
        Some(Ok(Event::FileLoaded)) => (Some(PlayerEvent { kind: "file-loaded", detail: None }), false),
        Some(Ok(Event::PlaybackRestart)) => (Some(PlayerEvent { kind: "playback-restart", detail: None }), false),
        Some(Ok(Event::EndFile(_))) => (Some(PlayerEvent { kind: "end-file", detail: None }), false),
        Some(Ok(Event::Shutdown)) => (Some(PlayerEvent { kind: "shutdown", detail: None }), true),
        _ => (None, false),
    }
};
if let Some(p) = payload { let _ = handle.emit("player-event", p); }
if shutdown { return; }
std::thread::sleep(Duration::from_millis(20));
```

**Why this shape:** `Event<'a>` borrows from `MutexGuard<Mpv>`. We must convert to `'static` data (owned `PlayerEvent`) **inside** the lock scope before dropping the guard. Otherwise borrow checker rejects with E0597.

### src/lib.rs — color_info command (uses PoC #6 finding)
```rust
#[tauri::command]
fn color_info(state: State<AppState>) -> Result<ColorInfo, String> {
    let p = state.player.lock();
    let cm = p.get_str("video-params/colormatrix").ok();
    let is_dv = cm.as_deref() == Some("dolbyvision");
    Ok(ColorInfo {
        is_dolby_vision: is_dv,
        colormatrix: cm,
        primaries: p.get_str("video-params/primaries").ok(),
        gamma: p.get_str("video-params/gamma").ok(),
        sig_peak: p.get_str("video-params/sig-peak").ok(),
    })
}
```

## Build & run

```sh
# Build (only once Tauri CLI is in $PATH; raw cargo build works too)
cd ~/Desktop/source/mpv-poc/velox-desktop/src-tauri
cargo build              # ~7s incremental, ~3min cold
# Run binary directly (Tauri 2 doesn't need cargo tauri dev for static frontends)
DYLD_LIBRARY_PATH=~/Desktop/source/mpv-poc/libmpv/lib ./target/debug/velox-desktop
```

## Gotchas hit & fixed

1. **Missing icon** — `tauri::generate_context!()` proc macro panics if `icons/icon.png` is absent. Created a 32×32 RGBA blue PNG via Python.
2. **Icon is not RGBA** — first attempt was an 8-byte 1×1 PNG (RGB). Tauri rejects: "icon … is not RGBA". Fixed by writing IHDR with color type=6 (RGBA).
3. **Borrow checker E0597** — `wait_event` returns `Option<Result<Event<'_>>>` borrowed from the guard. Restructured event loop to extract owned `PlayerEvent` inside the guard scope (above).
4. **`mut mpv` warning** — `libmpv2 5.x`'s `set_property` takes `&self` (interior locking), so binding doesn't need `mut`.

## What's NOT yet verified

- **End-to-end click-Play→video-renders flow.** Skeleton boots, webview loads, commands are wired. Next step is interactive: `cargo run` the binary, paste a DV file path into the test UI, click Play, see a separate libmpv top-level window appear with HDR/DV-correct rendering.
- **Embedded video into the webview.** Currently libmpv spawns its own window (cocoa-cb top-level NSWindow). For final UX, would need `mpv_render_context` + a Metal/OpenGL surface from the Tauri webview. Punt to a later phase.
- **Codesigning / bundle.** `bundle.active = false` during PoC.
- **Windows side.** Skeleton is Mac-only right now; needs cross-build setup (PoC #4 done for libmpv; Tauri side TBD).

## Next concrete steps
1. **Manual smoke test:** launch app, click Play on `regular.mkv`, confirm video window appears + HDR/DV looks correct.
2. **Wire to real Velox webapp** — replace `dist/index.html` with `webapp/dist/` build output, swap test UI for the actual webapp.
3. **Stream URL flow** — Tauri command that takes a Velox `media_id` + `api_key`, builds the `direct_url`, hands to libmpv `loadfile`.
4. **Window embedding research** — investigate `mpv_render_context` over Tauri's NSWindow handle. If too painful for v1, ship with cocoa-cb top-level video window (acceptable UX for media player).
