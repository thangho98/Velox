/**
 * TV-specific utilities for Android TV / large screen D-pad navigation.
 *
 * On Android TV there is no touch input -- users navigate via a remote with
 * D-pad arrows, Select/Enter, and Back.  Every interactive element must be
 * focusable and must show a clear visual indicator when focused.
 */

import { Platform, ViewStyle } from 'react-native'

// ── Constants ───────────────────────────────────────────────────────────────

/** `true` when the app is running on an Android TV (or any RN TV target). */
export const isTV: boolean = Platform.isTV

/** Velox brand red used for the focus ring on TV. */
const TV_ACCENT = '#e50914'

/** Focus ring style applied to the currently-focused element (TV only). */
export const TV_FOCUS_STYLE: ViewStyle = {
  borderWidth: 3,
  borderColor: TV_ACCENT,
  borderRadius: 8,
}

/** Minimum touch-target / focusable-area size for a 10-foot UI (dp). */
export const TV_MIN_TARGET = 64

/** Text scale multiplier for TV (matches responsive.ts TV fontScale). */
export const TV_FONT_SCALE = 1.4

/** Width of the TV sidebar navigation rail (dp). */
export const TV_SIDEBAR_WIDTH = 200

// ── Helpers ─────────────────────────────────────────────────────────────────

/**
 * Returns the style object only when running on TV; otherwise `undefined`.
 * Handy for conditional spread:  `style={[base, tvOnly(tvStyle)]}`.
 */
export function tvOnly<T>(style: T): T | undefined {
  return isTV ? style : undefined
}

/**
 * Scale a numeric value for TV (e.g. icon size, padding).
 * On non-TV platforms the value is returned unchanged.
 */
export function tvScale(value: number, scale = TV_FONT_SCALE): number {
  return isTV ? Math.round(value * scale) : value
}
