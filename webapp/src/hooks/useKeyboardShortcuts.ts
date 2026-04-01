import { useEffect, useEffectEvent } from 'react'

interface KeyboardShortcutActions {
  togglePlay: () => void
  seek: (seconds: number) => void
  applySeek: (time: number) => void
  changeVolume: (delta: number) => void
  toggleFullscreen: () => void
  toggleMute: () => void
  resetControlsTimeout: () => void
  isFullscreen: boolean
  isLocked: boolean
  setIsLocked: (v: boolean) => void
  seekStep: number
  volumeStep: number
}

export function useKeyboardShortcuts(actions: KeyboardShortcutActions) {
  const onKeyDown = useEffectEvent((e: KeyboardEvent) => {
    if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return
    // When locked, only allow 'l' to unlock (long-press not needed for keyboard)
    if (actions.isLocked) {
      if (e.key === 'l') {
        e.preventDefault()
        actions.setIsLocked(false)
        actions.resetControlsTimeout()
      }
      return
    }
    switch (e.key) {
      case ' ':
      case 'k':
        e.preventDefault()
        actions.togglePlay()
        break
      case 'ArrowLeft':
        e.preventDefault()
        actions.seek(-actions.seekStep)
        break
      case 'ArrowRight':
        e.preventDefault()
        actions.seek(actions.seekStep)
        break
      case 'ArrowUp':
        e.preventDefault()
        actions.changeVolume(actions.volumeStep)
        break
      case 'ArrowDown':
        e.preventDefault()
        actions.changeVolume(-actions.volumeStep)
        break
      case 'f':
        e.preventDefault()
        actions.toggleFullscreen()
        break
      case 'm':
        e.preventDefault()
        actions.toggleMute()
        break
      case 'j':
        e.preventDefault()
        actions.seek(-actions.seekStep * 2)
        break
      case 'l':
        e.preventDefault()
        actions.seek(actions.seekStep * 2)
        break
      case '0':
      case 'Home':
        e.preventDefault()
        actions.applySeek(0)
        break
      case 'Escape':
        if (actions.isFullscreen) {
          e.preventDefault()
          screen.orientation?.unlock?.()
          document.exitFullscreen()
        }
        break
    }
  })

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => onKeyDown(e)
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [])
}
