import { useRef, useEffect } from 'react'

interface YouTubeBackgroundProps {
  videoId: string
  muted: boolean
  onMutedChange?: (muted: boolean) => void
  className?: string
}

export function YouTubeBackground({ videoId, muted, className }: YouTubeBackgroundProps) {
  const iframeRef = useRef<HTMLIFrameElement>(null)

  // Control mute/unmute via postMessage (requires enablejsapi=1)
  useEffect(() => {
    const iframe = iframeRef.current
    if (!iframe?.contentWindow) return
    const cmd = muted ? 'mute' : 'unMute'
    iframe.contentWindow.postMessage(JSON.stringify({ event: 'command', func: cmd, args: '' }), '*')
  }, [muted])

  const src = `https://www.youtube.com/embed/${videoId}?autoplay=1&mute=1&controls=0&showinfo=0&rel=0&loop=1&playlist=${videoId}&modestbranding=1&iv_load_policy=3&disablekb=1&fs=0&playsinline=1&enablejsapi=1&origin=${window.location.origin}`

  return (
    <div className={className}>
      <div className="relative h-full w-full overflow-hidden">
        <iframe
          ref={iframeRef}
          src={src}
          allow="autoplay; encrypted-media"
          style={{
            border: 'none',
            position: 'absolute',
            top: '50%',
            left: '50%',
            transform: 'translate(-50%, -50%)',
            width: '200%',
            height: '200%',
            pointerEvents: 'none',
          }}
        />
      </div>
    </div>
  )
}
