/**
 * usePlayer — picks WebPlayer or DesktopPlayer based on the active platform,
 * and keeps the instance stable for the component's lifetime.
 *
 * The instance is destroyed on unmount.
 */

import { useEffect, useRef, useState } from 'react'
import { isTauri } from '@/platform'
import type { IPlayer } from './types'
import { WebPlayer } from './WebPlayer'
import { DesktopPlayer } from './DesktopPlayer'

export function usePlayer(): IPlayer {
  // Stable across renders within the same hook instance.
  const ref = useRef<IPlayer | null>(null)
  if (ref.current === null) {
    ref.current = isTauri() ? new DesktopPlayer() : new WebPlayer()
  }

  useEffect(() => {
    return () => {
      ref.current?.destroy()
      ref.current = null
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  return ref.current
}

/**
 * useIsDesktopPlayer — small convenience hook for branching UI when libmpv is
 * the active backend. Tied to the same isTauri() check used everywhere else.
 */
export function useIsDesktopPlayer(): boolean {
  const [v] = useState(() => isTauri())
  return v
}
