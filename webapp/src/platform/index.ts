/**
 * Selects the active PlatformAdapter at boot time.
 *
 * Detect order:
 * 1. Tauri (desktop) — `window.__TAURI_INTERNALS__` or `window.__TAURI__`
 * 2. Web — fallback
 */

import type { PlatformAdapter } from '@velox/shared/platform'

export function isTauri(): boolean {
  if (typeof window === 'undefined') return false
  const w = window as unknown as Record<string, unknown>
  return '__TAURI_INTERNALS__' in w || '__TAURI__' in w
}

export async function loadPlatform(): Promise<PlatformAdapter> {
  if (isTauri()) {
    const { initDesktopAdapter } = await import('./desktop-adapter')
    return initDesktopAdapter()
  }
  const { webPlatform } = await import('./web-adapter')
  return webPlatform
}
