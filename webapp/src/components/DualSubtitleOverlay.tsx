import { useEffect, useState } from 'react'

interface VTTCue {
  start: number // seconds
  end: number // seconds
  text: string
}

interface DualSubtitleOverlayProps {
  videoRef: React.RefObject<HTMLVideoElement | null>
  primaryUrl: string | null // VTT URL for primary subtitle (rendered at bottom)
  secondaryUrl?: string | null // VTT URL for secondary subtitle (rendered above primary)
  currentTime: number // synced from parent to avoid an extra timeupdate listener
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

export function DualSubtitleOverlay({
  primaryUrl,
  secondaryUrl,
  currentTime,
}: DualSubtitleOverlayProps) {
  const primaryCues = useVTTCues(primaryUrl)
  const secondaryCues = useVTTCues(secondaryUrl)

  const primaryText = activeCue(primaryCues, currentTime)
  const secondaryText = activeCue(secondaryCues, currentTime)

  if (!primaryText && !secondaryText) return null

  return (
    <div className="pointer-events-none absolute inset-x-0 bottom-20 flex flex-col items-center gap-1 px-8">
      {/* Secondary subtitle — smaller, yellow, above primary */}
      {secondaryText && (
        <p
          className="max-w-3xl rounded px-2 py-0.5 text-center text-sm font-medium leading-snug text-yellow-300 [text-shadow:0_1px_3px_rgba(0,0,0,0.9)]"
          style={{ background: 'rgba(0,0,0,0.45)' }}
          dangerouslySetInnerHTML={{ __html: secondaryText.replace(/\n/g, '<br/>') }}
        />
      )}
      {/* Primary subtitle — larger, white, at bottom */}
      {primaryText && (
        <p
          className="max-w-3xl rounded px-2 py-0.5 text-center text-base font-semibold leading-snug text-white [text-shadow:0_1px_4px_rgba(0,0,0,0.95)]"
          style={{ background: 'rgba(0,0,0,0.5)' }}
          dangerouslySetInnerHTML={{ __html: primaryText.replace(/\n/g, '<br/>') }}
        />
      )}
    </div>
  )
}

// Convenience ref — not used inside the component but exported so consumers can
// attach/pass the videoRef from the player without importing a separate type.
export type { DualSubtitleOverlayProps }
export { useVTTCues }

// Simple single-track overlay (wraps DualSubtitleOverlay with only primary)
interface SubtitleOverlayProps {
  videoRef: React.RefObject<HTMLVideoElement | null>
  subtitleUrl: string | null
  currentTime: number
}

export function SubtitleOverlay({ videoRef, subtitleUrl, currentTime }: SubtitleOverlayProps) {
  return (
    <DualSubtitleOverlay videoRef={videoRef} primaryUrl={subtitleUrl} currentTime={currentTime} />
  )
}
