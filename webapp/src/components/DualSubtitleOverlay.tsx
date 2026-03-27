import { useEffect, useRef, useState } from 'react'

interface VTTCue {
  start: number // seconds
  end: number // seconds
  text: string
}

interface SubtitleStyle {
  size: 'small' | 'medium' | 'large'
  color: string
  background: 'solid' | 'semi' | 'none'
}

interface DualSubtitleOverlayProps {
  videoRef: React.RefObject<HTMLVideoElement | null>
  primaryUrl: string | null // VTT URL for primary subtitle (rendered at bottom)
  secondaryUrl?: string | null // VTT URL for secondary subtitle (rendered above primary)
  currentTime: number // synced from parent to avoid an extra timeupdate listener
  offsetSeconds?: number // positive values delay the subtitle, negative values show it earlier
  primaryRenderedInVideo?: boolean
  style?: SubtitleStyle
}

/** Parse a VTT timestamp string ("HH:MM:SS.mmm" or "MM:SS.mmm") to seconds. */
function parseTimestamp(ts: string): number {
  const parts = ts.trim().split(':')
  if (parts.length === 3) {
    return Number(parts[0]) * 3600 + Number(parts[1]) * 60 + Number(parts[2])
  }
  return Number(parts[0]) * 60 + Number(parts[1])
}

/** Minimal VTT parser — handles the cue body only (ignores NOTE/STYLE/REGION). */
async function fetchVTT(url: string): Promise<VTTCue[]> {
  const res = await fetch(url)
  if (!res.ok) return []
  const text = await res.text()

  const cues: VTTCue[] = []
  // Split on blank lines
  const blocks = text.split(/\n\n+/)
  for (const block of blocks) {
    const lines = block.trim().split('\n')
    // Find the timestamp line
    const tsIdx = lines.findIndex((l) => l.includes('-->'))
    if (tsIdx === -1) continue
    const [startStr, endStr] = lines[tsIdx].split('-->').map((s) => s.trim())
    const start = parseTimestamp(startStr)
    const end = parseTimestamp(endStr)
    const text = lines
      .slice(tsIdx + 1)
      .join('\n')
      .trim()
    if (text) cues.push({ start, end, text })
  }
  return cues
}

function useVTTCues(url: string | null | undefined): VTTCue[] {
  const [cues, setCues] = useState<VTTCue[]>([])
  useEffect(() => {
    if (!url) {
      setCues([])
      return
    }
    fetchVTT(url).then(setCues)
  }, [url])
  return cues
}

function activeCue(cues: VTTCue[], time: number): string | null {
  const cue = cues.find((c) => time >= c.start && time <= c.end)
  return cue ? cue.text : null
}

// Font size mapping — Netflix uses ~32px for large on a 1080p screen
const SIZE_MAP = {
  small: 'text-2xl', // 24px
  medium: 'text-[32px]', // 32px
  large: 'text-[40px]', // 40px — Netflix-sized
} as const

const SECONDARY_SIZE_MAP = {
  small: 'text-base', // 16px
  medium: 'text-xl', // 20px
  large: 'text-2xl', // 24px
} as const

// Text stroke for no-background mode (Netflix/Emby style)
// Uses paint-order + stroke so the stroke appears behind the fill
function getTextStyle(color: string, background: 'solid' | 'semi' | 'none'): React.CSSProperties {
  const base: React.CSSProperties = { color }

  if (background === 'none') {
    return {
      ...base,
      WebkitTextStroke: '1.5px rgba(0,0,0,0.9)',
      paintOrder: 'stroke fill',
      textShadow: [
        '0 0 4px rgba(0,0,0,0.9)',
        '0 0 8px rgba(0,0,0,0.6)',
        '1px 1px 2px rgba(0,0,0,0.8)',
        '-1px -1px 2px rgba(0,0,0,0.8)',
      ].join(', '),
    }
  }

  if (background === 'semi') {
    return {
      ...base,
      background: 'rgba(0,0,0,0.5)',
      textShadow: '0 1px 3px rgba(0,0,0,0.8)',
    }
  }

  // solid
  return {
    ...base,
    background: 'rgba(0,0,0,0.8)',
    textShadow: '0 1px 2px rgba(0,0,0,0.6)',
  }
}

const DEFAULT_STYLE: SubtitleStyle = {
  size: 'large',
  color: '#ffffff',
  background: 'none',
}

/** Calculate the bottom letterbox height when video is letterboxed (portrait viewing). */
function useLetterboxBottom(videoRef: React.RefObject<HTMLVideoElement | null>): number {
  const [offset, setOffset] = useState(0)
  const rafRef = useRef(0)

  useEffect(() => {
    const video = videoRef.current
    if (!video) return

    const calculate = () => {
      const { videoWidth, videoHeight } = video
      if (!videoWidth || !videoHeight) {
        setOffset(0)
        return
      }
      const container = video.parentElement
      if (!container) return

      const cw = container.clientWidth
      const ch = container.clientHeight
      const videoAspect = videoWidth / videoHeight
      const containerAspect = cw / ch

      if (containerAspect < videoAspect) {
        // Letterboxing: black bars top/bottom
        const renderedHeight = cw / videoAspect
        setOffset(Math.round((ch - renderedHeight) / 2))
      } else {
        setOffset(0)
      }
    }

    const onUpdate = () => {
      cancelAnimationFrame(rafRef.current)
      rafRef.current = requestAnimationFrame(calculate)
    }

    video.addEventListener('loadedmetadata', onUpdate)
    window.addEventListener('resize', onUpdate)
    screen.orientation?.addEventListener('change', onUpdate)
    onUpdate()

    return () => {
      cancelAnimationFrame(rafRef.current)
      video.removeEventListener('loadedmetadata', onUpdate)
      window.removeEventListener('resize', onUpdate)
      screen.orientation?.removeEventListener('change', onUpdate)
    }
  }, [videoRef])

  return offset
}

export function DualSubtitleOverlay({
  videoRef,
  primaryUrl,
  secondaryUrl,
  currentTime,
  offsetSeconds = 0,
  primaryRenderedInVideo = false,
  style = DEFAULT_STYLE,
}: DualSubtitleOverlayProps) {
  const primaryCues = useVTTCues(primaryUrl)
  const secondaryCues = useVTTCues(secondaryUrl)
  const adjustedTime = currentTime - offsetSeconds
  const letterboxBottom = useLetterboxBottom(videoRef)

  const primaryText = activeCue(primaryCues, adjustedTime)
  const secondaryText = activeCue(secondaryCues, adjustedTime)

  if (!primaryText && !secondaryText) return null

  const primarySizeClass = SIZE_MAP[style.size]
  const secondarySizeClass = SECONDARY_SIZE_MAP[style.size]
  const needsBoxPadding = style.background !== 'none'

  // Base offset from video bottom edge (clears player controls)
  const BASE_BOTTOM = 112 // 7rem = bottom-28
  // For secondary-only (burned-in primary): higher positioning
  const secondaryLineCount = secondaryText ? secondaryText.split('\n').length : 0
  const secondaryOnlyBase = secondaryLineCount >= 3 ? 304 : secondaryLineCount === 2 ? 272 : 240

  const bottomPx =
    primaryRenderedInVideo && !primaryText
      ? letterboxBottom + secondaryOnlyBase
      : letterboxBottom + BASE_BOTTOM

  return (
    <div
      className="pointer-events-none absolute inset-x-0 flex flex-col items-center gap-1.5 px-8"
      style={{ bottom: `${bottomPx}px` }}
    >
      {/* Secondary subtitle — smaller, yellow, above primary */}
      {secondaryText && (
        <p
          className={`max-w-4xl text-center ${secondarySizeClass} font-medium leading-snug ${needsBoxPadding ? 'rounded px-3 py-1' : ''}`}
          style={getTextStyle('#fde047', style.background)}
          dangerouslySetInnerHTML={{ __html: secondaryText.replace(/\n/g, '<br/>') }}
        />
      )}
      {/* Primary subtitle — larger, at bottom */}
      {primaryText && (
        <p
          className={`max-w-4xl text-center ${primarySizeClass} font-bold leading-snug ${needsBoxPadding ? 'rounded px-3 py-1' : ''}`}
          style={getTextStyle(style.color, style.background)}
          dangerouslySetInnerHTML={{ __html: primaryText.replace(/\n/g, '<br/>') }}
        />
      )}
    </div>
  )
}

// Convenience ref — not used inside the component but exported so consumers can
// attach/pass the videoRef from the player without importing a separate type.
export type { DualSubtitleOverlayProps, SubtitleStyle }
export { useVTTCues }

// Simple single-track overlay (wraps DualSubtitleOverlay with only primary)
interface SubtitleOverlayProps {
  videoRef: React.RefObject<HTMLVideoElement | null>
  subtitleUrl: string | null
  currentTime: number
  offsetSeconds?: number
  primaryRenderedInVideo?: boolean
  style?: SubtitleStyle
}

export function SubtitleOverlay({
  videoRef,
  subtitleUrl,
  currentTime,
  offsetSeconds,
  primaryRenderedInVideo,
  style,
}: SubtitleOverlayProps) {
  return (
    <DualSubtitleOverlay
      videoRef={videoRef}
      primaryUrl={subtitleUrl}
      currentTime={currentTime}
      offsetSeconds={offsetSeconds}
      primaryRenderedInVideo={primaryRenderedInVideo}
      style={style}
    />
  )
}
