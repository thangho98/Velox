# PoC #5 + #6 — Render-to-window & DV file playback

**Status:** ✅ DONE (PoC #6 fully verified, PoC #5 deferred to Tauri host — see PoC #7)
**Date:** 2026-04-28
**Crate:** `~/Desktop/source/mpv-poc/dv-playback-test/`
**Test file:** `~/Desktop/source/mpv-poc/dovi_tool/assets/hevc_tests/regular.mkv` (DV Profile 8 BL+RPU)

## Goal
- PoC #5: Render video to its own native window via `vo=gpu-next` + `gpu-api=vulkan` (MoltenVK).
- PoC #6: Confirm Dolby Vision metadata is correctly detected by libmpv on a real DV file.

## Setup

### Cargo.toml
```toml
[package]
name = "dv-playback-test"
version = "0.1.0"
edition = "2021"

[dependencies]
libmpv2 = "5.0"
```

### build.rs (links libmpv from custom build, not Homebrew)
```rust
fn main() {
    let libdir = std::env::var("MPV_LIB_DIR")
        .unwrap_or_else(|_| "/Users/thawng/Desktop/source/mpv-poc/libmpv/lib".into());
    println!("cargo:rustc-link-search=native={libdir}");
    println!("cargo:rustc-link-lib=dylib=mpv");
    println!("cargo:rustc-link-arg=-Wl,-rpath,{libdir}");
}
```

### src/main.rs (key options)
```rust
mpv.set_property("hwdec", "videotoolbox")?;
if headless {
    mpv.set_property("vo", "null")?;       // PoC #6 path
} else {
    mpv.set_property("vo", "gpu-next")?;   // PoC #5 path (needs Cocoa runloop)
    mpv.set_property("gpu-api", "vulkan")?;
    mpv.set_property("force-window", "yes")?;
}
mpv.set_property("target-colorspace-hint", "yes")?;
mpv.set_property("tone-mapping", "bt.2390")?;
mpv.set_property("ao", "null")?;           // no audio output in PoC
mpv.set_property("keep-open", "yes")?;
mpv.command("loadfile", &[&path])?;
```

### Run
```sh
cd ~/Desktop/source/mpv-poc/dv-playback-test
DYLD_LIBRARY_PATH=~/Desktop/source/mpv-poc/libmpv/lib \
MPV_LIB_DIR=~/Desktop/source/mpv-poc/libmpv/lib \
HEADLESS=1 \
cargo run
```

## Result — PoC #6 ✅

```
Playing: regular.mkv
loadfile: Ok(())

== events ==
  ev[2]: StartFile
  ev[3]: FileLoaded
  ev[4]: VideoReconfig
  ev[5]: VideoReconfig    ← second reconfig is normal during DV reshape
  ev[6]: PlaybackRestart

  -- post-PlaybackRestart probe --
  video-format                     = hevc
  video-codec                      = H.265 / HEVC
  video-params/colormatrix         = dolbyvision   ← KEY DV SIGNAL
  video-params/primaries           = bt.2020
  video-params/gamma               = pq
  video-params/sig-peak            = 4.929096      ← HDR10 nits = sig_peak * 203
  video-params/dolby-vision-profile = ERR Raw(-8)  ← NOT exposed as property
  width / height                   = 256 / 144
  container-fps                    = 23.976025

== Final state ==
  time-pos        = 8.133000   ← decoded 8.13s of 10.8s in 8s window (faster than realtime)
  frame-drop-count = 0
```

### Critical findings
- **DV detection ≡ `colormatrix == "dolbyvision"`**, not `dolby-vision-profile` (Raw(-8) means PROPERTY_UNAVAILABLE).
  → Tauri `color_info` command must use `video-params/colormatrix` for DV gating.
- **HDR signaling intact:** primaries=bt.2020, gamma=pq, sig-peak=4.93 (≈1000 nits).
- **Decode performant:** 8.13s decoded in 8s wall (~1.0x realtime+ on small file), 0 dropped frames.
- **`hwdec-current=no` in headless mode** — VT only activates with a real VO. Hardware decode path will be exercised in Tauri (PoC #7).
- **`pixel-format` Raw(-8)** — also unavailable with `vo=null`. Not a problem; matters only when rendering.

## PoC #5 — VO=gpu-next windowed (deferred)

Initial run with `vo=gpu-next + gpu-api=vulkan + force-window=yes` returned `Raw(-16)` (`MPV_ERROR_VO_INIT_FAILED`):

```
ev[2]: StartFile
ev err: Raw(-16)
ev[3]: Deprecated(11)
```

### Root cause
`cocoa-cb` (the macOS window backend that libmpv uses for `vo=gpu-next` on standalone mode) needs:
1. Swift runtime + cocoa-cb compiled (we **have** this — `_cocoa_init_cocoa_cb` symbol present in `libmpv.2.dylib`).
2. `NSApplication` running on the main thread with an active event loop.

A plain `cargo run` binary doesn't bootstrap an NSApplication / runloop, so the Cocoa VO can't init. This is **not** a libmpv build defect — it's a host-process limitation.

### Resolution
PoC #5 is folded into **PoC #7 (Tauri 2 app)** — Tauri uses `tao` (which **does** create NSApplication on main thread), so cocoa-cb can spin up a top-level mpv window from there. Decision: don't write a separate winit/SDL harness — go straight to the real host.

## Verification commands

```sh
# Confirm cocoa-cb in libmpv build
nm ~/Desktop/source/mpv-poc/libmpv/lib/libmpv.2.dylib | grep _cocoa_init_cocoa_cb
# expected: 0000000000152494 t _cocoa_init_cocoa_cb

# Confirm Swift LibmpvHelper class
nm ~/Desktop/source/mpv-poc/libmpv/lib/libmpv.2.dylib | grep -c LibmpvHelper
# expected: > 0
```

## Takeaways for Tauri integration
1. **DV gating logic** in client should be `colormatrix === "dolbyvision"`, not a profile-number lookup.
2. **HDR detection:** `gamma === "pq"` ∧ `primaries === "bt.2020"`.
3. **VideoToolbox HW decode** has to be exercised in the real Tauri host where a window exists.
4. **Don't spawn libmpv on a non-main thread** if relying on cocoa-cb — but our pattern (libmpv on a worker thread that owns Mpv, Tauri main thread runs tao runloop) works because cocoa-cb internally dispatches to the main runloop via NSApplication. Confirmed in PoC #7.
