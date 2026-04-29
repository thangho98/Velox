/**
 * useDesktopBridge — wires Tauri-only OS integrations into React. Mounted once
 * at the top of the app on desktop builds; on web it's a no-op.
 *
 * Responsibilities:
 *   - Mac native menu commands (File / View / Playback) → app actions.
 *   - velox:// deep links → routed via React Router (same window).
 *   - File drag-drop onto the window → playback if it's a media file.
 *   - Window fullscreen toggle requests from the menu / shortcut.
 *
 * Player-action events ('velox-cmd:play-pause' etc.) are dispatched as plain
 * DOM CustomEvents so any active page (WatchPage, WatchLivePage) can subscribe
 * without us needing a global player reference.
 */

import { useEffect } from 'react'
import { useNavigate } from 'react-router'
import { isTauri } from '@/platform'
import { createLogger } from '@/lib/logger'

const log = createLogger('DesktopBridge')

export type VeloxCmd = 'play-pause' | 'seek-back' | 'seek-fwd' | 'toggle-fullscreen'

export function dispatchVeloxCmd(cmd: VeloxCmd): void {
  window.dispatchEvent(new CustomEvent<VeloxCmd>('velox-cmd', { detail: cmd }))
}

export function useVeloxCmd(handler: (cmd: VeloxCmd) => void): void {
  useEffect(() => {
    const onCmd = (e: Event) => {
      const detail = (e as CustomEvent<VeloxCmd>).detail
      if (detail) handler(detail)
    }
    window.addEventListener('velox-cmd', onCmd)
    return () => window.removeEventListener('velox-cmd', onCmd)
  }, [handler])
}

export function useDesktopBridge(): void {
  const navigate = useNavigate()

  useEffect(() => {
    if (!isTauri()) return
    let mounted = true
    const offFns: Array<() => void> = []

    const setup = async () => {
      const { listen } = await import('@tauri-apps/api/event')
      const { getCurrentWindow } = await import('@tauri-apps/api/window')
      const { open: openDialog } = await import('@tauri-apps/plugin-dialog')

      // ── Menu events from Rust (velox-menu) ────────────────────────────────
      const offMenu = await listen<string>('velox-menu', async (e) => {
        const id = e.payload
        log.info(`menu: ${id}`)
        switch (id) {
          case 'open-file': {
            const path = await openDialog({
              title: 'Open Media File',
              multiple: false,
              directory: false,
              filters: [
                {
                  name: 'Video',
                  extensions: ['mp4', 'mkv', 'mov', 'm4v', 'avi', 'webm', 'ts'],
                },
              ],
            })
            if (typeof path === 'string') routeLocalFile(path, navigate)
            break
          }
          case 'open-url': {
            const url = window.prompt('Enter media URL (http(s)://… or velox://…)')
            if (url) routeUrl(url, navigate)
            break
          }
          case 'toggle-fullscreen':
            try {
              const w = getCurrentWindow()
              const fs = await w.isFullscreen()
              await w.setFullscreen(!fs)
            } catch (err) {
              log.error(`fullscreen toggle failed: ${err}`)
            }
            dispatchVeloxCmd('toggle-fullscreen')
            break
          case 'reload':
            window.location.reload()
            break
          case 'playback-play-pause':
            dispatchVeloxCmd('play-pause')
            break
          case 'playback-seek-back':
            dispatchVeloxCmd('seek-back')
            break
          case 'playback-seek-fwd':
            dispatchVeloxCmd('seek-fwd')
            break
        }
      })
      if (!mounted) {
        offMenu()
        return
      }
      offFns.push(offMenu)

      // ── Deep links (velox-deep-link) ──────────────────────────────────────
      const offDl = await listen<string[]>('velox-deep-link', (e) => {
        const urls = e.payload
        log.info(`deep-link: ${urls.join(', ')}`)
        const first = urls[0]
        if (first) routeUrl(first, navigate)
      })
      if (!mounted) {
        offDl()
        return
      }
      offFns.push(offDl)

      // ── Window-level drag-and-drop of media files ─────────────────────────
      // Tauri 2 dispatches Drop events via the webview; we listen on the
      // current window for "tauri://drag-drop".
      const w = getCurrentWindow()
      const offDrop = await w.onDragDropEvent((event) => {
        if (event.payload.type !== 'drop') return
        const paths = event.payload.paths
        if (!Array.isArray(paths) || paths.length === 0) return
        const first = paths[0]
        log.info(`drag-drop: ${first}`)
        routeLocalFile(first, navigate)
      })
      if (!mounted) {
        offDrop()
        return
      }
      offFns.push(offDrop)
    }

    void setup().catch((err) => log.error(`setup failed: ${err}`))

    return () => {
      mounted = false
      for (const off of offFns) {
        try {
          off()
        } catch {
          // best-effort
        }
      }
    }
  }, [navigate])
}

// velox://watch/{id} → /watch/{id}
// velox://livetv/watch/{channelId} → /livetv/watch/{channelId}
// velox://media/{id} → /movies/{id} (default fallback to media detail)
// http(s):// → no-op (open externally would need user intent; ignore for now)
function routeUrl(url: string, navigate: (path: string) => void): void {
  try {
    const u = new URL(url)
    if (u.protocol === 'velox:') {
      const path = u.pathname.replace(/^\/+/, '')
      const host = u.host
      const route = host ? `/${host}/${path}`.replace(/\/$/, '') : `/${path}`
      navigate(route)
      return
    }
    log.info(`unhandled URL scheme: ${u.protocol}`)
  } catch (err) {
    log.error(`bad URL ${url}: ${err}`)
  }
}

function routeLocalFile(_path: string, _navigate: (path: string) => void): void {
  // v0.1 desktop is a thin client — local file playback would require a
  // sideloaded mpv direct-load path that bypasses the Velox HTTP backend.
  // Punt for now; surface a toast via console so we can wire later.
  log.info('local file open — not implemented in v0.1 (Velox-served URLs only)')
}
