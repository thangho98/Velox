import { useRef, useEffect, useState } from 'react'

declare global {
  interface Window {
    YT?: {
      Player: new (
        el: HTMLElement,
        config: {
          videoId: string
          playerVars?: Record<string, number | string>
          events?: Record<string, (event: { target: YTPlayer }) => void>
        },
      ) => YTPlayer
    }
    onYouTubeIframeAPIReady?: () => void
  }
}

interface YTPlayer {
  playVideo: () => void
  mute: () => void
  unMute: () => void
  isMuted: () => boolean
  destroy: () => void
}

interface YouTubeBackgroundProps {
  videoId: string
  muted: boolean
  onMutedChange?: (muted: boolean) => void
  className?: string
}

let apiLoaded = false
let apiReady = false
const readyCallbacks: Array<() => void> = []

function loadYTApi() {
  if (apiLoaded) return
  apiLoaded = true
  const tag = document.createElement('script')
  tag.src = 'https://www.youtube.com/iframe_api'
  document.head.appendChild(tag)
  window.onYouTubeIframeAPIReady = () => {
    apiReady = true
    readyCallbacks.forEach((cb) => cb())
    readyCallbacks.length = 0
  }
}

function onApiReady(cb: () => void) {
  if (apiReady) {
    cb()
  } else {
    readyCallbacks.push(cb)
    loadYTApi()
  }
}

export function YouTubeBackground({ videoId, muted, className }: YouTubeBackgroundProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const playerRef = useRef<YTPlayer | null>(null)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    if (!containerRef.current) return

    const el = document.createElement('div')
    containerRef.current.appendChild(el)

    onApiReady(() => {
      if (!window.YT || !containerRef.current) return
      playerRef.current = new window.YT.Player(el, {
        videoId,
        playerVars: {
          autoplay: 1,
          controls: 0,
          showinfo: 0,
          rel: 0,
          loop: 1,
          playlist: videoId,
          modestbranding: 1,
          iv_load_policy: 3,
          disablekb: 1,
          fs: 0,
          playsinline: 1,
          mute: 1,
        },
        events: {
          onReady: (event: { target: YTPlayer }) => {
            event.target.playVideo()
            event.target.mute()
            setReady(true)
          },
        },
      })
    })

    return () => {
      playerRef.current?.destroy()
      playerRef.current = null
    }
  }, [videoId])

  useEffect(() => {
    if (!ready || !playerRef.current) return
    if (muted) {
      playerRef.current.mute()
    } else {
      playerRef.current.unMute()
    }
  }, [muted, ready])

  return (
    <div className={className} style={{ overflow: 'hidden' }}>
      <div
        ref={containerRef}
        style={{
          width: '160%',
          height: '160%',
          marginLeft: '-30%',
          marginTop: '-15%',
          pointerEvents: 'none',
        }}
      />
    </div>
  )
}
