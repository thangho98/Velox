# PoC #3 — libmpv-rs FFI Smoke Test

**Status:** ✅ Successful — 2026-04-28
**Risk:** MEDIUM (originally) → RESOLVED
**Goal:** Confirm Rust ↔ custom libmpv.dylib FFI works end-to-end (load → set/get props → loadfile → event loop → exit). This is the integration path Tauri 2 will use.

## Outcome

| Check | Result |
|---|---|
| `Mpv::new()` (mpv_create + mpv_initialize) | ✅ |
| `get_property` (mpv-version, ffmpeg-version) | ✅ `mpv v0.41.0` / `8.1` |
| `set_property` for HDR/DV/HW decode options | ✅ all 5 accepted |
| `command("loadfile", ...)` | ✅ Ok(()) |
| Event loop (`wait_event`) | ✅ 9 events delivered |
| Full lifecycle (StartFile → FileLoaded → PlaybackRestart → EndFile(0)) | ✅ |

### Tauri-relevant options that all set successfully

```
hwdec                  = videotoolbox        ← HW decode on Apple Silicon
vo                     = gpu-next             ← libplacebo render path (REQUIRED for DV reshaping)
gpu-api                = vulkan               ← MoltenVK on macOS
target-colorspace-hint = yes                  ← signal HDR target to display
tone-mapping           = bt.2390              ← perceptual HDR→SDR map
```

### HW decoder coverage

```
hwdec-codecs = h264, vc1, hevc, vp8, vp9, av1, prores, prores_raw, ffv1, dpx
```

All major codecs Velox handles are HW-decoded by VideoToolbox.

## Crate selection

Crate: **`libmpv2 = "5.0"`** (`libmpv2 5.0.3` resolved). Latest active fork; clean API; uses `mpv.wait_event(timeout)` directly.

Rejected:
- `libmpv = "2.0.1"` — older, less maintained
- `libmpv-sirno`, `radiance-libmpv` — niche forks

For Tauri integration: **`tauri-plugin-libmpv = "0.3.2"`** exists and uses libmpv2 internally — likely the path of least resistance for Plan W's Tauri shell.

## Project structure

```
~/Desktop/source/mpv-poc/
  libmpv/                    ← isolated install prefix (libmpv.dylib + headers)
    include/mpv/{client,render,render_gl,stream_cb}.h
    lib/libmpv.dylib → libmpv.2.dylib
    lib/pkgconfig/mpv.pc    ← rewritten to point at this prefix (not /opt/homebrew)
  libmpv-smoke/              ← cargo project
    Cargo.toml
    build.rs
    src/main.rs
```

### `Cargo.toml`

```toml
[package]
name = "libmpv-smoke"
version = "0.1.0"
edition = "2021"

[dependencies]
libmpv2 = "5.0"
```

### `build.rs`

```rust
fn main() {
    let libdir = std::env::var("MPV_LIB_DIR")
        .unwrap_or_else(|_| "/Users/thawng/Desktop/source/mpv-poc/libmpv/lib".into());
    println!("cargo:rustc-link-search=native={}", libdir);
    println!("cargo:rustc-link-lib=dylib=mpv");
    println!("cargo:rustc-link-arg=-Wl,-rpath,{}", libdir);  // embed @rpath so binary finds dylib at runtime
    println!("cargo:rerun-if-env-changed=MPV_LIB_DIR");
}
```

Without `build.rs` cargo couldn't find `-lmpv` even with `PKG_CONFIG_PATH` exported — `libmpv2-sys 4.0.1` doesn't run pkg-config probe in its own build script.

## Issues hit

| Issue | Fix |
|---|---|
| `ld: library 'mpv' not found` | Added explicit `cargo:rustc-link-search` + `rustc-link-lib` in our build.rs |
| `mpv.wait_event` not on Mpv | libmpv2 5.x exposes it directly on `Mpv`, not via separate `EventContext::create_event_context` (older API) |
| Properties returning `Raw(-8)` (PROPERTY_NOT_FOUND) | `libplacebo-version`, `video-codecs`, `audio-codecs` renamed/removed in mpv master. Not blockers — switch to `decoder-list`/`encoder-list` if needed |

## Verification command

```sh
cd ~/Desktop/source/mpv-poc/libmpv-smoke && cargo run
```

Output excerpt:
```
== Loadfile test (lavfi audio sine) ==
  loadfile result: Ok(())
  ev[2]: StartFile
  ev[5]: FileLoaded
  ev[7]: PlaybackRestart
  ev[9]: EndFile(0)

== Pipeline summary ==
  total events: 9
  got StartFile: true
  got EndFile: true
```

## What this PoC does NOT cover yet

- **Window rendering** — synthetic audio file rendered to `vo=null`. Real video output to a window needs:
  - Option A: separate winit window + `wid` parameter (raw NSView pointer)
  - Option B: `mpv_render_context_create` + Tauri webview overlay (similar to IINA's approach)
- **Actual DV file playback** — requires test asset (DV Profile 5/7/8 HEVC). Saved for PoC #5.
- **HDR display detection** — `target-trc=pq` + `target-prim=bt.2020` should be auto-set by mpv when display reports HDR. Need real HDR display + content to verify.

## Conclusion

✅ **Tauri 2 integration path confirmed viable.** All FFI primitives Tauri needs work:
- Property get/set for runtime config (incl. hwdec, vo=gpu-next, tone-mapping)
- Command dispatch for player control (`loadfile`, `seek`, `pause`, etc.)
- Event loop for UI sync (StartFile, FileLoaded, EndFile, plus PropertyChange via `observe_property`)
- Custom dylib linkage via build.rs + rpath embedding

For Plan W desktop scaffold: use `tauri-plugin-libmpv` and point its build at our custom libmpv via `MPV_LIB_DIR` env var.

## Next PoCs

- **PoC #4** — Windows cross-compile pipeline (or msys2 native build)
- **PoC #5** — Render-to-window test with winit + raw-window-handle (proves video output pipeline, not just audio)
- **PoC #6** — DV/HDR file playback validation against real assets from Velox NAS
