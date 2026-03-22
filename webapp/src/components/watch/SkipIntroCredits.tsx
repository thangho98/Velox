import { useRef } from 'react'
import { LuSkipForward } from 'react-icons/lu'
import type { SkipSegment } from '@/types/api'
import { useTranslation } from '@/hooks/useTranslation'

interface SkipIntroCreditsProps {
  segments?: SkipSegment[]
  currentTime: number
  onSkip: (toTime: number) => void
  visible: boolean
  hideCredits?: boolean // Hide credits CTA when Up Next is showing
}

const BOUNDARY_THRESHOLD = 0.25 // seconds — prevents flicker at segment edges

export function SkipIntroCredits({
  segments,
  currentTime,
  onSkip,
  visible,
  hideCredits = false,
}: SkipIntroCreditsProps) {
  const { t } = useTranslation('watch')
  // Track which segments the user already skipped (by start time) to avoid re-showing
  const skippedRef = useRef<Set<number>>(new Set())
  const lastMediaSegmentsRef = useRef<SkipSegment[] | undefined>(undefined)

  // Reset skipped set when segments change (different media loaded)
  if (segments !== lastMediaSegmentsRef.current) {
    lastMediaSegmentsRef.current = segments
    skippedRef.current = new Set()
  }

  // Find active segment with boundary threshold to prevent flicker
  const activeSegment =
    segments?.find(
      (seg) =>
        !skippedRef.current.has(seg.start) &&
        !(seg.type === 'credits' && hideCredits) &&
        currentTime >= seg.start - BOUNDARY_THRESHOLD &&
        currentTime < seg.end - BOUNDARY_THRESHOLD,
    ) ?? null

  if (!activeSegment || !visible) return null

  const label = activeSegment.type === 'intro' ? t('controls.skipIntro') : t('controls.skipCredits')

  return (
    <div className="absolute bottom-56 right-6 z-30">
      <button
        onClick={() => {
          skippedRef.current.add(activeSegment.start)
          onSkip(activeSegment.end)
        }}
        className="flex items-center gap-2 rounded-lg bg-white/95 px-4 py-2.5 text-sm font-semibold text-black shadow-lg backdrop-blur-sm transition-all hover:bg-white hover:scale-105 active:scale-95"
      >
        <span>{label}</span>
        <LuSkipForward size={16} />
      </button>
    </div>
  )
}
