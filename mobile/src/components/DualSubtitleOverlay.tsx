/**
 * DualSubtitleOverlay — renders two subtitle tracks over the video.
 *
 * Primary subtitle: bottom of video, larger text (white by default).
 * Secondary subtitle: above primary, smaller text, yellow (#FFD700).
 *
 * Fetches and parses VTT files, syncs with playback currentTime.
 */

import { useState, useEffect } from 'react'
import { View, Text, StyleSheet } from 'react-native'

// ---------- Types ----------

interface VTTCue {
  start: number // seconds
  end: number // seconds
  text: string
}

export interface SubtitleAppearance {
  size: 'small' | 'medium' | 'large'
  color: string
  background: 'solid' | 'semi' | 'none'
}

interface DualSubtitleOverlayProps {
  primaryUrl: string | null
  secondaryUrl?: string | null
  currentTime: number
  visible: boolean
  offsetSeconds?: number
  appearance?: SubtitleAppearance
}

// ---------- VTT Parsing ----------

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
  try {
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
      const cueText = lines
        .slice(tsIdx + 1)
        .join('\n')
        .trim()
      if (cueText) cues.push({ start, end, text: cueText })
    }
    return cues
  } catch {
    return []
  }
}

// ---------- Hooks ----------

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

// ---------- Size maps (font sizes in px) ----------

const PRIMARY_SIZE_MAP = {
  small: 18,
  medium: 24,
  large: 32,
} as const

const SECONDARY_SIZE_MAP = {
  small: 13,
  medium: 16,
  large: 20,
} as const

// ---------- Styling helpers ----------

function getTextShadow(background: 'solid' | 'semi' | 'none'): string {
  if (background === 'none') {
    // Netflix-style text stroke via multiple text shadows
    return [
      '0px 0px 4px rgba(0,0,0,0.9)',
      '0px 0px 8px rgba(0,0,0,0.6)',
      '1px 1px 2px rgba(0,0,0,0.8)',
      '-1px -1px 2px rgba(0,0,0,0.8)',
      '1px -1px 2px rgba(0,0,0,0.8)',
      '-1px 1px 2px rgba(0,0,0,0.8)',
    ].join(', ')
  }
  return '0px 1px 3px rgba(0,0,0,0.8)'
}

function getBackgroundColor(background: 'solid' | 'semi' | 'none'): string | undefined {
  if (background === 'solid') return 'rgba(0,0,0,0.8)'
  if (background === 'semi') return 'rgba(0,0,0,0.5)'
  return undefined
}

// ---------- Default appearance ----------

const DEFAULT_APPEARANCE: SubtitleAppearance = {
  size: 'large',
  color: '#ffffff',
  background: 'none',
}

// ---------- Component ----------

export function DualSubtitleOverlay({
  primaryUrl,
  secondaryUrl,
  currentTime,
  visible,
  offsetSeconds = 0,
  appearance = DEFAULT_APPEARANCE,
}: DualSubtitleOverlayProps) {
  const primaryCues = useVTTCues(primaryUrl)
  const secondaryCues = useVTTCues(secondaryUrl)
  const adjustedTime = currentTime - offsetSeconds

  const primaryText = activeCue(primaryCues, adjustedTime)
  const secondaryText = activeCue(secondaryCues, adjustedTime)

  if (!visible || (!primaryText && !secondaryText)) return null

  const primaryFontSize = PRIMARY_SIZE_MAP[appearance.size]
  const secondaryFontSize = SECONDARY_SIZE_MAP[appearance.size]
  const bgColor = getBackgroundColor(appearance.background)
  const needsPadding = appearance.background !== 'none'
  const textShadow = getTextShadow(appearance.background)

  // Strip HTML tags from VTT text (some VTT files include <i>, <b> etc.)
  const stripTags = (text: string) => text.replace(/<[^>]+>/g, '')

  return (
    <View style={styles.container} pointerEvents="none">
      {/* Secondary subtitle (above primary, yellow, smaller) */}
      {secondaryText && (
        <View
          style={[
            styles.subtitleBox,
            needsPadding && styles.subtitlePadding,
            bgColor ? { backgroundColor: bgColor } : undefined,
          ]}
        >
          <Text
            style={[
              styles.subtitleText,
              {
                fontSize: secondaryFontSize,
                color: '#FFD700',
                fontWeight: '500',
                textShadowColor: 'rgba(0,0,0,0.9)',
                textShadowOffset: { width: 1, height: 1 },
                textShadowRadius: 4,
              },
            ]}
          >
            {stripTags(secondaryText)}
          </Text>
        </View>
      )}

      {/* Primary subtitle (bottom, larger, user-selected color) */}
      {primaryText && (
        <View
          style={[
            styles.subtitleBox,
            needsPadding && styles.subtitlePadding,
            bgColor ? { backgroundColor: bgColor } : undefined,
          ]}
        >
          <Text
            style={[
              styles.subtitleText,
              {
                fontSize: primaryFontSize,
                color: appearance.color,
                fontWeight: '700',
                textShadowColor: 'rgba(0,0,0,0.9)',
                textShadowOffset: { width: 1, height: 1 },
                textShadowRadius: 4,
              },
            ]}
          >
            {stripTags(primaryText)}
          </Text>
        </View>
      )}
    </View>
  )
}

// ---------- Styles ----------

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    left: 0,
    right: 0,
    bottom: 80, // Above bottom controls
    alignItems: 'center',
    paddingHorizontal: 16,
    gap: 4,
  },
  subtitleBox: {
    maxWidth: '90%',
    borderRadius: 4,
  },
  subtitlePadding: {
    paddingHorizontal: 12,
    paddingVertical: 4,
  },
  subtitleText: {
    textAlign: 'center',
    lineHeight: undefined, // Will be auto-calculated
  },
})

export type { DualSubtitleOverlayProps }
