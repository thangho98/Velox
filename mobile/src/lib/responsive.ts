/**
 * Responsive layout utilities for phone, tablet, and TV
 *
 * Breakpoints (matching Tailwind CSS):
 *   phone:    < 640px
 *   tablet:   640 - 1023px  (portrait & landscape)
 *   desktop:  1024 - 1279px (large tablet landscape, small TV)
 *   tv:       >= 1280px     (Android TV, large screens)
 */

import { useState, useEffect, useCallback } from 'react'
import { Dimensions, Platform, PixelRatio, ScaledSize } from 'react-native'

// ── Device Type Detection ────────────────────────────────────────────────────

export type DeviceType = 'phone' | 'tablet' | 'tv'
export type Orientation = 'portrait' | 'landscape'

const isTV = Platform.isTV

function detectDeviceType(width: number, height: number): DeviceType {
  if (isTV) return 'tv'
  const longer = Math.max(width, height)
  // Tablets typically have longer side >= 600dp
  return longer >= 900 ? 'tablet' : 'phone'
}

function detectOrientation(width: number, height: number): Orientation {
  return width >= height ? 'landscape' : 'portrait'
}

// ── Breakpoints ──────────────────────────────────────────────────────────────

export interface LayoutBreakpoint {
  /** Grid columns for media lists */
  gridColumns: number
  /** Grid columns for library cards */
  libraryColumns: number
  /** Card gap in pixels */
  cardGap: number
  /** Horizontal padding */
  screenPadding: number
  /** Base font scale (1.0 = phone default) */
  fontScale: number
  /** Whether to use side-by-side detail layout */
  sideBySideDetail: boolean
  /** Whether to show sidebar navigation (TV) */
  showSidebar: boolean
  /** Device type */
  device: DeviceType
  /** Current orientation */
  orientation: Orientation
  /** Screen width */
  width: number
  /** Screen height */
  height: number
  /** Whether controls should be larger (TV/tablet) */
  largeControls: boolean
}

function computeBreakpoint(width: number, height: number): LayoutBreakpoint {
  const device = detectDeviceType(width, height)
  const orientation = detectOrientation(width, height)

  if (device === 'tv' || (device === 'tablet' && width >= 1280)) {
    // TV / very large screen
    return {
      gridColumns: 6,
      libraryColumns: 4,
      cardGap: 16,
      screenPadding: 48,
      fontScale: 1.4,
      sideBySideDetail: true,
      showSidebar: true,
      device: isTV ? 'tv' : 'tablet',
      orientation,
      width,
      height,
      largeControls: true,
    }
  }

  if (device === 'tablet' && orientation === 'landscape') {
    // Tablet landscape (16:9, 16:10)
    return {
      gridColumns: 5,
      libraryColumns: 3,
      cardGap: 14,
      screenPadding: 32,
      fontScale: 1.2,
      sideBySideDetail: true,
      showSidebar: false,
      device,
      orientation,
      width,
      height,
      largeControls: false,
    }
  }

  if (device === 'tablet' && orientation === 'portrait') {
    // Tablet portrait
    return {
      gridColumns: 4,
      libraryColumns: 3,
      cardGap: 12,
      screenPadding: 24,
      fontScale: 1.1,
      sideBySideDetail: false,
      showSidebar: false,
      device,
      orientation,
      width,
      height,
      largeControls: false,
    }
  }

  if (width >= 480) {
    // Large phone / landscape phone
    return {
      gridColumns: 3,
      libraryColumns: 2,
      cardGap: 10,
      screenPadding: 16,
      fontScale: 1.0,
      sideBySideDetail: orientation === 'landscape',
      showSidebar: false,
      device,
      orientation,
      width,
      height,
      largeControls: false,
    }
  }

  // Small phone (default)
  return {
    gridColumns: 2,
    libraryColumns: 2,
    cardGap: 8,
    screenPadding: 12,
    fontScale: 1.0,
    sideBySideDetail: false,
    showSidebar: false,
    device,
    orientation,
    width,
    height,
    largeControls: false,
  }
}

// ── Hook ─────────────────────────────────────────────────────────────────────

/**
 * Dynamic responsive layout hook.
 * Updates automatically on orientation change / screen resize.
 */
export function useResponsiveLayout(): LayoutBreakpoint {
  const { width, height } = Dimensions.get('window')
  const [layout, setLayout] = useState(() => computeBreakpoint(width, height))

  const handleChange = useCallback(({ window }: { window: ScaledSize }) => {
    setLayout(computeBreakpoint(window.width, window.height))
  }, [])

  useEffect(() => {
    const sub = Dimensions.addEventListener('change', handleChange)
    return () => sub.remove()
  }, [handleChange])

  return layout
}

// ── Scaling Helpers ──────────────────────────────────────────────────────────

/**
 * Scale a font size based on device type.
 * Phone: returns as-is. Tablet: 10-20% larger. TV: 40% larger.
 */
export function scaledFont(base: number, fontScale: number): number {
  return Math.round(base * fontScale)
}

/**
 * Scale spacing (padding, margin) based on device.
 */
export function scaledSpacing(base: number, fontScale: number): number {
  return Math.round(base * fontScale)
}

/**
 * Calculate card width given columns, gap, and screen padding.
 */
export function cardWidth(
  columns: number,
  screenWidth: number,
  gap: number,
  padding: number,
): number {
  const totalGaps = gap * (columns - 1)
  const totalPadding = padding * 2
  return (screenWidth - totalPadding - totalGaps) / columns
}

/**
 * Get card dimensions for a specific size.
 * Poster ratio = 2:3, Backdrop ratio = 16:9
 */
export function cardDimensions(
  size: 'small' | 'medium' | 'large',
  columns: number,
  screenWidth: number,
  gap: number,
  padding: number,
) {
  const cw = cardWidth(columns, screenWidth, gap, padding)
  switch (size) {
    case 'small':
      return { width: cw, height: cw * 1.5 }
    case 'medium':
      return { width: cw, height: cw * 1.4 }
    case 'large':
      return { width: screenWidth - padding * 2, height: (screenWidth - padding * 2) * 0.56 }
  }
}
