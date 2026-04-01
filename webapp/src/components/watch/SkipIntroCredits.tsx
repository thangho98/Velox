import { useEffect, useRef, useState } from 'react'
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
  const [skippedSet, setSkippedSet] = useState<Set<number>>(() => new Set())

  // Reset skipped set when segments change (different media loaded)
  useEffect(() => {
    const newSet = new Set<number>()
    skippedRef.current = newSet
    setSkippedSet(newSet)
  }, [segments])

  // Find active segment with boundary threshold to prevent flicker
  const activeSegment =
    segments?.find(
      (seg) =>
        !skippedSet.has(seg.start) &&
        !(seg.type === 'credits' && hideCredits) &&
        currentTime >= seg.start - BOUNDARY_THRESHOLD &&
        currentTime < seg.end - BOUNDARY_THRESHOLD,
    ) ?? null

  if (!activeSegment || !visible) return null

  const label = activeSegment.type === 'intro' ? t('controls.skipIntro') : t('controls.skipCredits')

  return (
    <div className="absolute right-6 top-1/2 -translate-y-1/2 z-50 sm:translate-y-0 sm:top-auto sm:bottom-56">
      <button
        onClick={() => {
          const newSet = new Set(skippedRef.current)
          newSet.add(activeSegment.start)
          skippedRef.current = newSet
          setSkippedSet(newSet)
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
