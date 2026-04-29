#!/usr/bin/env bash
# Copy libmpv + transitive dylib deps into the .app's Frameworks/ folder
# and rewrite install names to @rpath, so the bundle runs on any Mac.
#
# Run AFTER `pnpm tauri build` — this rewrites the freshly produced .app
# in target/release/bundle/macos/Velox Desktop.app.
#
# Set MPV_LIB_DIR if libmpv lives somewhere other than the default.

set -euo pipefail

APP_DIR="${APP_DIR:-$(cd "$(dirname "$0")/.." && pwd)/src-tauri/target/release/bundle/macos/Velox Desktop.app}"
MPV_LIB_DIR="${MPV_LIB_DIR:-/Users/thawng/Desktop/source/mpv-poc/libmpv/lib}"
FRAMEWORKS="$APP_DIR/Contents/Frameworks"
EXEC="$APP_DIR/Contents/MacOS/velox-desktop"

if [[ ! -d "$APP_DIR" ]]; then
  echo "error: app bundle not found at $APP_DIR" >&2
  echo "       run 'pnpm tauri build' first" >&2
  exit 1
fi

if [[ ! -f "$MPV_LIB_DIR/libmpv.2.dylib" ]]; then
  echo "error: libmpv.2.dylib not found at $MPV_LIB_DIR" >&2
  exit 1
fi

mkdir -p "$FRAMEWORKS"

# Tracking set so we don't reprocess the same dylib twice
declare -a PROCESSED=()
processed() {
  local name="$1"
  for p in "${PROCESSED[@]}"; do
    [[ "$p" == "$name" ]] && return 0
  done
  return 1
}

# Skip system libraries — they're stable across Macs
is_system() {
  case "$1" in
    /System/*|/usr/lib/*) return 0 ;;
    *) return 1 ;;
  esac
}

copy_and_rewrite() {
  local src="$1"
  local name
  name="$(basename "$src")"

  if processed "$name"; then return; fi
  PROCESSED+=("$name")

  local dst="$FRAMEWORKS/$name"

  # Resolve the real file (some dylibs are symlinks)
  local real
  real="$(readlink -f "$src" 2>/dev/null || stat -f "%Y" "$src" 2>/dev/null || echo "$src")"
  if [[ -z "$real" || ! -e "$real" ]]; then real="$src"; fi

  if [[ ! -f "$dst" ]]; then
    cp -L "$real" "$dst"
    chmod +w "$dst"
    echo "  copied $name"
  fi

  # Rewrite the dylib's own install name to @rpath/<name>
  install_name_tool -id "@rpath/$name" "$dst" 2>/dev/null || true

  # Walk dependencies, copy each non-system dep, rewrite reference inside $dst
  local deps
  deps="$(otool -L "$dst" | tail -n +2 | awk '{print $1}')"
  while IFS= read -r dep; do
    [[ -z "$dep" ]] && continue
    if is_system "$dep"; then continue; fi
    # Self-reference (after id rewrite) — skip
    [[ "$dep" == "@rpath/$name" ]] && continue

    local dep_name
    dep_name="$(basename "$dep")"
    install_name_tool -change "$dep" "@rpath/$dep_name" "$dst" 2>/dev/null || true

    # Resolve dep on disk to recursively bring it in
    local dep_path="$dep"
    if [[ ! -f "$dep_path" ]]; then
      # Try MPV_LIB_DIR (some libmpv deps are local)
      [[ -f "$MPV_LIB_DIR/$dep_name" ]] && dep_path="$MPV_LIB_DIR/$dep_name"
    fi
    if [[ -f "$dep_path" ]]; then
      copy_and_rewrite "$dep_path"
    else
      echo "  warn: cannot find $dep on disk — runtime load may fail"
    fi
  done <<< "$deps"
}

echo "→ Bundling libmpv + dependencies into $APP_DIR"
copy_and_rewrite "$MPV_LIB_DIR/libmpv.2.dylib"

# Add rpath to the main executable so it can locate dylibs in Frameworks/
echo "→ Adding @executable_path/../Frameworks rpath to $EXEC"
install_name_tool -add_rpath "@executable_path/../Frameworks" "$EXEC" 2>/dev/null \
  || echo "  (rpath may already exist — that's fine)"

# Strip any baked-in absolute rpath from build.rs that points at /Users/...
old_rpath="$MPV_LIB_DIR"
install_name_tool -delete_rpath "$old_rpath" "$EXEC" 2>/dev/null || true

# Rewrite executable's libmpv reference if it still points at absolute path
exec_libmpv_ref="$(otool -L "$EXEC" | awk '{print $1}' | grep -E 'libmpv\.[0-9]+\.dylib$' | head -1 || true)"
if [[ -n "$exec_libmpv_ref" && "$exec_libmpv_ref" != "@rpath/"* ]]; then
  install_name_tool -change "$exec_libmpv_ref" "@rpath/libmpv.2.dylib" "$EXEC" 2>/dev/null || true
  echo "  rewrote executable libmpv ref → @rpath/libmpv.2.dylib"
fi

# Re-codesign ad-hoc since rewriting install_name on the dylibs invalidated
# whatever signature `tauri build` left behind. macOS won't launch a bundle
# whose Mach-O hashes don't match its embedded signature.
codesign --deep --force --sign - "$APP_DIR"
echo "✓ Bundling complete (ad-hoc signed)."
