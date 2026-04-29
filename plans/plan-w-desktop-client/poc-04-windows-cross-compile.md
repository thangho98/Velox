# PoC #4 — Windows Cross-Compile from macOS arm64

**Status:** ✅ Toolchain + critical libs verified — 2026-04-28
**Risk:** HIGH (originally) → LOW (toolchain proven; remaining work = mechanical)
**Goal:** Confirm the same DV-capable mpv stack we built for macOS can also be produced for Windows x86_64. Decide whether shipping Plan W on Windows is feasible from the same build host.

## Outcome — what cross-compiles cleanly from this Mac

| Component | Toolchain | Output | DV evidence |
|---|---|---|---|
| **libdovi 3.3.2** | cargo + MinGW linker | `dovi.dll` 1.8 MB (PE32+ x86-64) | The whole crate |
| **libplacebo 7.362.0** | meson cross-file + MinGW | `libplacebo-362.dll` 6.6 MB (PE32+ x86-64) | `pl_shader_dovi_reshape`, `pl_hdr_metadata_from_dovi_rpu`, `utils_dolbyvision.c.obj` all present |
| **FFmpeg n8.2-dev** | autoconf cross-compile + MinGW | 9 DLLs, 33 MB total | `av_dovi_alloc`, `av_dovi_metadata_alloc`, `av_dynamic_hdr_plus_*`, `av_dynamic_hdr_vivid_*` exported from `avcodec-62.dll` |

### FFmpeg artifacts produced

| DLL | Size |
|---|---|
| `avcodec-62.dll` | 19 MB |
| `avformat-62.dll` | 2.8 MB |
| `avfilter-11.dll` | 5.9 MB |
| `avutil-60.dll` | 1.2 MB |
| `swscale-9.dll` | 2.3 MB |
| `swresample-6.dll` | 176 KB |
| `avdevice-62.dll` | 152 KB |

All `PE32+ executable (DLL) x86-64, for MS Windows`. Total DV-capable media stack for Windows: **~40 MB** of DLLs cross-compiled from macOS arm64.

Plus, `pkg-config` resolution chain works end-to-end — FFmpeg's configure script located libplacebo via the cross-installed `.pc` files, libplacebo's meson located libdovi via its `.pc`, etc.

## Toolchain (one-time setup on macOS)

```sh
brew install mingw-w64        # 1.4 GB, gcc 15.2.0 + binutils + windows headers
rustup target add x86_64-pc-windows-gnu
```

Cargo cross-linker config (`~/.cargo/config.toml`):
```toml
[target.x86_64-pc-windows-gnu]
linker = "x86_64-w64-mingw32-gcc"
ar = "x86_64-w64-mingw32-ar"
```

Meson cross file (`/tmp/mingw-cross.ini`):
```ini
[binaries]
c = 'x86_64-w64-mingw32-gcc'
cpp = 'x86_64-w64-mingw32-g++'
ar = 'x86_64-w64-mingw32-ar'
strip = 'x86_64-w64-mingw32-strip'
windres = 'x86_64-w64-mingw32-windres'
dlltool = 'x86_64-w64-mingw32-dlltool'
pkg-config = 'pkg-config'

[host_machine]
system = 'windows'
cpu_family = 'x86_64'
cpu = 'x86_64'
endian = 'little'

[properties]
needs_exe_wrapper = true
```

External Vulkan headers (one-time):
```sh
curl -sL https://github.com/KhronosGroup/Vulkan-Headers/archive/refs/tags/v1.4.341.tar.gz | tar xz -C /tmp
cp -R /tmp/Vulkan-Headers-*/include/{vulkan,vk_video} ~/Desktop/source/mpv-poc/install_win/include/
```

## Build pipeline

Workspace addition:
```
~/Desktop/source/mpv-poc/install_win/   ← Windows install prefix
  bin/dovi.dll
  bin/libplacebo-362.dll
  include/{libdovi,libplacebo,vulkan,vk_video}/...
  lib/{dovi.dll.a, libplacebo.dll.a, libdovi.a}
  lib/pkgconfig/{dovi.pc, libplacebo.pc}
```

### Step 1 — libdovi for Windows

```sh
cd ~/Desktop/source/mpv-poc/dovi_tool/dolby_vision
cargo cinstall --release \
  --target x86_64-pc-windows-gnu \
  --prefix=$HOME/Desktop/source/mpv-poc/install_win
```

### Step 2 — libplacebo for Windows

```sh
cd ~/Desktop/source/mpv-poc/mpv-build/libplacebo
export PKG_CONFIG_LIBDIR="$HOME/Desktop/source/mpv-poc/install_win/lib/pkgconfig"
meson setup build_win \
  --cross-file /tmp/mingw-cross.ini \
  --prefix=$HOME/Desktop/source/mpv-poc/install_win \
  --default-library=shared \
  -Dvulkan=enabled -Dvk-proc-addr=disabled \
  -Dlibdovi=enabled \
  -Dshaderc=disabled -Dxxhash=disabled \
  -Ddemos=false -Dtests=false
meson compile -C build_win
meson install -C build_win
```

⚠️ `vk-proc-addr=disabled` for cross-build — the volk loader normally needed `pkg-config` glue we don't have on Mac. mpv will still resolve Vulkan via the Windows ICD at runtime through libplacebo's own loader.

### Step 3 — FFmpeg for Windows

```sh
cd ~/Desktop/source/mpv-poc/mpv-build/ffmpeg
mkdir build_win && cd build_win
export PKG_CONFIG_LIBDIR="$HOME/Desktop/source/mpv-poc/install_win/lib/pkgconfig"
../configure \
  --enable-cross-compile --target-os=mingw32 --arch=x86_64 \
  --cross-prefix=x86_64-w64-mingw32- --pkg-config=pkg-config \
  --extra-cflags="-I$HOME/Desktop/source/mpv-poc/install_win/include" \
  --extra-ldflags="-L$HOME/Desktop/source/mpv-poc/install_win/lib" \
  --enable-libplacebo \
  --enable-shared --disable-static --disable-programs \
  --disable-doc --disable-debug \
  --prefix=$HOME/Desktop/source/mpv-poc/install_win
make -j$(sysctl -n hw.ncpu) && make install
```

### Step 4 — libmpv.dll (TODO, not done in this PoC)

Remaining work (mechanical, not risk):
1. Cross-compile **libass** + **fribidi** + **harfbuzz** + **freetype** + **libpng** + **brotli** + ... (mpv runtime deps). Either:
   - Adapt `mpv-build/scripts/*` to invoke meson with `--cross-file` (small patch to `mpv-config`)
   - **OR** use [`shinchiro/mpv-winbuild-cmake`](https://github.com/shinchiro/mpv-winbuild-cmake) — production-grade Linux→Win cross-build script used by all major mpv Windows distributions. Has `enable_dovi=ON` flag.
2. Cross-compile **mpv** itself with meson cross-file targeting our `install_win/` prefix.

Recommended for Plan W shipping: GitHub Actions matrix on `ubuntu-latest` running `mpv-winbuild-cmake`. macOS host PoC proves the toolchain works; CI provides repeatability + cache.

## Verification commands

```sh
# Confirm Windows binaries
file ~/Desktop/source/mpv-poc/install_win/bin/*.dll
# → both report "PE32+ executable (DLL) ... x86-64, for MS Windows"

# DV symbols in Windows libplacebo
x86_64-w64-mingw32-nm \
  ~/Desktop/source/mpv-poc/install_win/bin/libplacebo-362.dll \
  | grep -iE "(dovi_reshape|hdr_metadata_from_dovi)"
# → pl_hdr_metadata_from_dovi_rpu, pl_shader_dovi_reshape

# DLL imports (sanity — should reference dovi.dll for libplacebo)
x86_64-w64-mingw32-objdump -p \
  ~/Desktop/source/mpv-poc/install_win/bin/libplacebo-362.dll \
  | grep "DLL Name"

# pkg-config feature flags
grep -E "pl_has_(dovi|libdovi|vulkan)" \
  ~/Desktop/source/mpv-poc/install_win/lib/pkgconfig/libplacebo.pc
# → all =1
```

## Issues hit (and fixes)

| Issue | Fix |
|---|---|
| meson "Program 'llvm-dlltool dlltool' not found" | Added `dlltool = 'x86_64-w64-mingw32-dlltool'` to cross-file |
| FFmpeg "libplacebo not found" then `vulkan/vulkan.h: No such file` | Installed Vulkan-Headers v1.4.341 from KhronosGroup tarball into install_win/include |
| Shared cargo config conflict | Used `--target` flag explicitly so Mac builds still default to native arm64 |
| FFmpeg `mfenc.c` fails with `ID3D11Texture2D` undeclared | Added `--disable-mediafoundation` to FFmpeg configure (we disabled `--disable-d3d11va` for cleanliness; mediafoundation encoder still pulls D3D11 types). Loss is only the MF-based H264/HEVC encoder; we don't need it for client playback (we DECODE on the client, no encoding). |

## Conclusion

✅ **Windows cross-compile path is viable from macOS arm64.** All three "non-trivial" components in the DV pipeline (libdovi Rust crate, libplacebo C+meson, FFmpeg autoconf) cross-compile cleanly. No source patches required.

Combined with PoC #1 (Mac arm64 native build), Plan W can ship:
- macOS arm64 (.app)
- macOS x86_64 (.app via second native build, then `lipo -create` for universal)
- Windows x86_64 (.exe)

All three target the same `libmpv.{dylib,dll}` API → Tauri 2 + libmpv2 Rust binding works on all of them.

**Recommended production setup:**
- macOS builds → GitHub Actions `macos-latest` runner (native)
- Windows builds → GitHub Actions `ubuntu-latest` runner with [shinchiro/mpv-winbuild-cmake](https://github.com/shinchiro/mpv-winbuild-cmake) (designed for this; has libdovi + libplacebo + cache)

This Mac host's MinGW pipeline is fine for ad-hoc testing but not the production build path — just slower than a dedicated Linux runner, and missing a few peripheral codecs we'd want (libxml2, libsoxr, dvbsubdec) without more cross-installed deps.

## Risk update for Plan W

| Risk | Status |
|---|---|
| Mac arm64 libmpv build | ✅ DONE (PoC #1) |
| Rust↔libmpv FFI | ✅ DONE (PoC #3) |
| Windows toolchain | ✅ DONE (this PoC) |
| Windows libmpv.dll full build | OPEN — mechanical only, use CI |
| winit window render path | OPEN (PoC #5) |
| Real DV file playback validation | OPEN (PoC #6) |
| Tauri webview overlay vs separate window | OPEN (Plan W phase 1 design) |
