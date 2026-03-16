import { useRef, useEffect } from 'react'
import { LuSkipForward, LuFilm } from 'react-icons/lu'
import type { CinemaItem } from '@/hooks/useCinemaMode'

interface CinemaOverlayProps {
  item: CinemaItem
  itemIndex: number
  totalItems: number
  onEnded: () => void
  onSkip: () => void
  onSkipAll: () => void
}

export function CinemaOverlay({
  item,
  itemIndex,
  totalItems,
  onEnded,
  onSkip,
  onSkipAll,
}: CinemaOverlayProps) {
  const videoRef = useRef<HTMLVideoElement>(null)
  const iframeRef = useRef<HTMLIFrameElement>(null)

  useEffect(() => {
    if (item.type === 'intro' && videoRef.current) {
      videoRef.current.play().catch(() => {})
    }
  }, [item])

  const isYouTube = item.url.includes('youtube.com') || item.url.includes('youtu.be')

  return (
    <div className="absolute inset-0 z-40 bg-black">
      {/* Video content */}
      {isYouTube ? (
        <iframe
          ref={iframeRef}
          src={item.url}
          className="h-full w-full"
          allow="autoplay; encrypted-media"
          allowFullScreen
        />
      ) : (
        <video
          ref={videoRef}
          src={item.url}
          className="h-full w-full object-contain"
          autoPlay
          onEnded={onEnded}
        />
      )}

      {/* Top bar: what's playing */}
      <div className="absolute left-0 right-0 top-0 flex items-center gap-3 bg-gradient-to-b from-black/70 to-transparent px-6 py-4">
        <LuFilm size={18} className="text-amber-400" />
        <span className="text-sm font-medium text-white/80">
          {item.type === 'intro' ? 'Cinema Intro' : `Trailer ${itemIndex + 1}`}
        </span>
        <span className="text-xs text-white/40">
          {itemIndex + 1} / {totalItems}
        </span>
      </div>

      {/* Bottom controls */}
      <div className="absolute bottom-0 left-0 right-0 flex items-center justify-between bg-gradient-to-t from-black/70 to-transparent px-6 py-4">
        <span className="text-sm text-white/60">{item.title}</span>
        <div className="flex items-center gap-3">
          {item.skippable && (
            <button
              onClick={onSkip}
              className="flex items-center gap-2 rounded-lg bg-white/20 px-4 py-2 text-sm font-medium text-white backdrop-blur-sm transition-colors hover:bg-white/30"
            >
              <LuSkipForward size={16} />
              Skip
            </button>
          )}
          <button
            onClick={onSkipAll}
            className="rounded-lg border border-white/30 px-4 py-2 text-sm text-white/70 transition-colors hover:border-white/60 hover:text-white"
          >
            Skip to Movie
          </button>
        </div>
      </div>
    </div>
  )
}
