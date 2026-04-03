/**
 * FocusableCard -- a wrapper that adds a TV-style focus highlight to any card.
 *
 * On phone/tablet it behaves as a plain `Pressable`.
 * On Android TV the component:
 *   - tracks focus/blur via `onFocus` / `onBlur`
 *   - renders a red border + slight scale-up when focused
 *   - forwards `hasTVPreferredFocus` so the first card in a list can
 *     auto-receive focus on mount
 */

import { useState } from 'react'
import { Pressable, Platform, type ViewStyle, type StyleProp } from 'react-native'

// ── Types ───────────────────────────────────────────────────────────────────

interface FocusableCardProps {
  children: React.ReactNode
  onPress?: () => void
  onLongPress?: () => void
  style?: StyleProp<ViewStyle>
  /** Give this card initial TV focus (only the first card in a row, typically). */
  hasTVPreferredFocus?: boolean
  /** Custom focus border color (defaults to Velox red #e50914). */
  focusColor?: string
  /** Disable the focus scale-up animation. */
  disableScale?: boolean
  /** Extra props forwarded to the underlying Pressable (e.g. nextFocusUp). */
  nextFocusUp?: number
  nextFocusDown?: number
  nextFocusLeft?: number
  nextFocusRight?: number
}

// ── Focused style ───────────────────────────────────────────────────────────

const FOCUS_SCALE = 1.05

function buildFocusStyle(color: string, disableScale: boolean): ViewStyle {
  const base: ViewStyle = {
    borderWidth: 3,
    borderColor: color,
  }
  if (!disableScale) {
    base.transform = [{ scale: FOCUS_SCALE }]
  }
  return base
}

// ── Component ───────────────────────────────────────────────────────────────

export function FocusableCard({
  children,
  onPress,
  onLongPress,
  style,
  hasTVPreferredFocus,
  focusColor = '#e50914',
  disableScale = false,
  nextFocusUp,
  nextFocusDown,
  nextFocusLeft,
  nextFocusRight,
}: FocusableCardProps) {
  const [focused, setFocused] = useState(false)

  return (
    <Pressable
      onPress={onPress}
      onLongPress={onLongPress}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      {...(Platform.isTV ? {
        hasTVPreferredFocus,
        nextFocusUp,
        nextFocusDown,
        nextFocusLeft,
        nextFocusRight,
      } as any : {})}
      style={[
        style,
        focused && Platform.isTV && buildFocusStyle(focusColor, disableScale),
      ]}
    >
      {children}
    </Pressable>
  )
}
