import { useEffect, useState } from 'react'
import { isTauri } from '@/platform'

export function useFullscreen(
  containerRef: React.RefObject<HTMLDivElement | null>,
  videoRef: React.RefObject<HTMLVideoElement | null>,
  showToastInfo: (msg: string) => void,
) {
  const [isFullscreen, setIsFullscreen] = useState(false)

  const toggleFullscreen = () => {
    if (isTauri()) {
      void (async () => {
        try {
          const { getCurrentWindow } = await import('@tauri-apps/api/window')
          const w = getCurrentWindow()
          const fs = await w.isFullscreen()
          await w.setFullscreen(!fs)
          setIsFullscreen(!fs)
        } catch (err) {
          showToastInfo(`Fullscreen: ${(err as Error).message}`)
        }
      })()
      return
    }
    type FullscreenDoc = Document & {
      webkitFullscreenElement?: Element
      webkitExitFullscreen?: () => void
    }
    type FullscreenEl = HTMLElement & {
      webkitRequestFullscreen?: () => Promise<void>
    }
    const doc = document as FullscreenDoc
    const isFs = !!(document.fullscreenElement || doc.webkitFullscreenElement)
    const container = containerRef.current as FullscreenEl | null
    const video = videoRef.current as
      | (HTMLVideoElement & {
          webkitEnterFullscreen?: () => void
          webkitExitFullscreen?: () => void
          webkitDisplayingFullscreen?: boolean
        })
      | null

    if (isFs || video?.webkitDisplayingFullscreen) {
      // Exit fullscreen
      screen.orientation?.unlock?.()
      if (video?.webkitDisplayingFullscreen) {
        video.webkitExitFullscreen?.()
      } else if (document.exitFullscreen) {
        document.exitFullscreen().catch(console.error)
      } else {
        doc.webkitExitFullscreen?.()
      }
    } else {
      // Enter fullscreen — use container element (keeps custom controls + subtitles visible)
      const lockLandscape = () => {
        const o = screen.orientation as ScreenOrientation & { lock?: (s: string) => Promise<void> }
        o?.lock?.('landscape').catch(() => {})
      }
      if (container?.requestFullscreen) {
        container
          .requestFullscreen()
          .then(lockLandscape)
          .catch((err: Error) => {
            // Fallback: iOS native video fullscreen (auto-rotates but loses custom UI)
            if (video?.webkitEnterFullscreen) {
              video.webkitEnterFullscreen()
            } else {
              showToastInfo(`Fullscreen: ${err.message}`)
              console.error('[fullscreen]', err)
            }
          })
      } else if (container?.webkitRequestFullscreen) {
        container.webkitRequestFullscreen()
        lockLandscape()
      } else if (video?.webkitEnterFullscreen) {
        video.webkitEnterFullscreen()
      } else {
        showToastInfo('Fullscreen not supported in this browser')
      }
    }
  }

  useEffect(() => {
    const onChange = () => {
      const nowFs = !!(
        document.fullscreenElement ??
        (document as Document & { webkitFullscreenElement?: Element }).webkitFullscreenElement
      )
      setIsFullscreen(nowFs)
      if (!nowFs) screen.orientation?.unlock?.()
    }
    document.addEventListener('fullscreenchange', onChange)
    document.addEventListener('webkitfullscreenchange', onChange)

    // Track iOS native video fullscreen (webkitEnterFullscreen fallback)
    const video = videoRef.current
    const onBeginFs = () => setIsFullscreen(true)
    const onEndFs = () => {
      setIsFullscreen(false)
      screen.orientation?.unlock?.()
    }
    video?.addEventListener('webkitbeginfullscreen', onBeginFs)
    video?.addEventListener('webkitendfullscreen', onEndFs)

    return () => {
      document.removeEventListener('fullscreenchange', onChange)
      document.removeEventListener('webkitfullscreenchange', onChange)
      video?.removeEventListener('webkitbeginfullscreen', onBeginFs)
      video?.removeEventListener('webkitendfullscreen', onEndFs)
    }
  }, [videoRef])

  return { isFullscreen, toggleFullscreen }
}
